package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/cache"
)

type KV struct {
	client *redis.Client
}

func NewKV(client *redis.Client) cache.KV {
	return &KV{client: client}
}

func (k *KV) Get(ctx context.Context, key string) ([]byte, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	val, err := k.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, cache.ErrCacheMiss
		}
		return nil, err
	}
	return val, nil
}

func (k *KV) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	return k.client.Set(ctx, key, val, ttl).Err()
}

func (k *KV) Delete(ctx context.Context, key string) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	return k.client.Del(ctx, key).Err()
}

func (k *KV) Exist(ctx context.Context, key string) (bool, error) {
	if err := cache.ValidateKey(key); err != nil {
		return false, err
	}
	count, err := k.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (k *KV) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	ttl, err := k.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// Redis TTL semantics:
	// -2 => key does not exist
	// -1 => key exists but has no expiration
	switch ttl {
	case -2 * time.Second:
		return 0, cache.ErrCacheMiss
	case -1 * time.Second:
		return 0, nil
	default:
		return ttl, nil
	}
}

func (k *KV) MGet(ctx context.Context, keys ...string) ([][]byte, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	if err := cache.ValidateKeys(keys...); err != nil {
		return nil, err
	}
	vals, err := k.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	result := make([][]byte, len(vals))
	for i, v := range vals {
		if v != nil {
			if s, ok := v.(string); ok {
				result[i] = []byte(s)
			}
		}
	}
	return result, nil
}

func (k *KV) MSet(ctx context.Context, pairs map[string][]byte, ttl time.Duration) error {
	if len(pairs) == 0 {
		return nil
	}
	if err := cache.ValidatePairs(pairs); err != nil {
		return err
	}
	pipe := k.client.Pipeline()
	for key, val := range pairs {
		if ttl > 0 {
			pipe.Set(ctx, key, val, ttl)
		} else {
			pipe.Set(ctx, key, val, 0)
		}
	}
	_, err := pipe.Exec(ctx)
	return err
}
