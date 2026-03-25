package headerforward

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// ClientFromServer forwards selected request headers from server context to client context.
func ClientFromServer(headerKeys ...string) middleware.Middleware {
	keys := normalizeKeys(headerKeys)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			serverTransport, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			clientTransport, ok := transport.FromClientContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			for _, key := range keys {
				value := strings.TrimSpace(serverTransport.RequestHeader().Get(key))
				if value == "" {
					continue
				}
				clientTransport.RequestHeader().Set(key, value)
			}
			return handler(ctx, req)
		}
	}
}

func normalizeKeys(headerKeys []string) []string {
	keys := make([]string, 0, len(headerKeys))
	for _, key := range headerKeys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		keys = append(keys, key)
	}
	return keys
}
