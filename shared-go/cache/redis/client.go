package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/cache/redis/otel"
)

// Config holds Redis client configuration.
type Config struct {
	Network            string
	Addr               string
	Password           string
	DB                 int
	PoolSize           int
	MinIdleConns       int
	ConnMaxLifetime    time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	EnableTracing      bool
	EnableLogging      bool
	Logger             log.Logger
	LogLevel           string
	SlowQuery          bool
	SlowQueryThreshold time.Duration
}

// NewRedisClient creates a new Redis client with optional configurations.
func NewRedisClient(cfg *Config) (*redis.Client, error) {
	redisOpts := &redis.Options{
		Network:         cfg.Network,
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
	}

	client := redis.NewClient(redisOpts)

	// Add hooks in order: tracing first, then logging
	if cfg.EnableTracing {
		client.AddHook(otel.NewTracingHook())
	}

	if cfg.EnableLogging && cfg.Logger != nil {
		client.AddHook(NewKratosClientWrapper(client, cfg.Logger, cfg.LogLevel, cfg.SlowQuery, cfg.SlowQueryThreshold))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return client, nil
}
