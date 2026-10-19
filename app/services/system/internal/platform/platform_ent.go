package platform

import (
	"fmt"
	"log/slog"

	"cyber-ecosystem/shared-go/kratos/observability"
	"cyber-ecosystem/shared-go/orm/ent/client"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/ent"
	_ "cyber-ecosystem/app/services/system/internal/ent/runtime"
)

func NewEntClient(c *conf.Data, logger *slog.Logger) (*ent.Client, func(), error) {
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
		// Wrap the SQL driver with otelsql (tracing + pool metrics; global providers).
		SQLOpener: observability.OpenSQL,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed opening connection to database: %w", err)
	}
	cl := ent.NewClient(ent.Driver(observability.WrapSlowQueryDriver(ec.Driver, logger)))
	return cl, func() { _ = cl.Close() }, nil
}
