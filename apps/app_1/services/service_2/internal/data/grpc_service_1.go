package data

import (
	"context"
	"fmt"

	jwt2 "github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/circuitbreaker"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	app1V1 "cyber-ecosystem/apps/app_1/gen/go/v1"
	"cyber-ecosystem/apps/app_1/services/service_2/internal/conf"
)

func NewGRPCClientService1(
	c *conf.Data,
	ca *conf.Auth,
	logger log.Logger,
	tp *tracesdk.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) (app1V1.BlogServiceClient, error) {
	var middlewares []middleware.Middleware
	middlewares = append(middlewares, recovery.Recovery(recovery.WithHandler(func(context.Context, any, any) error { return app1V1.ErrorErrorReasonUnspecified("") })))
	middlewares = append(middlewares, circuitbreaker.Client())
	middlewares = append(middlewares, metrics.Client(metrics.WithSeconds(_metricSeconds), metrics.WithRequests(_metricRequests)))
	if tp != nil {
		middlewares = append(middlewares, tracing.Client(tracing.WithTracerProvider(tp)))
	}
	middlewares = append(middlewares, metadata.Client())
	middlewares = append(middlewares, jwt.Client(
		func(token *jwt2.Token) (any, error) { return []byte(ca.Secret), nil },
		jwt.WithSigningMethod(jwt2.SigningMethodHS256),
		jwt.WithClaims(func() jwt2.Claims { return &jwt2.MapClaims{} }),
	))
	middlewares = append(middlewares, logging.Client(logger))

	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(c.GrpcService_1.Addr),
		grpc.WithTimeout(c.GrpcService_1.Timeout.AsDuration()),
		grpc.WithMiddleware(middlewares...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial service_1 gRPC service: %w", err)
	}
	return app1V1.NewBlogServiceClient(conn), nil
}
