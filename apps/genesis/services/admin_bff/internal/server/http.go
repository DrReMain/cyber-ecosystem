package server

import (
	"github.com/gorilla/handlers"
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
	krahttp "github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/middleware/error_report"
	"cyber-ecosystem/shared-go/kratos/middleware/i18n"
	"cyber-ecosystem/shared-go/kratos/middleware/validate"

	"cyber-ecosystem/apps/genesis/services/admin_bff/internal/conf"
	"cyber-ecosystem/apps/genesis/services/admin_bff/internal/service"
)

func NewHTTPServer(
	c *conf.Server,
	logger log.Logger,
	registrar []service.Registrar,
	tp *trace.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
	i18nBundle *i18n.Bundle,
) *krahttp.Server {
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

	var opts = []krahttp.ServerOption{
		krahttp.Middleware(middlewares...),
		krahttp.Filter(handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Content-Type", i18n.DefaultHeaderLang, "Authorization"}),
			handlers.MaxAge(86400),
		)),
	}
	if c.Http.Network != "" {
		opts = append(opts, krahttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, krahttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, krahttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := krahttp.NewServer(opts...)
	for _, r := range registrar {
		r.RegisterHTTP(srv)
	}
	return srv
}
