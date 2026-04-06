package server

import (
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"go.opentelemetry.io/otel/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"

	authmw "cyber-ecosystem/shared-go/kratos/middleware/auth"
	"cyber-ecosystem/shared-go/kratos/middleware/i18n"
	"cyber-ecosystem/shared-go/kratos/middleware/validate"

	"cyber-ecosystem/apps/app_1/services/service_1/internal/conf"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/service"
)

func NewHTTPServer(
	c *conf.Server,
	ca *conf.Auth,
	logger log.Logger,
	registrar []service.Registrar,
	tp *tracesdk.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
	i18nBundle *i18n.Bundle,
) *http.Server {
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
	middlewares = append(middlewares, selector.Server(
		jwt.Server(
			func(token *jwtv5.Token) (any, error) { return []byte(ca.Secret), nil },
			jwt.WithSigningMethod(jwtv5.SigningMethodHS256),
			jwt.WithClaims(func() jwtv5.Claims { return &jwtv5.MapClaims{} }),
		),
	).Match(authmw.NewWhiteListByPublicAccessInProtoMatcher()).Build())
	middlewares = append(middlewares, validate.ProtoValidate())

	var opts = []http.ServerOption{
		http.Middleware(middlewares...),
		http.Filter(handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Content-Type", "Accept-Language", "Authorization"}),
			// handlers.AllowCredentials(),
			handlers.MaxAge(86400), // 24 * 60 * 60
		)),
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
