package security

import (
	"context"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
)

// ErrNotBackendPlane and ErrNotMemberPlane are the plane guards' rejections;
// rebound onto catalog factories at service assembly.
var (
	ErrNotBackendPlane = errors.Forbidden("NOT_BACKEND_PLANE", "")
	ErrNotMemberPlane  = errors.Forbidden("NOT_MEMBER_PLANE", "")
)

// BackendPlaneGuard gates a rung to the operator world: administrative faces
// answer only to backend-plane seats, and the verdict precedes grant
// evaluation — a wrong-plane session is rejected as such, never as a
// missing grant.
func BackendPlaneGuard() middleware.Middleware {
	return planeGuard(PlaneBackend, ErrNotBackendPlane)
}

// MemberPlaneGuard gates a rung to the consumer world: a service's C-end
// faces answer only to member-plane seats, so an operator session cannot
// ride the session-only rung into member surfaces.
func MemberPlaneGuard() middleware.Middleware {
	return planeGuard(PlaneMember, ErrNotMemberPlane)
}

func planeGuard(plane string, rejection error) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			subject, ok := SubjectFromCtx(ctx)
			if !ok || subject.Plane != plane {
				return nil, rejection
			}
			return handler(ctx, req)
		}
	}
}
