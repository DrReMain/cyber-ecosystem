package data

import (
	"context"
	"os"

	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/conf"
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/data/ent"
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/data/ent/migrate"

	"github.com/DrReMain/cyber-ecosystem/go-shared/orm/ent/client"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewData, NewUserRP)

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
