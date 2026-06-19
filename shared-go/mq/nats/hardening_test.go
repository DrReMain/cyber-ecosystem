package nats

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"cyber-ecosystem/shared-go/mq"
)

// publishRecv subscribes to topic, publishes payload+headers, and returns the
// first received message (ok=false on a 5s timeout). The buffered done channel
// decouples the handler's signal from the select receiver, so delivery is
// deterministic under -count/-race.
func publishRecv(t *testing.T, m *mq.MQ, topic string, payload []byte, headers map[string]string) (mq.Message, bool) {
	t.Helper()
	done := make(chan mq.Message, 1)
	sub, err := m.Consumer.Subscribe(context.Background(), topic, "probe", func(_ context.Context, msg mq.Message) error {
		select {
		case done <- msg:
		default:
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = sub.Close() }()
	if _, err := m.Publisher.Publish(context.Background(), topic, &mq.Message{Payload: payload, Headers: headers}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	select {
	case msg := <-done:
		return msg, true
	case <-time.After(5 * time.Second):
		return mq.Message{}, false
	}
}

// Payload covering every byte value (0x00..0xFF, incl. null, 0xFF, control bytes)
// must survive a publish→subscribe round-trip byte-for-byte.
func TestBinaryPayloadRoundTrip(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	payload := make([]byte, 256)
	for i := range payload {
		payload[i] = byte(i)
	}
	msg, ok := publishRecv(t, m, uniqTopic(t, "bin"), payload, nil)
	if !ok {
		t.Fatal("did not receive binary payload")
	}
	if !bytes.Equal(msg.Payload, payload) {
		t.Fatalf("binary payload mismatch: got %d bytes, want %d", len(msg.Payload), len(payload))
	}
}

// A ~1 MiB payload (well under the 1 GiB stream cap) round-trips intact — no
// size regression and no silent truncation.
func TestLargePayloadRoundTrip(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	payload := make([]byte, 1<<20) // 1 MiB
	for i := range payload {
		payload[i] = byte(i * 7 % 251) // non-trivial pseudo-random content
	}
	msg, ok := publishRecv(t, m, uniqTopic(t, "big"), payload, nil)
	if !ok {
		t.Fatal("did not receive large payload")
	}
	if !bytes.Equal(msg.Payload, payload) {
		t.Fatalf("large payload mismatch: got %d bytes, want %d", len(msg.Payload), len(payload))
	}
}

// 500 messages to a single (competing) consumer: no loss, no duplicates.
func TestHighVolumeNoLoss(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	topic := uniqTopic(t, "vol")
	const n = 500

	var (
		mu    sync.Mutex
		seen  = make(map[int]struct{}, n)
		count atomic.Int32
		done  = make(chan struct{}, 1)
	)
	sub, err := m.Consumer.Subscribe(context.Background(), topic, "vol-group", func(_ context.Context, msg mq.Message) error {
		i := int(binary.BigEndian.Uint16(msg.Payload))
		mu.Lock()
		seen[i] = struct{}{}
		mu.Unlock()
		if count.Add(1) == n {
			select {
			case done <- struct{}{}:
			default:
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = sub.Close() }()

	for i := range n {
		var b [2]byte
		binary.BigEndian.PutUint16(b[:], uint16(i))
		if _, err := m.Publisher.Publish(context.Background(), topic, &mq.Message{Payload: b[:]}); err != nil {
			t.Fatalf("Publish %d: %v", i, err)
		}
	}
	select {
	case <-done:
	case <-time.After(20 * time.Second):
	}
	mu.Lock()
	defer mu.Unlock()
	if len(seen) != n {
		t.Fatalf("high-volume: received %d unique of %d", len(seen), n)
	}
}

// Headers with ASCII keys carry CJK/emoji/special-char VALUES and an empty
// value verbatim. (Keys must be ASCII — see TestNonASCIIHeaderKeyDropped.)
func TestHeaderFidelity(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	headers := map[string]string{
		"plain":  "value",
		"x-cjk":  "中文值-🎉",
		"x-spec": "a,b;c=d?e&f%20#",
		"empty":  "",
	}
	msg, ok := publishRecv(t, m, uniqTopic(t, "hdr"), []byte("h"), headers)
	if !ok {
		t.Fatal("did not receive headers")
	}
	for k, v := range headers {
		got, present := msg.Headers[k]
		if !present {
			t.Errorf("header %q: missing (got %v)", k, msg.Headers)
			continue
		}
		if got != v {
			t.Errorf("header %q: got %q, want %q", k, got, v)
		}
	}
}

// NATS header KEYS are restricted to ASCII (HTTP token chars); a non-ASCII key
// is silently dropped by the backend. Use ASCII keys and put unicode in the value.
func TestNonASCIIHeaderKeyDropped(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	msg, ok := publishRecv(t, m, uniqTopic(t, "hdrkey"), []byte("h"), map[string]string{"中文键": "v"})
	if !ok {
		t.Fatal("did not receive message")
	}
	if _, present := msg.Headers["中文键"]; present {
		t.Fatalf("non-ASCII header key unexpectedly preserved: %v", msg.Headers)
	}
}

// Concurrent publishers + a single competing consumer under -race: every
// message delivered exactly once, no loss/dup, no data race.
func TestConcurrentPubSub(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	topic := uniqTopic(t, "conc")
	const (
		pubs   = 4
		perPub = 50
		total  = pubs * perPub
	)

	var received atomic.Int32
	done := make(chan struct{}, 1)
	sub, err := m.Consumer.Subscribe(context.Background(), topic, "conc-group", func(_ context.Context, _ mq.Message) error {
		if received.Add(1) == total {
			select {
			case done <- struct{}{}:
			default:
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = sub.Close() }()

	var wg sync.WaitGroup
	for p := range pubs {
		wg.Go(func() {
			for range perPub {
				if _, err := m.Publisher.Publish(context.Background(), topic, &mq.Message{Payload: []byte("m")}); err != nil {
					t.Errorf("Publish p%d: %v", p, err)
					return
				}
			}
		})
	}
	wg.Wait()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
	}
	if got := received.Load(); got != total {
		t.Fatalf("concurrent: received %d of %d", got, total)
	}
}

// A dead endpoint surfaces as ErrUnavailable (nats.Connect fails synchronously:
// default RetryOnFailedConnect=false; MaxReconnects only governs post-connect).
func TestNewClientFaultUnavailable(t *testing.T) {
	_, _, err := NewClient(&Config{Endpoint: "nats://127.0.0.1:1"}) // port 1: refused
	if !errors.Is(err, mq.ErrUnavailable) {
		t.Fatalf("dead endpoint: got %v, want ErrUnavailable", err)
	}
}

// Nil config and empty endpoint both surface as ErrInvalidArgument.
func TestNewClientConfigValidation(t *testing.T) {
	cases := []struct {
		name string
		cfg  *Config
	}{
		{"nil", nil},
		{"empty endpoint", &Config{}},
	}
	for _, c := range cases {
		if _, _, err := NewClient(c.cfg); !errors.Is(err, mq.ErrInvalidArgument) {
			t.Fatalf("%s: got %v, want ErrInvalidArgument", c.name, err)
		}
	}
}

// A poison message carrying headers is retried to MaxRetries then dead-lettered;
// the DLQ entry preserves the original payload and carries the metadata headers
// (mq-original-topic / mq-delivered / mq-error / mq-orig-<key>) verbatim.
func TestDLQHeaderFidelity(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic := uniqTopic(t, "dlqhdr")
	t.Cleanup(func() { deleteStream(t, dlqStream) }) // isolate the shared DLQ stream

	var attempts atomic.Int32
	sub, err := m.Consumer.Subscribe(ctx, topic, "dlqhdr-group", func(_ context.Context, _ mq.Message) error {
		attempts.Add(1)
		return errBoom
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = sub.Close() }()

	headers := map[string]string{"trace": "abc123", "lang": "zh"}
	if _, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte("poison-with-headers"), Headers: headers}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	deadline := time.After(15 * time.Second)
	for attempts.Load() < 3 { // testConfig MaxRetries=3
		select {
		case <-time.After(150 * time.Millisecond):
		case <-deadline:
			t.Fatalf("expected >=3 attempts, got %d", attempts.Load())
		}
	}
	time.Sleep(500 * time.Millisecond) // let the DLQ publish + ack settle

	h, _, _ := NewClient(testConfig())
	defer func() { _ = h.nc.Drain() }()
	dlq, err := h.js.Stream(ctx, dlqStream)
	if err != nil {
		t.Fatalf("dlq stream: %v", err)
	}
	dc, err := dlq.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "hardening-dlq-reader",
		FilterSubject: dlqSubject(topic),
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		t.Fatalf("dlq consumer: %v", err)
	}
	batch, err := dc.Fetch(10, jetstream.FetchMaxWait(3*time.Second))
	if err != nil {
		t.Fatalf("dlq fetch: %v", err)
	}
	var found bool
	for mm := range batch.Messages() {
		hdr := mm.Headers()
		if string(mm.Data()) == "poison-with-headers" {
			found = true
			if got := hdr.Get("mq-original-topic"); got != topic {
				t.Errorf("mq-original-topic: got %q want %q", got, topic)
			}
			if hdr.Get("mq-delivered") == "" {
				t.Error("mq-delivered missing")
			}
			if hdr.Get("mq-error") == "" {
				t.Error("mq-error missing")
			}
			if got := hdr.Get("mq-orig-trace"); got != "abc123" {
				t.Errorf("mq-orig-trace: got %q want %q", got, "abc123")
			}
			if got := hdr.Get("mq-orig-lang"); got != "zh" {
				t.Errorf("mq-orig-lang: got %q want %q", got, "zh")
			}
		}
		_ = mm.Ack()
	}
	if !found {
		t.Fatal("poison message with headers not found in DLQ")
	}
}
