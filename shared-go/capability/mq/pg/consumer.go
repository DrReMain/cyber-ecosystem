package pg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"cyber-ecosystem/shared-go/capability/mq"
)

type consumer struct{ h *handle }

func newConsumer(h *handle) mq.Consumer { return &consumer{h: h} }

type subscription struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
	h      *handle
}

// stop cancels the poll context and waits for the poll goroutine to exit
// (bounded by ctx). Satisfies the stopper interface in client.go.
func (s *subscription) stop(ctx context.Context) {
	s.cancel()
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
	}
}

func (s *subscription) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.stop(ctx)
	if s.h != nil {
		s.h.unregister(s)
	}
	return nil
}

type delivery struct {
	deliveryID int64
	deliveries int // attempts BEFORE this fetch; this fetch makes it deliveries+1
	msgID      int64
	payload    []byte
	headers    map[string]string
	created    time.Time
}

func (c *consumer) Subscribe(ctx context.Context, topic, group string, handler func(context.Context, mq.Message) error) (mq.Subscription, error) {
	if err := mq.ValidateTopic(topic); err != nil {
		return nil, err
	}
	if err := mq.ValidateGroup(group); err != nil {
		return nil, err
	}
	if c.h.closed.Load() {
		return nil, fmt.Errorf("%w: mq client is closing", mq.ErrUnavailable)
	}
	if err := c.registerAndBackfill(ctx, topic, group); err != nil {
		return nil, mapError(err, "subscribe")
	}
	// The poll loop runs under the handle's lifetime ctx, NOT the caller's ctx, so a
	// long-lived subscription survives the originating request/RPC ending. The caller
	// ctx only gates the synchronous setup above.
	cctx, cancel := context.WithCancel(c.h.ctx)
	sub := &subscription{cancel: cancel, h: c.h}
	c.h.register(sub)
	sub.wg.Go(func() {
		c.pollLoop(cctx, topic, group, handler)
	})
	return sub, nil
}

// registerAndBackfill records the (group,topic) subscription; for a NEW group it
// backfills delivery rows for all retained messages on the topic (historical
// replay, matching a NATS durable's DeliverAll on first bind). An existing
// subscriber row ⇒ resume (its deliveries persist).
func (c *consumer) registerAndBackfill(ctx context.Context, topic, group string) error {
	tx, err := c.h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	ct, err := tx.Exec(ctx,
		`INSERT INTO subscribers(group_name,topic) VALUES($1,$2) ON CONFLICT DO NOTHING`, group, topic)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 1 {
		if err := backfillDeliveries(ctx, tx, topic, group); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// backfillDeliveries creates a delivery row for every retained message on the topic
// (historical replay, matching a NATS durable's DeliverAll on first bind). It is
// batched by id cursor so a high-volume topic doesn't turn the first Subscribe into
// one giant INSERT that locks messages/deliveries and spikes WAL.
func backfillDeliveries(ctx context.Context, tx pgx.Tx, topic, group string) error {
	const batch = 2000
	var cursor int64
	for {
		rows, err := tx.Query(ctx,
			`SELECT id FROM messages WHERE topic=$1 AND id > $2 ORDER BY id LIMIT $3`,
			topic, cursor, batch)
		if err != nil {
			return err
		}
		var ids []int64
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			ids = append(ids, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO deliveries(group_name,topic,message_id)
			 SELECT $1, $2, v FROM unnest($3::bigint[]) AS v
			 ON CONFLICT (group_name, message_id) DO NOTHING`,
			group, topic, ids); err != nil {
			return err
		}
		cursor = ids[len(ids)-1]
		if len(ids) < batch {
			return nil
		}
	}
}

func (c *consumer) pollLoop(ctx context.Context, topic, group string, handler func(context.Context, mq.Message) error) {
	cfg := c.h.cfg
	interval := pollIntervalOrDefault(cfg.PollInterval)
	vis := visibilityOrDefault(cfg.VisibilityTimeout)
	maxRetries := maxRetriesOrDefault(cfg.MaxRetries)
	batch := batchSizeOrDefault(cfg.BatchSize)
	log := slog.Default()
	var failStreak int
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollBackoff(interval, failStreak)):
		}
		ds, err := c.fetchBatch(ctx, group, topic, batch, vis)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			// Sustained fetch failure (PG down / pool exhausted): back off and log the
			// down→up transition only, so an outage is observable without log spam.
			if failStreak == 0 {
				log.Warn("mq/pg: poll fetch failed", "topic", topic, "group", group, "err", err)
			}
			failStreak++
			continue
		}
		if failStreak > 0 {
			log.Warn("mq/pg: poll fetch recovered", "topic", topic, "group", group, "streak", failStreak)
			failStreak = 0
		}
		for _, d := range ds {
			if ctx.Err() != nil {
				return
			}
			// Stall gate: a delivery already attempted maxRetries times and still
			// pending (handler stalled past visibility / never acked) is dead-lettered
			// at fetch — the PG analogue of NATS server-side MaxDeliver. d.deliveries is
			// the pre-fetch count, so >= maxRetries means the prior maxRetries attempts
			// did not ack.
			if maxRetries > 0 && d.deliveries >= maxRetries {
				c.dlqAtFetch(ctx, d, topic, group)
				continue
			}
			m := mq.Message{Topic: topic, Payload: d.payload, Headers: d.headers, ID: strconv.FormatInt(d.msgID, 10), Timestamp: d.created}
			herr := handler(ctx, m)
			c.settle(ctx, d, topic, group, herr, maxRetries)
		}
	}
}

// fetchBatch locks a batch of due deliveries (FOR UPDATE SKIP LOCKED), counts this
// delivery (deliveries+1 — a visibility-timeout redelivery counts toward maxRetries,
// matching NATS NumDelivered), and renews visibility so other consumers in the same
// group don't grab them while the handler runs.
func (c *consumer) fetchBatch(ctx context.Context, group, topic string, batch int, vis time.Duration) ([]delivery, error) {
	tx, err := c.h.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx,
		`SELECT d.id, d.deliveries, m.id, m.payload, m.headers, m.created_at
		 FROM deliveries d JOIN messages m ON d.message_id=m.id
		 WHERE d.group_name=$1 AND d.topic=$2 AND d.due_at <= now()
		 ORDER BY d.due_at, d.message_id LIMIT $3 FOR UPDATE SKIP LOCKED`, group, topic, batch)
	if err != nil {
		return nil, err
	}
	var ds []delivery
	for rows.Next() {
		var d delivery
		var hdr []byte
		if err := rows.Scan(&d.deliveryID, &d.deliveries, &d.msgID, &d.payload, &hdr, &d.created); err != nil {
			rows.Close()
			return nil, err
		}
		_ = json.Unmarshal(hdr, &d.headers)
		ds = append(ds, d)
	}
	rows.Close()
	if err := rows.Err(); err != nil { // surface stream errors; don't mask as an empty batch
		return nil, err
	}
	for _, d := range ds {
		if _, err := tx.Exec(ctx, `UPDATE deliveries SET due_at=now()+$1, deliveries=deliveries+1 WHERE id=$2`, vis, d.deliveryID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ds, nil
}

// settle applies the ack/Nak/DLQ policy (mirrors nats decideAck). Success → delete
// the delivery. Failure under the cap → redelay (deliveries was already counted at
// fetch, so no increment here). At the cap → DLQ+remove atomically; on DLQ failure
// → re-queue (the next fetch's stall gate will DLQ it). Swallowed settle errors are
// logged so a PG blip is observable rather than silently freezing retry state.
func (c *consumer) settle(ctx context.Context, d delivery, topic, group string, herr error, maxRetries int) {
	log := slog.Default()
	if herr == nil {
		_, _ = c.h.pool.Exec(ctx, `DELETE FROM deliveries WHERE id=$1`, d.deliveryID)
		return
	}
	attempt := d.deliveries + 1 // deliveries was incremented at fetch; this is the attempt number
	if attempt < maxRetries {
		if _, err := c.h.pool.Exec(ctx, `UPDATE deliveries SET due_at=now()+$1 WHERE id=$2`, nakBackoff(attempt), d.deliveryID); err != nil {
			log.Warn("mq/pg: nak redelay failed; message will be re-fetched at visibility expiry", "topic", topic, "group", group, "err", err)
		}
		return
	}
	if err := c.dlqAndRemove(ctx, d, topic, group, herr.Error()); err != nil {
		log.Warn("mq/pg: dlq failed; re-queueing (stall gate will DLQ on next fetch)", "topic", topic, "group", group, "err", err)
		_, _ = c.h.pool.Exec(ctx, `UPDATE deliveries SET due_at=now() WHERE id=$1`, d.deliveryID)
	}
}

// dlqAtFetch dead-letters a delivery that exhausted its attempts without acking
// (handler stalled past visibility on every attempt). The reason is synthetic
// since there is no handler error.
func (c *consumer) dlqAtFetch(ctx context.Context, d delivery, topic, group string) {
	reason := fmt.Sprintf("delivery attempts exhausted (%d) without ack", d.deliveries+1)
	if err := c.dlqAndRemove(ctx, d, topic, group, reason); err != nil {
		slog.Default().Warn("mq/pg: stall-gate dlq failed; will retry next fetch", "topic", topic, "group", group, "err", err)
		_, _ = c.h.pool.Exec(ctx, `UPDATE deliveries SET due_at=now() WHERE id=$1`, d.deliveryID)
	}
}

// dlqAndRemove inserts the dead-letter and removes the delivery in one txn
// (atomic — no DLQ-without-remove window that would duplicate on a crash; this
// is an atomicity PG gives that NATS's two-stream model structurally cannot).
func (c *consumer) dlqAndRemove(ctx context.Context, d delivery, topic, group, reason string) error {
	headers, _ := json.Marshal(d.headers)
	tx, err := c.h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx,
		`INSERT INTO dlq(topic,group_name,payload,headers,deliveries,error) VALUES($1,$2,$3,$4,$5,$6)`,
		topic, group, d.payload, headers, d.deliveries+1, reason); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM deliveries WHERE id=$1`, d.deliveryID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// pollBackoff returns the poll wait interval, exponentially backing off (capped)
// across consecutive fetch failures so a PG outage isn't hammered at poll rate.
func pollBackoff(base time.Duration, failures int) time.Duration {
	if failures <= 0 {
		return base
	}
	const capBackoff = 10 * time.Second
	shift := min(failures, 8)
	return min(base*time.Duration(1<<shift), capBackoff)
}

// nakBackoff is the linear redelivery delay after a handler failure, capped to
// match NATS's 1m Nak ceiling.
func nakBackoff(attempt int) time.Duration {
	const capBackoff = time.Minute
	return min(time.Duration(attempt)*100*time.Millisecond, capBackoff)
}
