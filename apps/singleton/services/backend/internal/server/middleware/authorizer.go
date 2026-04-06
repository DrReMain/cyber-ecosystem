package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"

	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/auth"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/security"
)

func Authorizer(enabled bool, logger log.Logger, authorizer security.Authorizer) middleware.Middleware {
	if !enabled || authorizer == nil {
		return nil
	}
	helper := log.NewHelper(log.With(logger, "module", "server/middleware/authorizer"))
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			ctx = ensureSecurityContext(ctx)

			claims, err := auth.IdentityFromContext(ctx)
			if err != nil {
				return nil, err
			}
			if claims == nil || claims.Subject == "" {
				return handler(ctx, req)
			}

			operation, ok := security.OperationFromContext(ctx)
			if !ok || operation == "" {
				helper.Warnf("authorization skipped: missing security operation")
				return handler(ctx, req)
			}

			if err := authorizer.Authorize(ctx, claims.Subject, operation); err != nil {
				return nil, err
			}

			return handler(ctx, req)
		}
	}
}
