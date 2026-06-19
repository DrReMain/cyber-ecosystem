package pg

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"cyber-ecosystem/shared-go/mq"
)

// stopper is the registry contract for active subscriptions: drainSubs stops
// each tracked one on shutdown. Defined as an interface (not *subscription) so
// client.go compiles independently of consumer.go.
type stopper interface {
	stop(ctx context.Context)
}

// NewClient 连独立 `mq` 库的 pgx 池，幂等建表，启动保留期 reaper，返回 handle + cleanup。
// cleanup：先 drain 所有订阅（停轮询），关 reaper，再关池。
func NewClient(cfg *Config) (*handle, func(), error) {
	if cfg == nil || cfg.DSN == "" {
		return nil, nil, fmt.Errorf("%w: dsn is required", mq.ErrInvalidArgument)
	}
	pool, err := pgxpool.New(context.Background(), cfg.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: pgxpool: %w", mq.ErrUnavailable, err)
	}
	for _, s := range schemaStmts {
		if _, err := pool.Exec(context.Background(), s); err != nil {
			pool.Close()
			return nil, nil, fmt.Errorf("%w: ensure schema: %w", mq.ErrUnavailable, err)
		}
	}
	h := &handle{pool: pool, cfg: *cfg}
	h.ctx, h.cancel = context.WithCancel(context.Background())
	stopReaper := h.startReaper()
	return h, func() {
		h.closed.Store(true) // reject new subscriptions racing shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		// Cancel the lifetime ctx first so every poll loop (derived from it, not from
		// any caller ctx) stops before we wait for them and close the pool.
		h.cancel()
		h.drainSubs(ctx)
		cancel()
		stopReaper()
		pool.Close()
	}, nil
}

// handle bundles the pgx pool + resolved config, and tracks active subscriptions
// so shutdown can drain them in order (mirrors shared-go/mq/nats handle).
type handle struct {
	pool *pgxpool.Pool
	cfg  Config

	// ctx is the handle/process lifetime context. Poll loops derive from it (not from
	// any Subscribe caller ctx) so long-lived subscriptions outlive the originating
	// request; cancelled by the returned cleanup.
	ctx    context.Context
	cancel context.CancelFunc

	// closed is set when cleanup begins; Subscribe rejects new subscriptions after it
	// so a Subscribe racing shutdown can't register a poll loop that escapes the drain.
	closed atomic.Bool

	mu   sync.Mutex
	subs map[stopper]struct{}
}

func (h *handle) register(s stopper) {
	h.mu.Lock()
	if h.subs == nil {
		h.subs = make(map[stopper]struct{})
	}
	h.subs[s] = struct{}{}
	h.mu.Unlock()
}

func (h *handle) unregister(s stopper) {
	h.mu.Lock()
	delete(h.subs, s)
	h.mu.Unlock()
}

// drainSubs gracefully stops every tracked subscription, bounded by ctx.
func (h *handle) drainSubs(ctx context.Context) {
	h.mu.Lock()
	subs := make([]stopper, 0, len(h.subs))
	for s := range h.subs {
		subs = append(subs, s)
	}
	h.mu.Unlock()
	for _, s := range subs {
		s.stop(ctx)
	}
}

// startReaper periodically (1) drops messages past retention (CASCADE removes their
// deliveries — NATS-MaxAge parity: an offline consumer's backlog ages out the same
// way on both backends; active messages are acked/DLQ'd long before retention), and
// (2) dead-letters deliveries that exhausted retries without acking and aren't being
// reached by a poll loop (e.g. a single wedged consumer goroutine, or a group with no
// active subscriber) — the backstop to the fetch-time stall gate, mirroring NATS
// server-side MaxDeliver.
func (h *handle) startReaper() func() {
	ctx, cancel := context.WithCancel(context.Background())
	retention := retentionOrDefault(h.cfg.Retention)
	maxRetries := maxRetriesOrDefault(h.cfg.MaxRetries)
	interval := max(retention/4, time.Minute)
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				cutoff := time.Now().Add(-retention)
				_, _ = h.pool.Exec(ctx, `DELETE FROM messages WHERE created_at < $1`, cutoff)
				h.reapStalled(ctx, maxRetries)
			}
		}
	}()
	return cancel
}

// reapStalled atomically dead-letters over-cap expired deliveries (deliveries >=
// maxRetries and past their visibility window) and removes them in one statement.
// No-op when maxRetries <= 0.
func (h *handle) reapStalled(ctx context.Context, maxRetries int) {
	if maxRetries <= 0 {
		return
	}
	const sql = `
WITH stalled AS (
  SELECT d.id AS did, d.topic, d.group_name, d.deliveries, m.payload, m.headers
  FROM deliveries d JOIN messages m ON d.message_id = m.id
  WHERE d.due_at <= now() AND d.deliveries >= $1
), ins AS (
  INSERT INTO dlq(topic, group_name, payload, headers, deliveries, error)
  SELECT topic, group_name, payload, headers, deliveries, $2 FROM stalled
)
DELETE FROM deliveries WHERE id IN (SELECT did FROM stalled)`
	if _, err := h.pool.Exec(ctx, sql, maxRetries, "delivery attempts exhausted without ack"); err != nil {
		slog.Default().Warn("mq/pg: reap stalled deliveries failed", "err", err)
	}
}
