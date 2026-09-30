package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/capability/cache"
)

type kv struct{ client *redis.Client }

func newKV(client *redis.Client) cache.KV { return &kv{client: client} }

func (k *kv) Get(ctx context.Context, key string) ([]byte, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	val, err := k.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, cache.ErrCacheMiss
		}
		return nil, mapErr(err)
	}
	return val, nil
}

func (k *kv) GetDel(ctx context.Context, key string) ([]byte, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	val, err := k.client.GetDel(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, cache.ErrCacheMiss
		}
		return nil, mapErr(err)
	}
	return val, nil
}

func (k *kv) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	return mapErr(k.client.Set(ctx, key, val, ttl).Err())
}

func (k *kv) Del(ctx context.Context, key string) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	return mapErr(k.client.Del(ctx, key).Err())
}

func (k *kv) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	// EXPIRE on a missing key reports false, not an error: a TTL slide races
	// only with the key dying (or being revoked), and the next read reports
	// that — the slide can never resurrect or un-revoke.
	return mapErr(k.client.Expire(ctx, key, ttl).Err())
}

func (k *kv) Exists(ctx context.Context, key string) (bool, error) {
	if err := cache.ValidateKey(key); err != nil {
		return false, err
	}
	n, err := k.client.Exists(ctx, key).Result()
	if err != nil {
		return false, mapErr(err)
	}
	return n > 0, nil
}

func (k *kv) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	ttl, err := k.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, mapErr(err)
	}
	return normalizeTTL(ttl)
}

// normalizeTTL maps go-redis TTL sentinels — encoded as raw nanoseconds
// (unscaled): -2ns = key absent, -1ns = no expiry — onto the cache contract.
func normalizeTTL(ttl time.Duration) (time.Duration, error) {
	switch ttl {
	case -2 * time.Nanosecond:
		return 0, cache.ErrCacheMiss
	case -1 * time.Nanosecond:
		return 0, nil
	default:
		return ttl, nil
	}
}

func (k *kv) MGet(ctx context.Context, keys ...string) ([][]byte, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	if err := cache.ValidateKeys(keys...); err != nil {
		return nil, err
	}
	vals, err := k.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([][]byte, len(vals))
	for i, v := range vals {
		if s, ok := v.(string); ok {
			out[i] = []byte(s)
		}
	}
	return out, nil
}
