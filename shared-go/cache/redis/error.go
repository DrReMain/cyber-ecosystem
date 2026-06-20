package redis

import (
	"context"
	"errors"
	"fmt"

	"cyber-ecosystem/shared-go/cache"
)

// mapErr classifies a raw go-redis / network error into the backend-agnostic
// cache sentinel contract. It is the single error boundary for the redis
// adapter: every impl routes its final error through it so callers can rely on
// errors.Is(err, cache.ErrXxx) regardless of backend.
//
// Classification rules (order matters):
//  1. nil → nil.
//  2. An error already carrying a cache sentinel (e.g. ErrCacheMiss mapped from
//     redis.Nil in the impls, or ErrInvalidArgument from Validate*) is returned
//     unchanged — the impl already chose the contract-correct sentinel.
//  3. A context error (Canceled / DeadlineExceeded) is returned unchanged. The
//     caller owns the ctx lifecycle and must distinguish "I cancelled" from
//     "redis is down"; wrapping it as ErrUnavailable would hide that.
//  4. Everything else — redis WRONGTYPE, dial/conn-refused, read/write timeout,
//     pool exhaustion, MOVED/ASK, arbitrary redis command errors — is wrapped as
//     ErrUnavailable. These are all "the cache operation could not be served"
//     from the caller's perspective, and leaving them raw defeats the sentinel
//     contract (callers cannot errors.Is a *net.OpError or proto.RedisError to
//     anything meaningful, and HandleCacheError silently drops them).
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	// Already classified by an impl (redis.Nil → ErrCacheMiss, Validate* →
	// ErrInvalidArgument, lock → ErrLockNotAcquired, etc.). Check the sentinel
	// set the contract actually exposes.
	switch {
	case errors.Is(err, cache.ErrCacheMiss),
		errors.Is(err, cache.ErrKeyNotFound),
		errors.Is(err, cache.ErrSessionNotFound),
		errors.Is(err, cache.ErrQuotaExceeded),
		errors.Is(err, cache.ErrInvalidArgument),
		errors.Is(err, cache.ErrLockNotAcquired):
		return err
	}
	// Context errors belong to the caller — pass through so callers can react to
	// their own cancellation/deadline rather than seeing ErrUnavailable.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	// Network / redis-runtime failure: wrap so callers see ErrUnavailable and
	// HandleCacheError's default branch gets a classifiable cause.
	return fmt.Errorf("%w: %w", cache.ErrUnavailable, err)
}
