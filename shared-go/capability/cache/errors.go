package cache

import "errors"

// Sentinel errors are part of the cache contract (backend-agnostic). The first
// six map 1:1 to the InfraError 32xx Cache codes (infra.proto); the service
// layer adapts them to application errors. ErrLockNotAcquired maps to
// INFRA_ERROR_CACHE_LOCK_NOT_ACQUIRED (3206).
var (
	ErrCacheMiss       = errors.New("cache miss")
	ErrKeyNotFound     = errors.New("cache key not found")
	ErrSessionNotFound = errors.New("cache session not found")
	ErrQuotaExceeded   = errors.New("cache quota exceeded")
	ErrInvalidArgument = errors.New("cache invalid argument")
	ErrUnavailable     = errors.New("cache unavailable")
	ErrLockNotAcquired = errors.New("cache lock not acquired")
)
