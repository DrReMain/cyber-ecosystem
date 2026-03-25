package response

import (
	"errors"

	connectrpc "connectrpc.com/connect"

	"github.com/DrReMain/cyber-ecosystem/gen/go/common"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Build normalizes a protocol-native RPC result into common.ResponseBody.
func Build(result proto.Message, err error) (*common.ResponseBody, error) {
	if err != nil {
		return &common.ResponseBody{
			Success: false,
			Err:     ExtractErrorBody(err),
		}, nil
	}
	anyResult, marshalErr := marshalResult(result)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return &common.ResponseBody{
		Success: true,
		Result:  anyResult,
	}, nil
}

// ExtractErrorBody converts gRPC/Connect/Kratos errors into common.ErrorBody.
func ExtractErrorBody(err error) *common.ErrorBody {
	if err == nil {
		return nil
	}
	if body := fromConnect(err); body != nil {
		return body
	}
	if body := fromGRPCStatus(err); body != nil {
		return body
	}
	se := kerrors.FromError(err)
	if se != nil {
		return &common.ErrorBody{
			Reason:  se.Reason,
			Message: se.Message,
		}
	}
	return &common.ErrorBody{Message: err.Error()}
}

func marshalResult(result proto.Message) (*anypb.Any, error) {
	if result == nil {
		return nil, nil
	}
	return anypb.New(result)
}

func fromConnect(err error) *common.ErrorBody {
	var ce *connectrpc.Error
	if !errors.As(err, &ce) {
		return nil
	}
	for _, detail := range ce.Details() {
		value, valueErr := detail.Value()
		if valueErr != nil {
			continue
		}
		if body, ok := value.(*common.ErrorBody); ok {
			return body
		}
	}
	return nil
}

func fromGRPCStatus(err error) *common.ErrorBody {
	s := status.Convert(err)
	if s == nil {
		return nil
	}
	for _, detail := range s.Details() {
		if body, ok := detail.(*common.ErrorBody); ok {
			return body
		}
	}
	return nil
}
