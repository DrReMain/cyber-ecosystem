package local_mixins

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"cyber-ecosystem/shared-go/kratos/security"

	gen "cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/intercept"
)

// TenantMixin adds a tenant_id column (always present, D2) and an interceptor
// that injects WHERE tenant_id from the authenticated Subject in ctx. It lives
// in local_mixins (like SoftDeleteMixin) because the interceptor depends on
// generated ent types.
type TenantMixin struct {
	mixin.Schema
}

func (TenantMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("tenant_id").MaxLen(20).Immutable(),
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

func applyTenant(w interface{ WhereP(...func(*sql.Selector)) }, tenantID string) {
	w.WhereP(sql.FieldEQ("tenant_id", tenantID))
}
