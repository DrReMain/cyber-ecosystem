package nats

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

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
		{"first attempt fails → nak (retry)", mkMeta(1), errors.New("boom"), nil, nakMsg},
		{"mid attempts fail → nak (retry)", mkMeta(2), errors.New("boom"), nil, nakMsg},
		{"cap reached, DLQ ok → term (poison isolated)", mkMeta(maxRetries), errors.New("boom"), nil, termMsg},
		{"cap reached, DLQ failed → nak (retain, no silent loss)", mkMeta(maxRetries), errors.New("boom"), dlqFail, nakMsg},
		{"over cap, DLQ failed → nak (retain)", mkMeta(maxRetries + 2), errors.New("boom"), dlqFail, nakMsg},
		{"meta nil + error → term (no infinite loop)", nil, errors.New("boom"), nil, termMsg},
		{"meta nil + success → ack", nil, nil, nil, ackMsg},
	}
	for _, c := range cases {
		if got := decideAck(c.meta, c.herr, maxRetries, c.dlqErr); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

// nakDelay is linear in deliveries (step from config), capped at 1m — the
// same ceiling PG's nakBackoff uses, so both backends retry in lockstep.
func TestNakDelay(t *testing.T) {
	c := &consumer{h: &handle{cfg: Config{NakBackoffStep: 500 * time.Millisecond}}}
	if d := c.nakDelay(1); d != 500*time.Millisecond {
		t.Errorf("nakDelay(1) = %v, want 500ms", d)
	}
	if d := c.nakDelay(4); d != 2*time.Second {
		t.Errorf("nakDelay(4) = %v, want 2s (linear)", d)
	}
	if d := c.nakDelay(1000); d != time.Minute {
		t.Errorf("nakDelay(1000) = %v, want 1m cap", d)
	}
	// Zero config steps through the default.
	c = &consumer{h: &handle{}}
	if d := c.nakDelay(2); d != time.Second {
		t.Errorf("default step nakDelay(2) = %v, want 1s", d)
	}
}

// deliveredCount clamps far-above-int counters; real counters sit far below.
func TestDeliveredCount(t *testing.T) {
	if got := deliveredCount(3); got != 3 {
		t.Errorf("deliveredCount(3) = %d", got)
	}
	if got := deliveredCount(math.MaxUint64); got != math.MaxInt32 {
		t.Errorf("deliveredCount(MaxUint64) = %d, want MaxInt32 clamp", got)
	}
}
