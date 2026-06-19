package pg

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

var errBoom = errBoomType{}

type errBoomType struct{}

func (errBoomType) Error() string { return "boom" }

// A handler that always fails is retried up to MaxRetries (3 in testConfig), then
// dead-lettered (inserted into the dlq table, delivery removed atomically).
func TestConsumerRetryThenDLQ(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	m := New(h)
	ctx := context.Background()
	topic, group := uniqTopic(t, "retry"), "retry-group"

	var attempts atomic.Int32
	sub, err := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, _ mq.Message) error {
		attempts.Add(1)
		return errBoom
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = sub.Close() }()

	if _, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte("poison")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	deadline := time.After(15 * time.Second)
	for attempts.Load() < 3 { // testConfig MaxRetries=3
		select {
		case <-time.After(100 * time.Millisecond):
		case <-deadline:
			t.Fatalf("expected >=3 attempts, got %d", attempts.Load())
		}
	}
	time.Sleep(500 * time.Millisecond) // let the DLQ insert settle

	var n int
	if err := h.pool.QueryRow(ctx, `SELECT count(*) FROM dlq WHERE topic=$1 AND payload='poison'`, topic).Scan(&n); err != nil {
		t.Fatalf("dlq query: %v", err)
	}
	if n != 1 {
		t.Fatalf("dlq count=%d, want 1", n)
	}
}
