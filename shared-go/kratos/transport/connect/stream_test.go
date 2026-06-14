package connect

import (
	"context"
	stderrors "errors"
	"io"
	"net/http"
	"sync/atomic"
	"testing"

	connectrpc "connectrpc.com/connect"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"cyber-ecosystem/shared-go/kratos/transport/connect/internal/matcher"
)

type fakeServerStream struct {
	grpc.ServerStream
}

func (fakeServerStream) Context() context.Context { return context.Background() }
func (fakeServerStream) SendMsg(any) error        { return nil }
func (fakeServerStream) RecvMsg(any) error        { return nil }

func TestMiddlewareStreamPerMessageSend(t *testing.T) {
	var count int32
	mw := func(h middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			atomic.AddInt32(&count, 1)
			return h(ctx, req)
		}
	}
	m := matcher.New()
	m.Add("/svc.X/Method", mw)

	ctx := transport.NewServerContext(context.Background(), &Transport{operation: "/svc.X/Method"})
	w := newMiddlewareStream(ctx, fakeServerStream{}, m)

	for i := 0; i < 3; i++ {
		if err := w.SendMsg(struct{}{}); err != nil {
			t.Fatalf("SendMsg: %v", err)
		}
	}
	if got := atomic.LoadInt32(&count); got != 3 {
		t.Fatalf("per-message send invocations = %d, want 3", got)
	}
}

func TestMiddlewareStreamPerMessageRecv(t *testing.T) {
	var count int32
	mw := func(h middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			atomic.AddInt32(&count, 1)
			return h(ctx, req)
		}
	}
	m := matcher.New()
	m.Add("/svc.X/Method", mw)

	ctx := transport.NewServerContext(context.Background(), &Transport{operation: "/svc.X/Method"})
	w := newMiddlewareStream(ctx, fakeServerStream{}, m)

	for i := 0; i < 2; i++ {
		if err := w.RecvMsg(struct{}{}); err != nil {
			t.Fatalf("RecvMsg: %v", err)
		}
	}
	if got := atomic.LoadInt32(&count); got != 2 {
		t.Fatalf("per-message recv invocations = %d, want 2", got)
	}
}

func TestMiddlewareStreamShortCircuitsWhenNoMatch(t *testing.T) {
	m := matcher.New() // no middleware registered
	ctx := transport.NewServerContext(context.Background(), &Transport{operation: "/svc.X/Method"})
	w := newMiddlewareStream(ctx, fakeServerStream{}, m)
	if err := w.SendMsg(struct{}{}); err != nil {
		t.Fatalf("SendMsg with no mw: %v", err)
	}
}

func TestMiddlewareStreamPassesContextThrough(t *testing.T) {
	ctx := transport.NewServerContext(context.Background(), &Transport{operation: "/svc.X/Method"})
	w := newMiddlewareStream(ctx, fakeServerStream{}, matcher.New())
	if w.Context() != ctx {
		t.Fatal("Context() must return the wrapped ctx (carries Transport)")
	}
}

func TestMiddlewareStreamPropagatesMiddlewareError(t *testing.T) {
	sentinel := stderrors.New("boom")
	mwErr := func(h middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			return nil, sentinel
		}
	}
	m := matcher.New()
	m.Add("/svc.X/Method", mwErr)

	ctx := transport.NewServerContext(context.Background(), &Transport{operation: "/svc.X/Method"})
	w := newMiddlewareStream(ctx, fakeServerStream{}, m)

	if err := w.SendMsg(struct{}{}); err != sentinel {
		t.Fatalf("SendMsg error = %v, want sentinel", err)
	}
	if err := w.RecvMsg(struct{}{}); err != sentinel {
		t.Fatalf("RecvMsg error = %v, want sentinel", err)
	}
}

// fakeHandlerConn is a minimal connect.StreamingHandlerConn whose
// ResponseHeader/ResponseTrailer back real, inspectable http.Header values.
type fakeHandlerConn struct {
	connectrpc.StreamingHandlerConn
	respHeader http.Header
	respTrail  http.Header
}

func (fakeHandlerConn) Spec() connectrpc.Spec          { return connectrpc.Spec{} }
func (fakeHandlerConn) Peer() connectrpc.Peer          { return connectrpc.Peer{} }
func (fakeHandlerConn) Receive(any) error              { return io.EOF }
func (fakeHandlerConn) Send(any) error                 { return nil }
func (f fakeHandlerConn) ResponseHeader() http.Header  { return f.respHeader }
func (f fakeHandlerConn) ResponseTrailer() http.Header { return f.respTrail }

// TestServerStreamSendHeaderFlushesImmediately proves that SendHeader writes to
// the conn's response headers eagerly (before any SendMsg), whereas SetHeader
// only buffers and leaves the conn header untouched.
func TestServerStreamSendHeaderFlushesImmediately(t *testing.T) {
	// SetHeader only buffers: conn response header must stay empty.
	connBuffered := fakeHandlerConn{respHeader: http.Header{}, respTrail: http.Header{}}
	ssBuffered := newServerStream(context.Background(), connBuffered)
	if err := ssBuffered.SetHeader(metadata.Pairs("x-custom", "buffered")); err != nil {
		t.Fatalf("SetHeader: %v", err)
	}
	if got := connBuffered.ResponseHeader().Get("X-Custom"); got != "" {
		t.Fatalf("SetHeader wrote to conn prematurely: X-Custom = %q, want empty", got)
	}

	// SendHeader flushes immediately: the value appears on conn.ResponseHeader()
	// without any SendMsg call.
	connFlushed := fakeHandlerConn{respHeader: http.Header{}, respTrail: http.Header{}}
	ssFlushed := newServerStream(context.Background(), connFlushed)
	if err := ssFlushed.SendHeader(metadata.Pairs("x-custom", "hv")); err != nil {
		t.Fatalf("SendHeader: %v", err)
	}
	if got := connFlushed.ResponseHeader().Get("X-Custom"); got != "hv" {
		t.Fatalf("SendHeader did not flush eagerly: X-Custom = %q, want hv", got)
	}
}
