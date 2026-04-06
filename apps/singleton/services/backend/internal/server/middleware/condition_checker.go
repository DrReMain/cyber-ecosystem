package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"

	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/auth"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/security"
)

// ConditionChecker returns middleware that evaluates ABAC conditions for the current operation.
func ConditionChecker(enabled bool, logger log.Logger, checker security.ConditionChecker) middleware.Middleware {
	if !enabled || checker == nil {
		return nil
	}
	helper := log.NewHelper(log.With(logger, "module", "server/middleware/condition_checker"))
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			ctx = ensureSecurityContext(ctx)

			claims, err := auth.IdentityFromContext(ctx)
			if err != nil || claims == nil || claims.Subject == "" {
				return handler(ctx, req)
			}

			operation, ok := security.OperationFromContext(ctx)
			if !ok || operation == "" {
				helper.Warnf("condition check skipped: missing security operation")
				return handler(ctx, req)
			}

			if err := checker.CheckConditions(ctx, claims.Subject, operation); err != nil {
				return nil, err
			}

			return handler(ctx, req)
		}
	}
}
