//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"

	"cyber-ecosystem/app/services/edge_mobile/internal/biz"
	"cyber-ecosystem/app/services/edge_mobile/internal/conf"
	"cyber-ecosystem/app/services/edge_mobile/internal/data"
	"cyber-ecosystem/app/services/edge_mobile/internal/platform"
	"cyber-ecosystem/app/services/edge_mobile/internal/server"
	"cyber-ecosystem/app/services/edge_mobile/internal/service"
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
		newApp,
	))
}
