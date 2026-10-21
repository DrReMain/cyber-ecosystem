package security_test

import (
	"context"
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"
)

func TestPlaneGuards(t *testing.T) {
	handler := func(ctx context.Context, _ any) (any, error) { return "ok", nil }

	cases := []struct {
		name  string
		guard func(ctx context.Context, req any) (any, error)
		sub   *security.Subject
		want  error
	}{
		{"backend seat passes the backend gate", security.BackendPlaneGuard()(handler), &security.Subject{UserID: "u1", Plane: security.PlaneBackend}, nil},
		{"member seat rejected at the backend gate", security.BackendPlaneGuard()(handler), &security.Subject{UserID: "u1", Plane: security.PlaneMember}, security.ErrNotBackendPlane},
		{"member seat passes the member gate", security.MemberPlaneGuard()(handler), &security.Subject{UserID: "u1", Plane: security.PlaneMember}, nil},
		{"backend seat rejected at the member gate", security.MemberPlaneGuard()(handler), &security.Subject{UserID: "u1", Plane: security.PlaneBackend}, security.ErrNotMemberPlane},
		{"no subject rejected at either gate", security.MemberPlaneGuard()(handler), nil, security.ErrNotMemberPlane},
	}
	for _, tc := range cases {
		ctx := context.Background()
		if tc.sub != nil {
			ctx = security.WithSubject(ctx, tc.sub)
		}
		_, err := tc.guard(ctx, nil)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
}
