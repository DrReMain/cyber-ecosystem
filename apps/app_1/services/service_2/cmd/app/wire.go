//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/app_1/services/service_2/internal/biz"
	"cyber-ecosystem/apps/app_1/services/service_2/internal/conf"
	"cyber-ecosystem/apps/app_1/services/service_2/internal/data"
	"cyber-ecosystem/apps/app_1/services/service_2/internal/i18n"
	"cyber-ecosystem/apps/app_1/services/service_2/internal/server"
	"cyber-ecosystem/apps/app_1/services/service_2/internal/service"
)

func wireApp(
	*conf.Server,
	*conf.Auth,
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
		newApp,
	))
}
