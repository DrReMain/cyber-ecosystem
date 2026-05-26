//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/genesis/services/base/internal/biz"
	"cyber-ecosystem/apps/genesis/services/base/internal/conf"
	"cyber-ecosystem/apps/genesis/services/base/internal/data"
	"cyber-ecosystem/apps/genesis/services/base/internal/i18n"
	"cyber-ecosystem/apps/genesis/services/base/internal/platform"
	"cyber-ecosystem/apps/genesis/services/base/internal/server"
	"cyber-ecosystem/apps/genesis/services/base/internal/service"
)

func wireApp(
	*conf.Server,
	*conf.Log,
	*conf.Data,
	*conf.Ops,
	log.Logger,
	*tracesdk.TracerProvider,
	*metricsdk.MeterProvider,
	metric.Int64Counter,
	metric.Float64Histogram,
) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
		i18n.ProviderSet,
		platform.ProviderSet,
		wire.Bind(new(biz.Transaction), new(*platform.Platform)),
		newApp,
	))
}
