package traceheader

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

const HeaderTraceID = "X-Trace-Id"

func Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			var traceID string
			sc := trace.SpanContextFromContext(ctx)
			if sc.IsValid() {
				traceID = sc.TraceID().String()
			}
			// Ensure trace ID is always set, even on panic
			// Use defer to guarantee header is set before returning to client
			defer func() {
				if traceID == "" {
					return
				}
				if tr, ok := transport.FromServerContext(ctx); ok {
					tr.ReplyHeader().Set(HeaderTraceID, traceID)
				}
			}()
			reply, err := handler(ctx, req)
			return reply, err
		}
	}
}
