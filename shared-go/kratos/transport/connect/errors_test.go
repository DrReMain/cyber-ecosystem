package connect

import (
	"testing"

	connectrpc "connectrpc.com/connect"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorToConnect(t *testing.T) {
	err := kerrors.BadRequest("INVALID", "bad request")
	got := ErrorToConnect(err)
	var ce *connectrpc.Error
	if !kerrors.As(got, &ce) {
		t.Fatalf("expect connect error, got %T", got)
	}
	if ce.Code() != connectrpc.CodeInvalidArgument {
		t.Fatalf("code=%v", ce.Code())
	}
}

func TestConnectToError(t *testing.T) {
	err := connectrpc.NewError(connectrpc.CodeUnauthenticated, kerrors.New(401, "UNAUTH", "unauthenticated"))
	got := ConnectToError(err)
	if got == nil {
		t.Fatal("expect kratos error")
	}
	if got.Code != 401 {
		t.Fatalf("code=%d", got.Code)
	}
}

func TestStatusToConnect(t *testing.T) {
	if StatusToConnect(nil) != nil {
		t.Fatal("nil status should return nil")
	}
	if StatusToConnect(status.New(codes.OK, "ok")) != nil {
		t.Fatal("ok status should return nil")
	}

	err := StatusToConnect(status.New(codes.NotFound, "not found"))
	var ce *connectrpc.Error
	if !kerrors.As(err, &ce) {
		t.Fatalf("expect connect error, got %T", err)
	}
	if ce.Code() != connectrpc.CodeNotFound {
		t.Fatalf("code=%v", ce.Code())
	}
}
