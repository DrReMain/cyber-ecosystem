// Package sanitize masks errors at service boundaries so implementation
// details do not leak across them.
package sanitize

import (
	"context"
	"errors"

	connect "connectrpc.com/connect"
	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// ErrUnexpected is the fallback for non-kratos errors masked by Server. It is
// a plain kratos error so the package depends on no project-specific error
// scheme; override it to customize.
var ErrUnexpected = kratoserrors.New(500, "INTERNAL", "internal server error")

// Server masks any non-kratos error from the handler as ErrUnexpected, keeping
// the original as the cause; kratos errors pass through unchanged.
func Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			reply, err := handler(ctx, req)
			if err == nil {
				return reply, nil
			}
			var ke *kratoserrors.Error
			if errors.As(err, &ke) {
				return reply, err
			}
			return nil, ErrUnexpected.WithCause(err)
		}
	}
}

// Client is the outbound counterpart: it remaps any upstream error into this
// service's error space so the provider's reason and internal detail never
// cross back to biz. Where Server trusts this service's own errors and passes
// them through, Client translates the provider's — same boundary intent,
// direction-specific policy.
func Client() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			reply, err := handler(ctx, req)
			if err != nil {
				return nil, MapUpstreamErr(err)
			}
			return reply, nil
		}
	}
}

// MapUpstreamErr maps an upstream error to a GeneralError by HTTP code,
// preserving the code's semantics (NotFound→NotFound, Unavailable→…) while
// regenerating a clean reason and message, so neither the provider's reason
// nor transport detail leaks.
func MapUpstreamErr(err error) error {
	switch extractHTTPCode(err) {
	case 400:
		return errorspb.ErrorGeneralErrorInvalidArgument("upstream rejected argument")
	case 401:
		return errorspb.ErrorGeneralErrorUnauthenticated("upstream unauthenticated")
	case 403:
		return errorspb.ErrorGeneralErrorPermissionDenied("upstream forbidden")
	case 404:
		return errorspb.ErrorGeneralErrorNotFound("upstream not found")
	case 409:
		return errorspb.ErrorGeneralErrorAlreadyExists("upstream conflict")
	case 412:
		return errorspb.ErrorGeneralErrorPreconditionFailed("upstream precondition failed")
	case 413:
		return errorspb.ErrorGeneralErrorPayloadTooLarge("upstream payload too large")
	case 429:
		return errorspb.ErrorGeneralErrorResourceExhausted("upstream rate limited")
	case 501:
		return errorspb.ErrorGeneralErrorNotImplemented("upstream not implemented")
	case 503:
		return errorspb.ErrorGeneralErrorUnavailable("upstream unavailable")
	case 504:
		return errorspb.ErrorGeneralErrorTimeout("upstream timeout")
	default:
		return errorspb.ErrorGeneralErrorInternal("upstream error")
	}
}

// extractHTTPCode returns the HTTP status code carried by err: a kratos
// *Error.Code directly, otherwise the grpc/connect status code mapped to HTTP.
func extractHTTPCode(err error) int {
	var ke *kratoserrors.Error
	if errors.As(err, &ke) {
		return int(ke.Code)
	}
	if c := status.Code(err); c != codes.Unknown {
		return grpcCodeToHTTP(int(c))
	}
	if c := connect.CodeOf(err); c != connect.CodeUnknown {
		return grpcCodeToHTTP(int(c))
	}
	return 500
}

// grpcCodeToHTTP maps a gRPC status code (shared by grpc and connect) to its
// HTTP status equivalent.
func grpcCodeToHTTP(c int) int {
	switch c {
	case int(codes.InvalidArgument), int(codes.FailedPrecondition), int(codes.OutOfRange):
		return 400
	case int(codes.Unauthenticated):
		return 401
	case int(codes.PermissionDenied):
		return 403
	case int(codes.NotFound):
		return 404
	case int(codes.AlreadyExists), int(codes.Aborted):
		return 409
	case int(codes.ResourceExhausted):
		return 429
	case int(codes.Unimplemented):
		return 501
	case int(codes.Unavailable):
		return 503
	case int(codes.DeadlineExceeded):
		return 504
	default:
		return 500
	}
}
