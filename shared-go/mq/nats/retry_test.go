package nats

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"cyber-ecosystem/shared-go/mq"
)

var errBoom = errBoomType{}

type errBoomType struct{}

func (errBoomType) Error() string { return "boom" }

// A handler that always fails is retried up to MaxRetries (3 in testConfig),
// then dead-lettered (acked off the original, published to mq-dlq).
func TestConsumerRetryThenDLQ(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic, group := uniqTopic(t, "retry"), "retry-group"
	t.Cleanup(func() { deleteStream(t, dlqStream) }) // isolate the shared DLQ stream too

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

	// wait for retries (testConfig MaxRetries=3)
	deadline := time.After(15 * time.Second)
	for attempts.Load() < 3 {
		select {
		case <-time.After(150 * time.Millisecond):
		case <-deadline:
			t.Fatalf("expected >=3 attempts, got %d", attempts.Load())
		}
	}
	time.Sleep(500 * time.Millisecond) // let DLQ publish + ack settle

	// inspect the DLQ stream for the poison message
	h, _, _ := NewClient(testConfig())
	defer func() { _ = h.nc.Drain() }()
	dlq, err := h.js.Stream(ctx, dlqStream)
	if err != nil {
		t.Fatalf("dlq stream: %v", err)
	}
	dc, err := dlq.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "retry-test-dlq-reader",
		FilterSubject: dlqSubject(topic),
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		t.Fatalf("dlq consumer: %v", err)
	}
	batch, err := dc.Fetch(10, jetstream.FetchMaxWait(3*time.Second))
	if err != nil {
		t.Fatalf("dlq fetch: %v", err)
	}
	var saw bool
	for mm := range batch.Messages() {
		if string(mm.Data()) == "poison" {
			saw = true
		}
		_ = mm.Ack()
	}
	if !saw {
		t.Errorf("DLQ did not contain the poison message")
	}
}
