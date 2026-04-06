package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"

	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/auth"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/datascope"
)

// ScopeInjector returns middleware that injects ScopeResolveFunc into context.
// Does NOT resolve immediately - Ent mixin calls it lazily.
// Must be placed AFTER the JWT middleware so claims are already in context.
func ScopeInjector(enabled bool, scopeResolver datascope.ScopeResolveFunc) middleware.Middleware {
	if !enabled || scopeResolver == nil {
		return nil
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			ctx = ensureSecurityContext(ctx)

			claims, err := auth.IdentityFromContext(ctx)
			if err != nil || claims == nil {
				return handler(ctx, req)
			}

			ctx = datascope.WithScopeResolver(ctx, scopeResolver)
			ctx = datascope.WithScopeUserID(ctx, claims.Subject)

			return handler(ctx, req)
		}
	}
}
