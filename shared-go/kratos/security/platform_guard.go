package security

import (
	"context"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
)

// ErrNotPlatformPlane is the platform guard's rejection; rebound onto a
// catalog factory at service assembly.
var ErrNotPlatformPlane = errors.Forbidden("NOT_PLATFORM_PLANE", "")

// PlatformGuard gates the platform plane: registry and other platform-scope
// faces answer only to inhabitants of the default tenant. Grants alone are
// not enough — a broadly granted tenant admin must not reach across tenants.
func PlatformGuard() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			subject, ok := SubjectFromCtx(ctx)
			if !ok || subject.TenantID != DefaultTenant {
				return nil, ErrNotPlatformPlane
			}
			return handler(ctx, req)
		}
	}
}
