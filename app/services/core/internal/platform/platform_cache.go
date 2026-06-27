package platform

import (
	"fmt"
	"log/slog"

	"cyber-ecosystem/shared-go/cache"
	cacheredis "cyber-ecosystem/shared-go/cache/redis"
	"cyber-ecosystem/shared-go/kratos/observability"

	"cyber-ecosystem/app/services/core/internal/conf"
)

func NewCache(c *conf.Data, logger *slog.Logger) (*cache.Cache, func(), error) {
	rc := c.GetRedis()
	if rc == nil {
		return nil, nil, fmt.Errorf("redis config is required")
	}
	client, closeFn, err := cacheredis.NewClient(toRedisConfig(rc))
	if err != nil {
		return nil, nil, err
	}
	// Attach redisotel tracing + metrics (reads global providers set by
	// observability.Init; no-op when trace/metrics disabled).
	if err := observability.InstrumentRedis(client); err != nil {
		closeFn()
		return nil, nil, fmt.Errorf("redis instrumentation: %w", err)
	}
	observability.AttachSlowRedisHook(client, logger)
	return cacheredis.New(client), closeFn, nil
}

func toRedisConfig(rc *conf.Data_Redis) *cacheredis.Config {
	network := rc.GetNetwork()
	if network == "" {
		network = "tcp"
	}
	return &cacheredis.Config{
		Network:         network,
		Addr:            rc.GetAddr(),
		Password:        rc.GetPassword(),
		DB:              int(rc.GetDb()),
		PoolSize:        int(rc.GetPoolSize()),
		MinIdleConns:    int(rc.GetMinIdleConns()),
		ConnMaxLifetime: rc.GetConnMaxLifetime().AsDuration(),
		DialTimeout:     rc.GetDialTimeout().AsDuration(),
		PoolTimeout:     rc.GetPoolTimeout().AsDuration(),
		ReadTimeout:     rc.GetReadTimeout().AsDuration(),
		WriteTimeout:    rc.GetWriteTimeout().AsDuration(),
	}
}
