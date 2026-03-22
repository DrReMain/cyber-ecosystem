package client

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
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
	MeterProvider   metric.MeterProvider
}

// EntClient wraps entsql.Driver and sql.DB for metrics registration
type EntClient struct {
	Driver *entsql.Driver
	DB     *sql.DB
}

func NewEntClient(cfg DBConfig) (*EntClient, error) {
	var (
		drvName string
		dsn     string
	)
	switch cfg.Driver {
	case dialect.MySQL:
		drvName = "mysql"
		// interpolateParams=true 提升批量插入性能
		// parseTime=true 处理 time.Time 转换
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=True&loc=Local&interpolateParams=true", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	case dialect.Postgres:
		drvName = "pgx"
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	default:
		return nil, fmt.Errorf("unsupported database driver %s", cfg.Driver)
	}

	opts := []otelsql.Option{
		otelsql.WithAttributes(
			semconv.DBSystemKey.String(cfg.Driver),
			attribute.String("db.name", cfg.DBName),
		),
		otelsql.WithSQLCommenter(true),
	}

	// Enable metrics if MeterProvider is provided
	if cfg.MeterProvider != nil {
		opts = append(opts, otelsql.WithMeterProvider(cfg.MeterProvider))
	}

	// Use otelsql to wrap the database driver for OpenTelemetry tracing
	// This automatically creates spans for each SQL operation
	db, err := otelsql.Open(drvName, dsn, opts...)
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

	// Register DB stats metrics for connection pool monitoring
	if cfg.MeterProvider != nil {
		if _, err := otelsql.RegisterDBStatsMetrics(db, otelsql.WithMeterProvider(cfg.MeterProvider)); err != nil {
			return nil, fmt.Errorf("failed registering db stats metrics: %w", err)
		}
	}

	return &EntClient{
		Driver: entsql.OpenDB(cfg.Driver, db),
		DB:     db,
	}, nil
}
