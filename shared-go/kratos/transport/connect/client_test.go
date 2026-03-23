package connect

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	template1V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"
	template1V1connect "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1/template1V1connect"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockBlogService struct {
	template1V1.UnimplementedBlogServiceServer
}

const streamProcedure = "/acme.echo.v1.EchoService/ServerStream"

func (m *mockBlogService) QueryBlog(context.Context, *template1V1.QueryBlogRequest) (*template1V1.QueryBlogResponse, error) {
	return &template1V1.QueryBlogResponse{}, nil
}

func (m *mockBlogService) GetBlog(context.Context, *template1V1.GetBlogRequest) (*template1V1.GetBlogResponse, error) {
	return &template1V1.GetBlogResponse{Id: "blog"}, nil
}

func TestDialInsecureAndInvoke(t *testing.T) {
	srv := NewServer(Address("127.0.0.1:0"))
	svc := &mockBlogService{}
	srv.Register(template1V1connect.NewBlogServiceHandler(svc, srv.HandlerOptions()...))

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
	cc, err := DialInsecure(context.Background(), WithEndpoint(u.Host))
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()

	client := template1V1connect.NewBlogServiceClient(cc.HTTPClient(), cc.Endpoint(), cc.ClientOptions()...)
	if _, err := client.QueryBlog(context.Background(), &template1V1.QueryBlogRequest{}); err != nil {
		t.Fatalf("QueryBlog() error = %v", err)
	}
}

func TestClientMiddlewareTransportContext(t *testing.T) {
	srv := NewServer(Address("127.0.0.1:0"))
	svc := &mockBlogService{}
	srv.Register(template1V1connect.NewBlogServiceHandler(svc, srv.HandlerOptions()...))

	called := false
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
	cc, err := DialInsecure(
		context.Background(),
		WithEndpoint(u.Host),
		WithMiddleware(func(next middleware.Handler) middleware.Handler {
			return func(ctx context.Context, req any) (any, error) {
				info, ok := transport.FromClientContext(ctx)
				if !ok {
					t.Fatal("missing client transport context")
				}
				if info.Kind() != KindConnect {
					t.Fatalf("kind=%q", info.Kind())
				}
				info.RequestHeader().Set("x-from-mw", "yes")
				if _, ok := req.(*template1V1.QueryBlogRequest); !ok {
					t.Fatalf("unexpected req type %T", req)
				}
				called = true
				return next(ctx, req)
			}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()

	client := template1V1connect.NewBlogServiceClient(cc.HTTPClient(), cc.Endpoint(), cc.ClientOptions()...)
	if _, err := client.QueryBlog(context.Background(), &template1V1.QueryBlogRequest{}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("middleware not called")
	}
}

func TestWithInterceptors(t *testing.T) {
	srv := newIPv4Server(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "{}")
	}))
	defer srv.Close()

	seen := false
	cc, err := DialInsecure(
		context.Background(),
		WithEndpoint(srv.URL),
		WithInterceptors(connectrpc.UnaryInterceptorFunc(func(next connectrpc.UnaryFunc) connectrpc.UnaryFunc {
			return func(ctx context.Context, req connectrpc.AnyRequest) (connectrpc.AnyResponse, error) {
				seen = true
				return next(ctx, req)
			}
		})),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()

	client := template1V1connect.NewBlogServiceClient(cc.HTTPClient(), cc.Endpoint(), cc.ClientOptions()...)
	_, _ = client.QueryBlog(context.Background(), &template1V1.QueryBlogRequest{})
	if !seen {
		t.Fatal("interceptor not called")
	}
}

type mockDiscovery struct {
	endpoints []string
}

func (m *mockDiscovery) GetService(_ context.Context, _ string) ([]*registry.ServiceInstance, error) {
	return nil, nil
}

func (m *mockDiscovery) Watch(ctx context.Context, _ string) (registry.Watcher, error) {
	return &mockWatcher{ctx: ctx, endpoints: m.endpoints}, nil
}

type mockWatcher struct {
	ctx       context.Context
	endpoints []string
	once      sync.Once
}

func (m *mockWatcher) Next() ([]*registry.ServiceInstance, error) {
	select {
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	default:
	}
	out := make([]*registry.ServiceInstance, 0, len(m.endpoints))
	for i, ep := range m.endpoints {
		out = append(out, &registry.ServiceInstance{
			ID:        fmt.Sprintf("%d", i+1),
			Name:      "demo",
			Version:   "v1",
			Endpoints: []string{ep},
		})
	}
	m.once.Do(func() {})
	time.Sleep(10 * time.Millisecond)
	return out, nil
}

func (m *mockWatcher) Stop() error { return nil }

func TestWithDiscoveryAndNodeFilter(t *testing.T) {
	s1 := newIPv4Server(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "one")
	}))
	defer s1.Close()
	s2 := newIPv4Server(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "two")
	}))
	defer s2.Close()
	addr1 := strings.TrimPrefix(s1.URL, "http://")
	addr2 := strings.TrimPrefix(s2.URL, "http://")

	cc, err := DialInsecure(
		context.Background(),
		WithEndpoint("discovery:///demo"),
		WithDiscovery(&mockDiscovery{endpoints: []string{
			fmt.Sprintf("connect://%s?isSecure=%s", addr1, strconv.FormatBool(false)),
			fmt.Sprintf("connect://%s?isSecure=%s", addr2, strconv.FormatBool(false)),
		}}),
		WithBlock(),
		WithNodeFilter(func(_ context.Context, nodes []selector.Node) []selector.Node {
			out := make([]selector.Node, 0, len(nodes))
			for _, n := range nodes {
				if n.Address() == addr2 {
					out = append(out, n)
				}
			}
			return out
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()

	req, _ := http.NewRequest(http.MethodGet, "http://demo/ping", nil)
	resp, err := cc.HTTPClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "two" {
		t.Fatalf("body = %q, want %q", string(body), "two")
	}
}

func TestH2CReflectionWithDialInsecure(t *testing.T) {
	srv := NewServer(Address("127.0.0.1:0"))
	svc := &mockBlogService{}
	srv.Register(template1V1connect.NewBlogServiceHandler(svc, srv.HandlerOptions()...))

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
	cc, err := DialInsecure(context.Background(), WithEndpoint(u.Host), WithH2C(true))
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()

	rc := grpcreflect.NewClient(cc.HTTPClient(), cc.Endpoint(), cc.ClientOptions()...)
	stream := rc.NewStream(context.Background())
	names, err := stream.ListServices()
	if err != nil {
		t.Fatal(err)
	}
	_, _ = stream.Close()
	if !containsServiceClient(names, protoreflect.FullName("api.template1.v1.BlogService")) {
		t.Fatalf("reflection services = %v, want api.template1.v1.BlogService", names)
	}
}

func TestDialWithTLSConfig(t *testing.T) {
	svc := &mockBlogService{}
	path, handler := template1V1connect.NewBlogServiceHandler(svc)
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, path) {
			handler.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	}))
	ts.EnableHTTP2 = true
	ts.StartTLS()
	defer ts.Close()

	endpoint := strings.TrimPrefix(ts.URL, "https://")
	cc, err := Dial(
		context.Background(),
		WithEndpoint(endpoint),
		WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()

	client := template1V1connect.NewBlogServiceClient(cc.HTTPClient(), cc.Endpoint(), cc.ClientOptions()...)
	if _, err := client.QueryBlog(context.Background(), &template1V1.QueryBlogRequest{}); err != nil {
		t.Fatalf("QueryBlog TLS call error = %v", err)
	}
}

func containsServiceClient(services []protoreflect.FullName, target protoreflect.FullName) bool {
	for _, name := range services {
		if name == target {
			return true
		}
	}
	return false
}

func TestClientOptionSetters(t *testing.T) {
	o := &clientOptions{}
	mockRT := roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, nil })
	mockD := &mockDiscovery{}
	filter := func(_ context.Context, _ []selector.Node) []selector.Node { return nil }
	WithEndpoint("127.0.0.1:13000")(o)
	WithTimeout(time.Second)(o)
	WithSubset(10)(o)
	WithBlock()(o)
	WithH2C(true)(o)
	WithTransport(mockRT)(o)
	WithDiscovery(mockD)(o)
	WithNodeFilter(filter)(o)
	WithStreamMiddleware(func(next middleware.Handler) middleware.Handler {
		return next
	})(o)
	if o.endpoint != "127.0.0.1:13000" {
		t.Fatalf("endpoint=%q", o.endpoint)
	}
	if o.timeout != time.Second {
		t.Fatalf("timeout=%v", o.timeout)
	}
	if o.subsetSize != 10 {
		t.Fatalf("subset=%d", o.subsetSize)
	}
	if !o.block {
		t.Fatal("block should be true")
	}
	if !o.h2c {
		t.Fatal("h2c should be true")
	}
	if o.transport == nil {
		t.Fatal("transport should be set")
	}
	if o.discovery == nil {
		t.Fatal("discovery should be set")
	}
	if len(o.nodeFilters) != 1 {
		t.Fatalf("node filter len=%d", len(o.nodeFilters))
	}
	if len(o.streamMw) != 1 {
		t.Fatalf("stream middleware len=%d", len(o.streamMw))
	}
}

func TestNodeFilterNoAvailableNode(t *testing.T) {
	s1 := newIPv4Server(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer s1.Close()
	addr1 := strings.TrimPrefix(s1.URL, "http://")

	cc, err := DialInsecure(
		context.Background(),
		WithEndpoint("discovery:///demo"),
		WithDiscovery(&mockDiscovery{endpoints: []string{
			fmt.Sprintf("connect://%s?isSecure=%s", addr1, strconv.FormatBool(false)),
		}}),
		WithBlock(),
		WithNodeFilter(func(_ context.Context, _ []selector.Node) []selector.Node {
			return nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()

	req, _ := http.NewRequest(http.MethodGet, "http://demo/ping", nil)
	_, err = cc.HTTPClient().Do(req)
	if err == nil {
		t.Fatal("expected NODE_NOT_FOUND error")
	}
}

func TestDialEndpointRequired(t *testing.T) {
	_, err := DialInsecure(context.Background())
	if err == nil {
		t.Fatal("expected endpoint required error")
	}
}

func TestDialWithDiscoveryAndDirectTarget(t *testing.T) {
	s1 := newIPv4Server(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s1.Close()

	cc, err := DialInsecure(
		context.Background(),
		WithEndpoint(s1.URL),
		WithDiscovery(&mockDiscovery{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()

	req, _ := http.NewRequest(http.MethodGet, s1.URL, nil)
	resp, err := cc.HTTPClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
}

func TestClientStreamMiddlewareTransportContext(t *testing.T) {
	srv := NewServer(Address("127.0.0.1:0"))
	srv.Register(
		streamProcedure,
		connectrpc.NewServerStreamHandlerSimple(
			streamProcedure,
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
	called := false
	cc, err := DialInsecure(
		context.Background(),
		WithEndpoint(u.Host),
		WithH2C(true),
		WithStreamMiddleware(func(next middleware.Handler) middleware.Handler {
			return func(ctx context.Context, req any) (any, error) {
				info, ok := transport.FromClientContext(ctx)
				if !ok {
					t.Fatal("missing client transport context")
				}
				if info.Operation() != streamProcedure {
					t.Fatalf("operation=%q", info.Operation())
				}
				called = true
				return next(ctx, req)
			}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()

	client := connectrpc.NewClient[emptypb.Empty, emptypb.Empty](cc.HTTPClient(), cc.Endpoint()+streamProcedure, cc.ClientOptions()...)
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

func newIPv4Server(handler http.Handler) *httptest.Server {
	ts := httptest.NewUnstartedServer(handler)
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	ts.Listener = l
	ts.Start()
	return ts
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
