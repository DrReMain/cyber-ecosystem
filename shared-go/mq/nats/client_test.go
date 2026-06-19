package nats

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

func testConfig() *Config {
	ep := os.Getenv("NATS_ENDPOINT")
	if ep == "" {
		ep = "nats://localhost:4222"
	}
	return &Config{Endpoint: ep, MaxRetries: 3, AckWait: 2 * time.Second, MaxAckPending: 10, NakBackoffStep: 20 * time.Millisecond}
}

func newTestMQ(t *testing.T) (*mq.MQ, func()) {
	t.Helper()
	h, cleanup, err := NewClient(testConfig())
	if err != nil {
		t.Skipf("nats unavailable: %v", err)
	}
	return New(h), cleanup
}

var testSeq atomic.Uint64

// uniqTopic returns a message subject unique across every test invocation —
// including -count repeats within one process — so each run gets its own
// JetStream stream and durable consumer with zero leftover state. JetStream
// streams and durables persist (file storage, MaxAge 7d); reusing a fixed
// subject lets a durable that acked a message in a prior run skip the
// re-published message, and async Consume teardown under -count can leave a
// same-named durable half-bound. A unique subject sidesteps both. The stream is
// deleted on cleanup so orphan streams don't accumulate across runs.
func uniqTopic(t *testing.T, base string) string {
	t.Helper()
	seq := testSeq.Add(1)
	topic := fmt.Sprintf("t-%s-%d", base, seq)
	t.Cleanup(func() { deleteStream(t, streamName(topic)) })
	return topic
}

// deleteStream removes a stream (and its consumers) if it exists; errors are
// ignored so cleanup is idempotent.
func deleteStream(t *testing.T, name string) {
	t.Helper()
	h, cleanup, err := NewClient(testConfig())
	if err != nil {
		return
	}
	defer cleanup()
	_ = h.js.DeleteStream(context.Background(), name)
}
