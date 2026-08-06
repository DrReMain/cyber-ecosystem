package server

import (
	"log/slog"

	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/logging"
	"github.com/go-kratos/kratos/v3/middleware/metadata"
	"github.com/go-kratos/kratos/v3/middleware/ratelimit"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/middleware/selector"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/middleware/sanitize"
	"cyber-ecosystem/shared-go/kratos/middleware/validator"
	"cyber-ecosystem/shared-go/kratos/observability"
	"cyber-ecosystem/shared-go/kratos/security"
	krauth "cyber-ecosystem/shared-go/kratos/security/auth"

	extv1 "cyber-ecosystem/gen/go/cyber/ext/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
)

func NewHTTPServer(
	c *conf.Server,
	logger *slog.Logger,
	registrar []Registrar,
	authn krauth.Authenticator,
) *http.Server {
	var middlewares []middleware.Middleware
	middlewares = append(middlewares, sanitize.Server())
	middlewares = append(middlewares, tracing.Server())
	middlewares = append(middlewares, observability.MetricsServer())
	middlewares = append(middlewares, logging.Server(logger))
	middlewares = append(middlewares, recovery.Recovery(recovery.WithLogger(observability.PrettyProto(logger))))
	middlewares = append(middlewares, ratelimit.Server())
	middlewares = append(middlewares, metadata.Server())
	middlewares = append(middlewares, selector.Server(
		krauth.BearerAuth(authn),
	).Match(security.MatchAccess(extv1.Access_ACCESS_ADMIN)).Build())
	middlewares = append(middlewares, selector.Server(security.DefaultGuard()).Match(security.MatchAccess(extv1.Access_ACCESS_UNSPECIFIED)).Build())
	middlewares = append(middlewares, validator.Server())

	var opts = []http.ServerOption{
		http.Middleware(middlewares...),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	for _, r := range registrar {
		r.RegisterHTTP(srv)
	}
	return srv
}
