package cache

import (
	"context"
	"time"
)

// KV is the generic string cache. Missing reads return ErrCacheMiss; TTL of 0
// means no expiration.
type KV interface {
	Get(ctx context.Context, key string) ([]byte, error)
	GetDel(ctx context.Context, key string) ([]byte, error) // atomic get + delete; miss → ErrCacheMiss
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
	GetTTL(ctx context.Context, key string) (time.Duration, error) // miss → ErrCacheMiss; no expiry → 0
	MGet(ctx context.Context, keys ...string) ([][]byte, error)    // missing entries are nil
}
