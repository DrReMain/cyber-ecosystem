package connect

import (
	"testing"

	connectrpc "connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	kerrors "github.com/go-kratos/kratos/v2/errors"
)

func TestErrorToConnect_Code(t *testing.T) {
	err := kerrors.BadRequest("INVALID", "bad request")
	got := ErrorToConnect(err)
	var ce *connectrpc.Error
	if !kerrors.As(got, &ce) {
		t.Fatalf("expected *connect.Error, got %T", got)
	}
	if ce.Code() != connectrpc.CodeInvalidArgument {
		t.Fatalf("expected CodeInvalidArgument, got %v", ce.Code())
	}
}

func TestErrorToConnect_CleanMessage(t *testing.T) {
	err := kerrors.NotFound("NOT_FOUND", "record not found")
	got := ErrorToConnect(err)
	var ce *connectrpc.Error
	if !kerrors.As(got, &ce) {
		t.Fatalf("expected *connect.Error, got %T", got)
	}
	if ce.Message() != "record not found" {
		t.Fatalf("expected clean message, got: %q", ce.Message())
	}
}

func TestErrorToConnect_ErrorInfoDetail(t *testing.T) {
	err := kerrors.BadRequest("ERROR_REASON_VALIDATOR", "validation error: id too short")
	got := ErrorToConnect(err)
	var ce *connectrpc.Error
	if !kerrors.As(got, &ce) {
		t.Fatalf("expected *connect.Error, got %T", got)
	}

	details := ce.Details()
	if len(details) == 0 {
		t.Fatal("expected at least one detail, got none")
	}

	msg, decodeErr := details[0].Value()
	if decodeErr != nil {
		t.Fatalf("failed to decode detail: %v", decodeErr)
	}
	info, ok := msg.(*errdetails.ErrorInfo)
	if !ok {
		t.Fatalf("expected *errdetails.ErrorInfo, got %T", msg)
	}
	if info.Reason != "ERROR_REASON_VALIDATOR" {
		t.Fatalf("expected reason ERROR_REASON_VALIDATOR, got %q", info.Reason)
	}
}

func TestErrorToConnect_PassThrough(t *testing.T) {
	original := connectrpc.NewError(connectrpc.CodeNotFound, nil)
	got := ErrorToConnect(original)
	if got != original {
		t.Fatal("expected connect.Error to pass through unchanged")
	}
}

func TestConnectToError_Basic(t *testing.T) {
	ce := connectrpc.NewError(connectrpc.CodeUnauthenticated, nil)
	got := ConnectToError(ce)
	if got == nil {
		t.Fatal("expected non-nil Kratos error")
	}
	if got.Code != 401 {
		t.Fatalf("expected HTTP 401, got %d", got.Code)
	}
}

func TestConnectToError_RoundTrip(t *testing.T) {
	original := kerrors.New(404, "MY_REASON", "not found")
	original.Metadata = map[string]string{"field": "id"}

	converted := ErrorToConnect(original)
	restored := ConnectToError(converted)

	if restored.Code != 404 {
		t.Fatalf("expected HTTP 404, got %d", restored.Code)
	}
	if restored.Reason != "MY_REASON" {
		t.Fatalf("expected reason MY_REASON, got %q", restored.Reason)
	}
	if restored.Message != "not found" {
		t.Fatalf("expected message 'not found', got %q", restored.Message)
	}
	if restored.Metadata["field"] != "id" {
		t.Fatalf("expected metadata field=id, got %v", restored.Metadata)
	}
}

func TestStatusToConnect(t *testing.T) {
	if StatusToConnect(nil) != nil {
		t.Fatal("nil status should return nil")
	}
	if StatusToConnect(status.New(codes.OK, "ok")) != nil {
		t.Fatal("OK status should return nil")
	}

	s := status.New(codes.NotFound, "not found")
	got := StatusToConnect(s)
	var ce *connectrpc.Error
	if !kerrors.As(got, &ce) {
		t.Fatalf("expected *connect.Error, got %T", got)
	}
	if ce.Code() != connectrpc.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", ce.Code())
	}
	if ce.Message() != "not found" {
		t.Fatalf("expected clean message, got %q", ce.Message())
	}
}
