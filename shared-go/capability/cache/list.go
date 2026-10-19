package cache

import "context"

// List is an ordered sequence (recent-N, lightweight queue, inbox). Pops on an
// empty list return ErrCacheMiss.
type List interface {
	LPush(ctx context.Context, key string, vals ...[]byte) (int64, error)
	RPush(ctx context.Context, key string, vals ...[]byte) (int64, error)
	LPop(ctx context.Context, key string) ([]byte, error)
	RPop(ctx context.Context, key string) ([]byte, error)
	LRange(ctx context.Context, key string, start, stop int64) ([][]byte, error)
	LLen(ctx context.Context, key string) (int64, error)
	LTrim(ctx context.Context, key string, start, stop int64) error
}
