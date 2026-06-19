package nats

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	natsclient "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"cyber-ecosystem/shared-go/mq"
)

// mapError classifies server-side timeouts (nats ErrTimeout, ctx deadline, a
// JetStream 504 API error) as ErrTimeout, and connection-class errors as
// ErrUnavailable.
func TestMapErrorTimeout(t *testing.T) {
	cases := []struct {
		name string
		in   error
		want error
	}{
		{"nats timeout", natsclient.ErrTimeout, mq.ErrTimeout},
		{"ctx deadline", context.DeadlineExceeded, mq.ErrTimeout},
		{"504 api error", &jetstream.APIError{Code: 504}, mq.ErrTimeout},
		{"no responders", natsclient.ErrNoResponders, mq.ErrUnavailable},
		{"conn closed", natsclient.ErrConnectionClosed, mq.ErrUnavailable},
	}
	for _, c := range cases {
		if !errors.Is(mapError(c.in, "op"), c.want) {
			t.Errorf("%s: mapError(%v) not %v", c.name, c.in, c.want)
		}
	}
	// a non-timeout API error must NOT be classified as Timeout.
	if errors.Is(mapError(&jetstream.APIError{Code: 500}, "op"), mq.ErrTimeout) {
		t.Error("500 APIError should not map to ErrTimeout")
	}
}

// The created consumer must carry the configured MaxAckPending (bounds in-flight
// memory + redelivery), not the server default of 1000.
func TestConsumerMaxAckPendingSet(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	topic := uniqTopic(t, "ackpend")
	sub, err := m.Consumer.Subscribe(context.Background(), topic, "g", func(context.Context, mq.Message) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = sub.Close() }()

	h, _, _ := NewClient(testConfig())
	defer func() { _ = h.nc.Drain() }()
	cons, err := h.js.Consumer(context.Background(), streamName(topic), "g-"+topic)
	if err != nil {
		t.Fatalf("consumer lookup: %v", err)
	}
	info, err := cons.Info(context.Background())
	if err != nil {
		t.Fatalf("consumer info: %v", err)
	}
	if want := int32(10); int32(info.Config.MaxAckPending) != want { // testConfig MaxAckPending=10
		t.Fatalf("MaxAckPending: got %d, want %d", info.Config.MaxAckPending, want)
	}
}

// After Close, no further handler callbacks run (graceful drain, then stopped).
func TestSubscriptionCloseStopsCallbacks(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	topic := uniqTopic(t, "close")
	var n atomic.Int32
	sub, err := m.Consumer.Subscribe(context.Background(), topic, "g", func(context.Context, mq.Message) error {
		n.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if _, err := m.Publisher.Publish(context.Background(), topic, &mq.Message{Payload: []byte("a")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	deadline := time.After(5 * time.Second)
	for n.Load() == 0 {
		select {
		case <-time.After(50 * time.Millisecond):
		case <-deadline:
			t.Fatal("no initial delivery")
		}
	}
	if err := sub.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	before := n.Load()
	for range 5 {
		_, _ = m.Publisher.Publish(context.Background(), topic, &mq.Message{Payload: []byte("b")})
	}
	time.Sleep(700 * time.Millisecond)
	if got := n.Load(); got != before {
		t.Fatalf("callback fired after Close: before=%d after=%d", before, got)
	}
}

// decideAck encodes the ack/Nak/Term policy. The DLQ-failure branch (→ Nak,
// retain) and the meta-nil guard (→ Term, no infinite loop) are covered here
// directly because a NATS-side DLQ failure isn't reliably triggerable.
func TestDecideAck(t *testing.T) {
	const maxRetries = 3
	mkMeta := func(delivered uint64) *jetstream.MsgMetadata { return &jetstream.MsgMetadata{NumDelivered: delivered} }
	dlqFail := errors.New("dlq write failed")
	cases := []struct {
		name   string
		meta   *jetstream.MsgMetadata
		herr   error
		dlqErr error
		want   ackDecision
	}{
		{"success → ack", mkMeta(1), nil, nil, ackMsg},
		{"first attempt fails → nak (retry)", mkMeta(1), errBoom, nil, nakMsg},
		{"mid attempts fail → nak (retry)", mkMeta(2), errBoom, nil, nakMsg},
		{"cap reached, DLQ ok → term (poison isolated)", mkMeta(maxRetries), errBoom, nil, termMsg},
		{"cap reached, DLQ failed → nak (retain, no silent loss)", mkMeta(maxRetries), errBoom, dlqFail, nakMsg},
		{"over cap, DLQ failed → nak (retain)", mkMeta(maxRetries + 2), errBoom, dlqFail, nakMsg},
		{"meta nil + error → term (no infinite loop)", nil, errBoom, nil, termMsg},
		{"meta nil + success → ack", nil, nil, nil, ackMsg},
	}
	for _, c := range cases {
		if got := decideAck(c.meta, c.herr, maxRetries, c.dlqErr); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
