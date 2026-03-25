package connect

import (
	"context"
	stderrors "errors"

	connectrpc "connectrpc.com/connect"

	"github.com/go-kratos/kratos/v2/errors"
	"google.golang.org/protobuf/proto"
)

type ErrorMessageResolver func(context.Context, *errors.Error) string

type ErrorDetailBuilder func(context.Context, error, *errors.Error, string) proto.Message

func NewErrorEncoder(resolveMessage ErrorMessageResolver, buildDetail ErrorDetailBuilder) func(context.Context, error) error {
	return func(ctx context.Context, sourceErr error) error {
		if sourceErr == nil {
			return nil
		}
		se := errors.FromError(sourceErr)
		message := se.Message
		if resolveMessage != nil {
			message = resolveMessage(ctx, se)
		}
		code := connectrpc.CodeUnknown
		if gs := se.GRPCStatus(); gs != nil {
			code = mapGRPCCodeToConnect(gs.Code())
		}
		encoded := connectrpc.NewError(code, stderrors.New(message))
		if buildDetail != nil {
			if detailMsg := buildDetail(ctx, sourceErr, se, message); detailMsg != nil {
				if detail, err := connectrpc.NewErrorDetail(detailMsg); err == nil {
					encoded.AddDetail(detail)
				}
			}
		}
		return encoded
	}
}
