package connect

import (
	"context"
	"testing"
)

func TestClientBaseURL(t *testing.T) {
	c, err := Dial(context.Background(), WithEndpoint("127.0.0.1:9999"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if c.BaseURL() == "" {
		t.Fatal("BaseURL() is empty")
	}
	if c.HTTPClient() == nil {
		t.Fatal("HTTPClient() is nil")
	}
}

func TestClientOptionsNonNil(t *testing.T) {
	c, err := Dial(context.Background(), WithEndpoint("127.0.0.1:9999"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if opts := c.ClientOptions(); len(opts) == 0 {
		t.Fatal("ClientOptions() empty; client interceptors (Transport injection + middleware) must be wired")
	}
}
