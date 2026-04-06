package validate

import (
	"context"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
)

var ErrVALIDATOR = errors.BadRequest("VALIDATOR", "verification failed")

func ProtoValidate() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			if msg, ok := req.(proto.Message); ok {
				if err := protovalidate.Validate(msg); err != nil {
					// if valErr, ok := err.(*protovalidate.ValidationError); ok {
					// 	for _, v := range valErr.Violations {
					// 		*v.Proto.Message
					// 	}
					// }
					return nil, ErrVALIDATOR.WithCause(err)
				}
			}
			return handler(ctx, req)
		}
	}
}
