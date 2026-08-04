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

var (
	ErrMissingToken = errors.Unauthorized("MISSING_TOKEN", "") // no Authorization header
	ErrTokenExpired = errors.Unauthorized("TOKEN_EXPIRED", "") // phase-1: access token missing/expired (client may refresh)
	ErrInvalidToken = errors.Unauthorized("INVALID_TOKEN", "") // phase-2: session gone (client must re-login)
)

func BearerAuth(auth Authenticator) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tok, ok := bearerToken(ctx)
			if !ok {
				return nil, ErrMissingToken
			}
			subject, err := auth.Authenticate(ctx, tok)
			if err != nil {
				return nil, ErrTokenExpired.WithCause(err)
			}
			if err := auth.CheckSession(ctx, subject.SessionID); err != nil {
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
