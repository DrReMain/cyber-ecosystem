package server

import (
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/service"

	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/middleware/validate"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"go.opentelemetry.io/otel/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func NewGRPCServer(
	c *conf.Server,
	logger log.Logger,
	services []service.Registrar,
	tp *tracesdk.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			metrics.Server(
				metrics.WithSeconds(_metricSeconds),
				metrics.WithRequests(_metricRequests),
			),
			tracing.Server(tracing.WithTracerProvider(tp)),
			logging.Server(logger),
			validate.ProtoValidate(validate.UseProtoMessage),
		),
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
	for _, svc := range services {
		svc.RegisterGRPC(srv)
	}
	return srv
}
