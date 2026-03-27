package traceheader

import (
	"context"
	"testing"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-kratos/kratos/v2/transport"
)

type mockHeader map[string][]string

func (h mockHeader) Get(key string) string {
	values := h[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (h mockHeader) Set(key, value string) {
	h[key] = []string{value}
}

func (h mockHeader) Add(key, value string) {
	h[key] = append(h[key], value)
}

func (h mockHeader) Keys() []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}

func (h mockHeader) Values(key string) []string {
	return h[key]
}

type mockTransporter struct {
	replyHeader mockHeader
}

func (m *mockTransporter) Kind() transport.Kind            { return transport.KindHTTP }
func (m *mockTransporter) Endpoint() string                { return "http://localhost" }
func (m *mockTransporter) Operation() string               { return "/test.op" }
func (m *mockTransporter) RequestHeader() transport.Header { return mockHeader{} }
func (m *mockTransporter) ReplyHeader() transport.Header   { return m.replyHeader }

func TestServer_SetsTraceIDHeader(t *testing.T) {
	tp := tracesdk.NewTracerProvider()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tr := &mockTransporter{replyHeader: mockHeader{}}
	ctx := transport.NewServerContext(context.Background(), tr)
	ctx, span := tp.Tracer("test").Start(ctx, "op")
	defer span.End()

	mw := Server()
	_, err := mw(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})(ctx, nil)
	if err != nil {
		t.Fatalf("middleware should not return error: %v", err)
	}

	got := tr.replyHeader.Get(HeaderTraceID)
	if got == "" {
		t.Fatalf("expected %s to be set", HeaderTraceID)
	}
	if _, parseErr := trace.TraceIDFromHex(got); parseErr != nil {
		t.Fatalf("trace id should be valid hex: %v", parseErr)
	}
}

func TestServer_NoSpanContextDoesNotSetHeader(t *testing.T) {
	tr := &mockTransporter{replyHeader: mockHeader{}}
	ctx := transport.NewServerContext(context.Background(), tr)

	mw := Server()
	_, err := mw(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})(ctx, nil)
	if err != nil {
		t.Fatalf("middleware should not return error: %v", err)
	}

	if got := tr.replyHeader.Get(HeaderTraceID); got != "" {
		t.Fatalf("did not expect %s, got %q", HeaderTraceID, got)
	}
}
