package data

import (
	"context"
	"os"

	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/biz"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent/migrate"
	_ "github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent/runtime"

	"github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/client"
	"github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/entutil"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/google/wire"
)

type Data struct {
	db *ent.Client
}

func NewData(c *conf.Data) (*Data, func(), error) {
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

var ProviderSet = wire.NewSet(
	NewData,
	wire.Bind(new(biz.Transaction), new(*Data)),
	NewBlogRP,
)
