package platform

import (
	"cyber-ecosystem/shared-go/capability/cache"

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
	if err := cache.ValidateCacheDefaultError(defaultCacheError); err != nil {
		return nil, err
	}
	return func(err error) error {
		return cache.HandleCacheError(err, defaultCacheError)
	}, nil
}
