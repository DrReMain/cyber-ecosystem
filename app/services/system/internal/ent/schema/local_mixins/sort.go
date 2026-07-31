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

// SortMixin equips a table with a fractional-index "sort" field for drag-to-
// reorder, and auto-assigns a key on create — appended after the current max —
// when the caller did not set one. Inserting a row between two others later is
// O(1) via fracdex (it never rewrites neighbours); the auto-assigned key merely
// gives new rows a stable "at the end" position until someone drags them.
//
// Table is the SQL table name: a generic mixin cannot derive it from the Go
// type, and it must match the schema's entsql table annotation.
//
// Concurrency — the create path is a non-atomic read-modify-write (query max →
// derive key → insert). Left unprotected, concurrent creates read the same max,
// derive the same key, and the loser is rejected by the unique index. This mixin
// serializes them instead with a PostgreSQL advisory lock. The conditions below
// decide whether the lock actually protects you:
//
//   - PostgreSQL only. Advisory locks are a PG feature; this mixin is already
//     PG-bound (its sort field relies on the "C" collation, and it queries the
//     max via raw SQL).
//   - The create MUST run inside a transaction. pg_advisory_xact_lock is
//     transaction-scoped: under autocommit it is released the moment the
//     statement ends and becomes a silent no-op, leaving the create unprotected
//     with no error. This repo guarantees a transaction by wrapping every UC
//     mutation in biz.Transaction.InTx; a direct client.Create() that bypasses
//     the UC silently loses protection. See biz.Transaction.
//   - Lock key = hashtext(table). Concurrent creates of the SAME table serialize;
//     different tables proceed in parallel. A hash collision is harmless — two
//     unrelated tables would merely wait on each other.
//
// Reach across replicas: the lock lives in PG shared memory, so it is shared by
// every connection to that PG, including different replicas of this service
// behind a load balancer. That is precisely where it outperforms an in-process
// mutex, which could only serialize within a single replica. Because sort is
// sovereign data of this service and lives in this one PG, this remains a local
// transaction — it does not involve distributed transactions. The one setting
// where the lock is insufficient is sharding sort writes across multiple PG
// instances, which is a separate, larger problem and out of scope here.
//
// Coverage — only the single-Create path (client.X.Create().Save()) is protected,
// which is all the UCs in this repo use:
//   - CreateBulk is NOT supported. ent runs every builder's hook BEFORE the
//     batch insert, so each builder's queryMaxSort reads the same pre-batch max
//     and derives the same key; the unique index then rejects the whole batch as
//     a ConstraintError (surfaced as 409). Fail-safe — no corrupt rows — but the
//     bulk fails, so set sort explicitly on each builder when bulk-inserting a
//     sorted table. The hook cannot detect bulk context, so this is a documented
//     boundary, not a runtime check.
//   - OnConflict (upsert) is only generated for schemas that opt into
//     FeatureUpsert; none here do. Where enabled, sort auto-assign under upsert
//     is undefined — set sort explicitly.
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
