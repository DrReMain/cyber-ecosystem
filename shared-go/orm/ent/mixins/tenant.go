package mixins

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"cyber-ecosystem/shared-go/kratos/security"
)

// TenantMixin scopes rows to the caller's tenant structurally: reads are
// filtered (via TenantScope below), writes are back-filled (inserts) and
// narrowed (updates, deletes) to the same tenant — business code never
// assigns tenants.
//
// The read half rides a per-service thin adapter, not this file: queries
// have no framework-level predicate door (entgo keeps WhereP on the
// per-service generated adapter by design), so each service wraps
// TenantScope in its own intercept.TraverseFunc — a dozen decision-free
// lines. Fields and hooks are fully shared here, riding framework types
// and structural assertions only.
type TenantMixin struct {
	mixin.Schema
}

func (TenantMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("tenant_id").MaxRuneLen(20).Immutable().
			Comment("tenant scope; back-filled on insert and filtered from the request subject"),
	}
}

func (TenantMixin) Indexes() []ent.Index {
	return nil
}

// TenantScope is the read-filter predicate for a context — nil when the
// read must not scope (registry reads, subject-less framework paths).
// The decision lives here, once; each service's thin local adapter feeds
// it to its generated interceptor.
func TenantScope(ctx context.Context) func(*sql.Selector) {
	if security.IsRegistryRead(ctx) {
		return nil
	}
	// An explicit domain switch wins over the subject — the same
	// precedence the per-tenant storage routing reads, so a domain pass
	// filters rows and buckets on one axis.
	if tenant, ok := security.ExplicitTenantFromCtx(ctx); ok {
		return sql.FieldEQ("tenant_id", tenant)
	}
	if subject, ok := security.SubjectFromCtx(ctx); ok && subject.TenantID != "" {
		return sql.FieldEQ("tenant_id", subject.TenantID)
	}
	return nil // no subject (framework/health) → skip
}

// Hooks back-fills tenant_id on inserts and narrows every update and delete
// to the same tenant the read filter above scopes by. Writes used to ride
// the biz-layer pre-read convention alone — correctly written modules
// could not be confused, but nothing structural answered for a raw
// cross-tenant mutation. An empty insert tenant (unset, or the zero value
// set by data layers) resolves via security.TenantFromCtx.
//
// The predicate guards all four write ops: soft-deleted tables rewrite
// deletes into updates that re-enter this chain (re-applying the predicate
// is idempotent), while hard-deleted tables keep native delete ops, which
// the narrower update-only shape would miss. Its precedence mirrors the
// read scope — explicit domain switch > subject > no subject (framework
// paths, registry sweeps) applies nothing.
func (TenantMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				// The generated hook.On's own gate, framework-native: only
				// creates back-fill.
				if !m.Op().Is(ent.OpCreate) {
					return next.Mutate(ctx, m)
				}
				if security.IsRegistryRead(ctx) {
					return nil, fmt.Errorf("tenant mixin: a registry-read context must not write")
				}
				if cur, ok := m.Field("tenant_id"); !ok || cur == "" {
					if err := m.SetField("tenant_id", security.TenantFromCtx(ctx)); err != nil {
						return nil, fmt.Errorf("tenant mixin: set tenant_id: %w", err)
					}
				}
				return next.Mutate(ctx, m)
			})
		},
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				// The generated hook.On's own gate, framework-native: the
				// four write ops take the tenant predicate.
				if !m.Op().Is(ent.OpUpdate | ent.OpUpdateOne | ent.OpDelete | ent.OpDeleteOne) {
					return next.Mutate(ctx, m)
				}
				mx, ok := m.(interface{ WhereP(...func(*sql.Selector)) })
				if !ok {
					return nil, fmt.Errorf("tenant mixin: unexpected mutation type %T", m)
				}
				if tenant, ok := security.ExplicitTenantFromCtx(ctx); ok {
					mx.WhereP(sql.FieldEQ("tenant_id", tenant))
				} else if subject, ok := security.SubjectFromCtx(ctx); ok && subject.TenantID != "" {
					mx.WhereP(sql.FieldEQ("tenant_id", subject.TenantID))
				}
				return next.Mutate(ctx, m)
			})
		},
	}
}
