package platform

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/orm/ent/client"
	"cyber-ecosystem/shared-go/orm/ent/logging"

	"cyber-ecosystem/apps/genesis/services/base/internal/conf"
	"cyber-ecosystem/apps/genesis/services/base/internal/ent"
	"cyber-ecosystem/apps/genesis/services/base/internal/ent/migrate"
	_ "cyber-ecosystem/apps/genesis/services/base/internal/ent/runtime"
)

func NewEntClient(
	c *conf.Data,
	cl *conf.Log,
	logger log.Logger,
	meterProvider *metricsdk.MeterProvider,
) (*ent.Client, error) {
	ec, err := client.NewEntClient(client.DBConfig{
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

	var drv dialect.Driver = ec.Driver
	if cl != nil && cl.Ent != nil && cl.Ent.Enabled {
		slowQueryThreshold := 200 * time.Millisecond
		if cl.Ent.SlowQueryThreshold != nil {
			slowQueryThreshold = cl.Ent.SlowQueryThreshold.AsDuration()
		}
		drv = logging.NewKratosDriverWrapper(ec.Driver, logger, cl.Ent.Level, cl.Ent.SlowQuery, slowQueryThreshold)
	}
	db := ent.NewClient(ent.Driver(drv))

	if c.Database.Migrate {
		if err := db.Schema.Create(
			context.Background(),
			migrate.WithForeignKeys(true),
			migrate.WithDropIndex(true),
		); err != nil {
			return nil, fmt.Errorf("failed running database schema migration: %w", err)
		}
	}

	return db, nil
}
