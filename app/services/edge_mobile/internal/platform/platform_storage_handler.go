package platform

import (
	"cyber-ecosystem/shared-go/storage"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

var defaultStorageError = &storage.StorageDefaultError{
	NotFound:        errorspb.ErrorInfraErrorStorageNotFound(""),
	Forbidden:       errorspb.ErrorInfraErrorStorageForbidden(""),
	SizeExceeded:    errorspb.ErrorInfraErrorStorageSizeExceed(""),
	InvalidArgument: errorspb.ErrorInfraErrorStorageInvalidArgument(""),
	Unavailable:     errorspb.ErrorInfraErrorStorageUnavailable(""),
}

func NewStorageErrorHandler() (StorageErrorHandler, error) {
	// Fail at wire boot (not on first request) if a sentinel slot is nil.
	if err := storage.ValidateStorageDefaultError(defaultStorageError); err != nil {
		return nil, err
	}
	return func(err error) error {
		return storage.HandleStorageError(err, defaultStorageError)
	}, nil
}
