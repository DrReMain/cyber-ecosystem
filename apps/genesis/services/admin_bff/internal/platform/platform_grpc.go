package platform

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/circuitbreaker"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	genesisV1 "cyber-ecosystem/apps/genesis/gen/go/v1"
	"cyber-ecosystem/apps/genesis/services/admin_bff/internal/conf"
)

func dialGRPC(
	c *conf.Data,
	logger log.Logger,
	tp *trace.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) (*grpc.ClientConn, func(), error) {
	var middlewares []middleware.Middleware
	middlewares = append(middlewares, recovery.Recovery())
	middlewares = append(middlewares, circuitbreaker.Client())
	middlewares = append(middlewares, metrics.Client(metrics.WithSeconds(_metricSeconds), metrics.WithRequests(_metricRequests)))
	if tp != nil {
		middlewares = append(middlewares, tracing.Client(tracing.WithTracerProvider(tp)))
	}
	middlewares = append(middlewares, metadata.Client())
	middlewares = append(middlewares, logging.Client(logger))

	conn, err := kgrpc.DialInsecure(
		context.Background(),
		kgrpc.WithEndpoint(c.BaseService.Addr),
		kgrpc.WithTimeout(c.BaseService.Timeout.AsDuration()),
		kgrpc.WithMiddleware(middlewares...),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial base service: %w", err)
	}

	helper := log.NewHelper(log.With(logger, "module", "platform/grpc"))
	helper.Infof("gRPC client connected to %s", c.BaseService.Addr)

	return conn,
		func() {
			if err := conn.Close(); err != nil {
				helper.Warnf("failed to close gRPC connection: %v", err)
			}
		},
		nil
}

func NewGRPCResourceClient(
	c *conf.Data,
	logger log.Logger,
	tp *trace.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) (genesisV1.ResourceServiceClient, func(), error) {
	conn, cleanup, err := dialGRPC(c, logger, tp, _metricRequests, _metricSeconds)
	if err != nil {
		return nil, nil, err
	}
	return genesisV1.NewResourceServiceClient(conn), cleanup, nil
}

func NewGRPCArticleClient(
	c *conf.Data,
	logger log.Logger,
	tp *trace.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) (genesisV1.ArticleServiceClient, func(), error) {
	conn, cleanup, err := dialGRPC(c, logger, tp, _metricRequests, _metricSeconds)
	if err != nil {
		return nil, nil, err
	}
	return genesisV1.NewArticleServiceClient(conn), cleanup, nil
}
