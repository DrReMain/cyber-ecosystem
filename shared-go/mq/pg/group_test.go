package pg

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

// Same group on two subscribers → each published message delivered exactly once total.
func TestGroupCompeting(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	m := New(h)
	ctx := context.Background()
	topic, group := uniqTopic(t, "comp"), "workers"

	var total atomic.Int32
	done := make(chan struct{}, 1)
	mk := func() func(context.Context, mq.Message) error {
		return func(context.Context, mq.Message) error {
			if total.Add(1) == 10 {
				select {
				case done <- struct{}{}:
				default:
				}
			}
			return nil
		}
	}
	s1, err := m.Consumer.Subscribe(ctx, topic, group, mk())
	if err != nil {
		t.Fatalf("s1: %v", err)
	}
	s2, err := m.Consumer.Subscribe(ctx, topic, group, mk())
	if err != nil {
		t.Fatalf("s2: %v", err)
	}
	defer func() { _ = s1.Close() }()
	defer func() { _ = s2.Close() }()
	time.Sleep(200 * time.Millisecond) // let both bind

	for i := range 10 {
		if _, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte{byte(i)}}); err != nil {
			t.Fatalf("Publish %d: %v", i, err)
		}
	}
	select {
	case <-done:
	case <-time.After(10 * time.Second):
	}
	if got := total.Load(); got != 10 {
		t.Fatalf("competing: total=%d, want 10 (each msg once)", got)
	}
}

// Different groups → every group gets every message.
func TestGroupBroadcast(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	m := New(h)
	ctx := context.Background()
	topic := uniqTopic(t, "bcast")

	var g1, g2 atomic.Int32
	wait := func(c *atomic.Int32, n int32) bool {
		deadline := time.After(5 * time.Second)
		for c.Load() < n {
			select {
			case <-time.After(100 * time.Millisecond):
			case <-deadline:
				return false
			}
		}
		return true
	}
	s1, err := m.Consumer.Subscribe(ctx, topic, "bcast-a", func(context.Context, mq.Message) error { g1.Add(1); return nil })
	if err != nil {
		t.Fatalf("s1: %v", err)
	}
	s2, err := m.Consumer.Subscribe(ctx, topic, "bcast-b", func(context.Context, mq.Message) error { g2.Add(1); return nil })
	if err != nil {
		t.Fatalf("s2: %v", err)
	}
	defer func() { _ = s1.Close() }()
	defer func() { _ = s2.Close() }()
	time.Sleep(200 * time.Millisecond)

	for range 3 {
		if _, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte("x")}); err != nil {
			t.Fatalf("Publish: %v", err)
		}
	}
	if !wait(&g1, 3) || !wait(&g2, 3) {
		t.Fatalf("broadcast: g1=%d g2=%d, want each 3", g1.Load(), g2.Load())
	}
}

// A subscriber acks msg 1 + closes; a later subscriber on the same group resumes
// — gets subsequent messages and does NOT re-receive msg 1 (its delivery was deleted).
func TestDurableResume(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	m := New(h)
	ctx := context.Background()
	topic, group := uniqTopic(t, "dur"), "resume-group"

	if _, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte{1}}); err != nil {
		t.Fatalf("Publish 1: %v", err)
	}
	got1 := make(chan struct{}, 1)
	s1, err := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, mm mq.Message) error {
		if len(mm.Payload) > 0 && mm.Payload[0] == 1 {
			select {
			case got1 <- struct{}{}:
			default:
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("s1: %v", err)
	}
	select {
	case <-got1:
	case <-time.After(5 * time.Second):
		t.Fatal("s1 did not get msg1")
	}
	_ = s1.Close() // msg1 acked (delivery deleted) before Close returns
	time.Sleep(200 * time.Millisecond)

	if _, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte{2}}); err != nil {
		t.Fatalf("Publish 2: %v", err)
	}
	if _, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte{3}}); err != nil {
		t.Fatalf("Publish 3: %v", err)
	}

	var seen []byte
	done := make(chan struct{}, 1)
	s2, err := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, mm mq.Message) error {
		seen = append(seen, mm.Payload...)
		if len(seen) >= 2 {
			select {
			case done <- struct{}{}:
			default:
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("s2: %v", err)
	}
	defer func() { _ = s2.Close() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
	if len(seen) != 2 || seen[0] != 2 || seen[1] != 3 {
		t.Fatalf("resume: s2 saw %v, want [2 3] (msg1 was acked)", seen)
	}
}
