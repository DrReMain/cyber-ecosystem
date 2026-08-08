// Tests verifying where the kratos transport ReplyHeader lands on the ERROR
// path — the known transport-parity discrepancy.
//
// kratos exposes only ReplyHeader() (no ReplyTrailer). On a handler error:
//   - grpc/http send ReplyHeader as a response header (kratos grpc does
//     grpc.SetHeader unconditionally; http writes w.Header()).
//   - connect attaches ReplyHeader to the *connect.Error's Meta() — a unary
//     error has no response header to write to (see connect interceptor.go
//     attachReplyHeadersToConnectError).
//
// So a client reading response headers after an error gets them on grpc/http
// but must read the *connect.Error.Meta() on connect. These tests pin the
// connect behavior; the grpc/http "response header" behavior is established in
// code (kratos transport/grpc/interceptor.go, transport/http).
package integration

import (
	"context"
	"errors"
	"testing"

	connectrpc "connectrpc.com/connect"
	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/transport"
	"google.golang.org/genproto/googleapis/api/httpbody"

	testpb "cyber-ecosystem/shared-go/kratos/transport/connect/testpb"
	testpbconnect "cyber-ecosystem/shared-go/kratos/transport/connect/testpb/testpbconnect"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
)

// replyHeaderErrSvc sets a reply header then returns an error.
type replyHeaderErrSvc struct{ testService }

func (replyHeaderErrSvc) Raw(ctx context.Context, _ *testpb.RawRequest) (*httpbody.HttpBody, error) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		tr.ReplyHeader().Add("x-custom-reply", "rv")
	}
	return nil, kerrors.NotFound("X", "boom")
}

// TestReplyHeaderOnErrorConnectUnary verifies that on a UNARY error the connect
// transport attaches the ReplyHeader to the returned *connect.Error's Meta()
// (a unary error has no response header). This diverges from grpc/http, which
// send ReplyHeader as a response header even on error.
//
// The ReplyHeader lands on the *connect.Error's Meta() at the WIRE level. The
// kratos unary interceptor that normalizes errors to *kerrors.Error drops
// Meta() by design, so this contract is visible only to raw connect clients —
// built without cli.ClientOptions() — and to non-kratos connect peers.
func TestReplyHeaderOnErrorConnectUnary(t *testing.T) {
	srv := connect.NewServer(connect.Address("127.0.0.1:0"), connect.Timeout(0))
	testpb.RegisterTransferServiceConnectServer(srv, replyHeaderErrSvc{})

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

	raw := testpbconnect.NewTransferServiceClient(cli.HTTPClient(), cli.BaseURL())
	_, err = raw.Raw(ctx, connectrpc.NewRequest(&testpb.RawRequest{Data: []byte("x")}))
	if err == nil {
		t.Fatal("expected an error from Raw")
	}
	var ce *connectrpc.Error
	if !errors.As(err, &ce) {
		t.Fatalf("expected *connect.Error, got %T: %v", err, err)
	}
	if got := ce.Meta().Get("X-Custom-Reply"); got != "rv" {
		t.Fatalf("unary error: connect attached ReplyHeader to error.Meta = %q, want %q", got, "rv")
	}
}
