package platform

import (
	errorspb "cyber-ecosystem/contracts/go/errors"
	"cyber-ecosystem/shared-go/cache"
)

func NewCacheErrorHandler() CacheErrorHandler {
	return func(err error) error {
		return cache.HandleCacheError(err, &cache.CacheDefaultError{
			CacheMiss:       errorspb.ErrorInfraErrorCacheMiss(""),
			KeyNotFound:     errorspb.ErrorInfraErrorCacheKeyNotFound(""),
			SessionNotFound: errorspb.ErrorInfraErrorCacheSessionNotFound(""),
			QuotaExceeded:   errorspb.ErrorInfraErrorCacheQuotaExceeded(""),
			InvalidArgument: errorspb.ErrorInfraErrorCacheInvalidArgument(""),
			Unavailable:     errorspb.ErrorInfraErrorCacheUnavailable(""),
		})
	}
}
