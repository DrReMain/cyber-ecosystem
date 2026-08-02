package nats

import (
	"context"
	"errors"
	"testing"

	natsclient "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"cyber-ecosystem/shared-go/capability/mq"
)

var errBoom = errors.New("boom")

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
	if errors.Is(mapError(&jetstream.APIError{Code: 500}, "op"), mq.ErrTimeout) {
		t.Error("500 APIError should not map to ErrTimeout")
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
