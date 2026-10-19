package local_mixins

import (
	"context"
	"fmt"
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"roci.dev/fracdex"

	gen "cyber-ecosystem/app/services/system/internal/ent"
)

// SortMixin equips a table with a fractional-index "sort" field for
// drag-to-reorder. Create auto-assigns a key appended after the current
// max; the read-modify-write is serialized by a PG advisory lock
// (hashtext(table), transaction-scoped) — the surrounding UC transaction
// MUST exist or the lock is a no-op. CreateBulk and upsert paths are not
// protected; set sort explicitly there. Lock reach is per-PG-instance,
// shared across replicas.
type SortMixin struct {
	mixin.Schema
	SoftDelete bool
	Table      string
}

func (s SortMixin) Fields() []ent.Field {
	return []ent.Field{
		// C collation = byte-order. fracdex keys are positional tokens and only order
		// correctly under byte-ordering; the DB default (e.g. en_US) mis-orders them
		// and treats case-only-differing keys as equal, breaking the unique index.
		field.String("sort").
			Annotations(entsql.Annotation{Collation: "C"}).
			Comment("fractional index for ordering"),
	}
}

func (s SortMixin) Indexes() []ent.Index {
	idx := index.Fields("sort").Unique()
	if s.SoftDelete {
		idx = idx.Annotations(entsql.IndexWhere("deleted_at IS NULL"))
	}
	return []ent.Index{idx}
}

func (s SortMixin) Hooks() []ent.Hook {
	return []ent.Hook{s.hook()}
}

// hook auto-assigns a sort key (appended after the current max) on create when
// the caller did not set one. A per-table advisory lock is taken before the max
// query so the read-modify-write runs serialized against other creates of the
// same table; see the SortMixin doc for why this needs a transaction and how it
// behaves across replicas.
func (s SortMixin) hook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if m.Op() != ent.OpCreate {
				return next.Mutate(ctx, m)
			}
			if _, ok := m.Field("sort"); ok {
				return next.Mutate(ctx, m)
			}

			mx, ok := m.(interface{ Client() *gen.Client })
			if !ok {
				return next.Mutate(ctx, m)
			}

			if err := validateTableName(s.Table); err != nil {
				return nil, err
			}

			if err := lockTableForSort(ctx, mx.Client(), s.Table); err != nil {
				return nil, err
			}

			maxSort, err := queryMaxSort(ctx, mx.Client(), s.Table, s.SoftDelete)
			if err != nil {
				return nil, err
			}

			newSort, err := fracdex.KeyBetween(maxSort, "")
			if err != nil {
				return nil, fmt.Errorf("sort mixin: generate key: %w", err)
			}

			if err := m.SetField("sort", newSort); err != nil {
				return nil, fmt.Errorf("sort mixin: set field: %w", err)
			}

			return next.Mutate(ctx, m)
		})
	}
}

// queryMaxSort returns the largest sort value in the table, excluding soft-deleted
// rows when softDelete is set.
func queryMaxSort(ctx context.Context, client *gen.Client, table string, softDelete bool) (string, error) {
	q := fmt.Sprintf(`SELECT sort FROM "%s" ORDER BY sort DESC LIMIT 1`, table)
	if softDelete {
		q = fmt.Sprintf(`SELECT sort FROM "%s" WHERE deleted_at IS NULL ORDER BY sort DESC LIMIT 1`, table)
	}

	rows, err := client.QueryContext(ctx, q)
	if err != nil {
		return "", fmt.Errorf("sort mixin: query max sort: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		var maxSort string
		if err := rows.Scan(&maxSort); err != nil {
			return "", fmt.Errorf("sort mixin: scan: %w", err)
		}
		return maxSort, nil
	}
	return "", nil
}

// lockTableForSort acquires a transaction-scoped advisory lock keyed by the table
// name, so concurrent creates of the same table run the read-modify-write one at
// a time. PG-only; a silent no-op outside a transaction — see the SortMixin doc.
func lockTableForSort(ctx context.Context, client *gen.Client, table string) error {
	if _, err := client.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", table); err != nil {
		return fmt.Errorf("sort mixin: acquire advisory lock: %w", err)
	}
	return nil
}

// safeIdent accepts only bare SQL identifiers. queryMaxSort interpolates table
// into SQL via Sprintf (a parameterized placeholder cannot be used for an
// identifier), so this is the one line that stops a malformed or hostile table
// name from breaking out of the string. Table names here are schema constants,
// but a reusable mixin must not trust the caller.
var safeIdent = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func validateTableName(table string) error {
	if !safeIdent.MatchString(table) {
		return fmt.Errorf("sort mixin: invalid table name %q", table)
	}
	return nil
}
