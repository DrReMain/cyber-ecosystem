package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport"

	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/security"
)

// ensureSecurityContext extracts Operation and ClientIP from the transport layer
// and writes them to security context for downstream consumers (biz/data layers).
// Idempotent: returns ctx unchanged if security context is already set.
func ensureSecurityContext(ctx context.Context) context.Context {
	if _, ok := security.RequestContextFromContext(ctx); ok {
		return ctx
	}
	tp, ok := transport.FromServerContext(ctx)
	if !ok {
		return ctx
	}
	clientIP := security.NormalizeClientIP(tp.RequestHeader().Get("X-Forwarded-For"))
	if clientIP == "" {
		clientIP = security.NormalizeClientIP(tp.RequestHeader().Get("X-Real-Ip"))
	}
	return security.WithRequestContext(ctx, security.RequestContext{
		Operation: tp.Operation(),
		ClientIP:  clientIP,
	})
}
