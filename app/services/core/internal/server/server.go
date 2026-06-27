package server

import (
	"github.com/go-kratos/kratos/v3/middleware/ratelimit"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/google/wire"

	"cyber-ecosystem/shared-go/kratos/middleware/validator"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// Map the validator's error onto the shared GENERAL_ERROR_VALIDATION_FAILED error.
func init() {
	recovery.ErrUnknownRequest = errorspb.ErrorGeneralErrorUnspecified("").WithCause(recovery.ErrUnknownRequest)
	ratelimit.ErrLimitExceed = errorspb.ErrorFlowErrorRateLimited("").WithCause(ratelimit.ErrLimitExceed)
	validator.ErrValidator = errorspb.ErrorGeneralErrorValidationFailed("").WithCause(validator.ErrValidator)
}

var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewConnectServer)
