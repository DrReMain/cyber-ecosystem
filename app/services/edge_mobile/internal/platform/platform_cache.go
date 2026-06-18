package platform

import (
	"fmt"

	"cyber-ecosystem/shared-go/cache"
	cacheredis "cyber-ecosystem/shared-go/cache/redis"

	"cyber-ecosystem/app/services/edge_mobile/internal/conf"
)

// NewCache builds the redis-backed cache container for edge_mobile and returns
// it with a cleanup that closes the redis client. The cleanup is part of the
// provider signature so google/wire registers it: wire invokes it on graceful
// shutdown AND automatically if a later provider fails during injection (e.g. DB
// unreachable at boot), so the redis connection pool is never orphaned.
func NewCache(c *conf.Data) (*cache.Cache, func(), error) {
	rc := c.GetRedis()
	if rc == nil {
		return nil, nil, fmt.Errorf("redis config is required")
	}
	client, closeFn, err := cacheredis.NewClient(toRedisConfig(rc))
	if err != nil {
		return nil, nil, err
	}
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
