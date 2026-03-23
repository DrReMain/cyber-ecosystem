package connect

import (
	"connectrpc.com/connect"
	"github.com/go-kratos/kratos/v2/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorToConnect converts Kratos/gRPC error to Connect error.
func ErrorToConnect(err error) error {
	if err == nil {
		return nil
	}

	// If already a connect.Error, return as is
	var ce *connect.Error
	if errors.As(err, &ce) {
		return ce
	}

	// Convert Kratos error
	ke := errors.FromError(err)

	// Get gRPC code from Kratos error via GRPCStatus
	var grpcCode codes.Code
	if gs := ke.GRPCStatus(); gs != nil {
		grpcCode = gs.Code()
	} else {
		grpcCode = codes.Unknown
	}

	// Map gRPC code to Connect code
	code := mapGRPCCodeToConnect(grpcCode)

	// Create Connect error with the message
	return connect.NewError(code, errors.New(int(ke.Code), ke.Reason, ke.Message))
}

// ConnectToError converts Connect error to Kratos error.
func ConnectToError(err error) *errors.Error {
	if err == nil {
		return nil
	}

	// If already a Kratos error, return as is
	var ke *errors.Error
	if errors.As(err, &ke) {
		return ke
	}

	// If Connect error, convert
	var ce *connect.Error
	if errors.As(err, &ce) {
		grpcCode := mapConnectCodeToGRPC(ce.Code())
		// Map gRPC code to HTTP code for Kratos error
		httpCode := mapGRPCCodeToHTTP(grpcCode)
		return errors.New(httpCode, "CONNECT_ERROR", ce.Message())
	}

	// Unknown error
	return errors.FromError(err)
}

// mapGRPCCodeToHTTP maps gRPC codes to HTTP status codes.
// This matches Kratos behavior.
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

// StatusToConnect converts a gRPC status to Connect error.
func StatusToConnect(s *status.Status) error {
	if s == nil || s.Code() == codes.OK {
		return nil
	}
	return connect.NewError(mapGRPCCodeToConnect(s.Code()), errors.New(500, s.Code().String(), s.Message()))
}
