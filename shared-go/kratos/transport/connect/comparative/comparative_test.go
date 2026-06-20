// Package comparative is the headline validation of the "three protocols treated
// equally" thesis.
//
// It brings up kratos gRPC + kratos HTTP + our Connect servers simultaneously,
// all sharing ONE middleware closure and ONE service implementation, then drives
// the SAME scenarios through each protocol's native client and asserts
// behavioral parity across:
//
//   - middleware invocation (one closure firing on all 3 transports)
//   - Transport.Kind()/Operation() values
//   - unary response data
//   - kratos error reason round-tripping
//   - streaming RPC parity (grpc vs connect only — transfer has no HTTP binding)
//
// The shared middleware is the crux: a single closure registered on all 3 servers
// must fire and observe a correctly-populated transport.Transporter on each. The
// only EXPECTED, documented difference is the Operation() string format (see
// TestTransportKindAndOperation); everything else must match.
package comparative

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	kerrors "github.com/go-kratos/kratos/v3/errors"
	kratosmiddleware "github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	kratosgrpc "github.com/go-kratos/kratos/v3/transport/grpc"
	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	mobilepb "cyber-ecosystem/gen/go/cyber/mobile/v1"
	mobilev1connect "cyber-ecosystem/gen/go/cyber/mobile/v1/v1connect"
)

// ---------------------------------------------------------------------------
// Shared recorder middleware
// ---------------------------------------------------------------------------

// recorder captures every middleware invocation across all 3 servers. The SAME
// instance is registered as the middleware on the grpc, http, and connect
// servers, so a call via any protocol appends exactly one entry.
type recorder struct {
	mu    sync.Mutex
	kinds []string
	ops   []string
}

func (r *recorder) record(kind, op string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.kinds = append(r.kinds, kind)
	r.ops = append(r.ops, op)
}

func (r *recorder) snapshot() (kinds, ops []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := func(s []string) []string {
		cp := make([]string, len(s))
		copy(cp, s)
		return cp
	}
	return out(r.kinds), out(r.ops)
}

// makeMW builds the shared middleware: it reads the transport from the server
// context (set by each transport's own interceptor) and records (kind, op).
// This is the literal "one closure works on all 3 transports" assertion.
func makeMW(r *recorder) kratosmiddleware.Middleware {
	return func(next kratosmiddleware.Handler) kratosmiddleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				r.record(tr.Kind().String(), tr.Operation())
			}
			return next(ctx, req)
		}
	}
}

// ---------------------------------------------------------------------------
// Shared services
// ---------------------------------------------------------------------------

// resourceSvc implements the MobileResourceService unary RPC for ALL 3 transports.
// mobilepb.MobileResourceServiceServer (grpc-style) is the canonical interface;
// both the HTTP and Connect generators dispatch to the same ListResource method.
// errOn flips the implementation to return a kratos NotFound for the error
// parity test.
type resourceSvc struct {
	mobilepb.UnimplementedMobileResourceServiceServer
	errOn bool
}

func (s *resourceSvc) ListResource(_ context.Context, _ *mobilepb.ListResourceRequest) (*mobilepb.ListResourceResponse, error) {
	if s.errOn {
		return nil, kerrors.NotFound("RES_NOT_FOUND", "nope")
	}
	return &mobilepb.ListResourceResponse{
		List: []*mobilepb.Service{{Name: "demo"}},
	}, nil
}

// transferSvc mirrors integration_test.go's testService: Subscribe→5 events,
// Echo→counts, Pipe→echo. Used for the grpc-vs-connect streaming parity test.
// (Transfer has no HTTP binding, so there is no http leg.)
type transferSvc struct {
	mobilepb.UnimplementedMobileTransferServiceServer
}

func (transferSvc) Subscribe(req *mobilepb.SubscribeRequest, stream grpc.ServerStreamingServer[mobilepb.SubscribeResponse]) error {
	for i := 0; i < 5; i++ {
		if err := stream.Send(&mobilepb.SubscribeResponse{
			EventId: fmt.Sprintf("%s-%d", req.GetTopic(), i+1),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (transferSvc) Echo(stream grpc.ClientStreamingServer[mobilepb.EchoRequest, mobilepb.EchoResponse]) error {
	var n int32
	for {
		_, err := stream.Recv()
		if err != nil {
			break
		}
		n++
	}
	return stream.SendAndClose(&mobilepb.EchoResponse{TotalMessages: n})
}

func (transferSvc) Pipe(stream grpc.BidiStreamingServer[mobilepb.PipeRequest, mobilepb.PipeResponse]) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return nil // EOF / client closed
		}
		if err := stream.Send(&mobilepb.PipeResponse{Data: req.GetData()}); err != nil {
			return err
		}
	}
}

// Raw is required by the interface but unused by the streaming parity test.
func (transferSvc) Raw(_ context.Context, req *mobilepb.RawRequest) (*httpbody.HttpBody, error) {
	ct := req.GetContentType()
	if ct == "" {
		ct = "application/octet-stream"
	}
	return &httpbody.HttpBody{ContentType: ct, Data: req.GetData()}, nil
}

// ---------------------------------------------------------------------------
// Harness: bring up all 3 servers + build all 3 clients
// ---------------------------------------------------------------------------

// clients holds the three typed clients built against the three servers.
type clients struct {
	grpcClient    mobilepb.MobileResourceServiceClient
	httpClient    mobilepb.MobileResourceServiceHTTPClient
	connectClient mobilev1connect.MobileResourceServiceClient

	// streaming clients (grpc + connect only)
	grpcStreamClient    mobilepb.MobileTransferServiceClient
	connectStreamClient mobilev1connect.MobileTransferServiceClient
}

// startAll spins up grpc, http, and connect servers on 127.0.0.1:0, each with
// the shared recorder middleware, each registering the resource service, and
// (grpc+connect) the transfer service. It then builds the three resource
// clients + two streaming clients and returns a stop func.
//
// The resourceSvc is shared by reference so the caller can flip errOn between
// tests (e.g. start a fresh harness with errOn=true for the error test).
func startAll(t *testing.T, res *resourceSvc) (*clients, *recorder, func()) {
	t.Helper()
	ctx := context.Background()
	rec := &recorder{}
	mw := makeMW(rec)
	cl := &clients{}

	var stops []func()
	cleanup := func() {
		for i := len(stops) - 1; i >= 0; i-- {
			stops[i]()
		}
	}
	// On failure mid-setup, tear down whatever we already brought up.
	fail := func(format string, args ...any) {
		cleanup()
		t.Fatalf(format, args...)
	}

	// --- gRPC server ---
	grpcSrv := kratosgrpc.NewServer(
		kratosgrpc.Address("127.0.0.1:0"),
		kratosgrpc.Timeout(0),
		kratosgrpc.Middleware(mw),
	)
	mobilepb.RegisterMobileResourceServiceServer(grpcSrv, res)
	mobilepb.RegisterMobileTransferServiceServer(grpcSrv, transferSvc{})
	// NOTE: kratos grpc.NewServer auto-registers grpc.health.v1.Health by
	// default (it creates its own grpchealth.Server), so the kratos grpc
	// client's health-aware path has a target without us re-registering.

	grpcEP, err := grpcSrv.Endpoint()
	if err != nil {
		fail("grpc endpoint: %v", err)
	}
	go func() { _ = grpcSrv.Start(ctx) }()
	waitReady(t, grpcEP.Host)

	grpcConn, err := kratosgrpc.NewClient(ctx,
		// direct resolver parses the address from the URL *path* (empty
		// authority), so the triple-slash form direct:///<host:port> is
		// required: direct://<host:port> would put host in the authority and
		// leave the path empty ("missing address").
		kratosgrpc.WithEndpoint("direct:///"+grpcEP.Host),
		kratosgrpc.WithTimeout(0),
	)
	if err != nil {
		fail("grpc client: %v", err)
	}
	cl.grpcClient = mobilepb.NewMobileResourceServiceClient(grpcConn)
	cl.grpcStreamClient = mobilepb.NewMobileTransferServiceClient(grpcConn)
	stops = append(stops, func() { _ = grpcConn.Close() })
	stops = append(stops, func() { _ = grpcSrv.Stop(ctx) })

	// --- HTTP server (kratos transport/http) ---
	httpSrv := kratoshttp.NewServer(
		kratoshttp.Address("127.0.0.1:0"),
		kratoshttp.Timeout(0),
		kratoshttp.Middleware(mw),
	)
	mobilepb.RegisterMobileResourceServiceHTTPServer(httpSrv, res)
	httpEP, err := httpSrv.Endpoint()
	if err != nil {
		fail("http endpoint: %v", err)
	}
	go func() { _ = httpSrv.Start(ctx) }()
	waitReady(t, httpEP.Host)

	httpClient, err := kratoshttp.NewClient(ctx,
		kratoshttp.WithEndpoint(httpEP.Host),
		kratoshttp.WithTimeout(0),
	)
	if err != nil {
		fail("http client: %v", err)
	}
	cl.httpClient = mobilepb.NewMobileResourceServiceHTTPClient(httpClient)
	stops = append(stops, func() { _ = httpClient.Close() })
	stops = append(stops, func() { _ = httpSrv.Stop(ctx) })

	// --- Connect server (our shared-go transport/connect) ---
	connSrv := connect.NewServer(
		connect.Address("127.0.0.1:0"),
		connect.Timeout(0),
		connect.Middleware(mw),
	)
	mobilepb.RegisterMobileResourceServiceConnectServer(connSrv, res)
	mobilepb.RegisterMobileTransferServiceConnectServer(connSrv, transferSvc{})
	connEP, err := connSrv.Endpoint()
	if err != nil {
		fail("connect endpoint: %v", err)
	}
	go func() { _ = connSrv.Start(ctx) }()
	waitReady(t, connEP.Host)

	connClient, err := connect.Dial(ctx,
		connect.WithEndpoint(connEP.Host),
		connect.WithH2C(true),
		connect.WithTimeout(0),
	)
	if err != nil {
		fail("connect dial: %v", err)
	}
	cl.connectClient = mobilev1connect.NewMobileResourceServiceClient(
		connClient.HTTPClient(), connClient.BaseURL(), connClient.ClientOptions()...,
	)
	cl.connectStreamClient = mobilev1connect.NewMobileTransferServiceClient(
		connClient.HTTPClient(), connClient.BaseURL(), connClient.ClientOptions()...,
	)
	stops = append(stops, func() { _ = connClient.Close() })
	stops = append(stops, func() { _ = connSrv.Stop(ctx) })

	return cl, rec, cleanup
}

// waitReady polls the loopback address with TCP dials until the listener
// accepts a connection (or the deadline passes), so the client never races
// ahead of an unstarted server. Mirrors the integration harness.
func waitReady(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server at %s never became ready", addr)
}

// ---------------------------------------------------------------------------
// Test 1: middleware parity — ONE closure fires on all 3 transports
// ---------------------------------------------------------------------------

func TestMiddlewareParityAllThree(t *testing.T) {
	cl, rec, stop := startAll(t, &resourceSvc{})
	defer stop()

	ctx := context.Background()
	if _, err := cl.grpcClient.ListResource(ctx, &mobilepb.ListResourceRequest{}); err != nil {
		t.Fatalf("grpc ListResource: %v", err)
	}
	if _, err := cl.httpClient.ListResource(ctx, &mobilepb.ListResourceRequest{}); err != nil {
		t.Fatalf("http ListResource: %v", err)
	}
	if _, err := cl.connectClient.ListResource(ctx, connectrpc.NewRequest(&mobilepb.ListResourceRequest{})); err != nil {
		t.Fatalf("connect ListResource: %v", err)
	}

	kinds, _ := rec.snapshot()
	if len(kinds) != 3 {
		t.Fatalf("expected 3 middleware invocations (one per protocol), got %d: %v", len(kinds), kinds)
	}
	got := map[string]bool{}
	for _, k := range kinds {
		got[k] = true
	}
	for _, want := range []string{"grpc", "http", "connect"} {
		if !got[want] {
			t.Errorf("expected middleware to fire for transport kind %q; kinds captured = %v", want, kinds)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 2: unary response data parity
// ---------------------------------------------------------------------------

func TestUnaryResponseParity(t *testing.T) {
	cl, _, stop := startAll(t, &resourceSvc{})
	defer stop()

	ctx := context.Background()
	const want = "demo"

	grpcResp, err := cl.grpcClient.ListResource(ctx, &mobilepb.ListResourceRequest{})
	if err != nil {
		t.Fatalf("grpc: %v", err)
	}
	httpResp, err := cl.httpClient.ListResource(ctx, &mobilepb.ListResourceRequest{})
	if err != nil {
		t.Fatalf("http: %v", err)
	}
	connResp, err := cl.connectClient.ListResource(ctx, connectrpc.NewRequest(&mobilepb.ListResourceRequest{}))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	gotGrpc := grpcResp.GetList()[0].GetName()
	gotHttp := httpResp.GetList()[0].GetName()
	gotConn := connResp.Msg.GetList()[0].GetName()

	if gotGrpc != want || gotHttp != want || gotConn != want {
		t.Fatalf("name mismatch: grpc=%q http=%q connect=%q, all want %q", gotGrpc, gotHttp, gotConn, want)
	}
	if gotGrpc != gotHttp || gotHttp != gotConn {
		t.Fatalf("protocols disagree: grpc=%q http=%q connect=%q", gotGrpc, gotHttp, gotConn)
	}
}

// ---------------------------------------------------------------------------
// Test 3: Transport.Kind() and Operation() parity (with a documented
// EXPECTED difference in the Operation() string format)
// ---------------------------------------------------------------------------

func TestTransportKindAndOperation(t *testing.T) {
	cl, rec, stop := startAll(t, &resourceSvc{})
	defer stop()

	ctx := context.Background()
	// one call per protocol so the recorder captures exactly one op per kind.
	if _, err := cl.grpcClient.ListResource(ctx, &mobilepb.ListResourceRequest{}); err != nil {
		t.Fatalf("grpc: %v", err)
	}
	if _, err := cl.httpClient.ListResource(ctx, &mobilepb.ListResourceRequest{}); err != nil {
		t.Fatalf("http: %v", err)
	}
	if _, err := cl.connectClient.ListResource(ctx, connectrpc.NewRequest(&mobilepb.ListResourceRequest{})); err != nil {
		t.Fatalf("connect: %v", err)
	}

	kinds, ops := rec.snapshot()
	if len(kinds) != 3 || len(ops) != 3 {
		t.Fatalf("expected 3 captured invocations, got kinds=%v ops=%v", kinds, ops)
	}

	// Build kind→op so we can assert per-protocol regardless of capture order.
	kind2op := map[string]string{}
	for i, k := range kinds {
		kind2op[k] = ops[i]
	}

	for _, want := range []string{"grpc", "http", "connect"} {
		if _, ok := kind2op[want]; !ok {
			t.Fatalf("missing captured invocation for kind %q (kinds=%v)", want, kinds)
		}
	}

	// Kind assertion: each transport reports its own canonical kind string.
	if kind2op["grpc"] == "" || kind2op["http"] == "" || kind2op["connect"] == "" {
		t.Fatalf("unexpected empty op map: %v", kind2op)
	}

	// Operation() parity. Initial hypothesis was that http would diverge to a
	// path-template form ("/api/v1/admin/resource") while grpc/connect kept the
	// gRPC FullMethod form. In practice ALL THREE report the identical gRPC
	// FullMethod — because protoc-gen-go-http emits an explicit
	// http.SetOperation(ctx, OperationMobileResourceServiceListResource) whose
	// constant is the FullMethod, overriding the mux path template the kratos
	// http filter would otherwise set. This is a STRONGER parity result than
	// expected: one middleware selector string ("…/ListResource") matches all 3.
	//
	// NOTE: had the http service NOT called http.SetOperation (e.g. a
	// hand-written handler), the http Operation() WOULD fall back to the path
	// template — a genuine transport-level difference to be aware of for
	// non-generated handlers. For generated services, parity holds exactly.
	const fullMethod = "/cyber.mobile.v1.MobileResourceService/ListResource"

	for _, kind := range []string{"grpc", "http", "connect"} {
		if op := kind2op[kind]; op != fullMethod {
			t.Errorf("%s Operation() = %q, want %q", kind, op, fullMethod)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 4: kratos error reason round-trips across all 3 protocols
// ---------------------------------------------------------------------------

func TestErrorParity(t *testing.T) {
	// errOn=true: ListResource returns kerrors.NotFound("RES_NOT_FOUND","nope").
	cl, _, stop := startAll(t, &resourceSvc{errOn: true})
	defer stop()

	ctx := context.Background()
	const wantReason = "RES_NOT_FOUND"

	// --- grpc ---
	_, grpcErr := cl.grpcClient.ListResource(ctx, &mobilepb.ListResourceRequest{})
	if grpcErr == nil {
		t.Fatal("grpc: expected error, got nil")
	}
	if code := grpcstatus.Code(grpcErr); code != grpccodes.NotFound {
		t.Errorf("grpc status code = %v, want NotFound", code)
	}
	// kratos client returns a gRPC status error; FromError recovers the reason
	// from the embedded ErrorInfo detail.
	grpcReason := kerrors.FromError(grpcErr).Reason
	if grpcReason != wantReason {
		t.Errorf("grpc reason = %q, want %q (err=%v)", grpcReason, wantReason, grpcErr)
	}

	// --- http ---
	_, httpErr := cl.httpClient.ListResource(ctx, &mobilepb.ListResourceRequest{})
	if httpErr == nil {
		t.Fatal("http: expected error, got nil")
	}
	// The kratos http client's DefaultErrorDecoder returns a *kerrors.Error
	// whose Code is the HTTP status and Reason is decoded from the body.
	var httpKE *kerrors.Error
	if !errors.As(httpErr, &httpKE) {
		t.Fatalf("http: expected *kerrors.Error, got %T: %v", httpErr, httpErr)
	}
	if httpKE.Code != http.StatusNotFound {
		t.Errorf("http Code = %d, want 404", httpKE.Code)
	}
	if httpKE.Reason != wantReason {
		t.Errorf("http reason = %q, want %q", httpKE.Reason, wantReason)
	}

	// --- connect ---
	_, connErr := cl.connectClient.ListResource(ctx, connectrpc.NewRequest(&mobilepb.ListResourceRequest{}))
	if connErr == nil {
		t.Fatal("connect: expected error, got nil")
	}
	var ce *connectrpc.Error
	if !errors.As(connErr, &ce) {
		t.Fatalf("connect: expected *connect.Error, got %T: %v", connErr, connErr)
	}
	if ce.Code() != connectrpc.CodeNotFound {
		t.Errorf("connect code = %v, want CodeNotFound", ce.Code())
	}
	// ConnectToError recovers the reason from the ErrorInfo detail we attached
	// in ErrorToConnect.
	connReason := connect.ConnectToError(connErr).Reason
	if connReason != wantReason {
		t.Errorf("connect reason = %q, want %q", connReason, wantReason)
	}

	// Headline assertion: the SAME kratos reason survives the round trip on all 3.
	if grpcReason != httpKE.Reason || httpKE.Reason != connReason {
		t.Fatalf("reason divergence: grpc=%q http=%q connect=%q", grpcReason, httpKE.Reason, connReason)
	}
}

// ---------------------------------------------------------------------------
// Test 5: streaming parity — grpc vs connect (transfer service has no HTTP leg)
// ---------------------------------------------------------------------------

// NOTE: TransferService exposes only grpc + connect bindings (no http). This is
// by design — streaming RPCs are not bound to the REST-style HTTP transport.
// So this test compares the two streaming-capable protocols only and documents
// that http is intentionally absent here.
func TestStreamingParityGRPCvsConnect(t *testing.T) {
	cl, _, stop := startAll(t, &resourceSvc{})
	defer stop()

	ctx := context.Background()

	// --- Subscribe (server-stream): expect 5 events on both ---
	grpcStream, err := cl.grpcStreamClient.Subscribe(ctx, &mobilepb.SubscribeRequest{Topic: "t"})
	if err != nil {
		t.Fatalf("grpc Subscribe open: %v", err)
	}
	grpcSubCount := 0
	for {
		_, rerr := grpcStream.Recv()
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			t.Fatalf("grpc Subscribe Recv: %v", rerr)
		}
		grpcSubCount++
	}

	connStream, err := cl.connectStreamClient.Subscribe(ctx, connectrpc.NewRequest(&mobilepb.SubscribeRequest{Topic: "t"}))
	if err != nil {
		t.Fatalf("connect Subscribe open: %v", err)
	}
	connSubCount := 0
	for connStream.Receive() {
		connSubCount++
	}
	if err := connStream.Err(); err != nil {
		t.Fatalf("connect Subscribe stream err: %v", err)
	}

	if grpcSubCount != 5 || connSubCount != 5 {
		t.Fatalf("Subscribe counts: grpc=%d connect=%d, both want 5", grpcSubCount, connSubCount)
	}
	if grpcSubCount != connSubCount {
		t.Fatalf("Subscribe count divergence: grpc=%d connect=%d", grpcSubCount, connSubCount)
	}

	// --- Echo (client-stream): send 3, expect TotalMessages=3 on both ---
	grpcEcho, eerr := cl.grpcStreamClient.Echo(ctx)
	if eerr != nil {
		t.Fatalf("grpc Echo open: %v", eerr)
	}
	for i := 0; i < 3; i++ {
		if err := grpcEcho.Send(&mobilepb.EchoRequest{Data: []byte("x")}); err != nil {
			t.Fatalf("grpc Echo Send[%d]: %v", i, err)
		}
	}
	grpcEchoResp, err := grpcEcho.CloseAndRecv()
	if err != nil {
		t.Fatalf("grpc Echo CloseAndRecv: %v", err)
	}

	connEcho := cl.connectStreamClient.Echo(ctx)
	for i := 0; i < 3; i++ {
		if err := connEcho.Send(&mobilepb.EchoRequest{Data: []byte("x")}); err != nil {
			t.Fatalf("connect Echo Send[%d]: %v", i, err)
		}
	}
	connEchoResp, err := connEcho.CloseAndReceive()
	if err != nil {
		t.Fatalf("connect Echo CloseAndReceive: %v", err)
	}

	if grpcEchoResp.GetTotalMessages() != 3 || connEchoResp.Msg.GetTotalMessages() != 3 {
		t.Fatalf("Echo counts: grpc=%d connect=%d, both want 3",
			grpcEchoResp.GetTotalMessages(), connEchoResp.Msg.GetTotalMessages())
	}
	if grpcEchoResp.GetTotalMessages() != connEchoResp.Msg.GetTotalMessages() {
		t.Fatalf("Echo count divergence: grpc=%d connect=%d",
			grpcEchoResp.GetTotalMessages(), connEchoResp.Msg.GetTotalMessages())
	}

	// --- Pipe (bidi): send 3, receive 3 echoes on both ---
	grpcPipe, perr := cl.grpcStreamClient.Pipe(ctx)
	if perr != nil {
		t.Fatalf("grpc Pipe open: %v", perr)
	}
	grpcPipeCount := 0
	for i := 0; i < 3; i++ {
		payload := fmt.Sprintf("m%d", i)
		if err := grpcPipe.Send(&mobilepb.PipeRequest{Data: []byte(payload)}); err != nil {
			t.Fatalf("grpc Pipe Send[%d]: %v", i, err)
		}
		resp, rerr := grpcPipe.Recv()
		if rerr != nil {
			t.Fatalf("grpc Pipe Recv[%d]: %v", i, rerr)
		}
		if string(resp.GetData()) != payload {
			t.Fatalf("grpc Pipe echo[%d] = %q, want %q", i, string(resp.GetData()), payload)
		}
		grpcPipeCount++
	}
	_ = grpcPipe.CloseSend()
	// Drain until server closes send side.
	for {
		_, rerr := grpcPipe.Recv()
		if rerr != nil {
			break
		}
	}

	connPipe := cl.connectStreamClient.Pipe(ctx)
	connPipeCount := 0
	for i := 0; i < 3; i++ {
		payload := fmt.Sprintf("m%d", i)
		if err := connPipe.Send(&mobilepb.PipeRequest{Data: []byte(payload)}); err != nil {
			t.Fatalf("connect Pipe Send[%d]: %v", i, err)
		}
		resp, rerr := connPipe.Receive()
		if rerr != nil {
			t.Fatalf("connect Pipe Receive[%d]: %v", i, rerr)
		}
		if string(resp.GetData()) != payload {
			t.Fatalf("connect Pipe echo[%d] = %q, want %q", i, string(resp.GetData()), payload)
		}
		connPipeCount++
	}
	if err := connPipe.CloseRequest(); err != nil {
		t.Fatalf("connect Pipe CloseRequest: %v", err)
	}
	for {
		_, rerr := connPipe.Receive()
		if rerr != nil {
			break
		}
	}
	_ = connPipe.CloseResponse()

	if grpcPipeCount != 3 || connPipeCount != 3 {
		t.Fatalf("Pipe counts: grpc=%d connect=%d, both want 3", grpcPipeCount, connPipeCount)
	}
	if grpcPipeCount != connPipeCount {
		t.Fatalf("Pipe count divergence: grpc=%d connect=%d", grpcPipeCount, connPipeCount)
	}
}
