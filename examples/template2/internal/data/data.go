package data

import (
	"context"
	"fmt"
	"time"

	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/biz"
	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent"
	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent/migrate"
	_ "github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent/runtime"

	template1V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"
	template1V1connect "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1/template1V1connect"
	entlog "github.com/DrReMain/cyber-ecosystem/shared-go/kratos/logging/ent"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect"
	"github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/client"
	"github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/entutil"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"entgo.io/ent/dialect"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type Data struct {
	db *ent.Client

	template1BlogService        template1V1.BlogServiceClient
	template1ConnectBlogService template1V1connect.BlogServiceClient
}

func NewData(
	c *conf.Data,
	logger log.Logger,
	db *ent.Client,
	template1BlogService template1V1.BlogServiceClient,
	template1ConnectBlogService template1V1connect.BlogServiceClient,
) (*Data, func(), error) {
	return &Data{
		db:                          db,
		template1BlogService:        template1BlogService,
		template1ConnectBlogService: template1ConnectBlogService,
	}, func() { db.Close() }, nil
}

func (d *Data) getClient(ctx context.Context) *ent.Client {
	return entutil.GetClientFromTx(ctx,
		ent.TxFromContext,
		func(tx *ent.Tx) *ent.Client { return tx.Client() },
		d.db,
	)
}

func (d *Data) InTx(ctx context.Context, fn func(context.Context) error) error {
	return entutil.InTx(ctx,
		ent.TxFromContext,
		ent.NewTxContext,
		d.db.Tx,
		fn,
	)
}

func NewEntClient(c *conf.Data, cLog *conf.Log, logger log.Logger, meterProvider *metricsdk.MeterProvider) (*ent.Client, error) {
	entClient, err := client.NewEntClient(client.DBConfig{
		Driver:          c.Database.Driver,
		Host:            c.Database.Host,
		Port:            int(c.Database.Port),
		User:            c.Database.User,
		Password:        c.Database.Password,
		DBName:          c.Database.DbName,
		MaxOpenConns:    int(c.Database.MaxOpenConns),
		MaxIdleConns:    int(c.Database.MaxIdleConns),
		ConnMaxLifetime: c.Database.ConnMaxLifetime.AsDuration(),
		MeterProvider:   meterProvider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to database: %w", err)
	}

	var finalDrv dialect.Driver = entClient.Driver
	if cLog != nil && cLog.Ent != nil && cLog.Ent.Enabled {
		slowQueryThreshold := 200 * time.Millisecond
		if cLog.Ent.SlowQueryThreshold != nil {
			slowQueryThreshold = cLog.Ent.SlowQueryThreshold.AsDuration()
		}
		finalDrv = entlog.NewDriverWrapper(entClient.Driver, logger, cLog.Ent.Level, cLog.Ent.SlowQuery, slowQueryThreshold)
	}
	db := ent.NewClient(ent.Driver(finalDrv))

	if c.Database.Migrate {
		if err := db.Schema.Create(
			context.Background(),
			migrate.WithDropIndex(true),
		); err != nil {
			return nil, err
		}
	}
	return db, nil
}

func NewTemplate1BlogService(
	c *conf.Data,
	logger log.Logger,
	tp *tracesdk.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) (template1V1.BlogServiceClient, error) {
	var middlewares = []middleware.Middleware{recovery.Recovery()}
	middlewares = append(middlewares, metrics.Client(metrics.WithSeconds(_metricSeconds), metrics.WithRequests(_metricRequests)))
	if tp != nil {
		middlewares = append(middlewares, tracing.Client(tracing.WithTracerProvider(tp)))
	}
	middlewares = append(middlewares, logging.Client(logger))
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(c.ServiceTemplate1.Addr),
		grpc.WithTimeout(c.ServiceTemplate1.Timeout.AsDuration()),
		grpc.WithMiddleware(middlewares...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial template1 gRPC service: %w", err)
	}
	return template1V1.NewBlogServiceClient(conn), nil
}

func NewTemplate1ConnectBlogService(
	c *conf.Data,
	logger log.Logger,
	tp *tracesdk.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) (template1V1connect.BlogServiceClient, error) {
	var middlewares = []middleware.Middleware{recovery.Recovery()}
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

var ProviderSet = wire.NewSet(
	NewData,
	NewEntClient,
	wire.Bind(new(biz.Transaction), new(*Data)),
	NewTemplate1BlogService,
	NewTemplate1ConnectBlogService,
	NewReadingRP,
)
