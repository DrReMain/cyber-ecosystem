package connect

import (
	stderrors "errors"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/go-kratos/kratos/v2/errors"
)

// ErrorToConnect converts a Kratos error to a Connect error.
// The error reason and metadata are carried in a google.rpc.ErrorInfo detail,
// which is consistent with how native gRPC encodes them in grpc-status-details-bin.
func ErrorToConnect(err error) error {
	if err == nil {
		return nil
	}

	var ce *connect.Error
	if errors.As(err, &ce) {
		return ce
	}

	ke := errors.FromError(err)

	var grpcCode codes.Code
	if gs := ke.GRPCStatus(); gs != nil {
		grpcCode = gs.Code()
	} else {
		grpcCode = codes.Unknown
	}

	connectErr := connect.NewError(mapGRPCCodeToConnect(grpcCode), stderrors.New(ke.Message))

	if ke.Reason != "" || len(ke.Metadata) > 0 {
		info := &errdetails.ErrorInfo{
			Reason:   ke.Reason,
			Metadata: ke.Metadata,
		}
		if detail, detailErr := connect.NewErrorDetail(info); detailErr == nil {
			connectErr.AddDetail(detail)
		}
	}

	return connectErr
}

// ConnectToError converts a Connect error to a Kratos error.
// It extracts reason and metadata from google.rpc.ErrorInfo detail if present.
func ConnectToError(err error) *errors.Error {
	if err == nil {
		return nil
	}

	var ke *errors.Error
	if errors.As(err, &ke) {
		return ke
	}

	var ce *connect.Error
	if !errors.As(err, &ce) {
		return errors.FromError(err)
	}

	grpcCode := mapConnectCodeToGRPC(ce.Code())
	httpCode := mapGRPCCodeToHTTP(grpcCode)

	reason := ""
	var metadata map[string]string
	for _, detail := range ce.Details() {
		msg, err := detail.Value()
		if err != nil {
			continue
		}
		if info, ok := msg.(*errdetails.ErrorInfo); ok {
			reason = info.Reason
			metadata = info.Metadata
			break
		}
	}

	ke = errors.New(httpCode, reason, ce.Message())
	if len(metadata) > 0 {
		ke.Metadata = metadata
	}
	return ke
}

func StatusToConnect(s *status.Status) error {
	if s == nil || s.Code() == codes.OK {
		return nil
	}

	connectErr := connect.NewError(mapGRPCCodeToConnect(s.Code()), stderrors.New(s.Message()))

	for _, detail := range s.Details() {
		if msg, ok := detail.(proto.Message); ok {
			if d, err := connect.NewErrorDetail(msg); err == nil {
				connectErr.AddDetail(d)
			}
		}
	}

	return connectErr
}

func mapGRPCCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.OK:
		return 200
	case codes.InvalidArgument:
		return 400
	case codes.NotFound:
		return 404
	case codes.AlreadyExists:
		return 409
	case codes.PermissionDenied:
		return 403
	case codes.Unauthenticated:
		return 401
	case codes.ResourceExhausted:
		return 429
	case codes.FailedPrecondition:
		return 400
	case codes.Aborted:
		return 409
	case codes.OutOfRange:
		return 400
	case codes.Unimplemented:
		return 501
	case codes.Unavailable:
		return 503
	case codes.DeadlineExceeded:
		return 504
	case codes.Canceled:
		return 499
	default:
		return 500
	}
}

func mapGRPCCodeToConnect(code codes.Code) connect.Code {
	switch code {
	case codes.Canceled:
		return connect.CodeCanceled
	case codes.Unknown:
		return connect.CodeUnknown
	case codes.InvalidArgument:
		return connect.CodeInvalidArgument
	case codes.DeadlineExceeded:
		return connect.CodeDeadlineExceeded
	case codes.NotFound:
		return connect.CodeNotFound
	case codes.AlreadyExists:
		return connect.CodeAlreadyExists
	case codes.PermissionDenied:
		return connect.CodePermissionDenied
	case codes.ResourceExhausted:
		return connect.CodeResourceExhausted
	case codes.FailedPrecondition:
		return connect.CodeFailedPrecondition
	case codes.Aborted:
		return connect.CodeAborted
	case codes.OutOfRange:
		return connect.CodeOutOfRange
	case codes.Unimplemented:
		return connect.CodeUnimplemented
	case codes.Internal:
		return connect.CodeInternal
	case codes.Unavailable:
		return connect.CodeUnavailable
	case codes.DataLoss:
		return connect.CodeDataLoss
	case codes.Unauthenticated:
		return connect.CodeUnauthenticated
	default:
		return connect.CodeUnknown
	}
}

func mapConnectCodeToGRPC(code connect.Code) codes.Code {
	switch code {
	case connect.CodeCanceled:
		return codes.Canceled
	case connect.CodeUnknown:
		return codes.Unknown
	case connect.CodeInvalidArgument:
		return codes.InvalidArgument
	case connect.CodeDeadlineExceeded:
		return codes.DeadlineExceeded
	case connect.CodeNotFound:
		return codes.NotFound
	case connect.CodeAlreadyExists:
		return codes.AlreadyExists
	case connect.CodePermissionDenied:
		return codes.PermissionDenied
	case connect.CodeResourceExhausted:
		return codes.ResourceExhausted
	case connect.CodeFailedPrecondition:
		return codes.FailedPrecondition
	case connect.CodeAborted:
		return codes.Aborted
	case connect.CodeOutOfRange:
		return codes.OutOfRange
	case connect.CodeUnimplemented:
		return codes.Unimplemented
	case connect.CodeInternal:
		return codes.Internal
	case connect.CodeUnavailable:
		return codes.Unavailable
	case connect.CodeDataLoss:
		return codes.DataLoss
	case connect.CodeUnauthenticated:
		return codes.Unauthenticated
	default:
		return codes.Unknown
	}
}
