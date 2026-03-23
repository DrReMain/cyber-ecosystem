package server

import (
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/service"

	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/middleware/validate"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"

	"go.opentelemetry.io/otel/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func NewConnectServer(
	c *conf.Server,
	logger log.Logger,
	services []service.Registrar,
	tp *tracesdk.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) *connect.Server {
	var middlewares = []middleware.Middleware{recovery.Recovery()}
	middlewares = append(middlewares, metrics.Server(metrics.WithSeconds(_metricSeconds), metrics.WithRequests(_metricRequests)))
	if tp != nil {
		middlewares = append(middlewares, tracing.Server(tracing.WithTracerProvider(tp)))
	}
	middlewares = append(middlewares, logging.Server(logger))
	middlewares = append(middlewares, validate.ProtoValidate(validate.UseProtoMessage))
	var opts = []connect.ServerOption{
		connect.Middleware(middlewares...),
	}
	if c.Connect.Network != "" {
		opts = append(opts, connect.Network(c.Connect.Network))
	}
	if c.Connect.Addr != "" {
		opts = append(opts, connect.Address(c.Connect.Addr))
	}
	if c.Connect.Timeout != nil {
		opts = append(opts, connect.Timeout(c.Connect.Timeout.AsDuration()))
	}
	srv := connect.NewServer(opts...)
	for _, svc := range services {
		svc.RegisterConnect(srv)
	}
	return srv
}
