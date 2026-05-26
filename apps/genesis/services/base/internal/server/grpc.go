package server

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"cyber-ecosystem/shared-go/kratos/middleware/error_report"
	"cyber-ecosystem/shared-go/kratos/middleware/i18n"
	"cyber-ecosystem/shared-go/kratos/middleware/validate"

	"cyber-ecosystem/apps/genesis/services/base/internal/conf"
	"cyber-ecosystem/apps/genesis/services/base/internal/service"
)

func NewGRPCServer(
	c *conf.Server,
	logger log.Logger,
	registrar []service.Registrar,
	tp *trace.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
	i18nBundle *i18n.Bundle,
) *grpc.Server {
	var middlewares []middleware.Middleware
	middlewares = append(middlewares, i18n.Server(i18nBundle))
	middlewares = append(middlewares, recovery.Recovery())
	middlewares = append(middlewares, ratelimit.Server())
	middlewares = append(middlewares, metrics.Server(metrics.WithSeconds(_metricSeconds), metrics.WithRequests(_metricRequests)))
	if tp != nil {
		middlewares = append(middlewares, tracing.Server(tracing.WithTracerProvider(tp)))
	}
	middlewares = append(middlewares, metadata.Server())
	middlewares = append(middlewares, logging.Server(logger))
	middlewares = append(middlewares, validate.ProtoValidate())
	middlewares = append(middlewares, error_report.Server())

	var opts = []grpc.ServerOption{
		grpc.Middleware(middlewares...),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	for _, r := range registrar {
		r.RegisterGRPC(srv)
	}
	return srv
}
