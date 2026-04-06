package server

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/circuitbreaker"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"

	"cyber-ecosystem/shared-go/kratos/middleware/validate"

	app1V1 "cyber-ecosystem/apps/app_1/gen/go/v1"
)

func init() {
	recovery.ErrUnknownRequest = app1V1.ErrorErrorReasonUnspecified("")

	ratelimit.ErrLimitExceed = app1V1.ErrorErrorReasonRatelimit("")

	circuitbreaker.ErrNotAllowed = app1V1.ErrorErrorReasonCircuitbreaker("")

	jwt.ErrMissingJwtToken = app1V1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrMissingJwtToken)
	jwt.ErrMissingKeyFunc = app1V1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrMissingKeyFunc)
	jwt.ErrTokenInvalid = app1V1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrTokenInvalid)
	jwt.ErrTokenExpired = app1V1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrTokenExpired)
	jwt.ErrTokenParseFail = app1V1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrTokenParseFail)
	jwt.ErrUnSupportSigningMethod = app1V1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrUnSupportSigningMethod)
	jwt.ErrWrongContext = app1V1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrWrongContext)
	jwt.ErrNeedTokenProvider = app1V1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrNeedTokenProvider)
	jwt.ErrSignToken = app1V1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrSignToken)
	jwt.ErrGetKey = app1V1.ErrorErrorReasonUnauthorized("").WithCause(jwt.ErrGetKey)

	validate.ErrVALIDATOR = app1V1.ErrorErrorReasonValidator("")
}

var ProviderSet = wire.NewSet(NewOpsServer, NewGRPCServer, NewHTTPServer, NewConnectServer, NewI18nBundle)
