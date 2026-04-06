package condition

import "context"

type clientIPKey struct{}

// WithClientIP stores the client IP in context for evaluator use.
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

// ClientIPFromContext retrieves the client IP from context.
func ClientIPFromContext(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey{}).(string)
	return ip
}

type userAttrsKey struct{}

// WithUserAttributes stores user attributes in context for evaluator use.
func WithUserAttributes(ctx context.Context, attrs map[string]string) context.Context {
	return context.WithValue(ctx, userAttrsKey{}, attrs)
}

// UserAttributesFromContext retrieves user attributes from context.
func UserAttributesFromContext(ctx context.Context) map[string]string {
	attrs, _ := ctx.Value(userAttrsKey{}).(map[string]string)
	return attrs
}
