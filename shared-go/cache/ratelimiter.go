package cache

import (
	"context"
	"time"
)

// RateResult is the outcome of a RateLimiter.Allow check.
type RateResult struct {
	Allowed    bool
	Remaining  int64
	RetryAfter time.Duration // 0 when allowed
}

// RateLimiter enforces a per-key quota over a window (GCRA under redis). It is
// distinct from the global transport-level BBR limiter: this gates per business
// key (per user, per IP, per resource).
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (*RateResult, error)
}
