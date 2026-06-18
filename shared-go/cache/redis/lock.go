package redis

import (
	"context"
	"errors"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/cache"
)

type lock struct{ locker *redislock.Client }

// NewLock returns the redis-backed Lock implementation (bsm/redislock).
func NewLock(client *redis.Client) cache.Lock {
	return &lock{locker: redislock.New(client)}
}

func (l *lock) Lock(ctx context.Context, key string, ttl time.Duration, opts ...cache.LockOption) (cache.Release, error) {
	o := applyLockOpts(opts)
	tries := o.Tries
	if tries <= 0 {
		tries = 1 << 30 // effectively unlimited; bounded by ctx
	}
	delay := o.RetryDelay
	if delay <= 0 {
		delay = 100 * time.Millisecond
	}
	return l.acquire(ctx, key, ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(delay), tries),
	})
}

func (l *lock) TryLock(ctx context.Context, key string, ttl time.Duration, _ ...cache.LockOption) (cache.Release, error) {
	return l.acquire(ctx, key, ttl, &redislock.Options{}) // no retry → single attempt
}

func (l *lock) acquire(ctx context.Context, key string, ttl time.Duration, opt *redislock.Options) (cache.Release, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}
	lk, err := l.locker.Obtain(ctx, key, ttl, opt)
	if err != nil {
		if errors.Is(err, redislock.ErrNotObtained) {
			if cerr := ctx.Err(); cerr != nil {
				return nil, cerr // ctx done while waiting
			}
			return nil, cache.ErrLockNotAcquired
		}
		return nil, err
	}
	return &release{lk: lk}, nil
}

type release struct{ lk *redislock.Lock }

func (r *release) Unlock(ctx context.Context) error {
	if err := r.lk.Release(ctx); err != nil {
		if errors.Is(err, redislock.ErrLockNotHeld) {
			return cache.ErrLockNotAcquired
		}
		return err
	}
	return nil
}

func (r *release) Extend(ctx context.Context, ttl time.Duration) error {
	// Refresh returns ErrNotObtained (not ErrLockNotHeld) when the lock has
	// expired/gone, so map both to the cache contract. A tiny retry absorbs a
	// transient redis blip while the lock is still validly held.
	if err := r.lk.Refresh(ctx, ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(10*time.Millisecond), 3),
	}); err != nil {
		if errors.Is(err, redislock.ErrLockNotHeld) || errors.Is(err, redislock.ErrNotObtained) {
			return cache.ErrLockNotAcquired
		}
		return err
	}
	return nil
}

func applyLockOpts(opts []cache.LockOption) *cache.LockOptions {
	o := &cache.LockOptions{}
	for _, fn := range opts {
		fn(o)
	}
	return o
}
