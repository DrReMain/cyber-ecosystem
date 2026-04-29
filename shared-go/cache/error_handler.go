package cache

import (
	stderrors "errors"

	kratoserrors "github.com/go-kratos/kratos/v2/errors"
)

type CacheDefaultError struct {
	CacheMiss       *kratoserrors.Error
	KeyNotFound     *kratoserrors.Error
	SessionNotFound *kratoserrors.Error
	QuotaExceeded   *kratoserrors.Error
	InvalidArgument *kratoserrors.Error
	Unavailable     *kratoserrors.Error
}

func HandleCacheError(err error, errs *CacheDefaultError) error {
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
	default:
		return errs.Unavailable.WithCause(err)
	}
}
