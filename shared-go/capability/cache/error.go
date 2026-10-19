package cache

import (
	stderrors "errors"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// Sentinel errors are the backend-agnostic cache contract, mapped 1:1 to the
// InfraError 32xx Cache codes (infra.proto); the service layer adapts them to
// application errors.
var (
	ErrCacheMiss       = stderrors.New("cache: miss")
	ErrKeyNotFound     = stderrors.New("cache: key not found")
	ErrSessionNotFound = stderrors.New("cache: session not found")
	ErrQuotaExceeded   = stderrors.New("cache: quota exceeded")
	ErrInvalidArgument = stderrors.New("cache: invalid argument")
	ErrLockNotAcquired = stderrors.New("cache: lock not acquired")
	ErrUnavailable     = stderrors.New("cache: unavailable")
)

// CacheDefaultError holds the app error instances a service maps the cache
// sentinels to: the mapping mechanism lives here, the concrete instances are
// supplied by each service's platform layer (same mechanism/instance split as
// entutil.HandleEntError), so this package never imports a service error proto.
type CacheDefaultError struct {
	CacheMiss       *kratoserrors.Error
	KeyNotFound     *kratoserrors.Error
	SessionNotFound *kratoserrors.Error
	QuotaExceeded   *kratoserrors.Error
	InvalidArgument *kratoserrors.Error
	LockNotAcquired *kratoserrors.Error
	Unavailable     *kratoserrors.Error // optional: nil → unknown errors pass through unchanged
}

// ValidateCacheDefaultError fails fast if a required slot is nil — a nil slot
// nil-derefs inside WithCause. Unavailable is optional.
func ValidateCacheDefaultError(errs *CacheDefaultError) error {
	for _, e := range []*kratoserrors.Error{
		errs.CacheMiss, errs.KeyNotFound, errs.SessionNotFound,
		errs.QuotaExceeded, errs.InvalidArgument, errs.LockNotAcquired,
	} {
		if e == nil {
			return stderrors.New("cache: CacheDefaultError has a nil sentinel slot")
		}
	}
	return nil
}

// HandleCacheError maps a cache error to the supplied app error, keeping the
// original as cause. The Unavailable branch is nil-safe: unknown errors pass
// through unchanged when it is unset.
func HandleCacheError(err error, errs *CacheDefaultError) error {
	if e := ValidateCacheDefaultError(errs); e != nil {
		return e
	}
	switch {
	case stderrors.Is(err, ErrCacheMiss):
		return errs.CacheMiss.WithCause(err)
	case stderrors.Is(err, ErrKeyNotFound):
		return errs.KeyNotFound.WithCause(err)
	case stderrors.Is(err, ErrSessionNotFound):
		return errs.SessionNotFound.WithCause(err)
	case stderrors.Is(err, ErrQuotaExceeded):
		return errs.QuotaExceeded.WithCause(err)
	case stderrors.Is(err, ErrInvalidArgument):
		return errs.InvalidArgument.WithCause(err)
	case stderrors.Is(err, ErrLockNotAcquired):
		return errs.LockNotAcquired.WithCause(err)
	default:
		if errs.Unavailable != nil {
			return errs.Unavailable.WithCause(err)
		}
		return err
	}
}
