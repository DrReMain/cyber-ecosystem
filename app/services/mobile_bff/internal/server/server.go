package server

import (
	"github.com/google/wire"

	"cyber-ecosystem/shared-go/kratos/middleware/validator"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// Map the validator's error onto the shared GENERAL_ERROR_VALIDATION_FAILED error.
func init() {
	validator.ErrValidator = errorspb.ErrorGeneralErrorValidationFailed("").WithCause(validator.ErrValidator)
}

var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewConnectServer)
