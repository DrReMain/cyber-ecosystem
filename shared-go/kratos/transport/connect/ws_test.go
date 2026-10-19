package connect_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	kratosmiddleware "github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	connect "cyber-ecosystem/shared-go/kratos/transport/connect"
	testpb "cyber-ecosystem/shared-go/kratos/transport/connect/testpb"
)

// The kratos WS stream dialect control frames (see ws.go for the wire spec).
const (
	wsControlEnd   = "\x1eend"
	wsControlError = "\x1eerror:"
)

// startWsTestServer brings up a connect Server with WS-bridged handlers behind
// an httptest server. Non-browser dialers send no Origin header, so the
// default same-origin upgrade check passes.
func startWsTestServer(t *testing.T, register func(*connect.Server)) *httptest.Server {
	t.Helper()
	srv := connect.NewServer()
	register(srv)
	ts := httptest.NewServer(http.HandlerFunc(srv.ServeHTTP))
	t.Cleanup(ts.Close)
	return ts
}

func wsDial(t *testing.T, ts *httptest.Server, procedure string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http")+procedure, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", procedure, err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func wsWriteProto(t *testing.T, conn *websocket.Conn, msg proto.Message) {
	t.Helper()
	frame, err := protojson.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal %T: %v", msg, err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, frame); err != nil {
		t.Fatalf("write frame: %v", err)
	}
}

// A middleware rejection must be announced over the dialect — control-error
// frame carrying the reason, then an abnormal close — not surface as a cold
// TCP drop the client cannot distinguish from a network failure.
func TestWsMiddlewareRejectAnnounced(t *testing.T) {
	const procedure = "/cyber.test.WsService/Pipe"
	srv := connect.NewServer()
	srv.Use(procedure, func(kratosmiddleware.Handler) kratosmiddleware.Handler {
		return func(context.Context, any) (any, error) {
			return nil, kratoserrors.Unauthorized("MISSING_SESSION", "")
		}
	})
	connect.HandleBidiStream(srv, procedure, func(grpc.BidiStreamingServer[testpb.PipeRequest, testpb.PipeResponse]) error {
		t.Error("handler must not run when middleware rejects")
		return nil
	})
	ts := httptest.NewServer(http.HandlerFunc(srv.ServeHTTP))
	t.Cleanup(ts.Close)

	conn := wsDial(t, ts, procedure)
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected control-error frame, got %v", err)
	}
	if text := string(data); !strings.HasPrefix(text, wsControlError) || !strings.Contains(text, "MISSING_SESSION") {
		t.Fatalf("frame = %q, want %s-prefixed carrying MISSING_SESSION", text, wsControlError)
	}
	if _, _, err := conn.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseInternalServerErr) {
		t.Fatalf("expected close 1011, got %v", err)
	}
}

func TestWsClientStreamEcho(t *testing.T) {
	const procedure = "/cyber.test.WsService/Echo"
	ts := startWsTestServer(t, func(srv *connect.Server) {
		connect.HandleClientStream(srv, procedure, func(stream grpc.ClientStreamingServer[testpb.EchoRequest, testpb.EchoResponse]) error {
			// Transport injection parity with the connect streaming legs.
			if tr, ok := transport.FromServerContext(stream.Context()); !ok || tr.Operation() != procedure {
				t.Errorf("operation missing or mismatched (ok=%v)", ok)
			}
			var (
				n    int32
				last int64
			)
			for {
				req, err := stream.Recv()
				if err != nil {
					break
				}
				n++
				last = req.GetSequence()
			}
			return stream.SendAndClose(&testpb.EchoResponse{TotalMessages: n, LastSequence: last})
		})
	})

	conn := wsDial(t, ts, procedure)
	for i := int64(1); i <= 3; i++ {
		wsWriteProto(t, conn, &testpb.EchoRequest{Data: []byte("x"), Sequence: i})
	}
	if err := conn.WriteMessage(websocket.TextMessage, []byte(wsControlEnd)); err != nil {
		t.Fatal(err)
	}

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	var res testpb.EchoResponse
	if err := protojson.Unmarshal(data, &res); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if res.GetTotalMessages() != 3 || res.GetLastSequence() != 3 {
		t.Fatalf("summary = %d msgs / last %d, want 3 / 3", res.GetTotalMessages(), res.GetLastSequence())
	}

	if _, _, err := conn.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		t.Fatalf("expected normal close, got %v", err)
	}
}

func TestWsBidiPipe(t *testing.T) {
	const procedure = "/cyber.test.WsService/Pipe"
	ts := startWsTestServer(t, func(srv *connect.Server) {
		connect.HandleBidiStream(srv, procedure, func(stream grpc.BidiStreamingServer[testpb.PipeRequest, testpb.PipeResponse]) error {
			for {
				req, err := stream.Recv()
				if err != nil {
					//nolint:nilerr // any Recv error ends the echo stream; close is the success path
					return nil
				}
				if err := stream.Send(&testpb.PipeResponse{Data: req.GetData(), Type: req.GetType(), Sequence: req.GetSequence()}); err != nil {
					return err
				}
			}
		})
	})

	conn := wsDial(t, ts, procedure)
	wsWriteProto(t, conn, &testpb.PipeRequest{Data: []byte("ping"), Type: "chat", Sequence: 1})

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read echo: %v", err)
	}
	var res testpb.PipeResponse
	if err := protojson.Unmarshal(data, &res); err != nil {
		t.Fatalf("unmarshal echo: %v", err)
	}
	if string(res.GetData()) != "ping" || res.GetSequence() != 1 {
		t.Fatalf("echo = %q / %d, want ping / 1", res.GetData(), res.GetSequence())
	}

	// Half-close: the server sees EOF and closes normally.
	if err := conn.WriteMessage(websocket.TextMessage, []byte(wsControlEnd)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := conn.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		t.Fatalf("expected normal close, got %v", err)
	}
}

// Handlers may Send from multiple goroutines (legal for gRPC-style stream
// interfaces); writes are serialized by wsServerStream.writeMu, mirroring
// kratos's http-binding stream. Without the mutex gorilla panics on the
// concurrent write.
func TestWsConcurrentSend(t *testing.T) {
	const procedure = "/cyber.test.WsService/Fanout"
	const senders = 8
	const perSender = 20
	ts := startWsTestServer(t, func(srv *connect.Server) {
		connect.HandleBidiStream(srv, procedure, func(stream grpc.BidiStreamingServer[testpb.PipeRequest, testpb.PipeResponse]) error {
			var wg sync.WaitGroup
			for i := range senders {
				wg.Add(1)
				go func(n int) {
					defer wg.Done()
					for range perSender {
						if err := stream.Send(&testpb.PipeResponse{Data: []byte("x"), Sequence: int64(n)}); err != nil {
							return
						}
					}
				}(i)
			}
			wg.Wait()
			// drain the client's half-close so the RPC ends cleanly
			for {
				if _, err := stream.Recv(); err != nil {
					//nolint:nilerr // drain ends at the half-close; EOF here is the success path
					return nil
				}
			}
		})
	})

	conn := wsDial(t, ts, procedure)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(wsControlEnd)); err != nil {
		t.Fatal(err)
	}
	received := 0
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				t.Fatalf("read: %v", err)
			}
			break
		}
		received++
	}
	if want := senders * perSender; received != want {
		t.Fatalf("received %d frames, want %d", received, want)
	}
}

func TestWsErrorFrame(t *testing.T) {
	const procedure = "/cyber.test.WsService/Fail"
	ts := startWsTestServer(t, func(srv *connect.Server) {
		connect.HandleBidiStream(srv, procedure, func(grpc.BidiStreamingServer[testpb.PipeRequest, testpb.PipeResponse]) error {
			return errWsTestFailure
		})
	})

	conn := wsDial(t, ts, procedure)
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read error frame: %v", err)
	}
	if !strings.HasPrefix(string(data), wsControlError) {
		t.Fatalf("frame %q lacks error prefix", data)
	}
	if _, _, err := conn.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseInternalServerErr) {
		t.Fatalf("expected abnormal close, got %v", err)
	}
}

// The unary middleware matcher must fire exactly once per stream-RPC on the
// WS leg (union alignment: http generators run ctx.Middleware per-RPC), not
// per message, with the Transport readable inside the middleware.
func TestWsUnaryMiddlewarePerRPC(t *testing.T) {
	const procedure = "/cyber.test.WsService/Echo"
	var invocations int
	var sawOperation bool
	ts := startWsTestServer(t, func(srv *connect.Server) {
		srv.Use("*", func(next kratosmiddleware.Handler) kratosmiddleware.Handler {
			return func(ctx context.Context, req any) (any, error) {
				invocations++
				if req != nil {
					t.Errorf("client-stream req = %T, want nil (http generator parity)", req)
				}
				if tr, ok := transport.FromServerContext(ctx); ok && tr.Operation() == procedure {
					sawOperation = true
				}
				return next(ctx, req)
			}
		})
		connect.HandleClientStream(srv, procedure, func(stream grpc.ClientStreamingServer[testpb.EchoRequest, testpb.EchoResponse]) error {
			var n int32
			for {
				if _, err := stream.Recv(); err != nil {
					break
				}
				n++
			}
			return stream.SendAndClose(&testpb.EchoResponse{TotalMessages: n})
		})
	})

	conn := wsDial(t, ts, procedure)
	for i := int64(1); i <= 3; i++ {
		wsWriteProto(t, conn, &testpb.EchoRequest{Data: []byte("x"), Sequence: i})
	}
	if err := conn.WriteMessage(websocket.TextMessage, []byte(wsControlEnd)); err != nil {
		t.Fatal(err)
	}
	// summary frame first, then the normal close
	if _, _, err := conn.ReadMessage(); err != nil {
		t.Fatalf("read summary: %v", err)
	}
	if _, _, err := conn.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		t.Fatalf("expected normal close, got %v", err)
	}

	if invocations != 1 {
		t.Fatalf("middleware invocations = %d, want exactly 1 (per-RPC, not per-message)", invocations)
	}
	if !sawOperation {
		t.Error("middleware could not read the Transport operation from ctx")
	}
}

var errWsTestFailure = grpcErrorf("ws test: deliberate failure")

func grpcErrorf(msg string) error { return &wsTestError{msg: msg} }

type wsTestError struct{ msg string }

func (e *wsTestError) Error() string { return e.msg }
