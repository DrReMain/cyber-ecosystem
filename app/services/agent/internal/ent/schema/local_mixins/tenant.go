package local_mixins

import (
	"context"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	gen "cyber-ecosystem/app/services/agent/internal/ent"
	"cyber-ecosystem/app/services/agent/internal/ent/intercept"
)

// TenantMixin: the fields and both write hooks live in the shared mixin
// (one copy, three services); this local half only feeds the shared read
// predicate to the generated interceptor — queries have no framework
// predicate door, so the adapter is per-service by entgo's design.
// Decision-free by construction: every rule lives in mixins.TenantScope.
type TenantMixin struct {
	mixins.TenantMixin
}

func (TenantMixin) Interceptors() []gen.Interceptor {
	return []gen.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			if p := mixins.TenantScope(ctx); p != nil {
				q.WhereP(p)
			}
			return nil
		}),
	}
}
