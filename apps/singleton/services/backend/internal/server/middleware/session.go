package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"

	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/auth"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/security"
)

// SessionValidator returns middleware that checks JWT session revocation.
// Must be placed AFTER the JWT middleware so claims are already in context.
func SessionValidator(enabled bool, logger log.Logger, validator security.SessionValidator) middleware.Middleware {
	if !enabled || validator == nil {
		return nil
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			claims, err := auth.IdentityFromContext(ctx)
			if err != nil {
				return nil, err
			}

			if err := validator.ValidateSession(ctx, claims.Sid); err != nil {
				return nil, err
			}

			return handler(ctx, req)
		}
	}
}
