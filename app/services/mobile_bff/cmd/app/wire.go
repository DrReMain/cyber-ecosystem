//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"

	"cyber-ecosystem/app/services/mobile_bff/internal/biz"
	"cyber-ecosystem/app/services/mobile_bff/internal/conf"
	"cyber-ecosystem/app/services/mobile_bff/internal/data"
	"cyber-ecosystem/app/services/mobile_bff/internal/platform"
	"cyber-ecosystem/app/services/mobile_bff/internal/server"
	"cyber-ecosystem/app/services/mobile_bff/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		platform.ProviderSet,
		wire.Bind(new(platform.Transaction), new(*platform.Platform)),
		newApp,
	))
}
