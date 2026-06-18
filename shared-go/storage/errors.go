package storage

import "errors"

// Sentinel errors are the storage contract (backend-agnostic). They map 1:1 to
// the InfraError 33xx Storage codes (infra.proto); the service platform layer
// adapts them to application errors via HandleStorageError. ErrSizeExceeded is
// produced in-process (by sizeLimitReader), never by the backend SDK.
var (
	ErrNotFound        = errors.New("storage: object not found")
	ErrForbidden       = errors.New("storage: access forbidden")
	ErrSizeExceeded    = errors.New("storage: size exceeds limit")
	ErrInvalidArgument = errors.New("storage: invalid argument")
	ErrUnavailable     = errors.New("storage: unavailable")
)
