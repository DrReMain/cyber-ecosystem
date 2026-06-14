package connect

import (
	"context"
	"errors"
	"net/http"

	connectrpc "connectrpc.com/connect"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
)

var errUnexpectedResponseType = errors.New("internal error: unexpected response type")

// kratosInterceptor bridges Kratos middleware/transport-context into Connect.
// It MUST run outermost (see Server.HandlerOptions) so the Transport is injected
// before any user interceptor reads transport.FromServerContext.
type kratosInterceptor struct{ server *Server }

func newKratosInterceptor(s *Server) connectrpc.Interceptor { return &kratosInterceptor{server: s} }

func (i *kratosInterceptor) WrapUnary(next connectrpc.UnaryFunc) connectrpc.UnaryFunc {
	return func(ctx context.Context, req connectrpc.AnyRequest) (connectrpc.AnyResponse, error) {
		ctx, cancel := Merge(ctx, i.server.baseCtx)
		defer cancel()
		ep := ""
		if i.server.endpoint != nil {
			ep = i.server.endpoint.String()
		}
		tr := &Transport{
			endpoint:    ep,
			operation:   req.Spec().Procedure,
			reqHeader:   newHeader(req.Header()),
			replyHeader: newHeader(http.Header{}),
			httpMethod:  "POST",
			remoteAddr:  req.Peer().Addr,
		}
		ctx = transport.NewServerContext(ctx, tr)
		if i.server.timeout > 0 {
			var cancel2 context.CancelFunc
			ctx, cancel2 = context.WithTimeout(ctx, i.server.timeout)
			defer cancel2()
		}
		h := func(ctx context.Context, _ any) (any, error) { return next(ctx, req) }
		if m := i.server.middleware.Match(tr.Operation()); len(m) > 0 {
			h = middleware.Chain(m...)(h)
		}
		resp, err := h(ctx, req.Any())
		if err != nil {
			enc := i.server.errorEncoder(ctx, err)
			attachReplyHeadersToConnectError(tr, enc)
			return nil, enc
		}
		if cr, ok := resp.(connectrpc.AnyResponse); ok {
			copyReplyHeaders(tr, cr.Header())
			return cr, nil
		}
		return nil, connectrpc.NewError(connectrpc.CodeInternal, errUnexpectedResponseType)
	}
}

func (i *kratosInterceptor) WrapStreamingClient(next connectrpc.StreamingClientFunc) connectrpc.StreamingClientFunc {
	return next
}

func (i *kratosInterceptor) WrapStreamingHandler(next connectrpc.StreamingHandlerFunc) connectrpc.StreamingHandlerFunc {
	return func(ctx context.Context, conn connectrpc.StreamingHandlerConn) error {
		ctx, cancel := Merge(ctx, i.server.baseCtx)
		defer cancel()
		ep := ""
		if i.server.endpoint != nil {
			ep = i.server.endpoint.String()
		}
		tr := &Transport{
			endpoint:    ep,
			operation:   conn.Spec().Procedure,
			reqHeader:   newHeader(conn.RequestHeader()),
			replyHeader: newHeader(http.Header{}),
			httpMethod:  "POST",
			remoteAddr:  conn.Peer().Addr,
		}
		ctx = transport.NewServerContext(ctx, tr)
		// Stream middleware is applied per-message inside middlewareStream (adapter.go),
		// mirroring gRPC where stream middleware is per-message, not per-RPC.
		err := next(ctx, conn)
		copyReplyHeadersToConn(tr, conn)
		if err != nil {
			enc := i.server.errorEncoder(ctx, err)
			attachReplyHeadersToConnectError(tr, enc)
			return enc
		}
		return nil
	}
}

func copyReplyHeaders(tr *Transport, h http.Header) {
	for _, k := range tr.replyHeader.Keys() {
		for _, v := range tr.replyHeader.Values(k) {
			h.Add(k, v)
		}
	}
}

func copyReplyHeadersToConn(tr *Transport, conn connectrpc.StreamingHandlerConn) {
	for _, k := range tr.replyHeader.Keys() {
		for _, v := range tr.replyHeader.Values(k) {
			conn.ResponseHeader().Add(k, v)
		}
	}
}

func attachReplyHeadersToConnectError(tr *Transport, err error) {
	ce, ok := err.(*connectrpc.Error)
	if !ok || tr == nil {
		return
	}
	for _, k := range tr.replyHeader.Keys() {
		for _, v := range tr.replyHeader.Values(k) {
			ce.Meta().Add(k, v)
		}
	}
}
