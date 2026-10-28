package local_mixins

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"cyber-ecosystem/shared-go/kratos/security"

	gen "cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/hook"
	"cyber-ecosystem/app/services/system/internal/ent/intercept"
)

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

func (TenantMixin) Interceptors() []gen.Interceptor {
	return []gen.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			if security.IsRegistryRead(ctx) {
				return nil
			}
			// An explicit domain switch wins over the subject — the same
			// precedence the per-tenant storage routing reads, so a domain
			// pass filters rows and buckets on one axis.
			if tenant, ok := security.ExplicitTenantFromCtx(ctx); ok {
				applyTenant(q, tenant)
				return nil
			}
			subject, ok := security.SubjectFromCtx(ctx)
			if !ok || subject.TenantID == "" {
				return nil // no subject (framework/health) → skip
			}
			applyTenant(q, subject.TenantID)
			return nil
		}),
	}
}

// Hooks back-fills tenant_id on inserts and narrows every update and delete
// to the same tenant the interceptor above filters reads by. Writes used to
// ride the biz-layer pre-read convention alone — correctly written modules
// could not be confused, but nothing structural answered for a raw
// cross-tenant mutation. An empty insert tenant (unset, or the zero value
// set by data layers) resolves via security.TenantFromCtx; business code
// never assigns tenants.
//
// The predicate guards all four write ops: soft-deleted tables rewrite
// deletes into updates that re-enter this chain (re-applying the predicate
// is idempotent), while hard-deleted tables keep native delete ops, which
// the narrower update-only shape would miss. Its precedence mirrors the
// interceptor — explicit domain switch > subject > no subject (framework
// paths, registry sweeps) applies nothing.
func (TenantMixin) Hooks() []gen.Hook {
	return []gen.Hook{
		hook.On(
			func(next gen.Mutator) gen.Mutator {
				return gen.MutateFunc(func(ctx context.Context, m gen.Mutation) (ent.Value, error) {
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
			gen.OpCreate,
		),
		hook.On(
			func(next gen.Mutator) gen.Mutator {
				return gen.MutateFunc(func(ctx context.Context, m gen.Mutation) (ent.Value, error) {
					mx, ok := m.(interface{ WhereP(...func(*sql.Selector)) })
					if !ok {
						return nil, fmt.Errorf("tenant mixin: unexpected mutation type %T", m)
					}
					if tenant, ok := security.ExplicitTenantFromCtx(ctx); ok {
						applyTenant(mx, tenant)
					} else if subject, ok := security.SubjectFromCtx(ctx); ok && subject.TenantID != "" {
						applyTenant(mx, subject.TenantID)
					}
					return next.Mutate(ctx, m)
				})
			},
			gen.OpUpdate|gen.OpUpdateOne|gen.OpDelete|gen.OpDeleteOne,
		),
	}
}

func applyTenant(w interface{ WhereP(...func(*sql.Selector)) }, tenantID string) {
	w.WhereP(sql.FieldEQ("tenant_id", tenantID))
}
