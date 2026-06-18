package cache

import (
	"context"
	"time"
)

// Release is a held distributed lock. Unlock releases it; Extend renews the TTL.
// Operations on an already-released/expired lock return ErrLockNotAcquired.
type Release interface {
	Unlock(ctx context.Context) error
	Extend(ctx context.Context, ttl time.Duration) error
}

// LockOptions tunes the blocking Lock acquire loop. The zero value works.
type LockOptions struct {
	Tries      int           // max attempts; <=0 retries while ctx/ttl allows
	RetryDelay time.Duration // backoff between attempts; <=0 defaults to 100ms
}

// LockOption configures LockOptions.
type LockOption func(*LockOptions)

// WithRetry sets the blocking Lock's retry policy.
func WithRetry(tries int, delay time.Duration) LockOption {
	return func(o *LockOptions) { o.Tries = tries; o.RetryDelay = delay }
}

// Lock is a distributed mutex over a single redis instance.
//
// TryLock makes a single attempt and returns ErrLockNotAcquired if the key is
// held. Lock retries until the lock is acquired, ctx is cancelled, OR — if ctx
// has no Deadline — the TTL elapses (the underlying redislock imposes a
// ttl-based deadline on a deadline-less ctx). Callers that need blocking beyond
// a single TTL MUST pass a ctx with a Deadline.
type Lock interface {
	Lock(ctx context.Context, key string, ttl time.Duration, opts ...LockOption) (Release, error)
	TryLock(ctx context.Context, key string, ttl time.Duration, opts ...LockOption) (Release, error)
}
