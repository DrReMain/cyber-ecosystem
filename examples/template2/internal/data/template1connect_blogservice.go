package data

import (
	"context"
	"fmt"

	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/conf"

	template1V1connect "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1/template1V1connect"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/i18n"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/middleware/headerforward"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"

	"go.opentelemetry.io/otel/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func NewTemplate1ConnectBlogService(
	c *conf.Data,
	logger log.Logger,
	tp *tracesdk.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) (template1V1connect.BlogServiceClient, error) {
	var middlewares = []middleware.Middleware{recovery.Recovery()}
	middlewares = append(middlewares, headerforward.ClientFromServer(i18n.HeaderAcceptLanguage))
	middlewares = append(middlewares, metrics.Client(metrics.WithSeconds(_metricSeconds), metrics.WithRequests(_metricRequests)))
	if tp != nil {
		middlewares = append(middlewares, tracing.Client(tracing.WithTracerProvider(tp)))
	}
	middlewares = append(middlewares, logging.Client(logger))
	conn, err := connect.DialInsecure(
		context.Background(),
		connect.WithEndpoint(c.ServiceTemplate1Connect.Addr),
		connect.WithTimeout(c.ServiceTemplate1Connect.Timeout.AsDuration()),
		connect.WithMiddleware(middlewares...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial template1 Connect service: %w", err)
	}
	return template1V1connect.NewBlogServiceClient(conn.HTTPClient(), conn.Endpoint(), conn.ClientOptions()...), nil
}
