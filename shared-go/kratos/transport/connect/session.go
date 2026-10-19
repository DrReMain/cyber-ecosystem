package connect

import (
	"context"

	connectrpc "connectrpc.com/connect"

	krauth "cyber-ecosystem/shared-go/kratos/security/auth"
)

type sessionTokenKey struct{}

// WithSessionToken carries a forwarded session token on ctx for outbound
// unary calls.
func WithSessionToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, sessionTokenKey{}, token)
}

// SessionCookieInterceptor stamps the WithSessionToken token as the session
// cookie header on every outbound unary request.
func SessionCookieInterceptor() connectrpc.Interceptor {
	return connectrpc.UnaryInterceptorFunc(func(next connectrpc.UnaryFunc) connectrpc.UnaryFunc {
		return func(ctx context.Context, req connectrpc.AnyRequest) (connectrpc.AnyResponse, error) {
			if token, ok := ctx.Value(sessionTokenKey{}).(string); ok && token != "" {
				req.Header().Set("Cookie", krauth.SessionCookieName+"="+token)
			}
			return next(ctx, req)
		}
	})
}
