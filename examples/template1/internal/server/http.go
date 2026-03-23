package server

import (
	"strings"
	"time"

	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/service"

	"github.com/DrReMain/cyber-ecosystem/gen/go/common"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/encoder"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/middleware/validate"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func NewHTTPServer(
	c *conf.Server,
	logger log.Logger,
	services []service.Registrar,
	tp *tracesdk.TracerProvider,
	cMetrics *conf.Metrics,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) *http.Server {
	var middlewares = []middleware.Middleware{recovery.Recovery()}
	middlewares = append(middlewares, metrics.Server(metrics.WithSeconds(_metricSeconds), metrics.WithRequests(_metricRequests)))
	if tp != nil {
		middlewares = append(middlewares, tracing.Server(tracing.WithTracerProvider(tp)))
	}
	middlewares = append(middlewares, logging.Server(logger))
	middlewares = append(middlewares, validate.ProtoValidate(validate.UseProtoMessage))
	var opts = []http.ServerOption{
		http.Middleware(middlewares...),
		http.ResponseEncoder(encoder.NewResponseEncoder(ResponseBuildBody)),
		http.ErrorEncoder(encoder.NewErrorEncoder(ErrorBuildBody)),
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
	if cMetrics != nil && len(cMetrics.Path) > 1 && strings.HasPrefix(cMetrics.Path, "/") {
		srv.Handle(cMetrics.Path, promhttp.Handler())
	}
	for _, svc := range services {
		svc.RegisterHTTP(srv)
	}
	return srv
}

func ResponseBuildBody(v any) (any, error) {
	reply := &common.Reply{
		T:       time.Now().UnixMilli(),
		Success: true,
		Msg:     "OK",
		Result:  nil,
	}
	if m, ok := v.(proto.Message); ok {
		if anyVal, err := anypb.New(m); err != nil {
			return nil, err
		} else {
			reply.Result = anyVal
		}
	}
	return reply, nil
}

func ErrorBuildBody(err *errors.Error) any {
	return &common.Reply{
		T:       time.Now().UnixMilli(),
		Success: false,
		Msg:     err.Message,
		Result:  nil,
	}
}
