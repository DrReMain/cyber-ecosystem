package data

import (
	"context"
	"time"

	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/biz"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent/migrate"
	_ "github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent/runtime"

	entlog "github.com/DrReMain/cyber-ecosystem/shared-go/kratos/logging/ent"
	"github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/client"
	"github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/entutil"

	"github.com/go-kratos/kratos/v2/log"

	"entgo.io/ent/dialect"
	"github.com/google/wire"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

type Data struct {
	db *ent.Client
}

func NewData(
	c *conf.Data,
	logger log.Logger,
	db *ent.Client,
) (*Data, func(), error) {
	return &Data{db: db}, func() { db.Close() }, nil
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
		log.Fatalf("failed opening connection to database: %v", err)
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

var ProviderSet = wire.NewSet(
	NewData,
	NewEntClient,
	wire.Bind(new(biz.Transaction), new(*Data)),
	NewBlogRP,
)
