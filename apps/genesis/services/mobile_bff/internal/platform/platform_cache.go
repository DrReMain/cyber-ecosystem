package platform

import (
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/cache"
	cachememory "cyber-ecosystem/shared-go/cache/memory"
	cacheredis "cyber-ecosystem/shared-go/cache/redis"

	"cyber-ecosystem/apps/genesis/services/mobile_bff/internal/conf"
)

func NewCache(c *conf.Data, cl *conf.Log, logger log.Logger) (*cache.Cache, error) {
	cfg := c.GetCache()
	if cfg == nil {
		return nil, fmt.Errorf("cache config is required")
	}

	switch cfg.GetType() {
	case "memory":
		return newMemoryCache(cfg.GetMemory(), cl, logger), nil
	case "redis":
		return newRedisCache(cfg.GetRedis(), cl, logger)
	default:
		return nil, fmt.Errorf("unsupported cache type: %q", cfg.GetType())
	}
}

func newRedisCache(cfg *conf.Data_Cache_Redis, cl *conf.Log, logger log.Logger) (*cache.Cache, error) {
	var enableLogging bool
	var logLevel string
	var slowQuery bool
	var slowQueryThreshold time.Duration

	if cl != nil && cl.Cache != nil && cl.Cache.Enabled {
		enableLogging = true
		logLevel = cl.Cache.Level
		slowQuery = cl.Cache.SlowQuery
		slowQueryThreshold = cl.Cache.SlowQueryThreshold.AsDuration()
	}

	c, err := cacheredis.NewRedisClient(&cacheredis.Config{
		Network:            cfg.GetNetwork(),
		Addr:               cfg.GetAddr(),
		Password:           cfg.GetPassword(),
		DB:                 int(cfg.GetDb()),
		PoolSize:           int(cfg.GetPoolSize()),
		MinIdleConns:       int(cfg.GetMinIdleConns()),
		ConnMaxLifetime:    cfg.GetConnMaxLifetime().AsDuration(),
		ReadTimeout:        cfg.GetReadTimeout().AsDuration(),
		WriteTimeout:       cfg.GetWriteTimeout().AsDuration(),
		EnableTracing:      cfg.GetOtelEnabled(),
		EnableLogging:      enableLogging,
		Logger:             logger,
		LogLevel:           logLevel,
		SlowQuery:          slowQuery,
		SlowQueryThreshold: slowQueryThreshold,
	})
	if err != nil {
		return nil, err
	}

	return &cache.Cache{
		Client:      c,
		KV:          cacheredis.NewKV(c),
		Counter:     cacheredis.NewCounter(c),
		Session:     cacheredis.NewSession(c),
		SortedSet:   cacheredis.NewSortedSet(c),
		RateLimiter: cacheredis.NewRateLimiter(c),
	}, nil
}

func newMemoryCache(cfg *conf.Data_Cache_Memory, cl *conf.Log, logger log.Logger) *cache.Cache {
	var enableLogging bool
	var logLevel string
	var slowQuery bool
	var slowQueryThreshold time.Duration

	if cl != nil && cl.Cache != nil && cl.Cache.Enabled {
		enableLogging = true
		logLevel = cl.Cache.Level
		slowQuery = cl.Cache.SlowQuery
		slowQueryThreshold = cl.Cache.SlowQueryThreshold.AsDuration()
	}

	return cachememory.NewMemoryClient(&cachememory.Config{
		EnableTracing:      cfg.GetOtelEnabled(),
		EnableLogging:      enableLogging,
		Logger:             logger,
		LogLevel:           logLevel,
		SlowQuery:          slowQuery,
		SlowQueryThreshold: slowQueryThreshold,
	})
}
