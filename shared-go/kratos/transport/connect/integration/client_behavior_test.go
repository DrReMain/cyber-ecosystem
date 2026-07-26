// Tests in this file exercise the client-side With* options where they have
// observable behavior (not just field assignment), all driven through the real
// loopback Connect server brought up by startServerWithClient. Coverage:
//   - TestClientMiddlewareRunsUnary   — WithMiddleware runs client middleware
//     on a unary RPC AND the middleware observes the injected client Transport
//     (Kind == "connect", Operation == the procedure).
//   - TestClientMiddlewareRunsStream  — WithStreamMiddleware runs client
//     middleware on a server-stream RPC.
//   - TestClientTimeout               — WithTimeout surfaces a client-side
//     deadline for a slow handler.
//   - TestClientCustomTransport       — WithTransport actually drives the
//     provided http.RoundTripper for outbound calls.
package integration

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	"golang.org/x/net/http2"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	testpb "cyber-ecosystem/shared-go/kratos/transport/connect/testpb"
)

const rawProcedure = "/connecttest.v1.TransferService/Raw"
const subscribeProcedure = "/connecttest.v1.TransferService/Subscribe"

// TestClientMiddlewareRunsUnary verifies that client-side middleware registered
// via connect.WithMiddleware runs for a unary RPC, and that the middleware can
// observe the client Transport injected by unaryClientInterceptor (via
// transport.FromClientContext). This proves client-side observability parity
// with the server: middleware sees Kind == "connect" and Operation == the RPC
// procedure, so tracing/metrics/logging middleware works client-side too.
func TestClientMiddlewareRunsUnary(t *testing.T) {
	var ran atomic.Int32
	var gotKind, gotOperation string

	// countingMW records the Transport it observes, then delegates. Because
	// unaryClientInterceptor injects the Transport via
	// transport.NewClientContext BEFORE chaining middleware, FromClientContext
	// MUST return a non-nil Transport with the procedure's operation.
	countingMW := func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			ran.Add(1)
			if tr, ok := transport.FromClientContext(ctx); ok {
				gotKind = string(tr.Kind())
				gotOperation = tr.Operation()
			}
			return handler(ctx, req)
		}
	}

	cli, stop := startServerWithClient(t, connect.WithMiddleware(countingMW))
	defer stop()

	resp, err := cli.Raw(context.Background(), connectrpc.NewRequest(&testpb.RawRequest{
		ContentType: "text/plain",
		Data:        []byte("hello"),
	}))
	if err != nil {
		t.Fatalf("Raw: %v", err)
	}
	if string(resp.Msg.Data) != "hello" {
		t.Fatalf("data = %q, want %q", string(resp.Msg.Data), "hello")
	}

	if c := ran.Load(); c != 1 {
		t.Fatalf("client middleware ran %d times, want 1", c)
	}
	if gotKind != "connect" {
		t.Fatalf("client Transport Kind = %q, want %q", gotKind, "connect")
	}
	if gotOperation != rawProcedure {
		t.Fatalf("client Transport Operation = %q, want %q", gotOperation, rawProcedure)
	}
}

// TestClientMiddlewareRunsStream verifies that client-side stream middleware
// registered via connect.WithStreamMiddleware runs for a server-streaming RPC.
// streamClientInterceptor chains the stream middleware around the construction
// of the StreamingClientConn, so this asserts the per-RPC (not per-message)
// client middleware hook fires.
func TestClientMiddlewareRunsStream(t *testing.T) {
	var ran atomic.Int32
	var gotKind, gotOperation string

	countingStreamMW := func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			ran.Add(1)
			if tr, ok := transport.FromClientContext(ctx); ok {
				gotKind = string(tr.Kind())
				gotOperation = tr.Operation()
			}
			return handler(ctx, req)
		}
	}

	cli, stop := startServerWithClient(t, connect.WithStreamMiddleware(countingStreamMW))
	defer stop()

	stream, err := cli.Subscribe(context.Background(), connectrpc.NewRequest(&testpb.SubscribeRequest{Topic: "mw"}))
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	var n int
	for stream.Receive() {
		n++
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream error: %v", err)
	}
	if n != 5 {
		t.Fatalf("received %d events, want 5", n)
	}

	if c := ran.Load(); c != 1 {
		t.Fatalf("client stream middleware ran %d times, want 1 (per-RPC)", c)
	}
	if gotKind != "connect" {
		t.Fatalf("client Transport Kind = %q, want %q", gotKind, "connect")
	}
	if gotOperation != subscribeProcedure {
		t.Fatalf("client Transport Operation = %q, want %q", gotOperation, subscribeProcedure)
	}
}

// TestClientTimeout verifies that connect.WithTimeout surfaces a client-side
// deadline. We dial with a 200ms timeout and call the "SLOW" Raw variant whose
// server handler sleeps 1s. The client must error out (it must NOT wait the
// full second). We assert honestly on whatever connect surfaces: a
// *connect.Error with CodeDeadlineExceeded, or a context deadline / url error
// wrapping one — any of these proves the timeout fired client-side.
func TestClientTimeout(t *testing.T) {
	cli, stop := startServerWithClient(t, connect.WithTimeout(200*time.Millisecond))
	defer stop()

	start := time.Now()
	_, err := cli.Raw(context.Background(), connectrpc.NewRequest(&testpb.RawRequest{
		Data: []byte("SLOW"),
	}))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected client-side timeout error, got nil")
	}
	// The client must return well before the handler's 1s sleep completes,
	// proving the deadline fired on the client side rather than the server
	// simply finishing.
	if elapsed > 900*time.Millisecond {
		t.Fatalf("client waited %v for SLOW call, expected to time out well under 1s", elapsed)
	}

	// connect-go surfaces client timeouts as *connect.Error(CodeDeadlineExceeded)
	// (the client interceptor's context.WithTimeout trips and the unary func
	// returns a deadline error that connect wraps). Accept that, or a raw
	// context.DeadlineExceeded / net deadline error as an honest fallback.
	var ce *connectrpc.Error
	if errors.As(err, &ce) {
		if ce.Code() != connectrpc.CodeDeadlineExceeded {
			t.Fatalf("connect code = %v, want %v", ce.Code(), connectrpc.CodeDeadlineExceeded)
		}
		return
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("client timeout did not surface as connect CodeDeadlineExceeded or context.DeadlineExceeded: %T: %v", err, err)
	}
}

// countingRoundTripper wraps an http.RoundTripper and counts RoundTrip calls.
// Used to prove connect.WithTransport actually drives the user-supplied
// transport for outbound requests.
type countingRoundTripper struct {
	next http.RoundTripper
	n    atomic.Int64
}

func (c *countingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	c.n.Add(1)
	return c.next.RoundTrip(req)
}

// TestClientCustomTransport verifies that connect.WithTransport wires the
// provided http.RoundTripper into the outbound path. We wrap a cleartext HTTP/2
// round tripper (mirroring defaultRoundTripper's insecure+h2c branch, since the
// test server is h2c-only) in a counter and assert RoundTrip is invoked for a
// unary call.
func TestClientCustomTransport(t *testing.T) {
	h2cRT := &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	}
	counter := &countingRoundTripper{next: h2cRT}

	cli, stop := startServerWithClient(t, connect.WithTransport(counter))
	defer stop()

	resp, err := cli.Raw(context.Background(), connectrpc.NewRequest(&testpb.RawRequest{
		ContentType: "text/plain",
		Data:        []byte("via-custom-rt"),
	}))
	if err != nil {
		t.Fatalf("Raw via custom transport: %v", err)
	}
	if string(resp.Msg.Data) != "via-custom-rt" {
		t.Fatalf("data = %q, want %q", string(resp.Msg.Data), "via-custom-rt")
	}

	if c := counter.n.Load(); c < 1 {
		t.Fatalf("custom transport RoundTrip called %d times, want >= 1", c)
	}
}
