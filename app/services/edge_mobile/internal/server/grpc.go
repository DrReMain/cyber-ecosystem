package server

import (
	"log/slog"

	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/logging"
	"github.com/go-kratos/kratos/v3/middleware/metadata"
	"github.com/go-kratos/kratos/v3/middleware/ratelimit"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/transport/grpc"

	"cyber-ecosystem/shared-go/kratos/middleware/validator"
	"cyber-ecosystem/shared-go/kratos/observability"

	"cyber-ecosystem/app/services/edge_mobile/internal/conf"
	"cyber-ecosystem/app/services/edge_mobile/internal/service"
)

func NewGRPCServer(
	c *conf.Server,
	logger *slog.Logger,
	registrar []service.Registrar,
) *grpc.Server {
	var middlewares []middleware.Middleware
	middlewares = append(middlewares, tracing.Server())
	middlewares = append(middlewares, observability.MetricsServer())
	middlewares = append(middlewares, logging.Server(logger))
	middlewares = append(middlewares, recovery.Recovery(recovery.WithLogger(observability.PrettyProto(logger))))
	middlewares = append(middlewares, ratelimit.Server())
	middlewares = append(middlewares, metadata.Server())
	middlewares = append(middlewares, validator.Server())

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
