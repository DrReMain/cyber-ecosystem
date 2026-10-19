package connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"cyber-ecosystem/shared-go/kratos/transport/connect/internal/matcher"
)

// Server ----------------------------------------------------------------------------------------------------------------

var _ grpc.ServerStream = (*serverStream)(nil)

// serverStream wraps a Connect StreamingHandlerConn to implement grpc.ServerStream.
// It is used together with grpc.GenericServerStream[Req, Res] to satisfy all three
// gRPC streaming server interfaces (ServerStreamingServer, ClientStreamingServer,
// BidiStreamingServer) without per-method adapter code.
type serverStream struct {
	conn       connect.StreamingHandlerConn
	ctx        context.Context
	header     metadata.MD
	trailer    metadata.MD
	sentHeader bool
}

func newServerStream(ctx context.Context, conn connect.StreamingHandlerConn) *serverStream {
	return &serverStream{
		conn:    conn,
		ctx:     ctx,
		header:  make(metadata.MD),
		trailer: make(metadata.MD),
	}
}

func (s *serverStream) Context() context.Context { return s.ctx }

func (s *serverStream) SetHeader(md metadata.MD) error {
	s.header = metadata.Join(s.header, md)
	return nil
}

func (s *serverStream) SendHeader(md metadata.MD) error {
	_ = s.SetHeader(md)
	s.flushHeader()
	return nil
}

func (s *serverStream) SetTrailer(md metadata.MD) {
	s.trailer = metadata.Join(s.trailer, md)
}

func (s *serverStream) SendMsg(m any) error {
	s.flushHeader()
	return s.conn.Send(m)
}

func (s *serverStream) RecvMsg(m any) error {
	return s.conn.Receive(m)
}

// flushHeader copies buffered headers to the Connect response headers on first call.
// Subsequent calls are no-ops.
func (s *serverStream) flushHeader() {
	if s.sentHeader {
		return
	}
	s.sentHeader = true
	h := s.conn.ResponseHeader()
	for k, vals := range s.header {
		for _, v := range vals {
			h.Add(k, v)
		}
	}
}

// flushTrailer copies buffered trailers to the Connect response trailers.
func (s *serverStream) flushTrailer() {
	t := s.conn.ResponseTrailer()
	for k, vals := range s.trailer {
		for _, v := range vals {
			t.Add(k, v)
		}
	}
}

// Compile-time assertions for stream wrappers -----------------------------------------------

var (
	_ grpc.ServerStreamingServer[struct{}]           = (*grpc.GenericServerStream[struct{}, struct{}])(nil)
	_ grpc.ClientStreamingServer[struct{}, struct{}] = (*clientStreamBridge[struct{}, struct{}])(nil)
	_ grpc.BidiStreamingServer[struct{}, struct{}]   = (*grpc.GenericServerStream[struct{}, struct{}])(nil)
	_ grpc.ServerStream                              = (*middlewareStream)(nil)
)

// middlewareStream wraps a grpc.ServerStream and applies the matched stream
// middleware on EVERY SendMsg/RecvMsg, mirroring gRPC's wrappedStream. When no
// middleware matches the operation, it short-circuits to the delegate.
type middlewareStream struct {
	grpc.ServerStream
	ctx context.Context
	m   matcher.Matcher
}

func newMiddlewareStream(ctx context.Context, ss grpc.ServerStream, m matcher.Matcher) *middlewareStream {
	return &middlewareStream{ServerStream: ss, ctx: ctx, m: m}
}

func (w *middlewareStream) Context() context.Context { return w.ctx }

func (w *middlewareStream) SendMsg(msg any) error {
	info, ok := transport.FromServerContext(w.ctx)
	if !ok {
		return w.ServerStream.SendMsg(msg)
	}
	next := w.m.Match(info.Operation())
	if len(next) == 0 {
		return w.ServerStream.SendMsg(msg)
	}
	h := func(ctx context.Context, req any) (any, error) {
		return req, w.ServerStream.SendMsg(msg)
	}
	_, err := middleware.Chain(next...)(h)(w.ctx, msg)
	return err
}

func (w *middlewareStream) RecvMsg(msg any) error {
	info, ok := transport.FromServerContext(w.ctx)
	if !ok {
		return w.ServerStream.RecvMsg(msg)
	}
	next := w.m.Match(info.Operation())
	if len(next) == 0 {
		return w.ServerStream.RecvMsg(msg)
	}
	h := func(ctx context.Context, req any) (any, error) {
		return req, w.ServerStream.RecvMsg(msg)
	}
	_, err := middleware.Chain(next...)(h)(w.ctx, msg)
	return err
}
