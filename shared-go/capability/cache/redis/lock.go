package redis

import (
	"context"
	"errors"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/capability/cache"
)

// lock is backed by bsm/redislock.
type lock struct{ locker *redislock.Client }

func newLock(client *redis.Client) cache.Lock {
	return &lock{locker: redislock.New(client)}
}

func (l *lock) Lock(ctx context.Context, key string, ttl time.Duration, opts ...cache.LockOption) (cache.Release, error) {
	tries, delay := resolveLockRetry(applyLockOpts(opts))
	return l.acquire(ctx, key, ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(delay), tries),
	})
}

// resolveLockRetry applies the LockOptions defaults: effectively unlimited
// attempts (bounded by ctx/ttl) and a 100ms backoff.
func resolveLockRetry(o *cache.LockOptions) (tries int, delay time.Duration) {
	if o.Tries <= 0 {
		tries = 1 << 30
	} else {
		tries = o.Tries
	}
	if o.RetryDelay <= 0 {
		delay = 100 * time.Millisecond
	} else {
		delay = o.RetryDelay
	}
	return tries, delay
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
		return nil, mapErr(err)
	}
	return &release{lk: lk}, nil
}

type release struct{ lk *redislock.Lock }

func (r *release) Unlock(ctx context.Context) error {
	if err := r.lk.Release(ctx); err != nil {
		if errors.Is(err, redislock.ErrLockNotHeld) {
			return cache.ErrLockNotAcquired
		}
		return mapErr(err)
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
		return mapErr(err)
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
