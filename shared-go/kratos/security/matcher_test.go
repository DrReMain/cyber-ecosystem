package security_test

import (
	"context"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"

	extv1 "cyber-ecosystem/gen/go/cyber/ext/v1"
)

func TestMatchAccess(t *testing.T) {
	// Annotation levels come from the generated descriptors registered by
	// this package's blank imports.
	tests := []struct {
		name  string
		op    string
		want  extv1.Access
		match bool
	}{
		{"admin op matches the admin audience", "/cyber.system.v1.UserService/ListUsers", extv1.Access_ACCESS_ADMIN, true},
		{"admin op does not match the public audience", "/cyber.system.v1.UserService/ListUsers", extv1.Access_ACCESS_PUBLIC, false},
		{"public op matches the public audience", "/cyber.system.v1.AuthService/Login", extv1.Access_ACCESS_PUBLIC, true},
		{"public op does not match the admin audience", "/cyber.system.v1.AuthService/Login", extv1.Access_ACCESS_ADMIN, false},
		{"unregistered op matches no declared audience", "/cyber.nope.v1.Ghost/Missing", extv1.Access_ACCESS_ADMIN, false},
		{"unregistered op falls to the unspecified branch", "/cyber.nope.v1.Ghost/Missing", extv1.Access_ACCESS_UNSPECIFIED, true},
		{"non-method descriptor falls to the unspecified branch", "/cyber.shared.common.v1/PageRequest", extv1.Access_ACCESS_UNSPECIFIED, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := security.MatchAccess(tt.want)
			// Second call exercises the per-operation cache path.
			if got := fn(context.Background(), tt.op); got != tt.match {
				t.Fatalf("match = %v, want %v", got, tt.match)
			}
			if got := fn(context.Background(), tt.op); got != tt.match {
				t.Fatalf("cached match = %v, want %v", got, tt.match)
			}
		})
	}
}
