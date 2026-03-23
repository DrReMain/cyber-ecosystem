package connect

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	template2V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template2/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestJSONCodecUsesCamelCase(t *testing.T) {
	srv := NewServer(Address("127.0.0.1:0"))
	srv.Register(
		"/acme.reading.v1.ReadingService/Get",
		connectrpc.NewUnaryHandlerSimple(
			"/acme.reading.v1.ReadingService/Get",
			func(context.Context, *emptypb.Empty) (*template2V1.GetBlogResponse, error) {
				return &template2V1.GetBlogResponse{
					Id:           "blog-1",
					ReadingCount: 7,
				}, nil
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
	req, err := http.NewRequest(http.MethodPost, "http://"+u.Host+"/acme.reading.v1.ReadingService/Get", bytes.NewBufferString(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connect-Protocol-Version", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	// After migration, we use camelCase field names
	if !strings.Contains(s, `"readingCount":`) {
		t.Fatalf("response should use camelCase field names, got %s", s)
	}
	if strings.Contains(s, `"reading_count":`) {
		t.Fatalf("response should not use snake_case proto names, got %s", s)
	}
}
