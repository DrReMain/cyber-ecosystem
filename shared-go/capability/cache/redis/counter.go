package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/capability/cache"
)

type counter struct{ client *redis.Client }

// NewCounter returns the redis-backed Counter implementation.
func NewCounter(client *redis.Client) cache.Counter { return &counter{client: client} }

func (c *counter) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	v, err := c.client.IncrBy(ctx, key, delta).Result()
	return v, mapErr(err)
}

func (c *counter) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	v, err := c.client.DecrBy(ctx, key, delta).Result()
	return v, mapErr(err)
}

func (c *counter) Get(ctx context.Context, key string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	v, err := c.client.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cache.ErrKeyNotFound
		}
		return 0, mapErr(err)
	}
	return v, nil
}

func (c *counter) Set(ctx context.Context, key string, value int64) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	return mapErr(c.client.Set(ctx, key, value, 0).Err())
}

func (c *counter) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	if ttl > 0 {
		return mapErr(c.client.Expire(ctx, key, ttl).Err())
	}
	return mapErr(c.client.Persist(ctx, key).Err())
}
