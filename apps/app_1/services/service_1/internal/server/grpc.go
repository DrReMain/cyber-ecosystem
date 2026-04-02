package server

import (
	"context"

	jwt2 "github.com/golang-jwt/jwt/v5"
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
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"cyber-ecosystem/shared-go/kratos/middleware/auth"
	"cyber-ecosystem/shared-go/kratos/middleware/i18n"
	"cyber-ecosystem/shared-go/kratos/middleware/traceheader"
	"cyber-ecosystem/shared-go/kratos/middleware/validate"

	app1V1 "cyber-ecosystem/apps/app_1/gen/go/v1"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/conf"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/service"
)

func NewGRPCServer(
	c *conf.Server,
	ca *conf.Auth,
	logger log.Logger,
	registrar []service.Registrar,
	tp *tracesdk.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
	i18nBundle *i18n.Bundle,
) *grpc.Server {
	var middlewares = []middleware.Middleware{}
	middlewares = append(middlewares, i18n.Server(i18nBundle))
	middlewares = append(middlewares, recovery.Recovery(recovery.WithHandler(func(context.Context, any, any) error { return app1V1.ErrorErrorReasonUnspecified("") })))
	middlewares = append(middlewares, ratelimit.Server())
	middlewares = append(middlewares, metrics.Server(metrics.WithSeconds(_metricSeconds), metrics.WithRequests(_metricRequests)))
	if tp != nil {
		middlewares = append(middlewares, tracing.Server(tracing.WithTracerProvider(tp)))
	}
	middlewares = append(middlewares, metadata.Server())
	middlewares = append(middlewares, traceheader.Server())
	middlewares = append(middlewares, logging.Server(logger))
	middlewares = append(middlewares, selector.Server(jwt.Server(
		func(token *jwt2.Token) (any, error) { return []byte(ca.Secret), nil },
		jwt.WithSigningMethod(jwt2.SigningMethodHS256),
		jwt.WithClaims(func() jwt2.Claims { return &jwt2.MapClaims{} })),
	).Match(auth.NewWhiteListByPublicAccessInProtoMatcher()).Build())
	middlewares = append(middlewares, validate.ProtoValidate(app1V1.ErrorErrorReasonValidator("")))

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
