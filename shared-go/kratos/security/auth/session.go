package auth

import (
	"context"
	stdhttp "net/http"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"

	"cyber-ecosystem/shared-go/kratos/security"
)

// Authenticator turns a session credential into the subject it belongs to.
// Implementations own credential storage and expiry; the middleware only
// transports the token.
type Authenticator interface {
	AuthenticateSession(ctx context.Context, token string) (*security.Subject, error)
}

// ErrMissingSession denies requests without a usable session credential.
var ErrMissingSession = errors.Unauthorized("MISSING_SESSION", "")

type Option func(*options)

type options struct {
	cookieName string
}

// SessionCookie sets the cookie the middleware reads the session token
// from; it is required.
func SessionCookie(name string) Option {
	return func(o *options) { o.cookieName = name }
}

// SessionAuth authenticates requests by their session cookie and injects
// the resolved subject into the context. Everything fails closed: missing
// transport, missing cookie, and a nil subject all deny.
func SessionAuth(authn Authenticator, opts ...Option) middleware.Middleware {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	// Without a cookie name every request would fail with a confusing
	// MISSING_SESSION; that is wiring breakage, so it fails loudly here.
	if o.cookieName == "" {
		panic("SessionAuth: SessionCookie(name) is required")
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, ErrMissingSession
			}
			token := cookieValue(tr.RequestHeader(), o.cookieName)
			if token == "" {
				return nil, ErrMissingSession
			}
			subject, err := authn.AuthenticateSession(ctx, token)
			if err != nil {
				return nil, err
			}
			if subject == nil {
				// nil-with-nil-error is an authenticator contract violation;
				// downstream middleware must never see a present-but-nil subject.
				return nil, ErrMissingSession
			}
			return handler(security.WithSubject(ctx, subject), req)
		}
	}
}

func cookieValue(header transport.Header, name string) string {
	// Multiple Cookie headers are legal on the wire (proxies split them);
	// parse every value rather than only the first.
	raw := stdhttp.Header{}
	for _, v := range header.Values("Cookie") {
		raw.Add("Cookie", v)
	}
	for _, c := range (&stdhttp.Request{Header: raw}).Cookies() {
		if c.Name == name && c.Value != "" {
			return c.Value
		}
	}
	return ""
}
