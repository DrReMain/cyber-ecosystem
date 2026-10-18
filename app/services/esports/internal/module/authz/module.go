package authz

import (
	"github.com/google/wire"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

var ProviderSet = wire.NewSet(
	NewAuthzUC,
	wire.Bind(new(kauthz.Authorizer), new(*AuthzUC)),
)
