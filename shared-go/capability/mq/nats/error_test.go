package nats

import (
	"context"
	"errors"
	"testing"

	natsclient "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"cyber-ecosystem/shared-go/capability/mq"
)

// mapError classifies server-side timeouts (nats ErrTimeout, ctx deadline, a
// JetStream 504 API error) as ErrTimeout, and connection-class errors as
// ErrUnavailable.
func TestMapError(t *testing.T) {
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
		{"unknown", errors.New("boom"), nil}, // passthrough, asserted below
	}
	for _, c := range cases {
		got := mapError(c.in, "op")
		if c.want == nil {
			if !errors.Is(got, c.in) {
				t.Errorf("%s: mapError(%v) lost the original", c.name, c.in)
			}
			continue
		}
		if !errors.Is(got, c.want) {
			t.Errorf("%s: mapError(%v) not %v", c.name, c.in, c.want)
		}
	}
	// Cancellation belongs to the caller — must not be reclassified.
	if got := mapError(context.Canceled, "op"); !errors.Is(got, context.Canceled) {
		t.Errorf("ctx canceled reclassified: %v", got)
	}
	if mapError(nil, "op") != nil {
		t.Error("nil should map to nil")
	}
	// A non-504 API error must not read as a timeout.
	if errors.Is(mapError(&jetstream.APIError{Code: 500}, "op"), mq.ErrTimeout) {
		t.Error("500 APIError should not map to ErrTimeout")
	}
}
