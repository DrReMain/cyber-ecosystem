package server

import (
	"context"

	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/service"

	"github.com/DrReMain/cyber-ecosystem/gen/go/common"
	template1V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/middleware/responsemeta"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/middleware/traceheader"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/middleware/validate"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"

	"go.opentelemetry.io/otel/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/protobuf/proto"
)

func NewConnectServer(
	c *conf.Server,
	logger log.Logger,
	services []service.Registrar,
	tp *tracesdk.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) *connect.Server {
	var middlewares = []middleware.Middleware{recovery.Recovery(recovery.WithHandler(func(context.Context, any, any) error {
		return errors.InternalServer(template1V1.ErrorReason_ERROR_REASON_UNSPECIFIED.String(), "")
	}))}
	middlewares = append(middlewares, metrics.Server(metrics.WithSeconds(_metricSeconds), metrics.WithRequests(_metricRequests)))
	if tp != nil {
		middlewares = append(middlewares, tracing.Server(tracing.WithTracerProvider(tp)))
	}
	middlewares = append(middlewares, traceheader.Server())
	middlewares = append(middlewares, responsemeta.Server())
	middlewares = append(middlewares, logging.Server(logger))
	middlewares = append(middlewares, localizeErrorMiddleware(template1V1.ErrorReason_ERROR_REASON_UNSPECIFIED.String()))
	middlewares = append(middlewares, validate.ProtoValidate(template1V1.ErrorReason_ERROR_REASON_VALIDATOR.String(), validate.UseDefaultError))
	var opts = []connect.ServerOption{
		connect.Middleware(middlewares...),
		connect.ErrorEncoder(func(ctx context.Context, err error) error {
			return connect.NewErrorEncoder(resolveErrorMessage, func(_ context.Context, sourceErr error, se *errors.Error, message string) proto.Message {
				return &common.ErrorBody{
					Reason:  se.Reason,
					Message: message,
					Details: extractErrorDetails(sourceErr),
				}
			})(ctx, err)
		}),
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
	for _, svc := range services {
		svc.RegisterConnect(srv)
	}
	return srv
}
