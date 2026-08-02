package cache

import (
	"context"
)

// Hash is a key holding multiple fields (object records, multi-dim counters,
// carts). Missing field reads return ErrCacheMiss.
type Hash interface {
	HSet(ctx context.Context, key, field string, val []byte) error
	HMSet(ctx context.Context, key string, fields map[string][]byte) error
	HGet(ctx context.Context, key, field string) ([]byte, error)
	HGetAll(ctx context.Context, key string) (map[string][]byte, error)
	HDel(ctx context.Context, key string, fields ...string) error
	HExists(ctx context.Context, key, field string) (bool, error)
	HIncrBy(ctx context.Context, key, field string, delta int64) (int64, error)
	HLen(ctx context.Context, key string) (int64, error)
}
