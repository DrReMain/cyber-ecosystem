package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/capability/cache"
)

type rateLimiter struct{ limiter *redis_rate.Limiter }

// NewRateLimiter returns the redis-backed RateLimiter (go-redis/redis_rate, GCRA).
func NewRateLimiter(client *redis.Client) cache.RateLimiter {
	return &rateLimiter{limiter: redis_rate.NewLimiter(client)}
}

func (r *rateLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (*cache.RateResult, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	if window <= 0 {
		return nil, cache.ErrInvalidArgument
	}
	if limit <= 0 {
		// A misconfigured (zero/negative) limit denies everything — safer than
		// feeding degenerate args into the GCRA lua.
		return &cache.RateResult{Allowed: false, Remaining: 0}, nil
	}
	// Burst = limit so a window may fill up to `limit` at once, then refills at
	// limit/window — matching "at most `limit` per `window`" semantics.
	res, err := r.limiter.Allow(ctx, key, redis_rate.Limit{
		Rate:   int(limit),
		Burst:  int(limit),
		Period: window,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &cache.RateResult{
		Allowed:    res.Allowed > 0,
		Remaining:  int64(res.Remaining),
		RetryAfter: res.RetryAfter,
	}, nil
}
