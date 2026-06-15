// Package validator provides a kratos middleware that runs protovalidate
// (buf.validate) rules on each unary request.
package validator

import (
	"context"

	"buf.build/go/protovalidate"
	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"google.golang.org/protobuf/proto"
)

// ErrValidator is the error returned on validation failure. Override it in an
// init() to map it onto the app's error enum.
var ErrValidator = errors.BadRequest("VALIDATOR", "verification failed")

// Server returns a middleware that validates each unary request against its
// buf.validate rules. i18n integration (violation extraction, direct-expose)
// is planned — see docs/design/i18n-error.md.
func Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			if msg, ok := req.(proto.Message); ok {
				if verr := protovalidate.Validate(msg); verr != nil {
					return nil, ErrValidator.WithCause(verr)
				}
			}
			return handler(ctx, req)
		}
	}
}
