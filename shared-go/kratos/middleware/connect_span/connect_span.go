package connect_span

import (
	"context"
	"net"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
)

func Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			ct, ok := tr.(*connecttransport.Transport)
			if !ok {
				return handler(ctx, req)
			}

			span := trace.SpanFromContext(ctx)
			if !span.IsRecording() {
				return handler(ctx, req)
			}

			span.SetAttributes(
				attribute.String("http.method", ct.HTTPMethod()),
				attribute.String("http.route", ct.Operation()),
				attribute.String("http.target", ct.Operation()),
			)

			if addr := ct.RemoteAddr(); addr != "" {
				host, portStr, splitErr := net.SplitHostPort(addr)
				if splitErr == nil {
					port, _ := strconv.Atoi(portStr)
					span.SetAttributes(
						attribute.String("net.peer.ip", host),
						attribute.Int("net.peer.port", port),
					)
				}
			}

			reply, err = handler(ctx, req)

			if reply != nil {
				if msg, ok := reply.(interface{ Any() any }); ok {
					if p, ok := msg.Any().(interface{ Size() int }); ok {
						span.SetAttributes(attribute.Int("send_msg.size", p.Size()))
					}
				}
			}

			return
		}
	}
}
