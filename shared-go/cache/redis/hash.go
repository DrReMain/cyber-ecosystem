package redis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/cache"
)

type hash struct{ client *redis.Client }

// NewHash returns the redis-backed Hash implementation.
func NewHash(client *redis.Client) cache.Hash { return &hash{client: client} }

func (h *hash) HSet(ctx context.Context, key, field string, val []byte) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	return mapErr(h.client.HSet(ctx, key, field, val).Err())
}

func (h *hash) HMSet(ctx context.Context, key string, fields map[string][]byte) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	if len(fields) == 0 {
		return nil
	}
	args := make([]any, 0, len(fields)*2)
	for f, v := range fields {
		args = append(args, f, v)
	}
	return mapErr(h.client.HSet(ctx, key, args...).Err())
}

func (h *hash) HGet(ctx context.Context, key, field string) ([]byte, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	val, err := h.client.HGet(ctx, key, field).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, cache.ErrCacheMiss
		}
		return nil, mapErr(err)
	}
	return val, nil
}

func (h *hash) HGetAll(ctx context.Context, key string) (map[string][]byte, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	m, err := h.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, mapErr(err)
	}
	out := make(map[string][]byte, len(m))
	for k, v := range m {
		out[k] = []byte(v)
	}
	return out, nil
}

func (h *hash) HDel(ctx context.Context, key string, fields ...string) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}
	if len(fields) == 0 {
		return nil
	}
	return mapErr(h.client.HDel(ctx, key, fields...).Err())
}

func (h *hash) HExists(ctx context.Context, key, field string) (bool, error) {
	if err := cache.ValidateKey(key); err != nil {
		return false, err
	}
	ok, err := h.client.HExists(ctx, key, field).Result()
	return ok, mapErr(err)
}

func (h *hash) HIncrBy(ctx context.Context, key, field string, delta int64) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	v, err := h.client.HIncrBy(ctx, key, field, delta).Result()
	return v, mapErr(err)
}

func (h *hash) HLen(ctx context.Context, key string) (int64, error) {
	if err := cache.ValidateKey(key); err != nil {
		return 0, err
	}
	v, err := h.client.HLen(ctx, key).Result()
	return v, mapErr(err)
}
