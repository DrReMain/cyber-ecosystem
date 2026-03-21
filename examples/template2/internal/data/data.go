package data

import (
	"context"
	"os"

	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/biz"
	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent"
	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent/migrate"
	_ "github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent/runtime"

	template1V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"
	"github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/client"
	"github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/entutil"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type Data struct {
	db *ent.Client

	template1BlogService template1V1.BlogServiceClient
}

func NewData(
	c *conf.Data,
	template1BlogService template1V1.BlogServiceClient,
) (*Data, func(), error) {
	drv, err := client.NewEntClient(client.DBConfig{
		Driver:          c.Database.Driver,
		Host:            c.Database.Host,
		Port:            int(c.Database.Port),
		User:            c.Database.User,
		Password:        c.Database.Password,
		DBName:          c.Database.DbName,
		MaxOpenConns:    int(c.Database.MaxOpenConns),
		MaxIdleConns:    int(c.Database.MaxIdleConns),
		ConnMaxLifetime: c.Database.ConnMaxLifetime.AsDuration(),
	})
	if err != nil {
		log.Fatalf("failed opening connection to database: %v", err)
	}

	db := ent.NewClient(ent.Driver(drv))
	if os.Getenv("DEPLOY_ENV") == "dev" {
		db = db.Debug()
		err := db.Schema.Create(
			context.Background(),
			migrate.WithDropIndex(true),
		)
		if err != nil {
			return nil, nil, err
		}
	}

	return &Data{
		db:                   db,
		template1BlogService: template1BlogService,
	}, func() { db.Close() }, nil
}

func NewTemplate1BlogService(
	c *conf.Data,
	logger log.Logger,
	tp *tracesdk.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) template1V1.BlogServiceClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(c.ServiceTemplate1.Addr),
		grpc.WithTimeout(c.ServiceTemplate1.Timeout.AsDuration()),
		grpc.WithHealthCheck(true),
		grpc.WithMiddleware(
			recovery.Recovery(),
			metrics.Client(
				metrics.WithSeconds(_metricSeconds),
				metrics.WithRequests(_metricRequests),
			),
			tracing.Client(tracing.WithTracerProvider(tp)),
			logging.Client(logger),
		),
	)
	if err != nil {
		panic(err)
	}
	return template1V1.NewBlogServiceClient(conn)
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

var ProviderSet = wire.NewSet(
	NewData,
	wire.Bind(new(biz.Transaction), new(*Data)),
	NewReadingRP,
	NewTemplate1BlogService,
)
