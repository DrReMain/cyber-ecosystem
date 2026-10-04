package connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/go-kratos/kratos/v3/middleware"
	"google.golang.org/grpc"
)

var errMissingSendAndClose = errors.New("connect: client streaming handler returned without calling SendAndClose")

// chainUnary runs the server's unary middleware matcher once per stream-RPC —
// the http-binding half of the union alignment policy: official kratos http
// generators wrap every streaming handler in ctx.Middleware (per-RPC), while
// grpc applies stream middleware per-message. Connect runs BOTH, so it is
// never weaker than either official binding. req mirrors the http generators:
// the bound request for server-streaming (validator sees it), nil otherwise.
// fn receives the middleware-enriched ctx; stream wrappers must be built
// inside it so stream.Context() exposes the injected values (Subject, …).
func (s *Server) chainUnary(ctx context.Context, operation string, req any, fn func(context.Context) error) error {
	m := s.middleware.Match(operation)
	if len(m) == 0 {
		return fn(ctx)
	}
	h := func(mctx context.Context, _ any) (any, error) {
		return nil, fn(mctx)
	}
	_, err := middleware.Chain(m...)(h)(ctx, req)
	return err
}

// Unary adapts a gRPC-style unary method to a Connect handler function.
func Unary[Req, Resp any](fn func(context.Context, *Req) (*Resp, error)) func(context.Context, *connect.Request[Req]) (*connect.Response[Resp], error) {
	return func(ctx context.Context, req *connect.Request[Req]) (*connect.Response[Resp], error) {
		resp, err := fn(ctx, req.Msg)
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(resp), nil
	}
}

// HandleUnary registers a gRPC-style unary method as a Connect handler.
func HandleUnary[Req, Resp any](srv *Server, procedure string, fn func(context.Context, *Req) (*Resp, error)) {
	handler := connect.NewUnaryHandler(procedure, Unary(fn), srv.HandlerOptions()...)
	srv.Register(procedure, handler)
}

// HandleServerStream registers a gRPC-style server-streaming method as a Connect handler.
// The stream is wrapped by middlewareStream so stream middleware runs per-message.
func HandleServerStream[Req, Res any](
	srv *Server,
	procedure string,
	fn func(*Req, grpc.ServerStreamingServer[Res]) error,
) {
	handler := connect.NewServerStreamHandler[Req, Res](
		procedure,
		func(ctx context.Context, req *connect.Request[Req], stream *connect.ServerStream[Res]) error {
			w := newServerStream(ctx, stream.Conn())
			defer w.flushTrailer()
			return srv.chainUnary(ctx, procedure, req.Msg, func(mctx context.Context) error {
				mw := newMiddlewareStream(mctx, w, srv.streamMatcher())
				return fn(req.Msg, &grpc.GenericServerStream[Req, Res]{ServerStream: mw})
			})
		},
		srv.HandlerOptions()...,
	)
	srv.Register(procedure, handler)
}

// HandleClientStream registers a gRPC-style client-streaming method as a Connect handler.
func HandleClientStream[Req, Res any](
	srv *Server,
	procedure string,
	fn func(grpc.ClientStreamingServer[Req, Res]) error,
) {
	handler := connect.NewClientStreamHandler[Req, Res](
		procedure,
		func(ctx context.Context, stream *connect.ClientStream[Req]) (*connect.Response[Res], error) {
			w := newServerStream(ctx, stream.Conn())
			defer w.flushTrailer()
			var closeMsg *Res
			if err := srv.chainUnary(ctx, procedure, nil, func(mctx context.Context) error {
				mw := newMiddlewareStream(mctx, w, srv.streamMatcher())
				bridge := &clientStreamBridge[Req, Res]{middlewareStream: mw, closeMsg: &closeMsg}
				return fn(bridge)
			}); err != nil {
				return nil, err
			}
			if closeMsg == nil {
				return nil, connect.NewError(connect.CodeInternal, errMissingSendAndClose)
			}
			return connect.NewResponse(closeMsg), nil
		},
		srv.HandlerOptions()...,
	)
	srv.Register(procedure, withWebSocket(handler, serveClientStreamWS(srv, procedure, fn)))
}

// clientStreamBridge adapts middlewareStream to grpc.ClientStreamingServer[Req, Res].
// It embeds *middlewareStream (per-message mw on Recv) and captures the SendAndClose
// response for the Connect handler to return.
type clientStreamBridge[Req, Res any] struct {
	*middlewareStream
	closeMsg **Res
}

// Recv reads the next request message (per-message stream middleware applies).
func (b *clientStreamBridge[Req, Res]) Recv() (*Req, error) {
	m := new(Req)
	if err := b.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// SendAndClose captures the response message for the Connect handler to return.
func (b *clientStreamBridge[Req, Res]) SendAndClose(m *Res) error {
	*b.closeMsg = m
	return nil
}

// HandleBidiStream registers a gRPC-style bidirectional-streaming method as a Connect handler.
func HandleBidiStream[Req, Res any](
	srv *Server,
	procedure string,
	fn func(grpc.BidiStreamingServer[Req, Res]) error,
) {
	handler := connect.NewBidiStreamHandler[Req, Res](
		procedure,
		func(ctx context.Context, stream *connect.BidiStream[Req, Res]) error {
			w := newServerStream(ctx, stream.Conn())
			defer w.flushTrailer()
			return srv.chainUnary(ctx, procedure, nil, func(mctx context.Context) error {
				mw := newMiddlewareStream(mctx, w, srv.streamMatcher())
				return fn(&grpc.GenericServerStream[Req, Res]{ServerStream: mw})
			})
		},
		srv.HandlerOptions()...,
	)
	srv.Register(procedure, withWebSocket(handler, serveBidiStreamWS(srv, procedure, fn)))
}
