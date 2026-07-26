// Package sanitize masks non-kratos errors as a generic internal error, so
// implementation details are not exposed.
package sanitize

import (
	"context"
	"errors"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
)

// ErrUnexpected is the fallback for any error that is not already a kratos
// *Error. It is a plain kratos error, so the package does not depend on any
// project-specific error scheme.
var ErrUnexpected = kratoserrors.New(500, "INTERNAL", "internal server error")

// Server returns middleware that masks any non-kratos error as ErrUnexpected,
// keeping the original error as the cause. Errors that are already kratos
// errors pass through unchanged.
func Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			reply, err := handler(ctx, req)
			if err == nil {
				return reply, nil
			}
			var ke *kratoserrors.Error
			if errors.As(err, &ke) {
				return reply, err
			}
			return nil, ErrUnexpected.WithCause(err)
		}
	}
}
