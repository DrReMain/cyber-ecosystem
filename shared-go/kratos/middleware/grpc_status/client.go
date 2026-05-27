package grpc_status

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-kratos/kratos/v2/middleware"

	errorspb "cyber-ecosystem/contracts/go/errors"
)

// Client returns a middleware that maps gRPC status errors to Kratos errors
// with proper reason codes, enabling i18n translation and structured error reporting.
// Place it as the innermost middleware (last in the chain) so all outer middlewares
// see the properly formatted error.
func Client() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			reply, err = handler(ctx, req)
			if err == nil {
				return
			}

			gs, ok := status.FromError(err)
			if !ok {
				return
			}

			switch gs.Code() {
			case codes.Unavailable:
				err = errorspb.ErrorInfraErrorNetworkConnection("").WithCause(err)
			case codes.DeadlineExceeded:
				err = errorspb.ErrorInfraErrorNetworkTimeout("").WithCause(err)
			}

			return
		}
	}
}
