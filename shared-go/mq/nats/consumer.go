package nats

import (
	"context"
	"strconv"
	"time"

	natsclient "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"cyber-ecosystem/shared-go/mq"
)

type consumer struct{ h *handle }

func newConsumer(h *handle) mq.Consumer { return &consumer{h: h} }

type subscription struct {
	consume jetstream.ConsumeContext
	cancel  context.CancelFunc
	h       *handle
}

// drain processes any in-flight/buffered messages then stops the consumer,
// bounded by ctx (on ctx.Done it force-Stops, discarding buffered messages).
// No callback runs after drain returns.
func (s *subscription) drain(ctx context.Context) {
	if s.consume != nil {
		s.consume.Drain()
		select {
		case <-s.consume.Closed():
		case <-ctx.Done():
			s.consume.Stop()
		}
	}
	s.cancel()
}

func (s *subscription) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.drain(ctx)
	s.h.unregister(s)
	return nil
}

// ackDecision is the terminal action to take on a delivered message.
type ackDecision int

const (
	ackMsg  ackDecision = iota // Ack  — handler succeeded
	nakMsg                     // Nak  — retry (handler failed, under the cap or DLQ write failed)
	termMsg                    // Term — poison, do not redeliver (cap reached + DLQ'd, or untrackable)
)

// decideAck is the pure ack/Nak/Term policy for a delivered message, given the
// handler result and (only relevant at the retry cap) the DLQ-publish result.
// Extracted so the policy is unit-testable without forcing NATS-side failures:
// on DLQ-publish failure it returns Nak (retain the message) rather than Term,
// and a metadata-less message is Terminated (no infinite redelivery loop).
func decideAck(meta *jetstream.MsgMetadata, herr error, maxRetries int, dlqErr error) ackDecision {
	if herr == nil {
		return ackMsg
	}
	if meta == nil {
		// No metadata ⇒ cannot count deliveries ⇒ cannot DLQ safely. Terminate to
		// avoid an infinite redelivery loop on a metadata-less poison message.
		return termMsg
	}
	if int(meta.NumDelivered) < maxRetries {
		return nakMsg // under the cap: retry
	}
	// Cap reached: dead-letter on success, otherwise retain (Nak) — never lose it.
	if dlqErr != nil {
		return nakMsg
	}
	return termMsg
}

func (c *consumer) Subscribe(ctx context.Context, topic, group string, handler func(context.Context, mq.Message) error) (mq.Subscription, error) {
	if err := mq.ValidateTopic(topic); err != nil {
		return nil, err
	}
	if err := mq.ValidateGroup(group); err != nil {
		return nil, err
	}
	if err := c.h.ensureStream(ctx, topic); err != nil {
		return nil, mapError(err, "ensure stream")
	}
	maxRetries := maxRetriesOrDefault(c.h.cfg.MaxRetries)
	// Durable name is group+topic so two topics that share a group name do not
	// collapse onto one durable, and the name is self-describing.
	ccfg := jetstream.ConsumerConfig{
		Durable:       group + "-" + topic, // same group → competing; different group → broadcast
		FilterSubject: topic,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       ackWaitOrDefault(c.h.cfg.AckWait),
		MaxDeliver:    -1, // unbounded; we gate via NumDelivered → DLQ, bounded by MaxAckPending
		MaxAckPending: maxAckPendingOrDefault(c.h.cfg.MaxAckPending),
	}
	cons, err := c.h.js.CreateOrUpdateConsumer(ctx, streamName(topic), ccfg)
	if err != nil {
		return nil, mapError(err, "create consumer")
	}

	cctx, cancel := context.WithCancel(ctx)
	cc, err := cons.Consume(func(msg jetstream.Msg) {
		meta, merr := msg.Metadata()
		m := mq.Message{Topic: topic, Payload: msg.Data(), Headers: headerToMap(msg.Headers())}
		if merr == nil && meta != nil {
			m.ID = strconv.FormatUint(meta.Sequence.Stream, 10)
			m.Timestamp = meta.Timestamp
		}
		herr := handler(cctx, m)
		// At the retry cap, attempt the DLQ write so decideAck can choose Term vs Nak.
		var dlqErr error
		if herr != nil && meta != nil && int(meta.NumDelivered) >= maxRetries {
			dlqErr = c.dlq(cctx, topic, m, meta.NumDelivered, herr)
		}
		switch decideAck(meta, herr, maxRetries, dlqErr) {
		case ackMsg:
			_ = msg.Ack()
		case termMsg:
			_ = msg.Term()
		case nakMsg:
			_ = msg.NakWithDelay(c.nakDelay(meta.NumDelivered))
		}
	})
	if err != nil {
		cancel()
		return nil, mapError(err, "consume")
	}
	sub := &subscription{consume: cc, cancel: cancel, h: c.h}
	c.h.register(sub)
	return sub, nil
}

// nakDelay is a small linear backoff for redelivery (step × deliveries, capped).
func (c *consumer) nakDelay(delivered uint64) time.Duration {
	d := nakBackoffStepOrDefault(c.h.cfg.NakBackoffStep) * time.Duration(delivered)
	return min(d, time.Minute)
}

func headerToMap(h natsclient.Header) map[string]string {
	if len(h) == 0 {
		return nil
	}
	out := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	return out
}

// dlq publishes a failed message to the shared DLQ stream with original metadata
// in headers, so it can be inspected/replayed without losing context.
func (c *consumer) dlq(ctx context.Context, topic string, m mq.Message, delivered uint64, herr error) error {
	if err := c.h.ensureDLQStream(ctx); err != nil {
		return mapError(err, "ensure dlq stream")
	}
	nmsg := &natsclient.Msg{Subject: dlqSubject(topic), Data: m.Payload}
	nmsg.Header = natsclient.Header{}
	nmsg.Header.Set("mq-original-topic", topic)
	nmsg.Header.Set("mq-delivered", strconv.FormatUint(delivered, 10))
	nmsg.Header.Set("mq-error", herr.Error())
	for k, v := range m.Headers {
		nmsg.Header.Set("mq-orig-"+k, v)
	}
	_, err := c.h.js.PublishMsg(ctx, nmsg)
	return mapError(err, "dlq publish")
}
