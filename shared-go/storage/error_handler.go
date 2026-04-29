package storage

import (
	"errors"

	kratoserrors "github.com/go-kratos/kratos/v2/errors"
)

type StorageDefaultError struct {
	NotFound    *kratoserrors.Error
	Forbidden   *kratoserrors.Error
	LimitExceed *kratoserrors.Error
	Unavailable *kratoserrors.Error
}

func HandleStorageError(err error, errs *StorageDefaultError) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return errs.NotFound.WithCause(err)
	case errors.Is(err, ErrForbidden):
		return errs.Forbidden.WithCause(err)
	case errors.Is(err, ErrLimitExceed):
		return errs.LimitExceed.WithCause(err)
	default:
		return errs.Unavailable.WithCause(err)
	}
}
