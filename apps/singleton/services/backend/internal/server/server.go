package server

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/circuitbreaker"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"

	"cyber-ecosystem/shared-go/kratos/middleware/validate"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
)

func init() {
	recovery.ErrUnknownRequest = singletonV1.ErrorErrorReasonUnspecified("")

	ratelimit.ErrLimitExceed = singletonV1.ErrorErrorReasonRatelimit("")

	circuitbreaker.ErrNotAllowed = singletonV1.ErrorErrorReasonCircuitbreaker("")

	jwt.ErrMissingJwtToken = singletonV1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrMissingJwtToken)
	jwt.ErrMissingKeyFunc = singletonV1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrMissingKeyFunc)
	jwt.ErrTokenInvalid = singletonV1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrTokenInvalid)
	jwt.ErrTokenExpired = singletonV1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrTokenExpired)
	jwt.ErrTokenParseFail = singletonV1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrTokenParseFail)
	jwt.ErrUnSupportSigningMethod = singletonV1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrUnSupportSigningMethod)
	jwt.ErrWrongContext = singletonV1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrWrongContext)
	jwt.ErrNeedTokenProvider = singletonV1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrNeedTokenProvider)
	jwt.ErrSignToken = singletonV1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrSignToken)
	jwt.ErrGetKey = singletonV1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrGetKey)

	validate.ErrVALIDATOR = singletonV1.ErrorErrorReasonValidator("")
}

var ProviderSet = wire.NewSet(NewOpsServer, NewGRPCServer, NewHTTPServer, NewConnectServer, NewI18nBundle)
