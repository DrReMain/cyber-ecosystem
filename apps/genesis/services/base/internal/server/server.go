package server

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"

	errorspb "cyber-ecosystem/contracts/go/errors"
	"cyber-ecosystem/shared-go/kratos/middleware/validate"
)

func init() {
	recovery.ErrUnknownRequest = errorspb.ErrorGeneralErrorUnspecified("").WithCause(recovery.ErrUnknownRequest)

	ratelimit.ErrLimitExceed = errorspb.ErrorFlowErrorRateLimited("").WithCause(ratelimit.ErrLimitExceed)

	validate.ErrVALIDATOR = errorspb.ErrorGeneralErrorValidationFailed("").WithCause(validate.ErrVALIDATOR)
}

var ProviderSet = wire.NewSet(
	NewOpsServer,
	NewGRPCServer,
	NewHTTPServer,
	NewConnectServer,
)
