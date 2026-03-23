package connect

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"connectrpc.com/connect"
)

// AdaptUnary creates a Connect unary handler from a gRPC-style handler.
// This is useful for manually adapting individual methods.
//
// Usage:
//
//	handler := connect.AdaptUnary(s.CreateBlog)
func AdaptUnary[Req, Resp any](handler func(context.Context, *Req) (*Resp, error)) func(context.Context, *connect.Request[Req]) (*connect.Response[Resp], error) {
	return func(ctx context.Context, req *connect.Request[Req]) (*connect.Response[Resp], error) {
		resp, err := handler(ctx, req.Msg)
		if err != nil {
			return nil, ErrorToConnect(err)
		}
		return connect.NewResponse(resp), nil
	}
}

// AdaptClientStream creates a Connect client streaming handler from a gRPC-style handler.
// The handler receives a receive function that returns messages until the stream ends.
func AdaptClientStream[Req, Resp any](handler func(context.Context, func() (*Req, error)) (*Resp, error)) func(context.Context, *connect.ClientStream[Req]) (*connect.Response[Resp], error) {
	return func(ctx context.Context, stream *connect.ClientStream[Req]) (*connect.Response[Resp], error) {
		recv := func() (*Req, error) {
			if !stream.Receive() {
				if err := stream.Err(); err != nil {
					return nil, err
				}
				return nil, nil // Stream ended
			}
			return stream.Msg(), nil
		}

		resp, err := handler(ctx, recv)
		if err != nil {
			return nil, ErrorToConnect(err)
		}
		return connect.NewResponse(resp), nil
	}
}

// AdaptServerStream creates a Connect server streaming handler from a gRPC-style handler.
func AdaptServerStream[Req, Resp any](handler func(context.Context, *Req, func(*Resp) error) error) func(context.Context, *connect.Request[Req], *connect.ServerStream[Resp]) error {
	return func(ctx context.Context, req *connect.Request[Req], stream *connect.ServerStream[Resp]) error {
		send := func(resp *Resp) error {
			return stream.Send(resp)
		}

		err := handler(ctx, req.Msg, send)
		if err != nil {
			return ErrorToConnect(err)
		}
		return nil
	}
}

// AdaptBidiStream creates a Connect bidirectional streaming handler from a gRPC-style handler.
func AdaptBidiStream[Req, Resp any](handler func(context.Context, func() (*Req, error), func(*Resp) error) error) func(context.Context, *connect.BidiStream[Req, Resp]) error {
	return func(ctx context.Context, stream *connect.BidiStream[Req, Resp]) error {
		recv := func() (*Req, error) {
			return stream.Receive()
		}

		send := func(resp *Resp) error {
			return stream.Send(resp)
		}

		err := handler(ctx, recv, send)
		if err != nil {
			return ErrorToConnect(err)
		}
		return nil
	}
}

// UnaryMethodAdapter creates a unary method adapter for a specific method name.
// This is used to adapt individual methods from gRPC to Connect.
func UnaryMethodAdapter[Req, Resp any](impl interface{}, methodName string) func(context.Context, *connect.Request[Req]) (*connect.Response[Resp], error) {
	return func(ctx context.Context, req *connect.Request[Req]) (*connect.Response[Resp], error) {
		method := reflect.ValueOf(impl).MethodByName(methodName)
		if !method.IsValid() {
			return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("method %s not found", methodName))
		}

		results := method.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(req.Msg),
		})

		if len(results) != 2 {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid method signature for %s", methodName))
		}

		resp := results[0].Interface()
		err, _ := results[1].Interface().(error)

		if err != nil {
			return nil, ErrorToConnect(err)
		}

		respTyped, ok := resp.(*Resp)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid response type for %s", methodName))
		}

		return connect.NewResponse(respTyped), nil
	}
}

// ServiceAdapter provides a generic way to adapt gRPC services to Connect handlers.
// It uses reflection to call the underlying gRPC methods.
type ServiceAdapter struct {
	impl interface{}
}

// NewServiceAdapter creates a new service adapter.
func NewServiceAdapter(impl interface{}) *ServiceAdapter {
	return &ServiceAdapter{impl: impl}
}

// InvokeMethod invokes a method by name using reflection.
func (a *ServiceAdapter) InvokeMethod(methodName string, ctx context.Context, req interface{}) (interface{}, error) {
	method := reflect.ValueOf(a.impl).MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	results := method.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(req),
	})

	if len(results) != 2 {
		return nil, fmt.Errorf("invalid method signature for %s", methodName)
	}

	resp := results[0].Interface()
	err, _ := results[1].Interface().(error)

	return resp, err
}

// BuildUnaryHandler creates a Connect handler for a unary method.
// This is the recommended way to adapt gRPC methods to Connect.
func BuildUnaryHandler[Req, Resp any](adapter *ServiceAdapter, methodName string) func(context.Context, *connect.Request[Req]) (*connect.Response[Resp], error) {
	return func(ctx context.Context, req *connect.Request[Req]) (*connect.Response[Resp], error) {
		resp, err := adapter.InvokeMethod(methodName, ctx, req.Msg)
		if err != nil {
			return nil, ErrorToConnect(err)
		}

		respTyped, ok := resp.(*Resp)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid response type for %s", methodName))
		}

		return connect.NewResponse(respTyped), nil
	}
}

// RegisterServiceHandler registers a service by creating handlers for each method.
// This is a helper function that works with the generated Connect code.
func RegisterServiceHandler(srv *Server, path string, handler http.Handler) {
	srv.Register(path, handler)
}
