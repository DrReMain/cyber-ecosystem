package local_mixins

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"roci.dev/fracdex"

	gen "cyber-ecosystem/app/services/edge_mobile/internal/ent"
)

// SortMixin adds a fractional-index "sort" field, auto-appended after the current
// max on create. Table is the SQL table name — required because a generic mixin
// cannot derive it from the Go type, and it must match the schema's entsql table
// annotation.
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
// the caller did not set one.
//
// TODO: the read-modify-write (max → key → insert) is not atomic; concurrent
// creates can derive the same key and the unique index rejects the loser. Fine
// for low-concurrency admin reordering; add a create-path retry-on-constraint if
// burst inserts become a real scenario.
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
	defer rows.Close()

	if rows.Next() {
		var maxSort string
		if err := rows.Scan(&maxSort); err != nil {
			return "", fmt.Errorf("sort mixin: scan: %w", err)
		}
		return maxSort, nil
	}
	return "", nil
}
