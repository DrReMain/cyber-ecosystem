package platform

import (
	"cyber-ecosystem/shared-go/cache"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

var defaultCacheError = &cache.CacheDefaultError{
	CacheMiss:       errorspb.ErrorInfraErrorCacheMiss(""),
	KeyNotFound:     errorspb.ErrorInfraErrorCacheKeyNotFound(""),
	SessionNotFound: errorspb.ErrorInfraErrorCacheSessionNotFound(""),
	QuotaExceeded:   errorspb.ErrorInfraErrorCacheQuotaExceeded(""),
	InvalidArgument: errorspb.ErrorInfraErrorCacheInvalidArgument(""),
	LockNotAcquired: errorspb.ErrorInfraErrorCacheLockNotAcquired(""),
	Unavailable:     errorspb.ErrorInfraErrorCacheUnavailable(""),
}

func NewCacheErrorHandler() (CacheErrorHandler, error) {
	// Fail at wire boot (not on first request) if a sentinel slot is nil — a
	// misconfigured CacheDefaultError would otherwise nil-panic in the
	// error-handling hot path.
	if err := cache.ValidateCacheDefaultError(defaultCacheError); err != nil {
		return nil, err
	}
	return func(err error) error {
		return cache.HandleCacheError(err, defaultCacheError)
	}, nil
}
