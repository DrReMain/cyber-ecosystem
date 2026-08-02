package cache

import (
	stderrors "errors"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// CacheDefaultError holds the application error instances a service maps the
// cache sentinel errors to. The mapping MECHANISM lives here in shared-go (so it
// is reusable and backend-agnostic), while the concrete *errors.Error instances
// are supplied by each service's platform layer, which owns the app error proto.
// Splitting mechanism from instances keeps the cache package free of any service
// error-proto dependency — the same separation entutil.HandleEntError uses.
type CacheDefaultError struct {
	CacheMiss       *kratoserrors.Error
	KeyNotFound     *kratoserrors.Error
	SessionNotFound *kratoserrors.Error
	QuotaExceeded   *kratoserrors.Error
	InvalidArgument *kratoserrors.Error
	LockNotAcquired *kratoserrors.Error
	Unavailable     *kratoserrors.Error
}

// ValidateCacheDefaultError fails fast at construction if a required sentinel
// slot is nil — otherwise the matching branch nil-derefs inside
// kratoserrors.WithCause, panicking in the error-handling hot path (the worst
// place to panic). Unavailable is intentionally optional: when nil, an unknown
// error is passed through unchanged rather than mapped, so a service can opt
// into raw pass-through for errors it has no specific code for.
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

// HandleCacheError maps a cache-layer error to the supplied application error,
// attaching the original as the cause (WithCause), so the specific reason
// reaches the caller while the underlying detail stays available for logging.
// It validates the sentinel slots on every call (mirroring entutil), so a
// misconfigured CacheDefaultError surfaces as an explicit error rather than a
// nil-deref panic. The Unavailable default branch is nil-safe: when Unavailable
// is unset, an unknown error is passed through unchanged.
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
