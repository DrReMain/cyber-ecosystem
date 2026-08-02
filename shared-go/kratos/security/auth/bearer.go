package auth

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"

	"cyber-ecosystem/shared-go/kratos/security"
)

const (
	prefix = "Bearer "
	header = "Authorization"
)

// ErrMissingToken is returned when the request carries no bearer token.
// Override it in the server's init() to map it onto the app's error enum.
var ErrMissingToken = errors.Unauthorized("MISSING_TOKEN", "")

// BearerAuth verifies the opaque bearer token carried in the Authorization
// header and injects the authenticated Subject into the request context.
func BearerAuth(auth Authenticator) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tok, ok := bearerToken(ctx)
			if !ok {
				return nil, ErrMissingToken
			}
			subject, err := auth.Authenticate(ctx, tok)
			if err != nil {
				e := errors.FromError(ErrInvalidToken.Unwrap())
				if e != nil {
					e.Message = err.Error()
					return nil, ErrInvalidToken.WithCause(e)
				}
				return nil, ErrInvalidToken.WithCause(err)
			}
			return handler(security.WithSubject(ctx, subject), req)
		}
	}
}

// bearerToken extracts the token from the "Authorization: Bearer <token>"
// header in the server transport context.
func bearerToken(ctx context.Context) (string, bool) {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return "", false
	}
	h := tr.RequestHeader().Get(header)
	if !strings.HasPrefix(h, prefix) {
		return "", false
	}
	return strings.TrimPrefix(h, prefix), true
}
