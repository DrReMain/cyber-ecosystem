package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

func TestRateLimiter(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	rl := c.RateLimiter
	resetKeys(t, c, "rate1")

	const limit = 3
	// First `limit` calls are allowed
	for i := 1; i <= limit; i++ {
		res, err := rl.Allow(ctx, "rate1", limit, time.Second)
		if err != nil {
			t.Fatalf("Allow %d: %v", i, err)
		}
		if !res.Allowed {
			t.Fatalf("Allow %d: want allowed, denied (remaining=%d retryAfter=%v)", i, res.Remaining, res.RetryAfter)
		}
	}

	// Next call exceeds the quota → denied, with a positive retry-after
	res, err := rl.Allow(ctx, "rate1", limit, time.Second)
	if err != nil {
		t.Fatalf("Allow over limit: %v", err)
	}
	if res.Allowed {
		t.Fatal("Allow over limit: want denied, got allowed")
	}
	if res.RetryAfter <= 0 {
		t.Fatalf("RetryAfter after deny: %v, want >0", res.RetryAfter)
	}
}

// TestRateLimiterGuards pins the defensive branches: limit<=0 denies (safe
// default for a misconfigured limiter) without error; window<=0 is an explicit
// ErrInvalidArgument.
func TestRateLimiterGuards(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	rl := c.RateLimiter
	resetKeys(t, c, "rateguard")

	res, err := rl.Allow(ctx, "rateguard", 0, time.Second)
	if err != nil || res.Allowed {
		t.Fatalf("limit=0: want denied/no-error, got allowed=%v err=%v", res.Allowed, err)
	}
	if _, err := rl.Allow(ctx, "rateguard", 5, 0); !errors.Is(err, cache.ErrInvalidArgument) {
		t.Fatalf("window=0: want ErrInvalidArgument, got %v", err)
	}
}
