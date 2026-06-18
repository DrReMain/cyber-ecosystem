package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/cache"
)

// NewClient builds a redis client, pings it to fail fast on misconfiguration,
// and returns the client plus a cleanup that closes it.
//
// Observability (command tracing/metrics via redisotel, and slow-query logging)
// is attached later through AddHook on this same client by the observability
// layer — intentionally not done here, so this constructor concerns itself only
// with connectivity and the cache interfaces never depend on it.
//
// DialTimeout/PoolTimeout of 0 fall back to go-redis defaults (5s and
// ReadTimeout+1s respectively).
func NewClient(cfg *Config) (*redis.Client, func(), error) {
	client := redis.NewClient(&redis.Options{
		Network:         cfg.Network,
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		DialTimeout:     cfg.DialTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, nil, fmt.Errorf("%w: redis ping failed: %w", cache.ErrUnavailable, err)
	}

	return client, func() { _ = client.Close() }, nil
}
