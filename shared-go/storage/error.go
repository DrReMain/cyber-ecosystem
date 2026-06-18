package storage

import (
	stderrors "errors"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// StorageDefaultError holds the application error instances a service maps the
// storage sentinel errors to. The mapping MECHANISM lives here in shared-go
// (reusable, backend-agnostic); the concrete *errors.Error instances are
// supplied by each service's platform layer, which owns the app error proto —
// the same mechanism/instance split cache.HandleCacheError uses, so this
// package never imports a service error proto.
type StorageDefaultError struct {
	NotFound        *kratoserrors.Error
	Forbidden       *kratoserrors.Error
	SizeExceeded    *kratoserrors.Error
	InvalidArgument *kratoserrors.Error
	Unavailable     *kratoserrors.Error // optional: nil → unknown errors pass through unchanged
}

// ValidateStorageDefaultError fails fast at construction if a required sentinel
// slot is nil — otherwise the matching branch nil-derefs inside
// kratoserrors.WithCause, panicking in the error-handling hot path. Unavailable
// is intentionally optional (nil-safe pass-through in HandleStorageError).
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

// HandleStorageError maps a storage-layer error to the supplied application
// error, attaching the original as the cause (WithCause). Validates slots on
// every call so a misconfigured StorageDefaultError surfaces explicitly rather
// than nil-panicking. The Unavailable branch is nil-safe.
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
