package connect

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

const testProcedure = "/acme.echo.v1.EchoService/Ping"

func TestServerUnaryMiddlewareAndHeaders(t *testing.T) {
	srv := NewServer(
		Address("127.0.0.1:0"),
		Middleware(func(next middleware.Handler) middleware.Handler {
			return func(ctx context.Context, req any) (any, error) {
				info, ok := transport.FromServerContext(ctx)
				if !ok {
					t.Fatal("server transport context not found")
				}
				if info.Operation() != testProcedure {
					t.Fatalf("operation = %q, want %q", info.Operation(), testProcedure)
				}
				info.ReplyHeader().Set("x-mw", "ok")
				return next(ctx, req)
			}
		}),
	)

	srv.Register(
		testProcedure,
		connectrpc.NewUnaryHandlerSimple(
			testProcedure,
			func(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
				info, ok := transport.FromServerContext(ctx)
				if !ok {
					t.Fatal("server transport context not found in handler")
				}
				if info.Kind() != KindConnect {
					t.Fatalf("kind = %q", info.Kind())
				}
				info.ReplyHeader().Set("x-handler", "ok")
				return &emptypb.Empty{}, nil
			},
			srv.HandlerOptions()...,
		),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- srv.Start(ctx)
	}()
	waitUntilServing(t, srv)

	u, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	client := connectrpc.NewClient[emptypb.Empty, emptypb.Empty](
		http.DefaultClient,
		"http://"+u.Host+testProcedure,
	)
	resp, err := client.CallUnary(context.Background(), connectrpc.NewRequest(&emptypb.Empty{}))
	if err != nil {
		t.Fatalf("CallUnary() error = %v", err)
	}
	if got := resp.Header().Get("x-mw"); got != "ok" {
		t.Fatalf("x-mw = %q", got)
	}
	if got := resp.Header().Get("x-handler"); got != "ok" {
		t.Fatalf("x-handler = %q", got)
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer stopCancel()
	if err := srv.Stop(stopCtx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start() returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestServerErrorMapping(t *testing.T) {
	srv := NewServer(
		Address("127.0.0.1:0"),
		Middleware(func(next middleware.Handler) middleware.Handler {
			return func(ctx context.Context, req any) (any, error) {
				info, ok := transport.FromServerContext(ctx)
				if ok {
					info.ReplyHeader().Set("x-mw", "err")
				}
				return next(ctx, req)
			}
		}),
	)
	srv.Register(
		testProcedure,
		connectrpc.NewUnaryHandlerSimple(
			testProcedure,
			func(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
				return nil, kerrors.BadRequest("INVALID_INPUT", "bad input")
			},
			srv.HandlerOptions()...,
		),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Start(ctx) }()
	waitUntilServing(t, srv)

	u, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	client := connectrpc.NewClient[emptypb.Empty, emptypb.Empty](
		http.DefaultClient,
		"http://"+u.Host+testProcedure,
	)
	_, err = client.CallUnary(context.Background(), connectrpc.NewRequest(&emptypb.Empty{}))
	if err == nil {
		t.Fatal("expected error")
	}
	var ce *connectrpc.Error
	if !kerrors.As(err, &ce) {
		t.Fatalf("expected connect error, got %T", err)
	}
	if ce.Code() != connectrpc.CodeInvalidArgument {
		t.Fatalf("code = %v, want %v", ce.Code(), connectrpc.CodeInvalidArgument)
	}
	if ce.Meta().Get("x-mw") != "err" {
		t.Fatalf("x-mw = %q", ce.Meta().Get("x-mw"))
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer stopCancel()
	_ = srv.Stop(stopCtx)
}

func TestReflectionDefaultEnabledAndDisableOption(t *testing.T) {
	client := &http.Client{Timeout: 500 * time.Millisecond}

	t.Run("enabled by default", func(t *testing.T) {
		srv := NewServer(Address("127.0.0.1:0"))
		srv.Register(
			testProcedure,
			connectrpc.NewUnaryHandlerSimple(
				testProcedure,
				func(context.Context, *emptypb.Empty) (*emptypb.Empty, error) { return &emptypb.Empty{}, nil },
				srv.HandlerOptions()...,
			),
		)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() { _ = srv.Start(ctx) }()
		waitUntilServing(t, srv)

		u, err := srv.Endpoint()
		if err != nil {
			t.Fatal(err)
		}
		h2cClient := newH2CClient()
		rClient := grpcreflect.NewClient(h2cClient, "http://"+u.Host)
		stream := rClient.NewStream(context.Background())
		names, err := stream.ListServices()
		if err != nil {
			t.Fatal(err)
		}
		_, _ = stream.Close()
		if !containsService(names, protoreflect.FullName("acme.echo.v1.EchoService")) {
			t.Fatalf("reflection services = %v, want acme.echo.v1.EchoService", names)
		}
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		_ = srv.Stop(stopCtx)
	})

	t.Run("disable reflection", func(t *testing.T) {
		srv := NewServer(Address("127.0.0.1:0"), DisableReflection())
		srv.Register(
			testProcedure,
			connectrpc.NewUnaryHandlerSimple(
				testProcedure,
				func(context.Context, *emptypb.Empty) (*emptypb.Empty, error) { return &emptypb.Empty{}, nil },
				srv.HandlerOptions()...,
			),
		)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() { _ = srv.Start(ctx) }()
		waitUntilServing(t, srv)

		u, err := srv.Endpoint()
		if err != nil {
			t.Fatal(err)
		}
		path, _ := ReflectionHandler("acme.echo.v1.EchoService")
		resp, err := client.Get("http://" + u.Host + path)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("reflection path %s should not be registered, status=%d", path, resp.StatusCode)
		}
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		_ = srv.Stop(stopCtx)
	})
}

func TestHealthEndpoint(t *testing.T) {
	client := &http.Client{Timeout: 500 * time.Millisecond}

	srv := NewServer(Address("127.0.0.1:0"), DisableReflection())
	srv.Register(
		testProcedure,
		connectrpc.NewUnaryHandlerSimple(
			testProcedure,
			func(context.Context, *emptypb.Empty) (*emptypb.Empty, error) { return &emptypb.Empty{}, nil },
			srv.HandlerOptions()...,
		),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Start(ctx) }()
	waitUntilServing(t, srv)

	u, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Get("http://" + u.Host + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/healthz status=%d", resp.StatusCode)
	}
	if !strings.Contains(string(body), "SERVING") {
		t.Fatalf("unexpected health body: %s", string(body))
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer stopCancel()
	_ = srv.Stop(stopCtx)
}

func TestServerFilterChain(t *testing.T) {
	srv := NewServer(
		Address("127.0.0.1:0"),
		Filter(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("x-filter", "yes")
				next.ServeHTTP(w, r)
			})
		}),
	)
	srv.Register(
		testProcedure,
		connectrpc.NewUnaryHandlerSimple(
			testProcedure,
			func(context.Context, *emptypb.Empty) (*emptypb.Empty, error) { return &emptypb.Empty{}, nil },
			srv.HandlerOptions()...,
		),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Start(ctx) }()
	waitUntilServing(t, srv)

	u, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	client := connectrpc.NewClient[emptypb.Empty, emptypb.Empty](
		http.DefaultClient,
		"http://"+u.Host+testProcedure,
	)
	resp, err := client.CallUnary(context.Background(), connectrpc.NewRequest(&emptypb.Empty{}))
	if err != nil {
		t.Fatal(err)
	}
	if got := resp.Header().Get("x-filter"); got != "yes" {
		t.Fatalf("x-filter=%q", got)
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer stopCancel()
	_ = srv.Stop(stopCtx)
}

func TestServerStreamMiddlewareTransportContext(t *testing.T) {
	called := false
	srv := NewServer(
		Address("127.0.0.1:0"),
		StreamMiddleware(func(next middleware.Handler) middleware.Handler {
			return func(ctx context.Context, req any) (any, error) {
				info, ok := transport.FromServerContext(ctx)
				if !ok {
					t.Fatal("server transport context not found")
				}
				if info.Operation() != testProcedure {
					t.Fatalf("operation = %q, want %q", info.Operation(), testProcedure)
				}
				called = true
				return next(ctx, req)
			}
		}),
	)
	srv.Register(
		testProcedure,
		connectrpc.NewServerStreamHandlerSimple(
			testProcedure,
			func(_ context.Context, _ *emptypb.Empty, stream *connectrpc.ServerStream[emptypb.Empty]) error {
				return stream.Send(&emptypb.Empty{})
			},
			srv.HandlerOptions()...,
		),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Start(ctx) }()
	waitUntilServing(t, srv)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		_ = srv.Stop(stopCtx)
	}()

	u, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	client := connectrpc.NewClient[emptypb.Empty, emptypb.Empty](newH2CClient(), "http://"+u.Host+testProcedure)
	stream, err := client.CallServerStream(context.Background(), connectrpc.NewRequest(&emptypb.Empty{}))
	if err != nil {
		t.Fatal(err)
	}
	for stream.Receive() {
	}
	if err := stream.Err(); err != nil {
		t.Fatal(err)
	}
	_ = stream.Close()
	if !called {
		t.Fatal("stream middleware not called")
	}
}

func TestHandlerOptionsMerge(t *testing.T) {
	seen := false
	custom := connectrpc.UnaryInterceptorFunc(func(next connectrpc.UnaryFunc) connectrpc.UnaryFunc {
		return func(ctx context.Context, req connectrpc.AnyRequest) (connectrpc.AnyResponse, error) {
			if req.Spec().Schema != "schema-test" {
				t.Fatalf("schema=%v", req.Spec().Schema)
			}
			seen = true
			return next(ctx, req)
		}
	})
	srv := NewServer(
		Address("127.0.0.1:0"),
		ConnectOptions(connectrpc.WithSchema("schema-test")),
		Interceptors(custom),
	)
	srv.Register(
		testProcedure,
		connectrpc.NewUnaryHandlerSimple(
			testProcedure,
			func(context.Context, *emptypb.Empty) (*emptypb.Empty, error) { return &emptypb.Empty{}, nil },
			srv.HandlerOptions()...,
		),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Start(ctx) }()
	waitUntilServing(t, srv)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		_ = srv.Stop(stopCtx)
	}()

	u, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	client := connectrpc.NewClient[emptypb.Empty, emptypb.Empty](
		http.DefaultClient,
		"http://"+u.Host+testProcedure,
	)
	_, err = client.CallUnary(context.Background(), connectrpc.NewRequest(&emptypb.Empty{}))
	if err != nil {
		t.Fatal(err)
	}
	if !seen {
		t.Fatal("custom interceptor not called")
	}
}

func waitUntilServing(t *testing.T, srv *Server) {
	t.Helper()
	client := &http.Client{Timeout: 200 * time.Millisecond}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if host := srv.listenerAddr(); host != "" {
			resp, reqErr := client.Get("http://" + host + "/healthz")
			if reqErr == nil {
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return
				}
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
	t.Fatal("server not ready")
}

func newH2CClient() *http.Client {
	return &http.Client{
		Timeout: time.Second,
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}
}

func containsService(services []protoreflect.FullName, target protoreflect.FullName) bool {
	for _, name := range services {
		if name == target {
			return true
		}
	}
	return false
}
