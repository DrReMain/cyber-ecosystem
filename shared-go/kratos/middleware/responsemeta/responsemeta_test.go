package responsemeta

import (
	"context"
	"net/http"
	"testing"

	connecttransport "github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport"
)

func TestServerSuccessHeaders(t *testing.T) {
	tr := newTestTransport()
	ctx := transport.NewServerContext(context.Background(), tr)
	h := Server()(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	_, err := h(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got := tr.reply.Get(HeaderResponseSuccess); got != "true" {
		t.Fatalf("X-Response-Success = %q", got)
	}
}

func TestServerErrorHeaders(t *testing.T) {
	tr := newTestTransport()
	ctx := transport.NewServerContext(context.Background(), tr)
	h := Server()(func(ctx context.Context, req any) (any, error) {
		return nil, errors.BadRequest("ERROR_REASON_VALIDATOR", "")
	})
	_, err := h(ctx, nil)
	if err == nil {
		t.Fatalf("expected err")
	}
	if got := tr.reply.Get(HeaderResponseSuccess); got != "false" {
		t.Fatalf("X-Response-Success = %q", got)
	}
	if got := tr.reply.Get(HeaderErrorReason); got != "ERROR_REASON_VALIDATOR" {
		t.Fatalf("X-Error-Reason = %q", got)
	}
}

type testTransport struct {
	req   transport.Header
	reply transport.Header
}

func newTestTransport() *testTransport {
	return &testTransport{
		req:   connecttransport.NewHeader(http.Header{}),
		reply: connecttransport.NewHeader(http.Header{}),
	}
}

func (t *testTransport) Kind() transport.Kind            { return transport.KindHTTP }
func (t *testTransport) Endpoint() string                { return "http://test" }
func (t *testTransport) Operation() string               { return "/test.v1.TestService/Test" }
func (t *testTransport) RequestHeader() transport.Header { return t.req }
func (t *testTransport) ReplyHeader() transport.Header   { return t.reply }
