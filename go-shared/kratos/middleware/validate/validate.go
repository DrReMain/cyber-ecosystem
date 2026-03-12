package validate

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
)

func ProtoValidate(formatError func(error) string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			if msg, ok := req.(proto.Message); ok {
				if err := protovalidate.Validate(msg); err != nil {
					return nil, errors.BadRequest("VALIDATOR", formatError(err)).WithCause(err)
				}
			}
			return handler(ctx, req)
		}
	}
}

// UseProtoMessage eg: validate.ProtoValidate(validate.UseProtoMessage)
func UseProtoMessage(err error) string {
	if valErr, ok := err.(*protovalidate.ValidationError); ok {
		for _, v := range valErr.Violations {
			return *v.Proto.Message
		}
	}
	return ""
}
