package server

import (
	"log/slog"

	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/logging"
	"github.com/go-kratos/kratos/v3/middleware/metadata"
	"github.com/go-kratos/kratos/v3/middleware/ratelimit"
	"github.com/go-kratos/kratos/v3/middleware/recovery"

	"cyber-ecosystem/shared-go/kratos/middleware/sanitize"
	"cyber-ecosystem/shared-go/kratos/middleware/validator"
	"cyber-ecosystem/shared-go/kratos/observability"
	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/kratos/transport/connect/health"
	"cyber-ecosystem/shared-go/kratos/transport/connect/reflection"

	"cyber-ecosystem/app/services/sample/internal/conf"
	"cyber-ecosystem/app/services/sample/internal/service"
)

func NewConnectServer(
	c *conf.Server,
	logger *slog.Logger,
	registrar []service.Registrar,
) *connect.Server {
	var middlewares []middleware.Middleware
	middlewares = append(middlewares, sanitize.Server())
	middlewares = append(middlewares, tracing.Server())
	middlewares = append(middlewares, observability.MetricsServer())
	middlewares = append(middlewares, logging.Server(logger))
	middlewares = append(middlewares, recovery.Recovery(recovery.WithLogger(observability.PrettyProto(logger))))
	middlewares = append(middlewares, ratelimit.Server())
	middlewares = append(middlewares, metadata.Server())
	middlewares = append(middlewares, validator.Server())

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
	for _, r := range registrar {
		r.RegisterConnect(srv)
	}
	if _, err := health.Register(srv); err != nil {
		logger.Warn("connect health registration failed", "error", err)
	}
	if err := reflection.Register(srv); err != nil {
		logger.Warn("connect reflection registration failed", "error", err)
	}
	return srv
}
