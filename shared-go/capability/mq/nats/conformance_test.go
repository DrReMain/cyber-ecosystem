package nats

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	natsclient "github.com/nats-io/nats.go"

	"cyber-ecosystem/shared-go/capability/mq"
)

// Live conformance probe against a real NATS JetStream backend; skipped unless
// NATS_CONFORMANCE_ENDPOINT is set. The probe list is shared with mq/pg's
// conformance suite: together they are the cross-backend parity check for the
// MQ semantic contract (at-least-once, group semantics, retry→DLQ, backfill,
// subscription lifetime). Each failure names the drift plus what breaks
// downstream. Drift is absorbed in config switches or platform instantiation —
// never inside this package.
//
//	NATS_CONFORMANCE_ENDPOINT=nats://localhost:4222 \
//	go test ./shared-go/capability/mq/nats/ -run TestLiveConformance -count=1 -v
func TestLiveConformance(t *testing.T) {
	endpoint := os.Getenv("NATS_CONFORMANCE_ENDPOINT")
	if endpoint == "" {
		t.Skip("NATS_CONFORMANCE_ENDPOINT not set; live conformance is opt-in")
	}
	// Values pinned to config.yaml's (commented) nats block: zeros would
	// silently resolve to compiled defaults and diverge this instantiation
	// from the service's.
	cfg := &Config{
		Endpoint:       endpoint,
		Creds:          os.Getenv("NATS_CONFORMANCE_CREDS"),
		MaxAge:         7 * 24 * time.Hour,
		MaxBytes:       1 << 30,
		MaxRetries:     5,
		AckWait:        30 * time.Second,
		MaxAckPending:  256,
		DLQMaxAge:      30 * 24 * time.Hour,
		DLQMaxBytes:    1 << 30,
		NakBackoffStep: 500 * time.Millisecond,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	h, cleanup, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c := New(h)
	suffix := fmt.Sprint(time.Now().UnixNano())
	topic := func(name string) string { return "conf" + name + suffix }
	group := func(name string) string { return "g" + name + suffix }

	t.Run("publish-roundtrip", func(t *testing.T) {
		tp := topic("rt")
		got := make(chan mq.Message, 1)
		sub, err := c.Consumer.Subscribe(ctx, tp, group("rt"), func(_ context.Context, m mq.Message) error {
			select {
			case got <- m:
			default:
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe: %v", err)
		}
		defer func() { _ = sub.Close() }()
		want := &mq.Message{Payload: []byte("hello"), Headers: map[string]string{"k": "v"}}
		if _, err := c.Publisher.Publish(ctx, tp, want); err != nil {
			t.Fatalf("Publish: %v", err)
		}
		select {
		case m := <-got:
			if string(m.Payload) != string(want.Payload) || m.Headers["k"] != "v" {
				t.Errorf("payload/headers drift: %+v", m)
			}
			if m.Topic != tp || m.ID == "" || m.Timestamp.IsZero() {
				t.Errorf("envelope drift: topic=%q id=%q ts=%v", m.Topic, m.ID, m.Timestamp)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("no delivery within 5s — publish/subscribe roundtrip drift")
		}
	})

	t.Run("competing-group", func(t *testing.T) {
		tp := topic("cg")
		var mu sync.Mutex
		seen := map[string]int{}
		count := func(_ context.Context, m mq.Message) error {
			mu.Lock()
			seen[m.ID]++
			mu.Unlock()
			return nil
		}
		s1, err := c.Consumer.Subscribe(ctx, tp, group("cg"), count)
		if err != nil {
			t.Fatalf("Subscribe(1): %v", err)
		}
		defer func() { _ = s1.Close() }()
		s2, err := c.Consumer.Subscribe(ctx, tp, group("cg"), count)
		if err != nil {
			t.Fatalf("Subscribe(2): %v", err)
		}
		defer func() { _ = s2.Close() }()
		const n = 10
		for i := 0; i < n; i++ {
			if _, err := c.Publisher.Publish(ctx, tp, &mq.Message{Payload: []byte{byte(i)}}); err != nil {
				t.Fatalf("Publish: %v", err)
			}
		}
		deadline := time.Now().Add(15 * time.Second)
		for {
			mu.Lock()
			total := len(seen)
			dups := 0
			for _, v := range seen {
				if v > 1 {
					dups++
				}
			}
			mu.Unlock()
			if total == n && dups == 0 {
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("same-group competing drift: %d/%d messages, %d redelivered — work-queue semantics broken", total, n, dups)
			}
			time.Sleep(100 * time.Millisecond)
		}
	})

	t.Run("broadcast-groups", func(t *testing.T) {
		tp := topic("bg")
		var a, b atomic.Int64
		s1, err := c.Consumer.Subscribe(ctx, tp, group("b1"), func(_ context.Context, m mq.Message) error {
			a.Add(1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe(b1): %v", err)
		}
		defer func() { _ = s1.Close() }()
		s2, err := c.Consumer.Subscribe(ctx, tp, group("b2"), func(_ context.Context, m mq.Message) error {
			b.Add(1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe(b2): %v", err)
		}
		defer func() { _ = s2.Close() }()
		const n = 3
		for i := 0; i < n; i++ {
			if _, err := c.Publisher.Publish(ctx, tp, &mq.Message{Payload: []byte{byte(i)}}); err != nil {
				t.Fatalf("Publish: %v", err)
			}
		}
		deadline := time.Now().Add(15 * time.Second)
		for (a.Load() < n || b.Load() < n) && time.Now().Before(deadline) {
			time.Sleep(100 * time.Millisecond)
		}
		if a.Load() != n || b.Load() != n {
			t.Errorf("broadcast drift: group1=%d group2=%d, want %d each — durable-per-group semantics broken", a.Load(), b.Load(), n)
		}
	})

	t.Run("retry-to-dlq", func(t *testing.T) {
		tp := topic("dlq")
		grp := group("dlq")
		var deliveries atomic.Int64
		// Core (non-JetStream) subscription on the DLQ subject catches the
		// dead-letter publish itself.
		dlqCh := make(chan *natsclient.Msg, 1)
		coreSub, err := h.nc.ChanSubscribe(dlqSubject(tp), dlqCh)
		if err != nil {
			t.Fatalf("core subscribe: %v", err)
		}
		defer func() { _ = coreSub.Unsubscribe() }()
		sub, err := c.Consumer.Subscribe(ctx, tp, grp, func(_ context.Context, _ mq.Message) error {
			deliveries.Add(1)
			return errors.New("always fails")
		})
		if err != nil {
			t.Fatalf("Subscribe: %v", err)
		}
		defer func() { _ = sub.Close() }()
		want := []byte("poison")
		if _, err := c.Publisher.Publish(ctx, tp, &mq.Message{Payload: want, Headers: map[string]string{"k": "v"}}); err != nil {
			t.Fatalf("Publish: %v", err)
		}
		select {
		case msg := <-dlqCh:
			if string(msg.Data) != string(want) {
				t.Errorf("DLQ payload drift: %q", msg.Data)
			}
			if got := msg.Header.Get("mq-original-topic"); got != tp {
				t.Errorf("mq-original-topic = %q, want %q — replay context lost", got, tp)
			}
			if msg.Header.Get("mq-delivered") == "" || msg.Header.Get("mq-error") == "" {
				t.Error("DLQ headers missing mq-delivered/mq-error — inspection/replay breaks")
			}
			if msg.Header.Get("mq-orig-k") != "v" {
				t.Error("original headers not carried under mq-orig-* — replay context lost")
			}
		case <-time.After(30 * time.Second):
			t.Fatal("message never reached the DLQ — poison isolation broken")
		}
		// Term must stop redelivery: exactly maxRetries deliveries, none after.
		time.Sleep(2 * time.Second)
		if d := deliveries.Load(); d != int64(cfg.MaxRetries) {
			t.Errorf("deliveries = %d, want %d — retry cap or Term drift", d, cfg.MaxRetries)
		}
	})

	t.Run("subscription-outlives-ctx", func(t *testing.T) {
		tp := topic("ctx")
		subCtx, subCancel := context.WithCancel(ctx)
		got := make(chan mq.Message, 1)
		sub, err := c.Consumer.Subscribe(subCtx, tp, group("ctx"), func(_ context.Context, m mq.Message) error {
			select {
			case got <- m:
			default:
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe: %v", err)
		}
		defer func() { _ = sub.Close() }()
		subCancel() // the caller ctx gates ONLY setup; the subscription survives
		time.Sleep(100 * time.Millisecond)
		if _, err := c.Publisher.Publish(ctx, tp, &mq.Message{Payload: []byte("x")}); err != nil {
			t.Fatalf("Publish: %v", err)
		}
		select {
		case <-got:
		case <-time.After(5 * time.Second):
			t.Fatal("subscription died with its setup ctx — long-lived consumers break")
		}
		if err := sub.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
		if _, err := c.Publisher.Publish(ctx, tp, &mq.Message{Payload: []byte("y")}); err != nil {
			t.Fatalf("Publish: %v", err)
		}
		select {
		case m := <-got:
			t.Errorf("delivery after Close: %+v — subscription leak", m)
		case <-time.After(2 * time.Second):
		}
	})

	t.Run("new-group-backfill", func(t *testing.T) {
		tp := topic("bf")
		// Publish BEFORE any subscriber exists on this topic.
		const n = 3
		for i := 0; i < n; i++ {
			if _, err := c.Publisher.Publish(ctx, tp, &mq.Message{Payload: []byte{byte(i)}}); err != nil {
				t.Fatalf("Publish: %v", err)
			}
		}
		var count atomic.Int64
		sub, err := c.Consumer.Subscribe(ctx, tp, group("bf"), func(_ context.Context, _ mq.Message) error {
			count.Add(1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe: %v", err)
		}
		defer func() { _ = sub.Close() }()
		deadline := time.Now().Add(15 * time.Second)
		for count.Load() < n && time.Now().Before(deadline) {
			time.Sleep(100 * time.Millisecond)
		}
		if count.Load() != n {
			t.Errorf("backfill = %d, want %d — DeliverAll-on-first-bind parity broken (pg backfills)", count.Load(), n)
		}
	})

	t.Run("close-drains", func(t *testing.T) {
		// The client cleanup drains subscriptions then the connection; it must
		// return promptly, and a post-cleanup publish must surface a sentinel
		// (not panic).
		start := time.Now()
		cleanup()
		if elapsed := time.Since(start); elapsed > 20*time.Second {
			t.Errorf("cleanup took %v — drain hang", elapsed)
		}
		_, err := c.Publisher.Publish(ctx, topic("cd"), &mq.Message{Payload: []byte("x")})
		if !errors.Is(err, mq.ErrUnavailable) {
			t.Errorf("post-cleanup Publish = %v, want ErrUnavailable", err)
		}
	})
}
