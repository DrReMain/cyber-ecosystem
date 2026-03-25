package traceheader

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"go.opentelemetry.io/otel/trace"
)

const HeaderTraceID = "X-Trace-Id"

func Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			reply, err := handler(ctx, req)
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return reply, err
			}
			sc := trace.SpanContextFromContext(ctx)
			if !sc.IsValid() {
				return reply, err
			}
			traceID := sc.TraceID().String()
			if traceID == "" {
				return reply, err
			}
			tr.ReplyHeader().Set(HeaderTraceID, traceID)
			return reply, err
		}
	}
}
