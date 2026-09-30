package storage

import (
	stderrors "errors"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// Sentinel errors are the backend-agnostic storage contract, mapped 1:1 to the
// InfraError 33xx Storage codes (infra.proto). ErrSizeExceeded is produced
// in-process (sizeLimitReader), never by the backend SDK.
var (
	ErrNotFound        = stderrors.New("storage: object not found")
	ErrForbidden       = stderrors.New("storage: access forbidden")
	ErrSizeExceeded    = stderrors.New("storage: size exceeds limit")
	ErrInvalidArgument = stderrors.New("storage: invalid argument")
	ErrUnavailable     = stderrors.New("storage: unavailable")
)

// StorageDefaultError holds the app error instances a service maps the storage
// sentinels to: the mapping mechanism lives here, the concrete instances are
// supplied by each service's platform layer (same mechanism/instance split as
// cache.HandleCacheError), so this package never imports a service error proto.
type StorageDefaultError struct {
	NotFound        *kratoserrors.Error
	Forbidden       *kratoserrors.Error
	SizeExceeded    *kratoserrors.Error
	InvalidArgument *kratoserrors.Error
	Unavailable     *kratoserrors.Error // optional: nil → unknown errors pass through unchanged
}

// ValidateStorageDefaultError fails fast if a required slot is nil — a nil
// slot nil-derefs inside WithCause. Unavailable is optional.
func ValidateStorageDefaultError(errs *StorageDefaultError) error {
	for _, e := range []*kratoserrors.Error{
		errs.NotFound, errs.Forbidden, errs.SizeExceeded, errs.InvalidArgument,
	} {
		if e == nil {
			return stderrors.New("storage: StorageDefaultError has a nil sentinel slot")
		}
	}
	return nil
}

// HandleStorageError maps a storage error to the supplied app error, keeping
// the original as cause. The Unavailable branch is nil-safe: unknown errors
// pass through unchanged when it is unset.
func HandleStorageError(err error, errs *StorageDefaultError) error {
	if e := ValidateStorageDefaultError(errs); e != nil {
		return e
	}
	switch {
	case stderrors.Is(err, ErrNotFound):
		return errs.NotFound.WithCause(err)
	case stderrors.Is(err, ErrForbidden):
		return errs.Forbidden.WithCause(err)
	case stderrors.Is(err, ErrSizeExceeded):
		return errs.SizeExceeded.WithCause(err)
	case stderrors.Is(err, ErrInvalidArgument):
		return errs.InvalidArgument.WithCause(err)
	default:
		if errs.Unavailable != nil {
			return errs.Unavailable.WithCause(err)
		}
		return err
	}
}
