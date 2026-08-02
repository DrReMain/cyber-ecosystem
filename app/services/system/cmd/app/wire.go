//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"

	krauth "cyber-ecosystem/shared-go/kratos/security/auth"

	"cyber-ecosystem/app/services/system/internal/biz"
	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/data"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/server"
	"cyber-ecosystem/app/services/system/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		platform.ProviderSet,
		wire.Bind(new(biz.Transaction), new(*platform.Platform)),
		wire.Bind(new(krauth.Authenticator), new(*biz.AuthUC)),
		newApp,
	))
}
