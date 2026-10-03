package local_mixins

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"cyber-ecosystem/shared-go/kratos/security"

	gen "cyber-ecosystem/app/services/agent/internal/ent"
	"cyber-ecosystem/app/services/agent/internal/ent/hook"
	"cyber-ecosystem/app/services/agent/internal/ent/intercept"
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
			subject, ok := security.SubjectFromCtx(ctx)
			if !ok || subject.TenantID == "" {
				return nil // no subject (framework/health) → skip
			}
			applyTenant(q, subject.TenantID)
			return nil
		}),
	}
}

// Hooks back-fills tenant_id on inserts: the write-side counterpart of the
// interceptor above. An empty tenant (unset, or the zero value set by data
// layers) means "unspecified" and resolves via security.TenantFromCtx.
// Business code never assigns tenants.
func (TenantMixin) Hooks() []gen.Hook {
	return []gen.Hook{
		hook.On(
			func(next gen.Mutator) gen.Mutator {
				return gen.MutateFunc(func(ctx context.Context, m gen.Mutation) (ent.Value, error) {
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
	}
}

func applyTenant(w interface{ WhereP(...func(*sql.Selector)) }, tenantID string) {
	w.WhereP(sql.FieldEQ("tenant_id", tenantID))
}
