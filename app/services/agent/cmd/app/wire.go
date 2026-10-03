//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"

	"cyber-ecosystem/app/services/agent/internal/bootstrap"
	"cyber-ecosystem/app/services/agent/internal/client"
	"cyber-ecosystem/app/services/agent/internal/conf"
	"cyber-ecosystem/app/services/agent/internal/module/agentconfig"
	"cyber-ecosystem/app/services/agent/internal/module/agentconfigadmin"
	"cyber-ecosystem/app/services/agent/internal/module/authz"
	"cyber-ecosystem/app/services/agent/internal/platform"
	"cyber-ecosystem/app/services/agent/internal/server"
	"cyber-ecosystem/app/services/agent/internal/shared"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Remote, *conf.Crypto, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		wire.Bind(new(shared.Transaction), new(*platform.Platform)),
		server.ProviderSet,
		platform.ProviderSet,
		bootstrap.ProviderSet,
		client.ProviderSet,
		authz.ProviderSet,
		agentconfig.ProviderSet,
		agentconfigadmin.ProviderSet,
		newApp,
	))
}
