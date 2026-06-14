package connect

import (
	"context"
	"errors"
	"fmt"
	"testing"

	connectrpc "connectrpc.com/connect"
	kerrors "github.com/go-kratos/kratos/v3/errors"
)

func TestErrorToConnectPreservesCode(t *testing.T) {
	ke := kerrors.NotFound("USER_NOT_FOUND", "missing")
	err := ErrorToConnect(ke)
	var ce *connectrpc.Error
	if !errors.As(err, &ce) {
		t.Fatalf("not a *connect.Error: %T", err)
	}
	if ce.Code() != connectrpc.CodeNotFound {
		t.Fatalf("code = %v, want NotFound", ce.Code())
	}
}

func TestErrorToConnectNil(t *testing.T) {
	if err := ErrorToConnect(nil); err != nil {
		t.Fatalf("nil in should give nil out, got %v", err)
	}
}

func TestConnectToErrorRoundtrip(t *testing.T) {
	ke := kerrors.BadRequest("BAD_REQ", "x")
	ce := ErrorToConnect(ke)
	back := ConnectToError(ce)
	if back.Reason != "BAD_REQ" {
		t.Fatalf("reason = %q, want BAD_REQ", back.Reason)
	}
	if back.Code != 400 {
		t.Fatalf("http code = %d, want 400", back.Code)
	}
}

func TestConnectToErrorNil(t *testing.T) {
	if got := ConnectToError(nil); got != nil {
		t.Fatalf("nil in should give nil out, got %v", got)
	}
}

func TestErrorToConnectPassesThroughConnectError(t *testing.T) {
	orig := connectrpc.NewError(connectrpc.CodeAlreadyExists, errors.New("dup"))
	out := ErrorToConnect(orig)
	if out != orig {
		t.Fatal("should pass through an existing *connect.Error unchanged")
	}
}

// TestGrpcToHTTPFailedPrecondition verifies that a FailedPrecondition Connect
// error round-trips back to HTTP 412 (http.StatusPreconditionFailed), matching
// Kratos's own gRPC→HTTP mapping. The 400 case must NOT swallow FailedPrecondition.
func TestGrpcToHTTPFailedPrecondition(t *testing.T) {
	ce := connectrpc.NewError(connectrpc.CodeFailedPrecondition, errors.New("precondition"))
	if ce.Code() != connectrpc.CodeFailedPrecondition {
		t.Fatalf("connect code = %v, want FailedPrecondition", ce.Code())
	}
	back := ConnectToError(ce)
	if back.Code != 412 {
		t.Fatalf("http code = %d, want 412", back.Code)
	}
}

// TestErrorToConnectContextDeadlineExceeded verifies that a bare
// context.DeadlineExceeded maps to connect.CodeDeadlineExceeded — matching the
// code a kratos gRPC server surfaces (gRPC-go's status.FromContextError maps
// bare ctx errors to codes.DeadlineExceeded). Without the context-error
// pre-check in ErrorToConnect, kratos FromError would map this to Internal,
// breaking 3-protocol parity.
func TestErrorToConnectContextDeadlineExceeded(t *testing.T) {
	out := ErrorToConnect(context.DeadlineExceeded)
	var ce *connectrpc.Error
	if !errors.As(out, &ce) {
		t.Fatalf("not a *connect.Error: %T", out)
	}
	if ce.Code() != connectrpc.CodeDeadlineExceeded {
		t.Fatalf("code = %v, want CodeDeadlineExceeded", ce.Code())
	}
}

// TestErrorToConnectWrappedDeadlineExceeded verifies the SAME mapping holds for
// a WRAPPED context error (errors.Is chain), since gRPC-go's
// status.FromContextError uses errors.Is too. A handler that returns
// fmt.Errorf("...: %w", context.DeadlineExceeded) must still surface
// DeadlineExceeded, not Internal.
func TestErrorToConnectWrappedDeadlineExceeded(t *testing.T) {
	wrapped := fmt.Errorf("upstream: %w", context.DeadlineExceeded)
	out := ErrorToConnect(wrapped)
	var ce *connectrpc.Error
	if !errors.As(out, &ce) {
		t.Fatalf("not a *connect.Error: %T", out)
	}
	if ce.Code() != connectrpc.CodeDeadlineExceeded {
		t.Fatalf("code = %v, want CodeDeadlineExceeded", ce.Code())
	}
}

// TestErrorToConnectContextCanceled verifies the canceled half of the
// context-error parity (matches gRPC-go's codes.Canceled for context.Canceled).
func TestErrorToConnectContextCanceled(t *testing.T) {
	out := ErrorToConnect(context.Canceled)
	var ce *connectrpc.Error
	if !errors.As(out, &ce) {
		t.Fatalf("not a *connect.Error: %T", out)
	}
	if ce.Code() != connectrpc.CodeCanceled {
		t.Fatalf("code = %v, want CodeCanceled", ce.Code())
	}
}

// TestErrorToConnectContextErrorDoesNotClobberKratosError guards against the
// context-error pre-check swallowing a real kratos error whose CAUSE happens
// to be a context error. A kratos error carries its own (reasonable) code and
// MUST take precedence over the context-error shortcut; only a BARE context
// error (no kratos error wrapping it) should hit the shortcut.
func TestErrorToConnectContextErrorDoesNotClobberKratosError(t *testing.T) {
	// A kratos NotFound whose cause is a context error: ErrorToConnect must
	// still surface CodeNotFound (the kratos code), NOT DeadlineExceeded.
	ke := kerrors.NotFound("DEADLOCK", "locked").WithCause(context.DeadlineExceeded)
	out := ErrorToConnect(ke)
	var ce *connectrpc.Error
	if !errors.As(out, &ce) {
		t.Fatalf("not a *connect.Error: %T", out)
	}
	if ce.Code() != connectrpc.CodeNotFound {
		t.Fatalf("code = %v, want CodeNotFound (kratos error code must win)", ce.Code())
	}
}
