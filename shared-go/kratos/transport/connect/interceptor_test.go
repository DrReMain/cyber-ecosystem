package connect

import (
	"context"
	"testing"

	connectrpc "connectrpc.com/connect"
	"github.com/go-kratos/kratos/v3/transport"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestWrapUnaryInjectsTransportOutermost(t *testing.T) {
	// Use NewServer so middleware/errorEncoder/etc. are initialized (a bare
	// &Server{} would nil-deref on middleware.Match).
	srv := NewServer()
	ki := newKratosInterceptor(srv).(*kratosInterceptor)

	var sawKind string
	next := func(ctx context.Context, _ connectrpc.AnyRequest) (connectrpc.AnyResponse, error) {
		if tr, ok := transport.FromServerContext(ctx); ok {
			sawKind = string(tr.Kind())
		}
		// Return a real response so the interceptor's happy path runs cleanly.
		return connectrpc.NewResponse(&emptypb.Empty{}), nil
	}
	wrapped := ki.WrapUnary(next)
	req := connectrpc.NewRequest(&emptypb.Empty{})
	if _, err := wrapped(context.Background(), req); err != nil {
		t.Fatalf("unexpected error from wrapped handler: %v", err)
	}
	if sawKind != "connect" {
		t.Fatalf("inner handler saw kind %q, want connect (kratos must inject Transport outermost)", sawKind)
	}
}

func TestHandlerOptionsKratosFirst(t *testing.T) {
	srv := NewServer()
	_ = srv.HandlerOptions() // smoke: builds without panic
}
