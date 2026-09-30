package cache

import (
	"context"
	"time"
)

// Counter is a standalone atomic integer. Get on a missing key returns
// ErrKeyNotFound; Expire with ttl 0 removes any expiration.
type Counter interface {
	Incr(ctx context.Context, key string, delta int64) (int64, error)
	Decr(ctx context.Context, key string, delta int64) (int64, error)
	Get(ctx context.Context, key string) (int64, error)
	Set(ctx context.Context, key string, value int64) error
	Expire(ctx context.Context, key string, ttl time.Duration) error
}
