package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/cache"
)

type Counter struct {
	client *redis.Client
}

func NewCounter(client *redis.Client) cache.Counter {
	return &Counter{client: client}
}

func (c *Counter) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	return c.client.IncrBy(ctx, key, delta).Result()
}

func (c *Counter) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	return c.client.DecrBy(ctx, key, delta).Result()
}

func (c *Counter) Get(ctx context.Context, key string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	val, err := c.client.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cache.ErrKeyNotFound
		}
		return 0, err
	}
	return val, nil
}

func (c *Counter) Set(ctx context.Context, key string, value int64) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	return c.client.Set(ctx, key, value, 0).Err()
}

func (c *Counter) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	if ttl == 0 {
		ok, err := c.client.Persist(ctx, key).Result()
		if err != nil {
			return err
		}
		if !ok {
			return cache.ErrKeyNotFound
		}
		return nil
	}
	ok, err := c.client.Expire(ctx, key, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return cache.ErrKeyNotFound
	}
	return nil
}
