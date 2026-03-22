//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/biz"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/server"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func wireApp(
	*conf.Server,
	*conf.Log,
	*conf.Data,
	*conf.Metrics,
	log.Logger,
	*tracesdk.TracerProvider,
	metric.Int64Counter,
	metric.Float64Histogram,
	*metricsdk.MeterProvider,
) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
		newApp,
	))
}
