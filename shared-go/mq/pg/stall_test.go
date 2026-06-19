package pg

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

// A delivery already attempted maxRetries times and still pending (handler stalled
// past visibility / never acked) is dead-lettered at fetch WITHOUT invoking the
// handler — the PG analogue of NATS server-side MaxDeliver.
func TestStallGateDLQ(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	m := New(h)
	ctx := context.Background()
	topic, group := uniqTopic(t, "stall"), "stall-group"
	maxRetries := 3

	// Seed a message plus a delivery already delivered maxRetries times, due now.
	var msgID int64
	if err := h.pool.QueryRow(ctx,
		`INSERT INTO messages(topic,payload,headers) VALUES($1,$2,$3) RETURNING id`,
		topic, []byte("stuck"), []byte("{}")).Scan(&msgID); err != nil {
		t.Fatalf("seed message: %v", err)
	}
	if _, err := h.pool.Exec(ctx,
		`INSERT INTO deliveries(group_name,topic,message_id,deliveries,due_at) VALUES($1,$2,$3,$4,now())`,
		group, topic, msgID, maxRetries); err != nil {
		t.Fatalf("seed delivery: %v", err)
	}

	var called atomic.Int32
	sub, err := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, _ mq.Message) error {
		called.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = sub.Close() }()

	deadline := time.After(5 * time.Second)
	for {
		var n int
		if err := h.pool.QueryRow(ctx, `SELECT count(*) FROM dlq WHERE topic=$1`, topic).Scan(&n); err != nil {
			t.Fatalf("dlq query: %v", err)
		}
		if n == 1 {
			break
		}
		select {
		case <-time.After(50 * time.Millisecond):
		case <-deadline:
			t.Fatalf("stalled message was not dead-lettered at fetch")
		}
	}
	if got := called.Load(); got != 0 {
		t.Fatalf("handler invoked %d times; stall gate must DLQ without invoking handler", got)
	}

	var d int
	if err := h.pool.QueryRow(ctx, `SELECT count(*) FROM deliveries WHERE topic=$1`, topic).Scan(&d); err != nil {
		t.Fatalf("delivery query: %v", err)
	}
	if d != 0 {
		t.Fatalf("delivery row count=%d, want 0 (DLQ+remove)", d)
	}
}

// The reaper dead-letters deliveries that exhausted retries without acking and are
// not being reached by a poll loop (e.g. a single wedged consumer goroutine) — the
// backstop to the fetch-time stall gate, mirroring NATS server-side MaxDeliver.
func TestReaperReapsStalled(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic, group := uniqTopic(t, "reap"), "reap-group"
	maxRetries := 3

	// Seed a message plus an over-cap, already-expired delivery that no poll loop
	// is bound to (no subscriber on this group/topic).
	var msgID int64
	if err := h.pool.QueryRow(ctx,
		`INSERT INTO messages(topic,payload,headers) VALUES($1,$2,$3) RETURNING id`,
		topic, []byte("wedged"), []byte("{}")).Scan(&msgID); err != nil {
		t.Fatalf("seed message: %v", err)
	}
	if _, err := h.pool.Exec(ctx,
		`INSERT INTO deliveries(group_name,topic,message_id,deliveries,due_at)
		 VALUES($1,$2,$3,$4,now()-interval '1 second')`,
		group, topic, msgID, maxRetries); err != nil {
		t.Fatalf("seed delivery: %v", err)
	}

	h.reapStalled(ctx, maxRetries)

	var n int
	if err := h.pool.QueryRow(ctx, `SELECT count(*) FROM dlq WHERE topic=$1`, topic).Scan(&n); err != nil {
		t.Fatalf("dlq query: %v", err)
	}
	if n != 1 {
		t.Fatalf("dlq count=%d, want 1 (reaper should dead-letter stalled delivery)", n)
	}
	var d int
	if err := h.pool.QueryRow(ctx, `SELECT count(*) FROM deliveries WHERE topic=$1`, topic).Scan(&d); err != nil {
		t.Fatalf("delivery query: %v", err)
	}
	if d != 0 {
		t.Fatalf("delivery row count=%d, want 0 (reaper removes after DLQ)", d)
	}
}
