package server

import (
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/ratelimit"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/middleware/selector"
	"github.com/google/wire"

	kaudit "cyber-ecosystem/shared-go/kratos/audit"
	"cyber-ecosystem/shared-go/kratos/middleware/sanitize"
	"cyber-ecosystem/shared-go/kratos/middleware/validator"
	"cyber-ecosystem/shared-go/kratos/security"
	krauth "cyber-ecosystem/shared-go/kratos/security/auth"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	extv1 "cyber-ecosystem/gen/go/cyber/ext/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

func init() {
	sanitize.ErrUnexpected = errorspb.ErrorGeneralErrorInternal("")

	recovery.ErrUnknownRequest = errorspb.ErrorGeneralErrorUnspecified("").WithCause(recovery.ErrUnknownRequest)
	ratelimit.ErrLimitExceed = errorspb.ErrorFlowErrorRateLimited("").WithCause(ratelimit.ErrLimitExceed)
	validator.ErrValidator = errorspb.ErrorGeneralErrorValidationFailed("").WithCause(validator.ErrValidator)

	security.ErrMissingAnnotation = errorspb.ErrorGeneralErrorInternal("").WithCause(security.ErrMissingAnnotation)
	security.ErrNotPlatformTenant = errorspb.ErrorGeneralErrorPermissionDenied("").WithCause(security.ErrNotPlatformTenant)
	security.ErrNotBackendPlane = errorspb.ErrorGeneralErrorPermissionDenied("").WithCause(security.ErrNotBackendPlane)
	krauth.ErrMissingSession = errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(krauth.ErrMissingSession)
	kauthz.ErrMissingSubject = errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(kauthz.ErrMissingSubject)
	kauthz.ErrPermissionDenied = errorspb.ErrorGeneralErrorPermissionDenied("").WithCause(kauthz.ErrPermissionDenied)
	kauthz.ErrPolicyDenied = systempb.ErrorSystemAuthzPolicyDenied("").WithCause(kauthz.ErrPolicyDenied)
}

func accessLadder(authn krauth.Authenticator, auditSink kaudit.Sink, authorizer kauthz.Authorizer) []middleware.Middleware {
	return []middleware.Middleware{
		selector.Server(security.DefaultGuard()).Match(security.MatchAccess(extv1.Access_ACCESS_UNSPECIFIED)).Build(),

		selector.Server(
			krauth.SessionAuth(authn),
			kaudit.Audit(auditSink),
		).Match(security.MatchAccess(extv1.Access_ACCESS_BUILTIN)).Build(),

		selector.Server(
			krauth.SessionAuth(authn),
			kaudit.Audit(auditSink),
			security.BackendPlaneGuard(),
			security.PlatformGuard(),
			kauthz.OperationAuthz(authorizer),
		).Match(security.MatchAccess(extv1.Access_ACCESS_PLATFORM)).Build(),

		selector.Server(
			krauth.SessionAuth(authn),
			kaudit.Audit(auditSink),
			security.BackendPlaneGuard(),
			kauthz.OperationAuthz(authorizer),
		).Match(security.MatchAccess(extv1.Access_ACCESS_ADMIN)).Build(),
	}
}

var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewConnectServer, NewRegistrarList)
