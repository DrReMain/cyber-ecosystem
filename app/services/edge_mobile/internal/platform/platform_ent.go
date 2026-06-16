package platform

import (
	"fmt"

	"cyber-ecosystem/shared-go/orm/ent/client"

	"cyber-ecosystem/app/services/edge_mobile/internal/conf"
	"cyber-ecosystem/app/services/edge_mobile/internal/ent"
	_ "cyber-ecosystem/app/services/edge_mobile/internal/ent/runtime"
)

// NewEntClient opens the ent client. Schema management is handled externally by
// Atlas versioned migrations (Nx targets migrate:diff / migrate:apply); the app
// performs no DDL on startup.
func NewEntClient(c *conf.Data) (*ent.Client, error) {
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
		return nil, fmt.Errorf("failed opening connection to database: %w", err)
	}
	return ent.NewClient(ent.Driver(ec.Driver)), nil
}
