package nats

import (
	"context"
	"sync"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

func TestPublishSubscribeRoundTrip(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic, group := uniqTopic(t, "rt"), "rt-group"

	var (
		mu   sync.Mutex
		got  *mq.Message
		done = make(chan struct{}, 1) // buffered: decouple the handler's signal from the receiver waiting in a select
	)
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

	id, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte("hello mq"), Headers: map[string]string{"k": "v"}})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if id == "" {
		t.Errorf("empty publish id")
	}
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("did not receive message within 10s")
	}
	mu.Lock()
	defer mu.Unlock()
	if got == nil || string(got.Payload) != "hello mq" || got.Headers["k"] != "v" {
		t.Errorf("received: %+v", got)
	}
}
