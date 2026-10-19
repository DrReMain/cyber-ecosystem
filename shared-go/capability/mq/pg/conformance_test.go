package pg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/mq"
)

// Live conformance probe against the standalone `mq` PG database; skipped
// unless PG_CONFORMANCE_DSN is set. The probe list is shared with mq/nats's
// conformance suite: together they are the cross-backend parity check for the
// MQ semantic contract (at-least-once, group semantics, retry→DLQ, backfill,
// subscription lifetime). Each failure names the drift plus what breaks
// downstream. Drift is absorbed in config switches or platform instantiation —
// never inside this package.
//
//	PG_CONFORMANCE_DSN=postgres://postgres:postgres@localhost:5432/mq?sslmode=disable \
//	go test ./shared-go/capability/mq/pg/ -run TestLiveConformance -count=1 -v
func TestLiveConformance(t *testing.T) {
	dsn := os.Getenv("PG_CONFORMANCE_DSN")
	if dsn == "" {
		t.Skip("PG_CONFORMANCE_DSN not set; live conformance is opt-in")
	}
	// Values pinned to config.yaml's mq.pg block: zeros would silently
	// resolve to compiled defaults and diverge this instantiation from the
	// service's.
	cfg := &Config{
		DSN:               dsn,
		PollInterval:      500 * time.Millisecond,
		VisibilityTimeout: 30 * time.Second,
		MaxRetries:        5,
		Retention:         7 * 24 * time.Hour,
		BatchSize:         16,
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
		var mu atomic.Int64
		count := func(_ context.Context, _ mq.Message) error {
			mu.Add(1)
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
		deadline := time.Now().Add(20 * time.Second)
		for mu.Load() < n && time.Now().Before(deadline) {
			time.Sleep(100 * time.Millisecond)
		}
		if d := mu.Load(); d != n {
			t.Errorf("same-group deliveries = %d, want exactly %d — SKIP LOCKED work-queue drift (dupes or loss)", d, n)
		}
		// Settled rows must be gone: ack deletes the delivery.
		var left int
		if err := h.pool.QueryRow(ctx, `SELECT count(*) FROM deliveries WHERE topic=$1 AND group_name=$2`, tp, group("cg")).Scan(&left); err != nil {
			t.Fatalf("count deliveries: %v", err)
		}
		if left != 0 {
			t.Errorf("%d delivery rows left after ack — settle drift", left)
		}
	})

	t.Run("broadcast-groups", func(t *testing.T) {
		tp := topic("bg")
		var a, b atomic.Int64
		s1, err := c.Consumer.Subscribe(ctx, tp, group("b1"), func(_ context.Context, _ mq.Message) error {
			a.Add(1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe(b1): %v", err)
		}
		defer func() { _ = s1.Close() }()
		s2, err := c.Consumer.Subscribe(ctx, tp, group("b2"), func(_ context.Context, _ mq.Message) error {
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
		deadline := time.Now().Add(20 * time.Second)
		for (a.Load() < n || b.Load() < n) && time.Now().Before(deadline) {
			time.Sleep(100 * time.Millisecond)
		}
		if a.Load() != n || b.Load() != n {
			t.Errorf("broadcast drift: group1=%d group2=%d, want %d each — fanout-per-group semantics broken", a.Load(), b.Load(), n)
		}
	})

	t.Run("retry-to-dlq", func(t *testing.T) {
		tp := topic("dlq")
		grp := group("dlq")
		var deliveries atomic.Int64
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
		// Poll the dlq table directly (the test lives in-package).
		deadline := time.Now().Add(30 * time.Second)
		var gotDeliveries int
		var gotErr string
		var gotPayload []byte
		for time.Now().Before(deadline) {
			err := h.pool.QueryRow(ctx,
				`SELECT deliveries, error, payload FROM dlq WHERE topic=$1 AND group_name=$2 ORDER BY id DESC LIMIT 1`,
				tp, grp).Scan(&gotDeliveries, &gotErr, &gotPayload)
			if err == nil {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		if gotPayload == nil {
			t.Fatal("message never reached the dlq table — poison isolation broken")
		}
		if string(gotPayload) != string(want) {
			t.Errorf("DLQ payload drift: %q", gotPayload)
		}
		if gotDeliveries != cfg.MaxRetries || gotErr == "" {
			t.Errorf("DLQ row deliveries=%d error=%q; want %d with a reason — replay context lost", gotDeliveries, gotErr, cfg.MaxRetries)
		}
		var left int
		if err := h.pool.QueryRow(ctx, `SELECT count(*) FROM deliveries WHERE topic=$1 AND group_name=$2`, tp, grp).Scan(&left); err != nil {
			t.Fatalf("count deliveries: %v", err)
		}
		if left != 0 {
			t.Errorf("%d delivery rows survived DLQ — atomic dlq+remove drift", left)
		}
		if d := deliveries.Load(); d != int64(cfg.MaxRetries) {
			t.Errorf("deliveries = %d, want %d — retry cap drift", d, cfg.MaxRetries)
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
		subCancel() // the caller ctx gates ONLY setup; the poll loop survives
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
			t.Errorf("delivery after Close: %+v — poll loop leak", m)
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
		deadline := time.Now().Add(20 * time.Second)
		for count.Load() < n && time.Now().Before(deadline) {
			time.Sleep(100 * time.Millisecond)
		}
		if count.Load() != n {
			t.Errorf("backfill = %d, want %d — DeliverAll-on-first-bind parity broken (nats delivers history)", count.Load(), n)
		}
	})

	t.Run("close-drains", func(t *testing.T) {
		// The client cleanup stops poll loops, the reaper, then the pool; it
		// must return promptly, and a post-cleanup publish must surface a
		// sentinel (not panic).
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
