//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"

	"cyber-ecosystem/app/services/system/internal/bootstrap"
	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/module/audit"
	"cyber-ecosystem/app/services/system/internal/module/auth"
	"cyber-ecosystem/app/services/system/internal/module/authz"
	"cyber-ecosystem/app/services/system/internal/module/dept"
	"cyber-ecosystem/app/services/system/internal/module/file"
	"cyber-ecosystem/app/services/system/internal/module/filepresign"
	"cyber-ecosystem/app/services/system/internal/module/fileproxy"
	"cyber-ecosystem/app/services/system/internal/module/policy"
	"cyber-ecosystem/app/services/system/internal/module/resource"
	"cyber-ecosystem/app/services/system/internal/module/role"
	"cyber-ecosystem/app/services/system/internal/module/transfer"
	"cyber-ecosystem/app/services/system/internal/module/user"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/server"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Authz, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		wire.Bind(new(shared.Transaction), new(*platform.Platform)),
		wire.Bind(new(shared.HookRegistry), new(*bootstrap.Lifecycle)),
		server.ProviderSet,
		platform.ProviderSet,
		authz.ProviderSet,
		bootstrap.ProviderSet,
		audit.ProviderSet,
		user.ProviderSet,
		auth.ProviderSet,
		dept.ProviderSet,
		file.ProviderSet,
		filepresign.ProviderSet,
		fileproxy.ProviderSet,
		policy.ProviderSet,
		resource.ProviderSet,
		role.ProviderSet,
		transfer.ProviderSet,
		newApp,
	))
}
