// Package integration provides end-to-end tests for the Connect transport.
//
// These tests spin up a real loopback Connect server (serving h2c — cleartext
// HTTP/2 — because the registered handlers need streaming, which requires HTTP/2)
// and hit it with a real generated v1connect client built on top of our own
// connect.Dial client. This validates the whole stack end-to-end: server bridge
// (gRPC-style impl -> Connect handler), the codec, and the client transport
// (including its h2c round tripper).
//
// All four method kinds are exercised against the connecttest.v1.TransferService:
//   - Raw         (unary,           returns google.api.HttpBody)
//   - Subscribe   (server-stream,   5 events)
//   - Echo        (client-stream,   counts 3 messages)
//   - Pipe        (bidi,            echoes 3 messages)
package integration

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	kerrors "github.com/go-kratos/kratos/v3/errors"
	kratosmiddleware "github.com/go-kratos/kratos/v3/middleware"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	testpb "cyber-ecosystem/shared-go/kratos/transport/connect/testpb"
	testpbconnect "cyber-ecosystem/shared-go/kratos/transport/connect/testpb/testpbconnect"
)

// testService implements testpb.TransferServiceServer (the gRPC-style interface)
// for the integration tests. RegisterTransferServiceConnectServer bridges it
// onto a Connect server.
type testService struct {
	testpb.UnimplementedTransferServiceServer
}

// Subscribe emits 5 events derived from the requested topic.
func (testService) Subscribe(req *testpb.SubscribeRequest, stream grpc.ServerStreamingServer[testpb.SubscribeResponse]) error {
	for i := 0; i < 5; i++ {
		if err := stream.Send(&testpb.SubscribeResponse{
			EventId: fmt.Sprintf("%s-%d", req.GetTopic(), i+1),
		}); err != nil {
			return err
		}
	}
	return nil
}

// Echo counts the incoming messages and returns a summary.
//
// Sentinel: a first message whose Data is exactly "NO_CLOSE" makes the handler
// misbehave — it returns nil WITHOUT calling SendAndClose, exercising the
// transport's CodeInternal error path (see adapter.HandleClientStream). Used by
// TestClientStreamMissingSendAndClose.
func (testService) Echo(stream grpc.ClientStreamingServer[testpb.EchoRequest, testpb.EchoResponse]) error {
	var n int32
	var totalBytes int64
	skipClose := false
	for {
		req, err := stream.Recv()
		if err != nil {
			break
		}
		if string(req.GetData()) == "NO_CLOSE" {
			skipClose = true
			continue
		}
		n++
		totalBytes += int64(len(req.GetData()))
	}
	if skipClose {
		return nil // misbehave: return without SendAndClose
	}
	return stream.SendAndClose(&testpb.EchoResponse{TotalMessages: n, TotalBytes: totalBytes})
}

// Pipe echoes each incoming message back to the client until the stream closes.
func (testService) Pipe(stream grpc.BidiStreamingServer[testpb.PipeRequest, testpb.PipeResponse]) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return nil // EOF / client closed send side
		}
		if err := stream.Send(&testpb.PipeResponse{Data: req.GetData()}); err != nil {
			return err
		}
	}
}

// Raw returns the request bytes wrapped in an HttpBody with the requested
// content type (defaulting to application/octet-stream).
//
// Sentinels:
//   - Data == "ERR"  returns a kratos NotFound error — used by TestErrorMapping
//     to verify kratos→connect code mapping end-to-end.
//   - Data == "SLOW" sleeps 1s before responding — used by TestClientTimeout
//     to trigger a client-side deadline. Guarded by the sentinel so other
//     tests are unaffected.
func (testService) Raw(ctx context.Context, req *testpb.RawRequest) (*httpbody.HttpBody, error) {
	if string(req.GetData()) == "ERR" {
		return nil, kerrors.NotFound("NOT_FOUND", "raw error sentinel")
	}
	if string(req.GetData()) == "SLOW" {
		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	ct := req.GetContentType()
	if ct == "" {
		ct = "application/octet-stream"
	}
	return &httpbody.HttpBody{ContentType: ct, Data: req.GetData()}, nil
}

// startServer brings up a loopback Connect server (h2c) with the test service
// registered and returns a typed v1connect client — also speaking h2c via our
// own connect.Dial — plus a stop func that tears everything down.
//
// h2c is mandatory: Connect streaming requires HTTP/2, and we're on cleartext.
// The server wraps its mux with h2c.NewHandler when tlsConf==nil (see
// NewServer), and the client uses an http2.Transport with AllowHTTP:true when
// insecure && WithH2C (see defaultRoundTripper).
func startServer(t *testing.T) (testpbconnect.TransferServiceClient, func()) {
	t.Helper()
	return startServerWithClient(t)
}

// startServerWithClient brings up a loopback Connect server (h2c) with the
// test service registered and returns a typed v1connect client built from
// connect.Dial with the given extra dial options (applied on top of the
// mandatory WithEndpoint + WithH2C). It lets client-behavior tests dial with
// specific With* options (WithMiddleware, WithTimeout, WithTransport, ...).
func startServerWithClient(t *testing.T, dialOpts ...connect.ClientOption) (testpbconnect.TransferServiceClient, func()) {
	t.Helper()

	// Timeout(0) disables the unary/overall deadline so long-lived streams
	// aren't killed by the server's default 1s timeout.
	srv := connect.NewServer(connect.Address("127.0.0.1:0"), connect.Timeout(0))
	testpb.RegisterTransferServiceConnectServer(srv, testService{})

	// Endpoint() creates the listener up front, so its Host is the real
	// address the client must dial.
	ep, err := srv.Endpoint()
	if err != nil {
		t.Fatalf("endpoint: %v", err)
	}

	ctx := context.Background()
	go func() { _ = srv.Start(ctx) }()

	// Wait until the listener actually accepts connections so the client
	// never races ahead of an unstarted server.
	waitReady(t, ep.Host)

	// h2c client via our own connect.Dial — this validates both the client
	// transport AND the server end-to-end.
	opts := make([]connect.ClientOption, 0, len(dialOpts)+2)
	opts = append(opts, connect.WithEndpoint(ep.Host), connect.WithH2C(true))
	opts = append(opts, dialOpts...)
	cli, err := connect.Dial(ctx, opts...)
	if err != nil {
		_ = srv.Stop(ctx)
		t.Fatalf("dial: %v", err)
	}
	client := testpbconnect.NewTransferServiceClient(cli.HTTPClient(), cli.BaseURL(), cli.ClientOptions()...)

	stop := func() {
		_ = cli.Close()
		_ = srv.Stop(ctx)
	}
	return client, stop
}

// waitReady polls the address with TCP dials until it accepts a connection or
// the deadline passes.
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

func TestUnaryRaw(t *testing.T) {
	cli, stop := startServer(t)
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
	if got := resp.Msg.GetContentType(); got != "text/plain" {
		t.Fatalf("content_type = %q, want %q", got, "text/plain")
	}
}

func TestServerStreamSubscribe(t *testing.T) {
	cli, stop := startServer(t)
	defer stop()

	stream, err := cli.Subscribe(context.Background(), connectrpc.NewRequest(&testpb.SubscribeRequest{Topic: "t"}))
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	var got int
	var wantFirst = "t-1"
	var firstOK bool
	for stream.Receive() {
		got++
		msg := stream.Msg()
		if got == 1 && msg.GetEventId() == wantFirst {
			firstOK = true
		}
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream error: %v", err)
	}
	if got != 5 {
		t.Fatalf("received %d events, want 5", got)
	}
	if !firstOK {
		t.Fatalf("first event id = %q, want %q", wantFirst, wantFirst)
	}
}

func TestClientStreamEcho(t *testing.T) {
	cli, stop := startServer(t)
	defer stop()

	stream := cli.Echo(context.Background())
	for i := 0; i < 3; i++ {
		if err := stream.Send(&testpb.EchoRequest{Data: []byte("x")}); err != nil {
			t.Fatalf("Send[%d]: %v", i, err)
		}
	}

	resp, err := stream.CloseAndReceive()
	if err != nil {
		t.Fatalf("CloseAndReceive: %v", err)
	}
	if got := resp.Msg.GetTotalMessages(); got != 3 {
		t.Fatalf("total_messages = %d, want 3", got)
	}
	if got := resp.Msg.GetTotalBytes(); got != 3 { // 3 * len("x")
		t.Fatalf("total_bytes = %d, want 3", got)
	}
}

// TestClientStreamMissingSendAndClose (I1) verifies that when a client-streaming
// handler returns WITHOUT calling SendAndClose (the "NO_CLOSE" sentinel in
// testService.Echo), the transport surfaces a *connect.Error with CodeInternal
// — the error path in adapter.HandleClientStream.
func TestClientStreamMissingSendAndClose(t *testing.T) {
	cli, stop := startServer(t)
	defer stop()

	stream := cli.Echo(context.Background())
	if err := stream.Send(&testpb.EchoRequest{Data: []byte("NO_CLOSE")}); err != nil {
		t.Fatal(err)
	}
	_, err := stream.CloseAndReceive() // connect-go v1.20 client method name (T10a)
	if err == nil {
		t.Fatal("expected error when handler returns without SendAndClose, got nil")
	}
	var ce *connectrpc.Error
	if !errors.As(err, &ce) {
		t.Fatalf("expected *connect.Error, got %T: %v", err, err)
	}
	if ce.Code() != connectrpc.CodeInternal {
		t.Fatalf("code = %v, want CodeInternal", ce.Code())
	}
}

func TestBidiPipe(t *testing.T) {
	cli, stop := startServer(t)
	defer stop()

	stream := cli.Pipe(context.Background())
	for i := 0; i < 3; i++ {
		payload := fmt.Sprintf("m%d", i)
		if err := stream.Send(&testpb.PipeRequest{Data: []byte(payload)}); err != nil {
			t.Fatalf("Send[%d]: %v", i, err)
		}
		resp, err := stream.Receive()
		if err != nil {
			t.Fatalf("Receive[%d]: %v", i, err)
		}
		if got := string(resp.GetData()); got != payload {
			t.Fatalf("echo[%d] = %q, want %q", i, got, payload)
		}
	}
	if err := stream.CloseRequest(); err != nil {
		t.Fatalf("CloseRequest: %v", err)
	}
	// Drain until the server closes its send side (EOF).
	for {
		_, err := stream.Receive()
		if err != nil {
			break
		}
	}
	if err := stream.CloseResponse(); err != nil {
		t.Fatalf("CloseResponse: %v", err)
	}
}

// TestUnaryJSONAndProto (F2) verifies that the SAME server accepts BOTH a
// JSON client (Content-Type application/json) and a proto client (Content-Type
// application/proto). The proto client uses connect-go's DEFAULT codec
// (application/proto — see connect-go newClientConfig); the JSON client
// injects connect.WithProtoJSON. Both must succeed and return the same data.
// This is the core F2 verification: a Connect server MUST accept both wire
// formats, selected by Content-Type — which only works once NewServer stops
// forcing a single codec (the F2 server fix).
func TestUnaryJSONAndProto(t *testing.T) {
	const want = "both-codecs"

	// proto client: connect-go's default codec is protoBinaryCodec, so
	// startServer's client already speaks application/proto.
	protoCli, protoStop := startServer(t)
	defer protoStop()
	resp, err := protoCli.Raw(context.Background(), connectrpc.NewRequest(&testpb.RawRequest{
		ContentType: "application/octet-stream",
		Data:        []byte(want),
	}))
	if err != nil {
		t.Fatalf("proto client Raw: %v", err)
	}
	if string(resp.Msg.Data) != want {
		t.Fatalf("proto client data = %q, want %q", string(resp.Msg.Data), want)
	}

	// json client: same test service, but the client forces application/json
	// via WithProtoJSON, injected through WithClientOptions on our dial.
	jsonCli, jsonStop := startClientWithCodec(t, connectrpc.WithProtoJSON())
	defer jsonStop()
	jresp, err := jsonCli.Raw(context.Background(), connectrpc.NewRequest(&testpb.RawRequest{
		ContentType: "application/octet-stream",
		Data:        []byte(want),
	}))
	if err != nil {
		t.Fatalf("json client Raw: %v", err)
	}
	if string(jresp.Msg.Data) != want {
		t.Fatalf("json client data = %q, want %q", string(jresp.Msg.Data), want)
	}
}

// startClientWithCodec brings up a loopback Connect server (h2c) with the test
// service and a client whose connect client options include the given
// connect-go ClientOptions (e.g. WithProtoJSON to force application/json).
func startClientWithCodec(t *testing.T, extraClientOpts ...connectrpc.ClientOption) (testpbconnect.TransferServiceClient, func()) {
	t.Helper()
	srv := connect.NewServer(connect.Address("127.0.0.1:0"), connect.Timeout(0))
	testpb.RegisterTransferServiceConnectServer(srv, testService{})
	ep, err := srv.Endpoint()
	if err != nil {
		t.Fatalf("endpoint: %v", err)
	}
	ctx := context.Background()
	go func() { _ = srv.Start(ctx) }()
	waitReady(t, ep.Host)

	cli, err := connect.Dial(ctx,
		connect.WithEndpoint(ep.Host),
		connect.WithH2C(true),
		connect.WithClientOptions(extraClientOpts...),
	)
	if err != nil {
		_ = srv.Stop(ctx)
		t.Fatalf("dial: %v", err)
	}
	client := testpbconnect.NewTransferServiceClient(cli.HTTPClient(), cli.BaseURL(), cli.ClientOptions()...)
	return client, func() {
		_ = cli.Close()
		_ = srv.Stop(ctx)
	}
}

// TestServerStreamCancel (F3 / lifecycle) verifies that cancelling the client
// context tears down a server-streaming RPC: after cancel, stream.Receive()
// returns a non-nil error (the client observes cancellation). The server's
// handler loop will also unblock because the send errors.
func TestServerStreamCancel(t *testing.T) {
	cli, stop := startServer(t)
	defer stop()

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := cli.Subscribe(ctx, connectrpc.NewRequest(&testpb.SubscribeRequest{Topic: "cancel"}))
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// Receive at least one event to confirm the stream is live.
	if !stream.Receive() {
		t.Fatalf("expected at least one event before cancel, got stream end: %v", stream.Err())
	}
	if first := stream.Msg(); first.GetEventId() == "" {
		t.Fatalf("first event id empty")
	}

	// Cancel the RPC from the client side.
	cancel()

	// After cancel, Receive() must surface an error (context canceled or the
	// wrapped transport error). Loop a few times to drain; the terminal call
	// must return false and Err() must be non-nil.
	var lastErr error
	for i := 0; i < 10; i++ {
		if !stream.Receive() {
			lastErr = stream.Err()
			break
		}
	}
	if lastErr == nil {
		t.Fatalf("expected non-nil error after cancel, got nil (stream ended cleanly)")
	}
}

// TestErrorMapping (errors) verifies that a kratos error returned from the
// service reaches the caller as a kratos error again: the server maps it onto
// the wire via ErrorToConnect, the client unary interceptor normalizes back
// via ConnectToError — Code = HTTP status, reason recovered from the
// ErrorInfo detail. We trigger kerrors.NotFound via the "ERR" sentinel in Raw.
func TestErrorMapping(t *testing.T) {
	cli, stop := startServer(t)
	defer stop()

	_, err := cli.Raw(context.Background(), connectrpc.NewRequest(&testpb.RawRequest{
		Data: []byte("ERR"),
	}))
	if err == nil {
		t.Fatalf("expected error from Raw(ERR), got nil")
	}

	var ke *kerrors.Error
	if !errors.As(err, &ke) {
		t.Fatalf("expected *kerrors.Error (normalized at the client boundary), got %T: %v", err, err)
	}
	if ke.Code != http.StatusNotFound {
		t.Fatalf("kratos Code = %d, want 404", ke.Code)
	}
	if ke.Reason != "NOT_FOUND" {
		t.Fatalf("reason = %q, want NOT_FOUND (round-tripped via the ErrorInfo detail)", ke.Reason)
	}
}

// TestPerMessageStreamMiddleware (F3) verifies that stream middleware
// registered via UseStream is invoked once PER MESSAGE (not once per RPC).
// We register a counting middleware on Subscribe, run a Subscribe that yields
// 5 events, and assert the middleware ran 5 times. This exercises the
// per-message middlewareStream wiring end-to-end.
func TestPerMessageStreamMiddleware(t *testing.T) {
	var count int
	mw := func(handler kratosmiddleware.Handler) kratosmiddleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			count++
			return handler(ctx, req)
		}
	}

	srv := connect.NewServer(connect.Address("127.0.0.1:0"), connect.Timeout(0))
	srv.UseStream("/connecttest.v1.TransferService/Subscribe", mw)
	testpb.RegisterTransferServiceConnectServer(srv, testService{})

	ep, err := srv.Endpoint()
	if err != nil {
		t.Fatalf("endpoint: %v", err)
	}
	ctx := context.Background()
	go func() { _ = srv.Start(ctx) }()
	waitReady(t, ep.Host)

	cli, err := connect.Dial(ctx, connect.WithEndpoint(ep.Host), connect.WithH2C(true))
	if err != nil {
		_ = srv.Stop(ctx)
		t.Fatalf("dial: %v", err)
	}
	defer func() {
		_ = cli.Close()
		_ = srv.Stop(ctx)
	}()

	client := testpbconnect.NewTransferServiceClient(cli.HTTPClient(), cli.BaseURL(), cli.ClientOptions()...)

	stream, err := client.Subscribe(ctx, connectrpc.NewRequest(&testpb.SubscribeRequest{Topic: "mw"}))
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	got := 0
	for stream.Receive() {
		got++
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream error: %v", err)
	}
	if got != 5 {
		t.Fatalf("received %d events, want 5", got)
	}
	if count != got {
		t.Fatalf("stream middleware invoked %d times, want %d (per-message)", count, got)
	}
}
