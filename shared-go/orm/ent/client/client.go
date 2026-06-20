package client

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBConfig struct {
	Driver          string
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	// SQLOpener opens the *sql.DB for (driverName, dsn). Defaults to sql.Open.
	// The platform layer sets this to observability.OpenSQL to wrap the driver
	// with OTel tracing + pool metrics. This package must NOT import the
	// observability package — the dependency is one-way (platform → orm).
	SQLOpener func(driverName, dsn string) (*sql.DB, error)
}

type EntClient struct {
	Driver *entsql.Driver
	DB     *sql.DB
}

func NewEntClient(cfg DBConfig) (*EntClient, error) {
	var drvName, dsn string
	switch cfg.Driver {
	case dialect.Postgres:
		drvName = "pgx"
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	default:
		return nil, fmt.Errorf("unsupported database driver %s", cfg.Driver)
	}

	opener := cfg.SQLOpener
	if opener == nil {
		opener = sql.Open
	}
	db, err := opener(drvName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed opening connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed pinging database: %w", err)
	}

	return &EntClient{
		Driver: entsql.OpenDB(cfg.Driver, db),
		DB:     db,
	}, nil
}
