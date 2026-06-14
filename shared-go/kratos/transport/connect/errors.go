package connect

import (
	"context"
	stderrors "errors"

	connectrpc "connectrpc.com/connect"
	"github.com/go-kratos/kratos/v3/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorToConnect converts a Kratos error into a Connect error.
// A nil error returns nil. An existing *connect.Error is passed through unchanged.
//
// Context-error parity: a bare (or wrapped) context.DeadlineExceeded /
// context.Canceled is mapped to connect.CodeDeadlineExceeded /
// connect.CodeCanceled instead of going through kerrors.FromError. Kratos's
// FromError does not special-case stdlib context errors — it would otherwise
// map them to HTTP 500 / codes.Internal. gRPC-go's server machinery, by
// contrast, maps a bare context error via status.FromContextError to
// codes.DeadlineExceeded / codes.Canceled. Without this branch, Connect would
// surface CodeInternal where a kratos gRPC server surfaces DeadlineExceeded —
// a 3-protocol consistency break. See comparative.TestTimeoutMappingGRPCvsConnect.
//
// Ordering matches gRPC-go: an error that carries its OWN gRPC status (kratos
// *errors.Error, or any *status.Status) wins over the context-error shortcut —
// gRPC's status.FromError resolves those before ever consulting
// status.FromContextError. So a kratos NotFound whose cause is a context error
// still surfaces CodeNotFound here, exactly as gRPC surfaces codes.NotFound.
func ErrorToConnect(err error) error {
	if err == nil {
		return nil
	}
	var ce *connectrpc.Error
	if errors.As(err, &ce) {
		return ce
	}
	// An error that carries its own gRPC status must keep that status, matching
	// gRPC-go (status.FromError resolves grpcstatus before FromContextError).
	// Only a context error WITHOUT its own status falls into the shortcut.
	if _, ok := asGRPCStatus(err); !ok {
		if ctxErr, code := contextErrorCode(err); code != connectrpc.CodeUnknown {
			return connectrpc.NewError(code, stderrors.New(ctxErr.Error()))
		}
	}
	ke := errors.FromError(err)
	grpcCode := codes.Unknown
	if gs := ke.GRPCStatus(); gs != nil {
		grpcCode = gs.Code()
	}
	out := connectrpc.NewError(grpcToConnect(grpcCode), stderrors.New(ke.Message))
	if ke.Reason != "" || len(ke.Metadata) > 0 {
		info := &errdetails.ErrorInfo{Reason: ke.Reason, Metadata: ke.Metadata}
		if d, e := connectrpc.NewErrorDetail(info); e == nil {
			out.AddDetail(d)
		}
	}
	return out
}

// asGRPCStatus reports whether err (or anything in its chain) implements the
// gRPC status interface (GRPCStatus() *status.Status), mirroring the
// grpcstatus check inside gRPC-go's status.FromError. A nil/empty status does
// NOT count (gRPC-go treats a nil GRPCStatus as Unknown, falling through to
// FromContextError for context errors).
func asGRPCStatus(err error) (*status.Status, bool) {
	type grpcstatus interface{ GRPCStatus() *status.Status }
	var gs grpcstatus
	if errors.As(err, &gs) {
		if s := gs.GRPCStatus(); s != nil {
			return s, true
		}
	}
	return nil, false
}

// contextErrorCode inspects err for a stdlib context error and returns the
// matching Connect code. A non-context error returns CodeUnknown so the caller
// falls through to the kratos FromError path. errors.Is is used so WRAPPED
// context errors (e.g. fmt.Errorf("...: %w", context.DeadlineExceeded)) map
// the same way, mirroring gRPC-go's status.FromContextError behavior.
func contextErrorCode(err error) (contextError error, code connectrpc.Code) {
	if stderrors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded, connectrpc.CodeDeadlineExceeded
	}
	if stderrors.Is(err, context.Canceled) {
		return context.Canceled, connectrpc.CodeCanceled
	}
	return err, connectrpc.CodeUnknown
}

// ConnectToError converts a Connect error into a Kratos error.
// A nil error returns nil. An existing *errors.Error is passed through.
func ConnectToError(err error) *errors.Error {
	if err == nil {
		return nil
	}
	var ke *errors.Error
	if errors.As(err, &ke) {
		return ke
	}
	var ce *connectrpc.Error
	if !errors.As(err, &ce) {
		return errors.FromError(err)
	}
	grpcCode := connectToGRPC(ce.Code())
	reason := ""
	var md map[string]string
	for _, d := range ce.Details() {
		if v, e := d.Value(); e == nil {
			if info, ok := v.(*errdetails.ErrorInfo); ok {
				reason = info.Reason
				md = info.Metadata
				break
			}
		}
	}
	out := errors.New(grpcToHTTP(grpcCode), reason, ce.Message())
	if len(md) > 0 {
		out.Metadata = md
	}
	return out
}

// grpcToConnect maps a gRPC code to the equivalent Connect code (1:1 by value).
func grpcToConnect(c codes.Code) connectrpc.Code {
	switch c {
	case codes.Canceled:
		return connectrpc.CodeCanceled
	case codes.Unknown:
		return connectrpc.CodeUnknown
	case codes.InvalidArgument:
		return connectrpc.CodeInvalidArgument
	case codes.DeadlineExceeded:
		return connectrpc.CodeDeadlineExceeded
	case codes.NotFound:
		return connectrpc.CodeNotFound
	case codes.AlreadyExists:
		return connectrpc.CodeAlreadyExists
	case codes.PermissionDenied:
		return connectrpc.CodePermissionDenied
	case codes.ResourceExhausted:
		return connectrpc.CodeResourceExhausted
	case codes.FailedPrecondition:
		return connectrpc.CodeFailedPrecondition
	case codes.Aborted:
		return connectrpc.CodeAborted
	case codes.OutOfRange:
		return connectrpc.CodeOutOfRange
	case codes.Unimplemented:
		return connectrpc.CodeUnimplemented
	case codes.Internal:
		return connectrpc.CodeInternal
	case codes.Unavailable:
		return connectrpc.CodeUnavailable
	case codes.DataLoss:
		return connectrpc.CodeDataLoss
	case codes.Unauthenticated:
		return connectrpc.CodeUnauthenticated
	default:
		return connectrpc.CodeUnknown
	}
}

// connectToGRPC maps a Connect code to the equivalent gRPC code (1:1 by value).
func connectToGRPC(c connectrpc.Code) codes.Code {
	switch c {
	case connectrpc.CodeCanceled:
		return codes.Canceled
	case connectrpc.CodeUnknown:
		return codes.Unknown
	case connectrpc.CodeInvalidArgument:
		return codes.InvalidArgument
	case connectrpc.CodeDeadlineExceeded:
		return codes.DeadlineExceeded
	case connectrpc.CodeNotFound:
		return codes.NotFound
	case connectrpc.CodeAlreadyExists:
		return codes.AlreadyExists
	case connectrpc.CodePermissionDenied:
		return codes.PermissionDenied
	case connectrpc.CodeResourceExhausted:
		return codes.ResourceExhausted
	case connectrpc.CodeFailedPrecondition:
		return codes.FailedPrecondition
	case connectrpc.CodeAborted:
		return codes.Aborted
	case connectrpc.CodeOutOfRange:
		return codes.OutOfRange
	case connectrpc.CodeUnimplemented:
		return codes.Unimplemented
	case connectrpc.CodeInternal:
		return codes.Internal
	case connectrpc.CodeUnavailable:
		return codes.Unavailable
	case connectrpc.CodeDataLoss:
		return codes.DataLoss
	case connectrpc.CodeUnauthenticated:
		return codes.Unauthenticated
	default:
		return codes.Unknown
	}
}

// grpcToHTTP maps a gRPC code to the HTTP status code Kratos uses.
// Explicit map (kratos does not export a public helper); matches Kratos's internal mapping.
func grpcToHTTP(c codes.Code) int {
	switch c {
	case codes.OK:
		return 200
	case codes.Canceled:
		return 499
	case codes.InvalidArgument, codes.OutOfRange:
		return 400
	case codes.FailedPrecondition:
		return 412
	case codes.Unauthenticated:
		return 401
	case codes.PermissionDenied:
		return 403
	case codes.NotFound:
		return 404
	case codes.AlreadyExists, codes.Aborted:
		return 409
	case codes.ResourceExhausted:
		return 429
	case codes.Unimplemented:
		return 501
	case codes.Unavailable:
		return 503
	case codes.DeadlineExceeded:
		return 504
	default:
		return 500
	}
}
