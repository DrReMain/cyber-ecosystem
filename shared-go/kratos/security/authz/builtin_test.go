package authz_test

import (
	"slices"
	"testing"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	// Register the generated descriptors the collector walks; system carries
	// the known builtin set the assertions below pin.
	_ "cyber-ecosystem/gen/go/cyber/system/v1"
)

func TestBuiltinOperations(t *testing.T) {
	ops := kauthz.BuiltinOperations("cyber/system/v1/")
	for _, want := range []string{
		"/cyber.system.v1.AuthService/Logout",
		"/cyber.system.v1.AuthService/GetCurrentUser",
		"/cyber.system.v1.IntrospectService/VerifySession",
		"/cyber.system.v1.IntrospectService/CheckOperation",
		"/cyber.system.v1.UserService/ChangePassword",
	} {
		if !slices.Contains(ops, want) {
			t.Errorf("builtin set missing %s", want)
		}
	}
	if slices.Contains(ops, "/cyber.system.v1.UserService/CreateUser") {
		t.Error("non-builtin operation collected")
	}
	if got := kauthz.BuiltinOperations("cyber/nothing/v1/"); len(got) != 0 {
		t.Errorf("unknown prefix returned %v", got)
	}
}
