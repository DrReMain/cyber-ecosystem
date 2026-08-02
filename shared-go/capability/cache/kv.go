package cache

import (
	"context"
	"time"
)

// KV is the generic string cache. Missing reads return ErrCacheMiss; TTL of 0
// means no expiration.
type KV interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Exist(ctx context.Context, key string) (bool, error)
	GetTTL(ctx context.Context, key string) (time.Duration, error) // miss → ErrCacheMiss; no expiry → 0
	MGet(ctx context.Context, keys ...string) ([][]byte, error)    // missing entries are nil
	MSet(ctx context.Context, pairs map[string][]byte, ttl time.Duration) error
}
