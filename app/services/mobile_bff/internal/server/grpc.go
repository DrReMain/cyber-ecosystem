package server

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/logging"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/transport/grpc"

	"cyber-ecosystem/app/services/mobile_bff/internal/conf"
	"cyber-ecosystem/app/services/mobile_bff/internal/service"
)

func NewGRPCServer(
	c *conf.Server,
	logger *slog.Logger,
	registrar []service.Registrar,
) *grpc.Server {
	var middlewares []middleware.Middleware
	middlewares = append(middlewares, recovery.Recovery())
	middlewares = append(middlewares, logging.Server(logger))

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
