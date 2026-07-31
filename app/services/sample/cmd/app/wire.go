//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"

	"cyber-ecosystem/app/services/sample/internal/biz"
	"cyber-ecosystem/app/services/sample/internal/client"
	"cyber-ecosystem/app/services/sample/internal/conf"
	"cyber-ecosystem/app/services/sample/internal/server"
	"cyber-ecosystem/app/services/sample/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Remote, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		client.ProviderSet,
		newApp,
	))
}
