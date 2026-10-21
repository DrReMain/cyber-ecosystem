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

// SessionCookieName is the session cookie's wire contract across every
// server, outbound forwarder, and client in the ecosystem — the single
// home. Consumers reference it; a re-spelled literal drifts silently.
const SessionCookieName = "session_token"

// SessionNamespaceHeader declares which client world's credential slot the
// request presents. The declaration is an untrusted selection among
// cookies the browser already sent — the plane always derives from the
// verified seat, never from this header.
const SessionNamespaceHeader = "x-session-namespace"

// SessionCookieNameFor resolves the session cookie name for one client
// namespace: undeclared callers (service-to-service forwards, probes) read
// the base slot; a declared namespace gets its own slot, so multiple
// client worlds hold concurrent credentials on a single host.
func SessionCookieNameFor(namespace string) string {
	if namespace == "" {
		return SessionCookieName
	}
	return SessionCookieName + "_" + namespace
}

// SessionAuth authenticates requests by their session cookie and injects
// the resolved subject into the context. Everything fails closed: missing
// transport, missing cookie, and a nil subject all deny.
func SessionAuth(authn Authenticator) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, ErrMissingSession
			}
			token := cookieValue(tr.RequestHeader(), SessionCookieNameFor(tr.RequestHeader().Get(SessionNamespaceHeader)))
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
