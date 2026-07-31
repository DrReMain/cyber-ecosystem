package server

import (
	"github.com/go-kratos/kratos/v3/middleware/ratelimit"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/google/wire"

	"cyber-ecosystem/shared-go/kratos/middleware/sanitize"
	"cyber-ecosystem/shared-go/kratos/middleware/validator"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// Map framework middleware errors onto the shared error scheme.
func init() {
	sanitize.ErrUnexpected = errorspb.ErrorGeneralErrorInternal("")

	recovery.ErrUnknownRequest = errorspb.ErrorGeneralErrorUnspecified("").WithCause(recovery.ErrUnknownRequest)
	ratelimit.ErrLimitExceed = errorspb.ErrorFlowErrorRateLimited("").WithCause(ratelimit.ErrLimitExceed)
	validator.ErrValidator = errorspb.ErrorGeneralErrorValidationFailed("").WithCause(validator.ErrValidator)
}

var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewConnectServer)
