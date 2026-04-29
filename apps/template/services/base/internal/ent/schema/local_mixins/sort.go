package local_mixins

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"roci.dev/fracdex"

	gen "cyber-ecosystem/apps/template/services/base/internal/ent"
)

// SortMixin provides fractional-index ordering for schemas.
// On create, it queries the max sort key and generates a new key at the end of the list.
// It detects soft-delete via SetDeletedAt type assertion and filters deleted rows.
//
// When SoftDelete is true, the unique index on sort uses a WHERE deleted_at IS NULL clause
// so soft-deleted rows do not conflict with active sort keys.
//
// Concurrency note: under high-concurrency creates, two transactions may read the same
// maxSort and produce duplicate keys. The UNIQUE index on sort acts as the
// final safety net — one INSERT will fail with a constraint violation.
type SortMixin struct {
	mixin.Schema
	SoftDelete bool
}

func (s SortMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("sort").
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
	return []ent.Hook{sortHook}
}

var sortHook ent.Hook = func(next ent.Mutator) ent.Mutator {
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

		maxSort, err := queryMaxSort(ctx, mx.Client(), m)
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

func queryMaxSort(ctx context.Context, client *gen.Client, m ent.Mutation) (string, error) {
	table := strings.ToLower(m.Type())

	q := fmt.Sprintf(`SELECT sort FROM "%s" ORDER BY sort DESC LIMIT 1`, table)
	if _, hasSD := m.(interface{ SetDeletedAt(time.Time) }); hasSD {
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
