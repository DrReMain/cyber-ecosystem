package platform

import (
	errorspb "cyber-ecosystem/contracts/go/errors"
	"cyber-ecosystem/shared-go/storage"
)

func NewStorageErrorHandler() StorageErrorHandler {
	return func(err error) error {
		return storage.HandleStorageError(err, &storage.StorageDefaultError{
			NotFound:    errorspb.ErrorInfraErrorStorageNotFound(""),
			Forbidden:   errorspb.ErrorInfraErrorStorageForbidden(""),
			LimitExceed: errorspb.ErrorInfraErrorStorageSizeExceed(""),
			Unavailable: errorspb.ErrorInfraErrorStorageUnavailable(""),
		})
	}
}
