package auth

import (
	"github.com/google/wire"

	krauth "cyber-ecosystem/shared-go/kratos/security/auth"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// risk: irreversible
var ProviderSet = wire.NewSet(
	NewAuthUC,
	NewTokenRP,
	NewIdentityProviders,
	NewSmsProvider,
	NewWechatProvider,
	NewAuthService,
	wire.Bind(new(krauth.Authenticator), new(*AuthUC)),
	wire.Bind(new(shared.PasswordPolicy), new(*AuthUC)),
)
