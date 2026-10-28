package auth

import (
	"github.com/google/wire"

	krauth "cyber-ecosystem/shared-go/kratos/security/auth"
)

var ProviderSet = wire.NewSet(
	NewAuthUC,
	NewTokenRP,
	NewIdentityProviders,
	NewSmsProvider,
	NewWechatProvider,
	NewAuthService,
	wire.Bind(new(krauth.Authenticator), new(*AuthUC)),
)
