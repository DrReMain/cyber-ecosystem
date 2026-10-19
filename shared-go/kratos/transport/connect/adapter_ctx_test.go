package connect_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	connectrpc "connectrpc.com/connect"
	"github.com/go-kratos/kratos/v3/middleware"
	"google.golang.org/grpc"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	testpb "cyber-ecosystem/shared-go/kratos/transport/connect/testpb"
	testpbconnect "cyber-ecosystem/shared-go/kratos/transport/connect/testpb/testpbconnect"
)

type ctxInjectKey struct{}

// Regression: unary-middleware ctx injections (Subject, transport headers) must
// reach the streaming handler through stream.Context(). The connect legs once
// built the stream wrapper with the pre-chain context, so handlers saw a bare
// ctx and SubjectFromCtx failed — mirroring the http generators, which pass the
// enriched context via stream.SetContext.
func TestServerStreamHandlerSeesMiddlewareContext(t *testing.T) {
	srv := connect.NewServer(connect.Middleware(func(h middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			return h(context.WithValue(ctx, ctxInjectKey{}, "subject-injected"), req)
		}
	}))
	connect.HandleServerStream(srv, "/connecttest.v1.TransferService/Subscribe",
		func(in *testpb.SubscribeRequest, stream grpc.ServerStreamingServer[testpb.SubscribeResponse]) error {
			v, _ := stream.Context().Value(ctxInjectKey{}).(string)
			return stream.Send(&testpb.SubscribeResponse{EventId: v})
		})
	ts := httptest.NewServer(http.HandlerFunc(srv.ServeHTTP))
	defer ts.Close()

	client := testpbconnect.NewTransferServiceClient(ts.Client(), ts.URL)
	stream, err := client.Subscribe(context.Background(), connectrpc.NewRequest(&testpb.SubscribeRequest{Topic: "ctx"}))
	if err != nil {
		t.Fatalf("Subscribe open: %v", err)
	}
	var got string
	for stream.Receive() {
		got = stream.Msg().GetEventId()
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("Subscribe stream: %v", err)
	}
	if got != "subject-injected" {
		t.Fatalf("stream.Context() missed the unary-middleware injection: got %q", got)
	}
}
