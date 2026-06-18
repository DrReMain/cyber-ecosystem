package platform

import (
	"fmt"

	"cyber-ecosystem/shared-go/orm/ent/client"

	"cyber-ecosystem/app/services/edge_mobile/internal/conf"
	"cyber-ecosystem/app/services/edge_mobile/internal/ent"
	_ "cyber-ecosystem/app/services/edge_mobile/internal/ent/runtime"
)

// NewEntClient opens the ent client and returns it with a cleanup that closes
// the underlying database pool. Wire registers the cleanup so the pool is
// closed on shutdown and on any later provider failure during injection. Schema
// management is external (Atlas versioned migrations via the migrate:diff /
// migrate:apply Nx targets); the app performs no DDL on startup.
func NewEntClient(c *conf.Data) (*ent.Client, func(), error) {
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
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed opening connection to database: %w", err)
	}
	cl := ent.NewClient(ent.Driver(ec.Driver))
	return cl, func() { _ = cl.Close() }, nil
}
