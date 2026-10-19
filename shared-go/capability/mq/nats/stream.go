package nats

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
)

// streamName returns the JetStream stream name backing a topic.
func streamName(topic string) string { return "mq-" + topic }

// dlqStream is the single shared dead-letter stream.
const dlqStream = "mq-dlq"

// dlqSubject returns the DLQ subject for a topic (carried by the dlq stream).
func dlqSubject(topic string) string { return "dlq." + topic }

// ensureStream creates-or-updates the stream for a topic with the configured
// retention cap. Idempotent.
func (h *handle) ensureStream(ctx context.Context, topic string) error {
	_, err := h.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      streamName(topic),
		Subjects:  []string{topic},
		Retention: jetstream.LimitsPolicy,
		MaxAge:    maxAgeOrDefault(h.cfg.MaxAge),
		MaxBytes:  maxBytesOrDefault(h.cfg.MaxBytes),
	})
	return err
}

// ensureDLQStream creates-or-updates the shared DLQ stream (subjects dlq.>) with
// dedicated, longer retention so poison messages aren't evicted by live-traffic
// pressure on the source streams.
func (h *handle) ensureDLQStream(ctx context.Context) error {
	_, err := h.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      dlqStream,
		Subjects:  []string{"dlq.>"},
		Retention: jetstream.LimitsPolicy,
		MaxAge:    dlqMaxAgeOrDefault(h.cfg.DLQMaxAge),
		MaxBytes:  dlqMaxBytesOrDefault(h.cfg.DLQMaxBytes),
	})
	return err
}
