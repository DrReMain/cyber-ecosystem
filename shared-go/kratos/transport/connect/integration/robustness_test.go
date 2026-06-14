// Robustness / edge-case tests for the Connect transport.
//
// These complement integration_test.go by exercising behavior-critical paths
// that were previously uncovered:
//   - TestHeaderTrailerPropagation (F10): gRPC-style SetHeader/SetTrailer on a
//     server-stream handler round-trip to the client via ResponseHeader() /
//     ResponseTrailer().
//   - TestUnaryTimeout: connect.Timeout on the server enforces a real unary
//     deadline that surfaces to the client.
//   - TestLargeMessageRoundTrip: a 1 MiB payload survives the codec + framing.
//   - TestConcurrentStreams: N concurrent server-streams on one client don't
//     corrupt per-stream state (run with -race).
//   - TestConnectToErrorE2E: a wire-received *connect.Error round-trips through
//     ConnectToError into a kratos *errors.Error with the right Reason/Code.
package integration

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	kerrors "github.com/go-kratos/kratos/v3/errors"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	connect "cyber-ecosystem/shared-go/kratos/transport/connect"

	mobilepb "cyber-ecosystem/gen/go/cyber/mobile/v1"
	mobilev1connect "cyber-ecosystem/gen/go/cyber/mobile/v1/v1connect"
	v1 "cyber-ecosystem/gen/go/cyber/transfer/v1"
)

// headerService is a testService variant whose server-stream Subscribe sets a
// custom response header before the first Send and a custom trailer before
// returning. It exists solely for TestHeaderTrailerPropagation; it shares the
// other method bodies with testService via embedding.
type headerService struct {
	testService
}

// Subscribe sets x-custom-header before sending any event and x-custom-trailer
// before returning, exercising serverStream.SetHeader/SetTrailer -> Connect
// ResponseHeader/ResponseTrailer propagation (the F10 gap).
func (headerService) Subscribe(req *v1.SubscribeRequest, stream grpc.ServerStreamingServer[v1.SubscribeResponse]) error {
	// Buffer a header; it is flushed on the first Send (see serverStream.SendMsg
	// -> flushHeader).
	if err := stream.SetHeader(metadata.Pairs("x-custom-header", "hv")); err != nil {
		return err
	}
	for i := 0; i < 5; i++ {
		if err := stream.Send(&v1.SubscribeResponse{
			EventId: bytesHeaderEventID(req.GetTopic(), i),
		}); err != nil {
			return err
		}
	}
	// Buffer a trailer; flushTrailer runs via defer in HandleServerStream when
	// the handler returns.
	stream.SetTrailer(metadata.Pairs("x-custom-trailer", "tv"))
	return nil
}

// bytesHeaderEventID builds the event id used by headerService so the message
// count check reuses the same "topic-N" convention as testService.
func bytesHeaderEventID(topic string, i int) string {
	return topic + "-" + itoa(i+1)
}

// itoa is a tiny dependency-free int->string (avoids pulling fmt into a hot
// loop body unnecessarily; fmt is fine here but kept explicit for clarity).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// startServerWithService is the generalized form of startServer: it registers
// an arbitrary TransferServiceServer implementation. Used for the header/trailer
// test (headerService) and any other variant that needs a custom handler.
func startServerWithService(t *testing.T, svc mobilepb.MobileTransferServiceServer, opts ...connect.ServerOption) (mobilev1connect.MobileTransferServiceClient, func()) {
	t.Helper()

	serverOpts := append([]connect.ServerOption{connect.Address("127.0.0.1:0"), connect.Timeout(0)}, opts...)
	srv := connect.NewServer(serverOpts...)
	mobilepb.RegisterMobileTransferServiceConnectServer(srv, svc)

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
	client := mobilev1connect.NewMobileTransferServiceClient(cli.HTTPClient(), cli.BaseURL(), cli.ClientOptions()...)

	stop := func() {
		_ = cli.Close()
		_ = srv.Stop(ctx)
	}
	return client, stop
}

// TestHeaderTrailerPropagation (F10) verifies that a server-stream handler's
// gRPC-style SetHeader/SetTrailer are flushed through to the connect-go client:
//   - ResponseHeader() must contain x-custom-header == "hv" (set before first Send).
//   - ResponseTrailer() must contain x-custom-trailer == "tv" (set before return).
//
// connect-go's ServerStreamForClient.ResponseHeader() blocks until the first
// Receive() returns, and ResponseTrailer() is only fully populated after the
// stream ends (Receive returns an EOF-wrapping error). So we Receive once to
// unblock the header, then drain to EOF before reading the trailer.
func TestHeaderTrailerPropagation(t *testing.T) {
	cli, stop := startServerWithService(t, headerService{})
	defer stop()

	stream, err := cli.Subscribe(context.Background(), connectrpc.NewRequest(&v1.SubscribeRequest{Topic: "hdr"}))
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// First Receive unblocks ResponseHeader() on the client side.
	if !stream.Receive() {
		t.Fatalf("first Receive returned false before any data: %v", stream.Err())
	}

	// Header keys are canonicalized by http.Header; read canonical form.
	hdr := stream.ResponseHeader()
	if got := hdr.Get("X-Custom-Header"); got != "hv" {
		t.Fatalf("ResponseHeader X-Custom-Header = %q, want %q", got, "hv")
	}

	// Drain the rest of the stream so the trailer is fully populated.
	var got int
	for stream.Receive() {
		got++
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream error: %v", err)
	}
	// got counts events AFTER the first one consumed above.
	if got != 4 {
		t.Fatalf("received %d trailing events, want 4 (5 total - 1 consumed)", got)
	}

	// Trailer should now be available after EOF.
	trl := stream.ResponseTrailer()
	if got := trl.Get("X-Custom-Trailer"); got != "tv" {
		t.Fatalf("ResponseTrailer X-Custom-Trailer = %q, want %q", got, "tv")
	}
}

// slowService is a testService whose Raw sleeps 1s before responding. Used to
// trigger the server's unary timeout in TestUnaryTimeout.
type slowService struct {
	testService
}

// Raw sleeps 1 second (longer than the 500ms server timeout) so the
// server-side context deadline fires before the handler returns.
func (slowService) Raw(ctx context.Context, _ *v1.RawRequest) (*httpbody.HttpBody, error) {
	// Wait until the server's unary timeout (set below to 500ms) cancels ctx.
	select {
	case <-time.After(1 * time.Second):
		// Handler completed normally — the timeout did NOT fire. Still return a
		// valid body so the test can fail meaningfully below.
		return &httpbody.HttpBody{ContentType: "application/octet-stream", Data: []byte("late")}, nil
	case <-ctx.Done():
		// Surface the deadline/cancellation as-is; the transport maps it via
		// ErrorToConnect to a connect code.
		return nil, ctx.Err()
	}
}

// TestUnaryTimeout verifies that connect.Timeout on the server enforces a real
// unary deadline: a handler that outlives it is killed and the client observes
// an error well before the handler would have returned.
//
// Surfaced code: the server's unary interceptor wraps ctx with a WithTimeout
// (see interceptor.go WrapUnary), so the slow handler returns
// context.DeadlineExceeded. ErrorToConnect's context-error branch maps that to
// connect.CodeDeadlineExceeded — matching what a kratos gRPC server surfaces
// (gRPC-go's status.FromContextError maps a bare ctx error to
// codes.DeadlineExceeded). This 3-protocol parity is pinned by
// comparative.TestTimeoutMappingGRPCvsConnect.
//
// What this test proves: (1) the timeout fires (elapsed << the handler's 1s
// sleep), (2) the failure reaches the client as a *connect.Error, and (3) the
// surfaced code is DeadlineExceeded (parity with gRPC).
func TestUnaryTimeout(t *testing.T) {
	// 500ms server timeout; handler sleeps 1s.
	cli, stop := startServerWithService(t, slowService{}, connect.Timeout(500*time.Millisecond))
	defer stop()

	start := time.Now()
	_, err := cli.Raw(context.Background(), connectrpc.NewRequest(&v1.RawRequest{
		Data: []byte("slow"),
	}))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error from Raw when the server unary timeout fires, got nil")
	}

	var ce *connectrpc.Error
	if !errors.As(err, &ce) {
		t.Fatalf("expected *connect.Error, got %T: %v", err, err)
	}

	// The surfaced code MUST be DeadlineExceeded — matching the kratos gRPC
	// server, which surfaces codes.DeadlineExceeded for the same bare-ctx.Err()
	// scenario. A regression to CodeInternal would mean ErrorToConnect's
	// context-error branch regressed and Connect would diverge from gRPC again.
	if ce.Code() != connectrpc.CodeDeadlineExceeded {
		t.Fatalf("connect code = %v, want CodeDeadlineExceeded (parity with gRPC-go's status.FromContextError)", ce.Code())
	}

	// The client must observe the failure well before the 1s handler would have
	// returned — this is the real proof the timeout fired rather than the
	// handler completing. Bound generously to avoid CI flakes.
	if elapsed > 900*time.Millisecond {
		t.Fatalf("client waited %v for a timeout that should fire near 500ms (handler sleep is 1s)", elapsed)
	}
}

// TestLargeMessageRoundTrip verifies a 1 MiB payload survives the codec + framing
// end-to-end with no truncation or size-related failure.
func TestLargeMessageRoundTrip(t *testing.T) {
	cli, stop := startServer(t)
	defer stop()

	want := bytes.Repeat([]byte("a"), 1<<20) // 1 MiB

	resp, err := cli.Raw(context.Background(), connectrpc.NewRequest(&v1.RawRequest{
		ContentType: "application/octet-stream",
		Data:        want,
	}))
	if err != nil {
		t.Fatalf("Raw (1MiB): %v", err)
	}
	if got := len(resp.Msg.Data); got != len(want) {
		t.Fatalf("data length = %d, want %d", got, len(want))
	}
	// Spot-check first and last bytes (catches off-by-one / head truncation).
	if resp.Msg.Data[0] != 'a' {
		t.Fatalf("first byte = %q, want 'a'", resp.Msg.Data[0])
	}
	if resp.Msg.Data[len(resp.Msg.Data)-1] != 'a' {
		t.Fatalf("last byte = %q, want 'a'", resp.Msg.Data[len(resp.Msg.Data)-1])
	}
	// Full equality as the strongest correctness check.
	if !bytes.Equal(resp.Msg.Data, want) {
		t.Fatalf("data mismatch: bytes not equal to 1MiB of 'a'")
	}
}

// TestConcurrentStreams verifies that N concurrent server-stream RPCs on a
// single client each receive their full 5-event payload with no cross-stream
// corruption. Run under -race to also catch data races in per-stream state.
func TestConcurrentStreams(t *testing.T) {
	const n = 10
	const wantEvents = 5

	cli, stop := startServer(t)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(n)
	errCh := make(chan error, n)

	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			topic := itoa(id)
			stream, err := cli.Subscribe(context.Background(), connectrpc.NewRequest(&v1.SubscribeRequest{Topic: topic}))
			if err != nil {
				errCh <- err
				return
			}
			var got int
			for stream.Receive() {
				msg := stream.Msg()
				// Each event id must be "<id>-<n>"; a mismatch implies cross-stream
				// corruption (e.g. a shared buffer reused across streams).
				expectID := topic + "-" + itoa(got+1)
				if msg.GetEventId() != expectID {
					errCh <- errors.New("goroutine " + itoa(id) + " event " + itoa(got) + " id = " + msg.GetEventId() + ", want " + expectID)
					return
				}
				got++
			}
			if err := stream.Err(); err != nil {
				errCh <- err
				return
			}
			if got != wantEvents {
				errCh <- errors.New("goroutine " + itoa(id) + " received " + itoa(got) + " events, want " + itoa(wantEvents))
				return
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent stream error: %v", err)
	}
}

// errReasonService is a testService whose Raw returns a kratos NotFound with a
// known Reason ("BIG") when the request Data is the sentinel "ERR". Used by
// TestConnectToErrorE2E to verify the wire-received *connect.Error round-trips
// through ConnectToError back into a kratos *errors.Error.
type errReasonService struct {
	testService
}

// Raw returns kerrors.NotFound("BIG", ...) for the "ERR" sentinel so the
// Reason survives the kratos->connect->kratos round trip.
func (errReasonService) Raw(_ context.Context, req *v1.RawRequest) (*httpbody.HttpBody, error) {
	if string(req.GetData()) == "ERR" {
		return nil, kerrors.NotFound("BIG", "nope")
	}
	ct := req.GetContentType()
	if ct == "" {
		ct = "application/octet-stream"
	}
	return &httpbody.HttpBody{ContentType: ct, Data: req.GetData()}, nil
}

// TestConnectToErrorE2E verifies that a *connect.Error actually received over
// the wire can be converted back to a kratos *errors.Error with the original
// Reason and an HTTP code matching NotFound (404). This exercises the REVERSE
// mapping (connect -> kratos) on a real wire error, not a constructed one.
func TestConnectToErrorE2E(t *testing.T) {
	cli, stop := startServerWithService(t, errReasonService{})
	defer stop()

	_, err := cli.Raw(context.Background(), connectrpc.NewRequest(&v1.RawRequest{
		Data: []byte("ERR"),
	}))
	if err == nil {
		t.Fatal("expected error from Raw(ERR), got nil")
	}

	// Round-trip the wire error through ConnectToError.
	ke := connect.ConnectToError(err)
	if ke == nil {
		t.Fatal("ConnectToError returned nil for a non-nil error")
	}
	if ke.Reason != "BIG" {
		t.Fatalf("kratos Reason = %q, want %q", ke.Reason, "BIG")
	}
	if ke.Code != 404 {
		t.Fatalf("kratos Code = %d, want 404 (NotFound)", ke.Code)
	}
}
