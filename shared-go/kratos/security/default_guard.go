package security

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
)

// protectedNamespace is the proto namespace all business RPCs live under. RPCs
// outside it are treated as framework built-ins (see DefaultGuard).
const protectedNamespace = "cyber."

var ErrMissingANNOTATION = errors.ServiceUnavailable("MISSING_ANNOTATION", "")

// DefaultGuard is the middleware for the ACCESS_UNSPECIFIED selector branch —
// the catch-all for business RPCs missing the cyber.ext.v1.access annotation.
// It rejects them so a forgotten annotation fails loudly instead of silently
// passing.
//
// Framework handlers pass through: grpc's built-in health/reflection/channelz
// (and any future transport built-in) cannot carry the access annotation, and
// on the grpc transport they reach this chain via the interceptor. They are
// told apart from business RPCs by namespace — any operation not under
// protectedNamespace is assumed to be a framework RPC.
//
// Load-bearing assumption: every business RPC's operation is
// /<protectedNamespace><domain>.v1.Svc/Method (e.g. /cyber.system.v1.UserService/GetUser).
// This holds because monorepo protos live under proto/cyber/* and buf lint
// (PACKAGE_DIRECTORY_MATCH) keeps package == directory. A business RPC defined
// outside cyber.* would be wrongly treated as a framework RPC and let through
// unannotated — keep business protos under proto/cyber/*.
func DefaultGuard() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				if !strings.HasPrefix(strings.TrimPrefix(tr.Operation(), "/"), protectedNamespace) {
					return handler(ctx, req) // framework built-in: annotation not applicable
				}
			}
			return nil, ErrMissingANNOTATION
		}
	}
}
