package connect

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// WebSocket upgrade leg for client-streaming and bidi procedures.
//
// Browsers cannot stream fetch request bodies, so the connect protocol's
// client/bidi legs are unreachable from the browser (connect-web hard-blocks
// them client-side). The connect server therefore also accepts a WebSocket
// upgrade on the SAME procedure path, speaking kratos's WS stream dialect
// (see kratos v3 transport/http/stream.go): every data frame is one WS text
// message holding a protojson-encoded message, the client half-closes with
// the control frame "\x1eend", the server signals errors with a
// "\x1eerror:" frame followed by an abnormal close, and a normal close ends
// the stream. One port serves all four streaming primitives: unary and
// server-streaming over the connect protocol, client/bidi over this dialect.
//
// Middleware parity: this leg mirrors the connect protocol's streaming legs
// exactly — Transport injection, the unary middleware matcher once per RPC
// (union policy, see adapter.go chainUnary), stream middleware per message
// via middlewareStream, and error encoding at the boundary. Non-upgrade
// requests are untouched.

const (
	wsControlPrefix = "\x1e"
	wsControlEnd    = wsControlPrefix + "end"
	wsControlError  = wsControlPrefix + "error:"
)

var wsUpgrader = websocket.Upgrader{}

// withWebSocket dispatches WebSocket upgrades to the WS bridge and everything
// else to the connect handler.
func withWebSocket(next http.Handler, ws http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if websocket.IsWebSocketUpgrade(r) {
			ws(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// wsServerStream implements grpc.ServerStream over a WebSocket connection
// using the kratos WS dialect. Header/trailer metadata is not carried by the
// dialect; those methods satisfy the interface as no-ops.
//
// Concurrency contract (same as kratos's http-binding stream): writes are
// serialized by writeMu so handlers may call Send from any goroutine —
// gorilla panics on concurrent writes. Reads (RecvMsg) must stay on one
// goroutine, the one driving the handler loop.
type wsServerStream struct {
	conn    *websocket.Conn
	ctx     context.Context
	writeMu sync.Mutex
}

func (s *wsServerStream) Context() context.Context     { return s.ctx }
func (s *wsServerStream) SetHeader(metadata.MD) error  { return nil }
func (s *wsServerStream) SendHeader(metadata.MD) error { return nil }
func (s *wsServerStream) SetTrailer(metadata.MD)       {}

func (s *wsServerStream) SendMsg(m any) error {
	data, err := protojson.Marshal(m.(proto.Message))
	if err != nil {
		return err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.conn.WriteMessage(websocket.TextMessage, data)
}

// close ends the dialect: an error is announced with a control frame and an
// abnormal close; a clean return closes normally (the client reads EOF).
func (s *wsServerStream) close(err error) {
	deadline := time.Now().Add(time.Second)
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if err != nil {
		_ = s.conn.WriteMessage(websocket.TextMessage, []byte(wsControlError+err.Error()))
		_ = s.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error()),
			deadline,
		)
		return
	}
	_ = s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		deadline,
	)
}

func (s *wsServerStream) RecvMsg(m any) error {
	for {
		_, data, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return io.EOF
			}
			return err
		}
		text := string(data)
		switch {
		case text == wsControlEnd:
			return io.EOF
		case strings.HasPrefix(text, wsControlError):
			return errors.New(strings.TrimPrefix(text, wsControlError))
		default:
			return protojson.Unmarshal(data, m.(proto.Message))
		}
	}
}

// wsClientStream adapts the stream to the gRPC client-streaming server
// interface (Recv + SendAndClose), mirroring adapter.go's clientStreamBridge.
type wsClientStream[Req, Res any] struct {
	*middlewareStream
}

func (b *wsClientStream[Req, Res]) Recv() (*Req, error) {
	m := new(Req)
	if err := b.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (b *wsClientStream[Req, Res]) SendAndClose(m *Res) error {
	return b.SendMsg(m)
}

// serveClientStreamWS runs a client-streaming handler over an upgraded
// connection.
func serveClientStreamWS[Req, Res any](
	srv *Server,
	procedure string,
	fn func(grpc.ClientStreamingServer[Req, Res]) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		var stream *wsServerStream
		streamErr := runWsCall(srv, r, procedure, func(ctx context.Context) error {
			stream = &wsServerStream{conn: conn, ctx: ctx}
			mw := newMiddlewareStream(ctx, stream, srv.streamMatcher())
			return fn(&wsClientStream[Req, Res]{middlewareStream: mw})
		})
		closeWsCall(conn, stream, streamErr)
	}
}

// serveBidiStreamWS is the bidi counterpart; the middlewareStream satisfies
// both directions directly.
func serveBidiStreamWS[Req, Res any](
	srv *Server,
	procedure string,
	fn func(grpc.BidiStreamingServer[Req, Res]) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		var stream *wsServerStream
		streamErr := runWsCall(srv, r, procedure, func(ctx context.Context) error {
			stream = &wsServerStream{conn: conn, ctx: ctx}
			mw := newMiddlewareStream(ctx, stream, srv.streamMatcher())
			return fn(&grpc.GenericServerStream[Req, Res]{ServerStream: mw})
		})
		closeWsCall(conn, stream, streamErr)
	}
}

func closeWsCall(conn *websocket.Conn, stream *wsServerStream, streamErr error) {
	if stream != nil {
		stream.close(streamErr)
		return
	}
	if streamErr != nil {
		(&wsServerStream{conn: conn}).close(streamErr)
	}
}

// runWsCall injects the kratos Transport, merges the server base context, and
// runs the unary middleware matcher once per stream-RPC — the WS mirror of the
// connect legs' chainUnary (see adapter.go for the union alignment policy).
// The per-request context's cancellation is dropped (streams outlive the HTTP
// exchange, mirroring kratos's detachStreamContext) while its values are kept
// and passed to the handler via the callback. The per-request timeout does not
// apply: streams are long-lived by design.
func runWsCall(srv *Server, r *http.Request, procedure string, fn func(context.Context) error) error {
	ctx, cancel := Merge(context.WithoutCancel(r.Context()), srv.baseCtx)
	defer cancel()
	srv.mu.RLock()
	ep := ""
	if srv.endpoint != nil {
		ep = srv.endpoint.String()
	}
	srv.mu.RUnlock()
	tr := &Transport{
		endpoint:    ep,
		operation:   procedure,
		reqHeader:   newHeader(r.Header),
		replyHeader: newHeader(http.Header{}),
		httpMethod:  http.MethodGet,
		remoteAddr:  r.RemoteAddr,
	}
	ctx = transport.NewServerContext(ctx, tr)
	if m := srv.middleware.Match(tr.Operation()); len(m) > 0 {
		h := func(ctx context.Context, _ any) (any, error) {
			return nil, fn(ctx)
		}
		if _, err := (middleware.Chain(m...))(h)(ctx, nil); err != nil {
			return err
		}
		return nil
	}
	if err := fn(ctx); err != nil {
		return err
	}
	return nil
}
