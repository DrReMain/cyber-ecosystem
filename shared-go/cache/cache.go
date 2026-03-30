package cache

import (
	"context"
	"io"
	"time"
)

// KV defines the interface for a generic key-value cache.
// Implementations can use Redis, memory, or any other storage backend.
type KV interface {
	// Get retrieves the cached value for the given key.
	// Returns ErrCacheMiss if the key does not exist or has expired.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in the cache with the given TTL.
	// If ttl is 0, the value has no expiration.
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error

	// Delete removes the cached value for the given key.
	Delete(ctx context.Context, key string) error

	// Exist checks if the key exists in the cache (including non-expired entries).
	Exist(ctx context.Context, key string) (bool, error)

	// GetTTL returns the remaining TTL for the key.
	// Returns 0 if the key has no expiration.
	// Returns ErrCacheMiss if the key does not exist or has expired.
	GetTTL(ctx context.Context, key string) (time.Duration, error)

	// MGet retrieves multiple cached values at once.
	// Missing or expired keys will return nil bytes for that key.
	MGet(ctx context.Context, keys ...string) ([][]byte, error)

	// MSet sets multiple cache values at once with the same TTL.
	MSet(ctx context.Context, pairs map[string][]byte, ttl time.Duration) error
}

// Counter defines the interface for an atomic counter.
// Implementations should handle concurrent access safely.
type Counter interface {
	// Incr increments the counter by delta and returns the new value.
	// Creates the key with value delta if it does not exist.
	Incr(ctx context.Context, key string, delta int64) (int64, error)

	// Decr decrements the counter by delta and returns the new value.
	// Creates the key with value -delta if it does not exist.
	Decr(ctx context.Context, key string, delta int64) (int64, error)

	// Get retrieves the current counter value.
	// Returns ErrKeyNotFound if the key does not exist or has expired.
	Get(ctx context.Context, key string) (int64, error)

	// Set sets the counter to a specific value.
	Set(ctx context.Context, key string, value int64) error

	// Expire sets an expiration time on the counter.
	// If ttl is 0, removes the expiration.
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

// Session defines the interface for session storage.
// It organizes data using sessionID + key pattern internally.
type Session interface {
	// Get retrieves the session value for the given sessionID and key.
	// Returns ErrSessionNotFound if the key does not exist or has expired.
	Get(ctx context.Context, sessionID, key string) ([]byte, error)

	// Set stores a session value with the given TTL.
	// If ttl is 0, the session has no expiration.
	Set(ctx context.Context, sessionID, key string, val []byte, ttl time.Duration) error

	// Delete removes the session value for the given sessionID and key.
	Delete(ctx context.Context, sessionID, key string) error

	// Exists checks if the session exists (including non-expired).
	Exists(ctx context.Context, sessionID string) (bool, error)

	// Refresh extends the TTL of an existing session.
	// Returns ErrSessionNotFound if the session does not exist or has expired.
	Refresh(ctx context.Context, sessionID string, ttl time.Duration) error

	// Destroy destroys the entire session.
	Destroy(ctx context.Context, sessionID string) error

	// Keys returns all user-level keys for the given sessionID.
	// The returned keys are the user-provided key names, not the internal storage keys.
	Keys(ctx context.Context, sessionID string) ([]string, error)
}

// Member represents a member in a sorted set with its score.
type Member struct {
	Score  float64
	Member string
}

// SortedSet defines the interface for a sorted set (zset) data structure.
// Members are ordered by their scores.
type SortedSet interface {
	// Add adds one or more members to a sorted set.
	// Updates the score if the member already exists.
	Add(ctx context.Context, key string, members ...Member) error

	// Incr increments the score of a member by delta.
	// Creates the sorted set if it does not exist.
	Incr(ctx context.Context, key string, member string, delta float64) (float64, error)

	// Rank returns the rank of a member in ascending order (0-indexed).
	// Returns ErrKeyNotFound if the key or member does not exist.
	Rank(ctx context.Context, key string, member string) (int64, error)

	// RevRank returns the rank of a member in descending order (0-indexed).
	// Returns ErrKeyNotFound if the key or member does not exist.
	RevRank(ctx context.Context, key string, member string) (int64, error)

	// Score returns the score of a member.
	// Returns ErrKeyNotFound if the key or member does not exist.
	Score(ctx context.Context, key string, member string) (float64, error)

	// Range returns members with scores between start and stop (inclusive).
	// Supports negative indices (e.g., -1 for last element).
	Range(ctx context.Context, key string, start, stop int64) ([]Member, error)

	// RevRange returns members with scores between start and stop in descending order.
	// Supports negative indices.
	RevRange(ctx context.Context, key string, start, stop int64) ([]Member, error)

	// Remove removes one or more members from a sorted set.
	Remove(ctx context.Context, key string, members ...string) error
}

// RateLimiter defines the interface for a rate limiter.
// Implements a sliding window algorithm to control request rates.
type RateLimiter interface {
	// Allow checks if a request should be allowed.
	// Returns true if quota is available, false if rate limit exceeded.
	// Returns ErrQuotaExceeded when the request is denied.
	Allow(ctx context.Context, key string, quota int64, window time.Duration) (bool, error)
}

// Cache is a container that holds all cache-related interfaces and the underlying client.
// The Client field holds the underlying client for closing (e.g., *redis.Client).
// For memory cache, Client is nil since no cleanup is needed.
type Cache struct {
	// Client is the underlying client. It implements io.Closer.
	// For Redis, this is *redis.Client.
	// For Memory, this is nil.
	Client      io.Closer
	KV          KV
	Counter     Counter
	Session     Session
	SortedSet   SortedSet
	RateLimiter RateLimiter
}
