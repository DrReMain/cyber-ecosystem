package security

import (
	"context"
	"strings"
)

type RequestContext struct {
	Operation string
	ClientIP  string
}

type requestContextKey struct{}

func WithRequestContext(ctx context.Context, request RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey{}, request)
}

func RequestContextFromContext(ctx context.Context) (RequestContext, bool) {
	request, ok := ctx.Value(requestContextKey{}).(RequestContext)
	return request, ok
}

func OperationFromContext(ctx context.Context) (string, bool) {
	request, ok := RequestContextFromContext(ctx)
	if !ok || request.Operation == "" {
		return "", false
	}
	return request.Operation, true
}

func ClientIPFromContext(ctx context.Context) (string, bool) {
	request, ok := RequestContextFromContext(ctx)
	if !ok || request.ClientIP == "" {
		return "", false
	}
	return request.ClientIP, true
}

func NormalizeClientIP(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if idx := strings.Index(raw, ","); idx >= 0 {
		raw = raw[:idx]
	}
	return strings.TrimSpace(raw)
}
