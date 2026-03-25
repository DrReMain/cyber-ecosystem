package responsemeta

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

const (
	HeaderResponseSuccess = "X-Response-Success"
	HeaderErrorReason     = "X-Error-Reason"
)

func Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			reply, err := handler(ctx, req)
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return reply, err
			}
			if err == nil {
				tr.ReplyHeader().Set(HeaderResponseSuccess, "true")
				return reply, nil
			}
			se := errors.FromError(err)
			tr.ReplyHeader().Set(HeaderResponseSuccess, "false")
			if se.Reason != "" {
				tr.ReplyHeader().Set(HeaderErrorReason, se.Reason)
			}
			return reply, err
		}
	}
}
