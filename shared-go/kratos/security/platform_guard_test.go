package security_test

import (
	"context"
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"
)

func TestPlatformGuard(t *testing.T) {
	guard := security.PlatformGuard()
	handler := func(ctx context.Context, _ any) (any, error) { return "ok", nil }

	cases := []struct {
		name string
		sub  *security.Subject
		want error
	}{
		{"platform inhabitant passes", &security.Subject{UserID: "u1", TenantID: security.DefaultTenant}, nil},
		{"foreign tenant rejected", &security.Subject{UserID: "u1", TenantID: "acme"}, security.ErrNotPlatformPlane},
		{"no subject rejected", nil, security.ErrNotPlatformPlane},
	}
	for _, tc := range cases {
		ctx := context.Background()
		if tc.sub != nil {
			ctx = security.WithSubject(ctx, tc.sub)
		}
		_, err := guard(handler)(ctx, nil)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
}
