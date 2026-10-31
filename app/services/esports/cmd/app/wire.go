//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"

	"cyber-ecosystem/app/services/esports/internal/bootstrap"
	"cyber-ecosystem/app/services/esports/internal/client"
	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/module/authz"
	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/entitleadmin"
	"cyber-ecosystem/app/services/esports/internal/module/game"
	"cyber-ecosystem/app/services/esports/internal/module/match"
	"cyber-ecosystem/app/services/esports/internal/module/member"
	"cyber-ecosystem/app/services/esports/internal/module/order"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/module/report"
	"cyber-ecosystem/app/services/esports/internal/module/sku"
	"cyber-ecosystem/app/services/esports/internal/module/test"
	"cyber-ecosystem/app/services/esports/internal/module/testadmin"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/server"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Remote, *conf.Payment, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		wire.Bind(new(shared.Transaction), new(*platform.Platform)),
		wire.Bind(new(shared.HookRegistry), new(*bootstrap.Lifecycle)),
		wire.Bind(new(shared.DomainEventBus), new(*bootstrap.EventSubscriptions)),
		server.ProviderSet,
		platform.ProviderSet,
		bootstrap.ProviderSet,
		client.ProviderSet,
		authz.ProviderSet,
		profile.ProviderSet,
		game.ProviderSet,
		match.ProviderSet,
		member.ProviderSet,
		order.ProviderSet,
		sku.ProviderSet,
		test.ProviderSet,
		testadmin.ProviderSet,
		entitle.ProviderSet,
		entitleadmin.ProviderSet,
		report.ProviderSet,
		newApp,
	))
}
