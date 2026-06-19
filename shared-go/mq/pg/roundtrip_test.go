package pg

import (
	"context"
	"sync"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

func TestRoundTrip(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	m := New(h)
	ctx := context.Background()
	topic, group := uniqTopic(t, "rt"), "rt-group"

	var mu sync.Mutex
	var got *mq.Message
	done := make(chan struct{}, 1)
	sub, err := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, msg mq.Message) error {
		mu.Lock()
		got = &msg
		mu.Unlock()
		select {
		case done <- struct{}{}:
		default:
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = sub.Close() }()

	if _, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte("hello pg"), Headers: map[string]string{"k": "v"}}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("did not receive within 5s")
	}
	mu.Lock()
	defer mu.Unlock()
	if got == nil || string(got.Payload) != "hello pg" || got.Headers["k"] != "v" {
		t.Fatalf("got %+v", got)
	}
}
