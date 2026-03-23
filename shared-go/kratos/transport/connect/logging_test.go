package connect

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	template1V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"
	template1V1connect "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1/template1V1connect"
	"github.com/go-kratos/kratos/v2/log"
	klogging "github.com/go-kratos/kratos/v2/middleware/logging"
)

func TestClientLoggingArgsUseProtoRequest(t *testing.T) {
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
	l := &captureLogger{}
	cc, err := DialInsecure(
		context.Background(),
		WithEndpoint(u.Host),
		WithMiddleware(klogging.Client(l)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()

	client := template1V1connect.NewBlogServiceClient(cc.HTTPClient(), cc.Endpoint(), cc.ClientOptions()...)
	_, err = client.GetBlog(context.Background(), &template1V1.GetBlogRequest{Id: "blog-1"})
	if err != nil {
		t.Fatal(err)
	}

	args := l.get("args")
	if strings.Contains(args, "Method:POST") {
		t.Fatalf("args should not be raw http request: %s", args)
	}
	if !strings.Contains(args, "blog-1") {
		t.Fatalf("args should contain request payload: %s", args)
	}
}

func TestServerLoggingArgsUseProtoRequest(t *testing.T) {
	l := &captureLogger{}
	srv := NewServer(Address("127.0.0.1:0"), Middleware(klogging.Server(l)))
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
	_, err = client.GetBlog(context.Background(), &template1V1.GetBlogRequest{Id: "blog-2"})
	if err != nil {
		t.Fatal(err)
	}

	args := l.get("args")
	if args == "" {
		t.Fatal("server args should not be empty")
	}
	if !strings.Contains(args, "blog-2") {
		t.Fatalf("args should contain request payload: %s", args)
	}
}

type captureLogger struct {
	mu     sync.Mutex
	values map[string]string
}

func (l *captureLogger) Log(_ log.Level, keyvals ...any) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.values == nil {
		l.values = make(map[string]string)
	}
	for i := 0; i+1 < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}
		l.values[key] = toString(keyvals[i+1])
	}
	return nil
}

func (l *captureLogger) get(key string) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.values[key]
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	default:
		return fmt.Sprint(x)
	}
}
