//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"

	krauth "cyber-ecosystem/shared-go/kratos/security/auth"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/module/auth"
	"cyber-ecosystem/app/services/system/internal/module/dept"
	"cyber-ecosystem/app/services/system/internal/module/item"
	"cyber-ecosystem/app/services/system/internal/module/resource"
	"cyber-ecosystem/app/services/system/internal/module/user"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/server"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		platform.ProviderSet,
		wire.Bind(new(shared.Transaction), new(*platform.Platform)),
		wire.Bind(new(krauth.Authenticator), new(*auth.AuthUC)),
		auth.ProviderSet,
		user.ProviderSet,
		dept.ProviderSet,
		item.ProviderSet,
		resource.ProviderSet,
		newApp,
	))
}
