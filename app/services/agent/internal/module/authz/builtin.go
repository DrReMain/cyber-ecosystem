package authz

import (
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	_ "cyber-ecosystem/gen/go/cyber/agent/v1"
)

var builtinOperations = kauthz.BuiltinOperations("cyber/agent/v1/")
