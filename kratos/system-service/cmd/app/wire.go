//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/biz"
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/data"
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/server"
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
		newApp,
	))
}
