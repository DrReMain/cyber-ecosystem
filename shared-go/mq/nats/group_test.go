package nats

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

// Same group on two subscribers → each published message delivered exactly once total.
func TestGroupCompetingWorkQueue(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic, group := uniqTopic(t, "compete"), "workers"

	var a, b atomic.Int32
	mk := func(c *atomic.Int32) func(context.Context, mq.Message) error {
		return func(context.Context, mq.Message) error { c.Add(1); return nil }
	}
	s1, _ := m.Consumer.Subscribe(ctx, topic, group, mk(&a))
	s2, _ := m.Consumer.Subscribe(ctx, topic, group, mk(&b))
	defer func() { _ = s1.Close(); _ = s2.Close() }()
	time.Sleep(500 * time.Millisecond) // let both bind the durable

	for i := range 10 {
		_, _ = m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte{byte(i)}})
	}
	time.Sleep(3 * time.Second) // drain

	total := a.Load() + b.Load()
	if total != 10 {
		t.Fatalf("competing: total deliveries=%d (a=%d b=%d), want 10 (each msg once)", total, a.Load(), b.Load())
	}
}

// Different groups → each gets every message.
func TestGroupBroadcast(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic := uniqTopic(t, "bcast")

	var g1, g2 atomic.Int32
	wait := func(c *atomic.Int32, n int32, timeout time.Duration) bool {
		deadline := time.After(timeout)
		for c.Load() < n {
			select {
			case <-time.After(100 * time.Millisecond):
			case <-deadline:
				return false
			}
		}
		return true
	}
	s1, _ := m.Consumer.Subscribe(ctx, topic, "bcast-a", func(context.Context, mq.Message) error { g1.Add(1); return nil })
	s2, _ := m.Consumer.Subscribe(ctx, topic, "bcast-b", func(context.Context, mq.Message) error { g2.Add(1); return nil })
	defer func() { _ = s1.Close(); _ = s2.Close() }()
	time.Sleep(500 * time.Millisecond)

	for range 3 {
		_, _ = m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte("x")})
	}
	if !wait(&g1, 3, 5*time.Second) || !wait(&g2, 3, 5*time.Second) {
		t.Fatalf("broadcast: g1=%d g2=%d, want each 3", g1.Load(), g2.Load())
	}
}

// A durable consumer acks msg 1 + closes; a later subscriber on the same
// durable resumes — gets subsequent messages and does NOT re-receive msg 1.
func TestDurableResume(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic, group := uniqTopic(t, "dur"), "resume-group"

	// msg 1: publish + consume (ack) with s1, then close s1.
	_, _ = m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte{1}})
	got1 := make(chan struct{}, 1)
	s1, _ := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, mm mq.Message) error {
		if len(mm.Payload) > 0 && mm.Payload[0] == 1 {
			select {
			case got1 <- struct{}{}:
			default:
			}
		}
		return nil
	})
	select {
	case <-got1:
	case <-time.After(5 * time.Second):
		t.Fatal("s1 did not get msg 1")
	}
	_ = s1.Close()
	time.Sleep(800 * time.Millisecond) // let the ack register on the durable

	// msgs 2,3 published AFTER s1 closed (queue in the stream).
	_, _ = m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte{2}})
	_, _ = m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte{3}})

	// s2 (same durable) resumes: should get 2,3 — NOT re-deliver acked msg 1.
	var seen []byte
	done := make(chan struct{}, 1) // buffered: decouple handler signal from select receiver
	s2, _ := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, mm mq.Message) error {
		seen = append(seen, mm.Payload...)
		if len(seen) >= 2 {
			select {
			case done <- struct{}{}:
			default:
			}
		}
		return nil
	})
	defer func() { _ = s2.Close() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
	if len(seen) != 2 || seen[0] != 2 || seen[1] != 3 {
		t.Fatalf("durable resume: s2 saw %v, want [2 3] (msg 1 was acked by s1)", seen)
	}
}
