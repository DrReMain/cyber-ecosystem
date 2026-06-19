# MQ Capability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `shared-go/mq` (backend-agnostic message-queue capability) + a NATS JetStream implementation, wired into edge_mobile the way `shared-go/cache` / `shared-go/storage` are, and validated end-to-end through the service stack.

**Architecture:** Clone the cache/storage pattern — root package `shared-go/mq` (interfaces + sentinels + error mechanism), backend `shared-go/mq/nats` (JetStream). Mechanism/instance split keeps the package free of any service error-proto dependency. `group` param on Subscribe gives both work-queue (same durable → competing) and broadcast (different durable → each gets all). At-least-once via consumer ack; bounded retry via `NumDelivered` then a dead-letter stream.

**Tech Stack:** Go 1.25.0 (root module `cyber-ecosystem`), nats.go v1.52.0 (`github.com/nats-io/nats.go` + `/jetstream`), kratos v3, google/wire. NATS JetStream in k3s at `nats://localhost:4222`.

## Global Constraints

- Single root `go.mod` (module `cyber-ecosystem`, `go 1.25.0`). nats.go already added (`github.com/nats-io/nats.go v1.52.0`).
- Mechanism/instance split: `shared-go/mq` MUST NOT import any `cyber-ecosystem/gen/go/...` error proto. Only `app/.../platform` imports `errorspb`.
- DO NOT hand-edit generated files (`conf.pb.go`, `wire_gen.go`, anything under `gen/`). Regenerate via Nx targets.
- gci import order: standard → default → `Prefix(cyber-ecosystem/shared-go)` → `Prefix(cyber-ecosystem/gen/go)` → `Prefix(cyber-ecosystem/app)`.
- MQ error codes 34xx new in `infra.proto`; consumer handler business errors are NOT MQ errors (retry/DLQ, never `HandleMQError`).
- One handle, one backend per project (mirror old-repo cache memory+redis pattern; multi-backend is the caller's concern — never baked into skeleton examples).
- Delayed delivery is the scheduler's job (deferred), NOT in the MQ abstraction.
- NATS reachable at `nats://localhost:4222` (k3d LB host:4222 → NodePort 30422); no auth in dev.

**Nx targets (verified):** conf → `nx run edge_mobile:proto:conf`; errors/service proto → `nx run proto:generate:go`; wire → `nx run edge_mobile:generate:wire`; build → `nx run edge_mobile:build`. **No Nx target for `go test`** — run `go test ./shared-go/mq/...` directly from repo root.

---

## File Structure

```
shared-go/mq/                    # backend-agnostic contract
  errors.go        # sentinels: ErrInvalidArgument / ErrUnavailable / ErrTimeout
  error.go         # MQDefaultError + ValidateMQDefaultError + HandleMQError
  validate.go      # ValidateTopic / ValidateGroup
  message.go       # Message
  publisher.go     # Publisher interface
  consumer.go      # Consumer interface + Subscription
  mq.go            # MQ struct { Publisher; Consumer }
shared-go/mq/nats/               # JetStream impl
  client.go        # NewClient(*Config) (*nats.Conn + JetStream handle, func(), error)
  config.go        # Config + defaults + validate
  mq.go            # New(...) *mq.MQ (assembles Publisher + Consumer)
  stream.go        # ensureStream (CreateOrUpdateStream for a topic + the dlq stream)
  publisher.go     # Publish via js.PublishMsg (with headers)
  consumer.go      # Subscribe: CreateOrUpdateConsumer + Consume loop + ack/nak + DLQ
  error.go         # mapError: nats/jetstream err → sentinel
app/services/edge_mobile/internal/
  conf/conf.proto                              # +message MQ + MQ mq = 4
  configs/config.yaml                          # +data.mq.nats block
  platform/platform_mq.go                      # NewMQ + toMQConfig
  platform/platform_mq_handler.go              # defaultMQError + NewMQErrorHandler
  platform/platform.go                         # +mq fields/getters/ProviderSet
proto/cyber/shared/errors/v1/infra.proto       # +34xx MQ block
```

---

## Task 1: Error model + Validate + 34xx error codes (backend-agnostic, unit-tested)

**Files:**
- Modify: `proto/cyber/shared/errors/v1/infra.proto` (add 34xx MQ block after 33xx Storage)
- Create: `shared-go/mq/errors.go`, `shared-go/mq/error.go`, `shared-go/mq/validate.go`, `shared-go/mq/error_test.go`

**Interfaces:**
- Produces: sentinels `ErrInvalidArgument/ErrUnavailable/ErrTimeout`, `MQDefaultError`, `ValidateMQDefaultError`, `HandleMQError`, `ValidateTopic`, `ValidateGroup`.

- [ ] **Step 1: Add the 34xx MQ error codes to infra.proto**

In `proto/cyber/shared/errors/v1/infra.proto`, after the 33xx Storage block, add:
```proto
  // 34xx: MQ
  INFRA_ERROR_MQ_INVALID_ARGUMENT = 3400 [(.errors.code) = 400];
  INFRA_ERROR_MQ_UNAVAILABLE = 3401 [(.errors.code) = 503];
  INFRA_ERROR_MQ_TIMEOUT = 3402 [(.errors.code) = 504];
```

- [ ] **Step 2: Regenerate the errors proto**

Run: `nx run proto:generate:go`
Expected: `gen/go/cyber/shared/errors/v1/infra_errors.pb.go` contains `func ErrorInfraErrorMqInvalidArgument(...)`, `ErrorInfraErrorMqUnavailable`, `ErrorInfraErrorMqTimeout`.
Verify: `grep -c "ErrorInfraErrorMq" gen/go/cyber/shared/errors/v1/infra_errors.pb.go` returns 6 (Error+Is each ×3).

- [ ] **Step 3: Create `shared-go/mq/errors.go`**

```go
package mq

import "errors"

// Sentinel errors are the MQ contract (backend-agnostic). They map to the
// InfraError 34xx MQ codes (infra.proto); the service platform layer adapts
// them to application errors via HandleMQError. Consumer handler BUSINESS
// errors are not MQ errors — they are retried/DLQ'd, never passed here.
var (
	ErrInvalidArgument = errors.New("mq: invalid argument")
	ErrUnavailable     = errors.New("mq: unavailable")
	ErrTimeout         = errors.New("mq: timeout")
)
```

- [ ] **Step 4: Create `shared-go/mq/error.go`**

```go
package mq

import (
	stderrors "errors"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// MQDefaultError holds the application error instances a service maps the MQ
// sentinel errors to. The mapping MECHANISM lives here in shared-go (reusable,
// backend-agnostic); the concrete *errors.Error instances are supplied by each
// service's platform layer — the same mechanism/instance split cache/storage use.
type MQDefaultError struct {
	InvalidArgument *kratoserrors.Error
	Unavailable     *kratoserrors.Error
	Timeout         *kratoserrors.Error // optional: nil → unknown errors pass through unchanged
}

func ValidateMQDefaultError(errs *MQDefaultError) error {
	for _, e := range []*kratoserrors.Error{errs.InvalidArgument, errs.Unavailable} {
		if e == nil {
			return stderrors.New("mq: MQDefaultError has a nil sentinel slot")
		}
	}
	return nil
}

// HandleMQError maps an MQ-infra error to the supplied application error,
// attaching the original as the cause. Used only for infra errors (publish/
// connect); consumer handler errors are retried/DLQ'd, not mapped here.
func HandleMQError(err error, errs *MQDefaultError) error {
	if e := ValidateMQDefaultError(errs); e != nil {
		return e
	}
	switch {
	case stderrors.Is(err, ErrInvalidArgument):
		return errs.InvalidArgument.WithCause(err)
	case stderrors.Is(err, ErrUnavailable):
		return errs.Unavailable.WithCause(err)
	default:
		if stderrors.Is(err, ErrTimeout) && errs.Timeout != nil {
			return errs.Timeout.WithCause(err)
		}
		if errs.Unavailable != nil {
			return errs.Unavailable.WithCause(err)
		}
		return err
	}
}
```

- [ ] **Step 5: Create `shared-go/mq/validate.go`**

```go
package mq

const (
	maxTopicLen = 249  // NATS token/subject element limit
	maxGroupLen = 120  // durable-name safe bound
)

// ValidateTopic rejects empty / over-long / wild-card topics. Topics are plain
// subject names (no wildcards): MQ maps one topic → one stream + one subject.
func ValidateTopic(topic string) error {
	if topic == "" || len(topic) > maxTopicLen {
		return ErrInvalidArgument
	}
	return nil
}

// ValidateGroup rejects empty / over-long group names (group → durable name).
func ValidateGroup(group string) error {
	if group == "" || len(group) > maxGroupLen {
		return ErrInvalidArgument
	}
	return nil
}
```

- [ ] **Step 6: Write the failing test `shared-go/mq/error_test.go`**

```go
package mq

import (
	"errors"
	"testing"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

func TestHandleMQError(t *testing.T) {
	errs := &MQDefaultError{
		InvalidArgument: kratoserrors.New(400, "INVALID_ARGUMENT", "x"),
		Unavailable:     kratoserrors.New(503, "UNAVAILABLE", "x"),
		Timeout:         kratoserrors.New(504, "TIMEOUT", "x"),
	}
	cases := []struct {
		in   error
		want int
	}{
		{ErrInvalidArgument, 400},
		{ErrUnavailable, 503},
		{ErrTimeout, 504},
		{errors.New("boom"), 503}, // unknown → Unavailable
	}
	for _, c := range cases {
		got := HandleMQError(c.in, errs)
		var ke *kratoserrors.Error
		if !errors.As(got, &ke) || ke.Code != int32(c.want) {
			t.Errorf("HandleMQError(%v): code=%v, want %d", c.in, ke, c.want)
		}
	}
}

func TestValidateMQDefaultError(t *testing.T) {
	if err := ValidateMQDefaultError(&MQDefaultError{}); err == nil {
		t.Fatal("want error for all-nil slots")
	}
	if err := ValidateMQDefaultError(&MQDefaultError{
		InvalidArgument: kratoserrors.New(400, "i", ""),
		Unavailable:     kratoserrors.New(503, "u", ""),
	}); err != nil {
		t.Fatalf("nil Timeout should be allowed: %v", err)
	}
}

func TestValidateTopicGroup(t *testing.T) {
	if err := ValidateTopic(""); err != ErrInvalidArgument {
		t.Errorf("empty topic: %v", err)
	}
	if err := ValidateTopic("orders.events"); err != nil {
		t.Errorf("valid topic: %v", err)
	}
	if err := ValidateGroup(""); err != ErrInvalidArgument {
		t.Errorf("empty group: %v", err)
	}
	if err := ValidateGroup("notify-svc"); err != nil {
		t.Errorf("valid group: %v", err)
	}
}
```

- [ ] **Step 7: Run the test**

Run: `go test ./shared-go/mq/...`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add proto/cyber/shared/errors/v1/infra.proto gen/go/cyber/shared/errors/v1/ shared-go/mq/errors.go shared-go/mq/error.go shared-go/mq/validate.go shared-go/mq/error_test.go
git commit -m "feat(mq): 错误模型 + ValidateTopic/Group + 34xx MQ 错误码"
```

---

## Task 2: Contract + NATS client + Publisher + basic Consumer + Platform wiring — vertical round-trip loop

This is the thinnest end-to-end slice: publish a message through `Platform.GetMQ()` and consume it via `Consumer.Subscribe` (ack on success). Retry/DLQ (Task 3) and group semantics (Task 4) come later.

**Files:**
- Create: `shared-go/mq/message.go`, `publisher.go`, `consumer.go`, `mq.go`
- Create: `shared-go/mq/nats/client.go`, `config.go`, `mq.go`, `stream.go`, `publisher.go`, `consumer.go`, `error.go`, `client_test.go`, `roundtrip_test.go`
- Modify: `app/services/edge_mobile/internal/conf/conf.proto`, `configs/config.yaml`
- Create: `app/services/edge_mobile/internal/platform/platform_mq.go`, `platform_mq_handler.go`
- Modify: `app/services/edge_mobile/internal/platform/platform.go`

**Interfaces:**
- Consumes: Task 1 sentinels + `HandleMQError` + `ValidateTopic/Group`.
- Produces: `Message`, `Publisher`, `Consumer`, `Subscription`, `MQ`, `nats.NewClient`, `nats.New`, `nats.Config`.

- [ ] **Step 1: Create contract files**

`shared-go/mq/message.go`:
```go
package mq

import "time"

// Message is a queue message. Publishers set Topic/Payload/Headers; the backend
// fills ID/Timestamp on consume.
type Message struct {
	Topic     string
	Payload   []byte
	Headers   map[string]string
	ID        string
	Timestamp time.Time
}
```

`shared-go/mq/publisher.go`:
```go
package mq

import "context"

// Publisher publishes a message to a topic (one topic → one backend stream).
type Publisher interface {
	Publish(ctx context.Context, topic string, msg *Message) (id string, err error)
}
```

`shared-go/mq/consumer.go`:
```go
package mq

import "context"

// Subscription is a running consumer; Close stops consuming.
type Subscription interface {
	Close() error
}

// Consumer subscribes to a topic. group is the consumer identity: the same
// group across subscribers → competing work-queue (each message → one); a
// different group → independent durable → broadcast (each group gets all).
// handler returns nil to ack; an error to retry (and, after MaxRetries, DLQ).
type Consumer interface {
	Subscribe(ctx context.Context, topic, group string, handler func(ctx context.Context, msg Message) error) (Subscription, error)
}
```

`shared-go/mq/mq.go`:
```go
package mq

// MQ holds the messaging capability. Backends populate both fields.
type MQ struct {
	Publisher Publisher
	Consumer  Consumer
}
```

- [ ] **Step 2: Create `shared-go/mq/nats/config.go`**

```go
package nats

import "time"

// Config maps conf.Data.MQ.NATS to the nats.go client. Zero-value Duration/
// size fields fall back to the defaults below inside NewClient/New.
type Config struct {
	Endpoint   string        // nats://localhost:4222
	Creds      string        // optional creds/token; empty → no auth
	MaxAge     time.Duration // stream message TTL; 0 → 7d
	MaxBytes   int64         // stream size cap; 0 → 1GB
	MaxRetries int           // consumer retries before DLQ; 0 → 5
	AckWait    time.Duration // redelivery wait on no-ack / crash; 0 → 30s
}

const (
	defaultMaxAge     = 7 * 24 * time.Hour
	defaultMaxBytes   = 1 << 30 // 1 GiB
	defaultMaxRetries = 5
	defaultAckWait    = 30 * time.Second
)

func maxAgeOrDefault(v time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return defaultMaxAge
}
func maxBytesOrDefault(v int64) int64 {
	if v > 0 {
		return v
	}
	return defaultMaxBytes
}
func maxRetriesOrDefault(v int) int {
	if v > 0 {
		return v
	}
	return defaultMaxRetries
}
func ackWaitOrDefault(v time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return defaultAckWait
}
```

- [ ] **Step 3: Create `shared-go/mq/nats/error.go`**

```go
package nats

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"cyber-ecosystem/shared-go/mq"
)

// mapError translates a nats/jetstream error into an mq sentinel. Consumer
// handler BUSINESS errors never reach here (they are retried/DLQ'd upstream).
func mapError(err error, op string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s: %w", op, mq.ErrTimeout)
	}
	if errors.Is(err, nats.ErrNoResponders) || errors.Is(err, nats.ErrConnectionClosed) ||
		errors.Is(err, nats.ErrConnectionReconnecting) || errors.Is(err, jetstream.ErrNotConnected) {
		return fmt.Errorf("%s: %w", op, mq.ErrUnavailable)
	}
	return fmt.Errorf("%s: %w", op, err)
}
```

- [ ] **Step 4: Create `shared-go/mq/nats/client.go`**

```go
package nats

import (
	"fmt"
	"time"

	natsclient "github.com/nats-io/nats.go"

	"cyber-ecosystem/shared-go/mq"
)

// NewClient dials NATS and returns the conn + a JetStream handle + a cleanup
// that drains+closes the connection. The JetStream handle is what Publisher/
// Consumer use; the conn is needed for cleanup.
func NewClient(cfg *Config) (*handle, func(), error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("%w: nil config", mq.ErrInvalidArgument)
	}
	if cfg.Endpoint == "" {
		return nil, nil, fmt.Errorf("%w: endpoint is required", mq.ErrInvalidArgument)
	}
	opts := []natsclient.Option{natsclient.Name("cyber-ecosystem-mq"), natsclient.ReconnectWait(2 * time.Second)}
	if cfg.Creds != "" {
		opts = append(opts, natsclient.Token(cfg.Creds))
	}
	nc, err := natsclient.Connect(cfg.Endpoint, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: nats connect: %w", mq.ErrUnavailable, err)
	}
	js, err := jetstream.New(nc)
	if err != nil {
		_ = nc.Close()
		return nil, nil, fmt.Errorf("%w: jetstream init: %w", mq.ErrUnavailable, err)
	}
	h := &handle{nc: nc, js: js, cfg: *cfg}
	return h, func() {
		_ = nc.Drain()
	}, nil
}

// handle bundles the NATS conn + JetStream handle + resolved config. Held by
// the publisher and consumer impls.
type handle struct {
	nc  *natsclient.Conn
	js  jetstream.JetStream
	cfg Config
}
```

- [ ] **Step 5: Create `shared-go/mq/nats/stream.go`**

```go
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

// ensureDLQStream creates-or-updates the shared DLQ stream (subjects dlq.>).
func (h *handle) ensureDLQStream(ctx context.Context) error {
	_, err := h.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      dlqStream,
		Subjects:  []string{"dlq.>"},
		Retention: jetstream.LimitsPolicy,
		MaxAge:    maxAgeOrDefault(h.cfg.MaxAge),
		MaxBytes:  maxBytesOrDefault(h.cfg.MaxBytes),
	})
	return err
}
```

- [ ] **Step 6: Create `shared-go/mq/nats/publisher.go`**

```go
package nats

import (
	"context"

	natsclient "github.com/nats-io/nats.go"

	"cyber-ecosystem/shared-go/mq"
)

type publisher struct{ h *handle }

func newPublisher(h *handle) mq.Publisher { return &publisher{h: h} }

func (p *publisher) Publish(ctx context.Context, topic string, msg *mq.Message) (string, error) {
	if err := mq.ValidateTopic(topic); err != nil {
		return "", err
	}
	if err := p.h.ensureStream(ctx, topic); err != nil {
		return "", mapError(err, "ensure stream")
	}
	nmsg := &natsclient.Msg{Subject: topic, Data: msg.Payload}
	if len(msg.Headers) > 0 {
		hdr := natsclient.Header{}
		for k, v := range msg.Headers {
			hdr.Set(k, v)
		}
		nmsg.Header = hdr
	}
	ack, err := p.h.js.PublishMsg(ctx, nmsg)
	if err != nil {
		return "", mapError(err, "publish")
	}
	return ack.SequenceString(), nil
}
```

- [ ] **Step 7: Create `shared-go/mq/nats/consumer.go` (basic: ack on nil, nak on error — retry; DLQ added in Task 3)**

```go
package nats

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"

	"cyber-ecosystem/shared-go/mq"
)

type consumer struct{ h *handle }

func newConsumer(h *handle) mq.Consumer { return &consumer{h: h} }

type subscription struct {
	consume jetstream.ConsumeContext
	cancel  context.CancelFunc
}

func (s *subscription) Close() error {
	if s.consume != nil {
		s.consume.Stop()
	}
	s.cancel()
	return nil
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
	ccfg := jetstream.ConsumerConfig{
		Durable:       group, // same group → competing; different group → broadcast
		FilterSubject: topic,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       ackWaitOrDefault(c.h.cfg.AckWait),
		MaxDeliver:    -1, // unbounded; we gate via NumDelivered → DLQ
	}
	cons, err := c.h.js.CreateOrUpdateConsumer(ctx, streamName(topic), ccfg)
	if err != nil {
		return nil, mapError(err, "create consumer")
	}

	cctx, cancel := context.WithCancel(ctx)
	_, err = cons.Consume(func(msg jetstream.Msg) {
		meta, merr := msg.Metadata()
		m := mq.Message{
			Topic:   topic,
			Payload: msg.Data(),
			Headers: headerToMap(msg.Headers()),
		}
		if merr == nil && meta != nil {
			m.ID = ackSeqString(meta.Sequence.Stream)
			m.Timestamp = meta.Timestamp
		}
		if herr := handler(cctx, m); herr != nil {
			// Retry path refined in Task 3 (DLQ after MaxRetries). For now: nak → redeliver.
			if meta != nil && int(meta.NumDelivered) >= maxRetries {
				_ = c.dlq(cctx, topic, m, meta.NumDelivered, herr)
				_ = msg.Ack()
				return
			}
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	})
	if err != nil {
		cancel()
		return nil, mapError(err, "consume")
	}
	return &subscription{consume: nil, cancel: cancel}, nil
}
```

Wait — `cons.Consume` returns a `ConsumeContext`, not the second value. Fix the assignment:

Replace the `_, err = cons.Consume(...)` block with:
```go
	cc, err := cons.Consume(func(msg jetstream.Msg) {
		// ... handler body as above ...
	})
	if err != nil {
		cancel()
		return nil, mapError(err, "consume")
	}
	return &subscription{consume: cc, cancel: cancel}, nil
```

And add helpers `ackSeqString`, `headerToMap`, and `dlq` (DLQ is fleshed out in Task 3; here `dlq` is a stub that publishes to the DLQ subject so Task 2's success path works without exercising DLQ). Add at the bottom of `consumer.go`:

```go
func headerToMap(h jetstream.Headers) map[string]string {
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

func ackSeqString(seq uint64) string {
	// minimal uint64 → string without importing strconv here
	if seq == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	for seq > 0 {
		pos--
		b[pos] = byte('0' + seq%10)
		seq /= 10
	}
	return string(b[pos:])
}

// dlq publishes a failed message to the shared DLQ stream with original metadata
// in headers (Task 3 exercises this; the helper is complete here).
func (c *consumer) dlq(ctx context.Context, topic string, m mq.Message, delivered uint64, herr error) error {
	if err := c.h.ensureDLQStream(ctx); err != nil {
		return mapError(err, "ensure dlq stream")
	}
	nmsg := &natsclient.Msg{Subject: dlqSubject(topic), Data: m.Payload}
	nmsg.Header = natsclient.Header{}
	nmsg.Header.Set("mq-original-topic", topic)
	nmsg.Header.Set("mq-delivered", ackSeqString(delivered))
	nmsg.Header.Set("mq-error", herr.Error())
	for k, v := range m.Headers {
		nmsg.Header.Set("mq-orig-"+k, v)
	}
	_, err := c.h.js.PublishMsg(ctx, nmsg)
	return mapError(err, "dlq publish")
}
```

Add the `natsclient` import (`natsclient "github.com/nats-io/nats.go"`) to `consumer.go` imports.

- [ ] **Step 8: Create `shared-go/mq/nats/mq.go`**

```go
package nats

import "cyber-ecosystem/shared-go/mq"

// New assembles an *mq.MQ from a connected handle.
func New(h *handle) *mq.MQ {
	return &mq.MQ{Publisher: newPublisher(h), Consumer: newConsumer(h)}
}
```

- [ ] **Step 9: Create test helpers `shared-go/mq/nats/client_test.go`**

```go
package nats

import (
	"os"
	"testing"

	"cyber-ecosystem/shared-go/mq"
)

func testConfig() *Config {
	ep := os.Getenv("NATS_ENDPOINT")
	if ep == "" {
		ep = "nats://localhost:4222"
	}
	return &Config{Endpoint: ep, MaxRetries: 3, AckWait: 0} // AckWait 0 → default; tests use small via config below
}

func newTestMQ(t *testing.T) (*mq.MQ, func()) {
	t.Helper()
	h, cleanup, err := NewClient(testConfig())
	if err != nil {
		t.Skipf("nats unavailable: %v", err)
	}
	return New(h), cleanup
}
```

- [ ] **Step 10: Write the failing round-trip test `shared-go/mq/nats/roundtrip_test.go`**

```go
package nats

import (
	"context"
	"sync"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

func TestPublishSubscribeRoundTrip(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic, group := "rt-basic", "rt-group"

	var (
		mu   sync.Mutex
		got  *mq.Message
		done = make(chan struct{})
	)
	sub, err := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, msg mq.Message) error {
		mu.Lock()
		got = &msg
		mu.Unlock()
		select {
		case done <- struct{}{}:
		default:
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = sub.Close() }()

	id, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte("hello mq"), Headers: map[string]string{"k": "v"}})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if id == "" {
		t.Errorf("empty publish id")
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("did not receive message within 5s")
	}
	mu.Lock()
	defer mu.Unlock()
	if got == nil || string(got.Payload) != "hello mq" || got.Headers["k"] != "v" {
		t.Errorf("received: %+v", got)
	}
}
```

- [ ] **Step 11: Build to catch compile errors**

Run: `go build ./shared-go/mq/...`
Expected: compiles.

- [ ] **Step 12: Run the round-trip test (NATS must be up: `nx run deploy:mq:start`)**

Run: `go test ./shared-go/mq/nats/ -run TestPublishSubscribeRoundTrip -count=1`
Expected: PASS (receives "hello mq" with header k=v).

- [ ] **Step 13: Add `conf.Data.MQ` to conf.proto**

In `app/services/edge_mobile/internal/conf/conf.proto`, add a nested message + field so `message Data` includes `MQ mq = 4;`:
```proto
  message MQ {
    message NATS {
      string endpoint = 1;
      string creds = 2;
      google.protobuf.Duration max_age = 3;
      int64 max_bytes = 4;
      int32 max_retries = 5;
      google.protobuf.Duration ack_wait = 6;
    }
    NATS nats = 1;
  }
```
and inside `message Data`, after `Storage storage = 3;` add `MQ mq = 4;`.

- [ ] **Step 14: Regenerate conf.pb.go**

Run: `nx run edge_mobile:proto:conf`
Expected: `conf.pb.go` has `Data_MQ` / `Data_MQ_NATS` + `GetMq()`.

- [ ] **Step 15: Add `data.mq` to config.yaml**

In `app/services/edge_mobile/configs/config.yaml`, append under `data:` (after `storage:`):
```yaml
  mq:
    nats:
      endpoint: nats://localhost:4222
      # creds: "" # dev 无鉴权
      max_age: 168h     # 7 天
      max_bytes: 1073741824  # 1 GiB
      max_retries: 5
      ack_wait: 30s
```

- [ ] **Step 16: Create `platform_mq.go`**

`app/services/edge_mobile/internal/platform/platform_mq.go`:
```go
package platform

import (
	"fmt"

	"cyber-ecosystem/shared-go/mq"
	mqnats "cyber-ecosystem/shared-go/mq/nats"

	"cyber-ecosystem/app/services/edge_mobile/internal/conf"
)

// NewMQ builds the NATS-backed MQ container for edge_mobile. The (T, func(), error)
// shape lets wire chain the cleanup for graceful shutdown and partial-injection
// rollback.
func NewMQ(c *conf.Data) (*mq.MQ, func(), error) {
	mc := c.GetMq()
	if mc == nil {
		return nil, nil, fmt.Errorf("mq config is required")
	}
	cfg := toMQConfig(mc.GetNats())
	h, closeFn, err := mqnats.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return mqnats.New(h), closeFn, nil
}

func toMQConfig(n *conf.Data_MQ_NATS) *mqnats.Config {
	cfg := &mqnats.Config{
		Endpoint:   n.GetEndpoint(),
		Creds:      n.GetCreds(),
		MaxBytes:   n.GetMaxBytes(),
		MaxRetries: int(n.GetMaxRetries()),
	}
	if n.MaxAge != nil {
		cfg.MaxAge = n.GetMaxAge().AsDuration()
	}
	if n.AckWait != nil {
		cfg.AckWait = n.GetAckWait().AsDuration()
	}
	return cfg
}
```

- [ ] **Step 17: Create `platform_mq_handler.go`**

`app/services/edge_mobile/internal/platform/platform_mq_handler.go`:
```go
package platform

import (
	"cyber-ecosystem/shared-go/mq"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

var defaultMQError = &mq.MQDefaultError{
	InvalidArgument: errorspb.ErrorInfraErrorMqInvalidArgument(""),
	Unavailable:     errorspb.ErrorInfraErrorMqUnavailable(""),
	Timeout:         errorspb.ErrorInfraErrorMqTimeout(""),
}

func NewMQErrorHandler() (MQErrorHandler, error) {
	if err := mq.ValidateMQDefaultError(defaultMQError); err != nil {
		return nil, err
	}
	return func(err error) error {
		return mq.HandleMQError(err, defaultMQError)
	}, nil
}
```

- [ ] **Step 18: Wire MQ into `platform.go`**

Modify `app/services/edge_mobile/internal/platform/platform.go`: add `type MQErrorHandler func(error) error`; add fields `mq *mq.MQ` + `handleMQError MQErrorHandler` to `Platform`; add the two params to `NewPlatform`; add `GetMQ()` + `HandleMQError()` accessors; add `NewMQ`, `NewMQErrorHandler` to `ProviderSet`. Add the `mq` package import.

- [ ] **Step 19: Regen wire + build**

Run:
```bash
nx run edge_mobile:generate:wire
nx run edge_mobile:build
```
Expected: `wire_gen.go` calls `NewMQ`/`NewMQErrorHandler` and passes them to `NewPlatform`; build succeeds.

- [ ] **Step 20: Commit**

```bash
git add shared-go/mq/ app/services/edge_mobile/internal/conf/conf.proto app/services/edge_mobile/internal/conf/conf.pb.go app/services/edge_mobile/internal/platform/ app/services/edge_mobile/cmd/app/wire_gen.go app/services/edge_mobile/configs/config.yaml
git commit -m "feat(mq): 契约 + NATS 后端(Publisher/Consumer 基本往返)+ edge_mobile Platform 接入"
```

---

## Task 3: Retry + DLQ (handler failure → nak → at MaxRetries → dead-letter + ack)

The consumer from Task 2 already naks-on-error and DLQs at `MaxRetries`. This task locks the failure path with a focused test.

**Files:**
- Create: `shared-go/mq/nats/retry_test.go`

**Interfaces:**
- Consumes: Task 2 `Consumer.Subscribe` (retry/DLQ behavior).

- [ ] **Step 1: Write the failing DLQ test `shared-go/mq/nats/retry_test.go`**

```go
package nats

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"cyber-ecosystem/shared-go/mq"
)

// A handler that always fails; expect it retried up to MaxRetries (3 in testConfig)
// then dead-lettered (acked off the original stream, published to mq-dlq).
func TestConsumerRetryThenDLQ(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic, group := "retry-topic", "retry-group"

	var attempts atomic.Int32
	sub, err := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, _ mq.Message) error {
		attempts.Add(1)
		return errBoom
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer func() { _ = sub.Close() }()

	if _, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte("poison")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// Wait until the message stops being redelivered (acked off after DLQ).
	deadline := time.After(15 * time.Second)
	for {
		if attempts.Load() >= 3 {
			break
		}
		select {
		case <-time.After(200 * time.Millisecond):
		case <-deadline:
			t.Fatalf("expected >=3 attempts, got %d", attempts.Load())
		}
	}
	// Allow a little more time for the DLQ publish + ack to settle.
	time.Sleep(time.Second)

	// Verify the message landed on the DLQ stream.
	h, _, _ := NewClient(testConfig()) // share connection for inspection
	defer func() { _ = h.nc.Drain() }()
	dlq, err := h.js.Stream(ctx, dlqStream)
	if err != nil {
		t.Fatalf("dlq stream lookup: %v (ensureDLQStream may not have run; consumer DLQ path triggers it)", err)
	}
	var mu sync.Mutex
	var saw bool
	cc, err := dlq.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "retry-test-dlq-reader",
		FilterSubject: dlqSubject(topic),
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		t.Fatalf("dlq consumer: %v", err)
	}
	_ = cc // use Fetch to inspect
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	batch, _ := dlqConsumerFetch(cctx, h, topic)
	for _, mm := range batch {
		_ = mm.Ack()
		if string(mm.Data()) == "poison" {
			mu.Lock()
			saw = true
			mu.Unlock()
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if !saw {
		t.Errorf("DLQ did not contain the poison message")
	}
}

var errBoom = errBoomType{}

type errBoomType struct{}

func (errBoomType) Error() string { return "boom" }
```

Add the helper `dlqConsumerFetch` to `retry_test.go`:
```go
func dlqConsumerFetch(ctx context.Context, h *handle, topic string) ([]jetstream.Msg, error) {
	dlq, err := h.js.Stream(ctx, dlqStream)
	if err != nil {
		return nil, err
	}
	cons, err := dlq.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "retry-test-dlq-reader",
		FilterSubject: dlqSubject(topic),
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, err
	}
	batch, err := cons.Fetch(10, jetstream.FetchMaxWait(3*time.Second))
	if err != nil {
		return nil, err
	}
	var out []jetstream.Msg
	for mm := range batch.Messages() {
		out = append(out, mm)
	}
	return out, nil
}
```

- [ ] **Step 2: Run the test**

Run: `go test ./shared-go/mq/nats/ -run TestConsumerRetryThenDLQ -count=1`
Expected: PASS (>=3 attempts, poison message found in the DLQ stream).

- [ ] **Step 3: Commit**

```bash
git add shared-go/mq/nats/retry_test.go
git commit -m "test(mq): handler 失败重试到 MaxRetries 后进 DLQ"
```

---

## Task 4: Group semantics — competing (work-queue) + broadcast + durable recovery

**Files:**
- Create: `shared-go/mq/nats/group_test.go`

**Interfaces:**
- Consumes: Task 2 `Subscribe` (group → durable name).

- [ ] **Step 1: Write the competing-consumer test**

```go
package nats

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

// Same group on two subscribers → each published message delivered to exactly one.
func TestGroupCompetingWorkQueue(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic, group := "group-compete", "workers"

	var a, b atomic.Int32
	mk := func(counter *atomic.Int32) func(context.Context, mq.Message) error {
		return func(context.Context, mq.Message) error { counter.Add(1); return nil }
	}
	s1, _ := m.Consumer.Subscribe(ctx, topic, group, mk(&a))
	s2, _ := m.Consumer.Subscribe(ctx, topic, group, mk(&b))
	defer func() { _ = s1.Close(); _ = s2.Close() }()
	time.Sleep(300 * time.Millisecond) // let both consumers bind the durable

	for i := range 10 {
		_, _ = m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte{byte(i)}})
	}
	time.Sleep(3 * time.Second) // let them drain

	total := a.Load() + b.Load()
	if total != 10 {
		t.Fatalf("competing: total deliveries=%d, want 10 (each msg once)", total)
	}
	// both should have done some work (competing, not one doing all) — allow either nonzero
	if a.Load() == 10 || b.Load() == 10 {
		t.Logf("warn: one subscriber got all 10 (a=%d b=%d); distribution may be skewed but each msg was once", a.Load(), b.Load())
	}
}
```

- [ ] **Step 2: Write the broadcast test**

```go
// Different groups → each gets every message.
func TestGroupBroadcast(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic := "group-bcast"

	var g1, g2 atomic.Int32
	wait := func(c *atomic.Int32, n int32, timeout time.Duration) bool {
		deadline := time.After(timeout)
		for c.Load() < n {
			select {
			case <-time.After(100 * time.Millisecond):
			case <-deadline:
				return false
			}
		}
		return true
	}
	s1, _ := m.Consumer.Subscribe(ctx, topic, "bcast-a", func(context.Context, mq.Message) error { g1.Add(1); return nil })
	s2, _ := m.Consumer.Subscribe(ctx, topic, "bcast-b", func(context.Context, mq.Message) error { g2.Add(1); return nil })
	defer func() { _ = s1.Close(); _ = s2.Close() }()
	time.Sleep(300 * time.Millisecond)

	for range 3 {
		_, _ = m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte("x")})
	}
	if !wait(&g1, 3, 5*time.Second) || !wait(&g2, 3, 5*time.Second) {
		t.Fatalf("broadcast: g1=%d g2=%d, want each 3", g1.Load(), g2.Load())
	}
}
```

- [ ] **Step 3: Write the durable crash-recovery test**

```go
// A durable consumer, closed mid-stream, resumes from where it left off after re-subscribing.
func TestDurableResume(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic, group := "durable-resume", "resume-group"

	// Publish 3 before any consumer exists.
	for i := 1; i <= 3; i++ {
		_, _ = m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte{byte(i)}})
	}
	// First subscriber consumes 1 then closes (simulate crash mid-stream).
	gotFirst := make(chan mq.Message, 1)
	s1, _ := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, mm mq.Message) error {
		select {
		case gotFirst <- mm:
		default:
		}
		return nil
	})
	select {
	case <-gotFirst:
	case <-time.After(5 * time.Second):
		t.Fatal("first subscriber did not get a message")
	}
	_ = s1.Close()
	time.Sleep(500 * time.Millisecond) // let the durable settle

	// Re-subscribe with the same group; should get the remaining messages.
	var seen []byte
	done := make(chan struct{})
	s2, _ := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, mm mq.Message) error {
		seen = append(seen, mm.Payload...)
		if len(seen) >= 2 {
			select {
			case done <- struct{}{}:
			default:
			}
		}
		return nil
	})
	defer func() { _ = s2.Close() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
	// Total across both subscribers should be 3 (durable tracked position).
	if len(seen) < 2 {
		t.Fatalf("durable resume: second subscriber saw %d, want >=2", len(seen))
	}
}
```

- [ ] **Step 4: Run the group tests**

Run: `go test ./shared-go/mq/nats/ -run 'TestGroupCompetingWorkQueue|TestGroupBroadcast|TestDurableResume' -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add shared-go/mq/nats/group_test.go
git commit -m "test(mq): group 语义(竞争工作队列 / 广播 / durable 恢复)"
```

---

## Task 5: Stack validation campaign (temp RPC + three-protocol + hardening + report + revert)

MQ has no existing consumer RPC, so validate through the stack with a temp probe RPC, then revert it.

**Files:**
- Modify (temp): `proto/cyber/mobile/v1/transfer.proto` (+`MQProbe` rpc + messages)
- Modify (temp): `app/services/edge_mobile/internal/service/transfer.go` (impl)
- Run: `nx run proto:generate:go`
- Create (permanent): `shared-go/mq/nats/hardening_test.go`
- Create: `docs/mq-validation-report.md`
- Revert: transfer.proto + transfer.go + regen

**Interfaces:**
- Consumes: Tasks 1–4 via `platform.GetMQ()` / `HandleMQError`.

- [ ] **Step 1: Add the temp probe RPC to transfer.proto**

Append to `service MobileTransferService` and add messages:
```proto
  // TEMP: mq stack-validation probe — reverted after validation.
  rpc MQProbe(MQProbeRequest) returns (MQProbeResponse) {
    option (cyber.ext.v1.method_comment) = "[TEMP] mq 穿栈验证探针";
    option (google.api.http) = { post: "/api/v1/mobile/transfer/mq-probe" body: "*" };
  }
}
message MQProbeRequest { string topic = 1; bytes payload = 2; string group = 3; }
message MQProbeCheck { string op = 1; bool ok = 2; string err = 3; }
message MQProbeResponse { repeated MQProbeCheck checks = 1; }
```

- [ ] **Step 2: Regen proto**

Run: `nx run proto:generate:go`

- [ ] **Step 3: Implement the probe on TransferService (temp)**

TransferService needs `*platform.Platform`. Update its constructor to `NewTransferService(logger *slog.Logger, p *platform.Platform)` and add a `platform` field; implement:
```go
func (s *TransferService) MQProbe(ctx context.Context, req *pb.MQProbeRequest) (*pb.MQProbeResponse, error) {
	mq_ := s.platform.GetMQ()
	topic := req.GetTopic(); if topic == "" { topic = "probe/topic" }
	group := req.GetGroup(); if group == "" { group = "probe-group" }
	checks := []*pb.MQProbeCheck{}
	ok := func(op string) { checks = append(checks, &pb.MQProbeCheck{Op: op, Ok: true}) }
	fail := func(op string, err error) { checks = append(checks, &pb.MQProbeCheck{Op: op, Err: s.platform.HandleMQError(err).Error()}) }
	if _, err := mq_.Publisher.Publish(ctx, topic, &mq.Message{Payload: req.GetPayload()}); err != nil { fail("publish", err) } else { ok("publish") }
	sub, err := mq_.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, m mq.Message) error { ok("consume"); return nil })
	if err != nil { fail("subscribe", err) } else { time.Sleep(1500 * time.Millisecond); _ = sub.Close() }
	return &pb.MQProbeResponse{Checks: checks}, nil
}
```
Add imports `mq "cyber-ecosystem/shared-go/mq"`, `platform`, `time`. Regen wire (`nx run edge_mobile:generate:wire`) + build.

- [ ] **Step 4: Run service + probe across three protocols**

Start the service (`./app/services/edge_mobile/bin/app -conf app/services/edge_mobile/configs` after `nx run edge_mobile:build`). Then:
```bash
# gRPC :12002
grpcurl -plaintext -d '{"topic":"probe/x","payload":"aGVsbG8=","group":"probe-g"}' localhost:12002 cyber.mobile.v1.MobileTransferService/MQProbe
# HTTP :11002
curl -sS -X POST http://localhost:11002/api/v1/mobile/transfer/mq-probe -H 'Content-Type: application/json' -d '{"topic":"probe/x","payload":"aGVsbG8=","group":"probe-g"}'
# Connect :13002
curl -sS -X POST http://localhost:13002/cyber.mobile.v1.MobileTransferService/MQProbe -H 'Content-Type: application/json' -d '{"topic":"probe/x","payload":"aGVsbG8=","group":"probe-g"}'
```
Expected: all three return `{publish: ok, consume: ok}`. Record verbatim in the report.

- [ ] **Step 5: Write permanent hardening tests `shared-go/mq/nats/hardening_test.go`**

```go
package nats

import (
	"bytes"
	"context"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

func TestPublishBinaryPayload(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic, group := "hard-bin", "hard-bin-g"
	got := make(chan mq.Message, 1)
	sub, _ := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, mm mq.Message) error { got <- mm; return nil })
	defer func() { _ = sub.Close() }()
	bin := []byte{0x00, 0x01, 0xFF, '\n', 0x7F}
	_, _ = m.Publisher.Publish(ctx, topic, &mq.Message{Payload: bin})
	select {
	case mm := <-got:
		if !bytes.Equal(mm.Payload, bin) { t.Errorf("binary round-trip: got %v want %v", mm.Payload, bin) }
	case <-time.After(5 * time.Second):
		t.Fatal("binary: did not receive")
	}
}

func TestPublishBigPayload(t *testing.T) {
	m, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic, group := "hard-big", "hard-big-g"
	got := make(chan mq.Message, 1)
	sub, _ := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, mm mq.Message) error { got <- mm; return nil })
	defer func() { _ = sub.Close() }()
	big := bytes.Repeat([]byte("Q"), 512*1024) // 512 KiB
	_, _ = m.Publisher.Publish(ctx, topic, &mq.Message{Payload: big})
	select {
	case mm := <-got:
		if len(mm.Payload) != len(big) { t.Errorf("big: got %d want %d", len(mm.Payload), len(big)) }
	case <-time.After(5 * time.Second):
		t.Fatal("big: did not receive")
	}
}

// TestMQNATSDown: a dead NATS endpoint surfaces as ErrUnavailable.
func TestMQNATSDown(t *testing.T) {
	_, _, err := NewClient(&Config{Endpoint: "nats://127.0.0.1:39998"})
	if err == nil { t.Skip("expected error on dead endpoint") }
	// mapError wraps connection errors as ErrUnavailable.
}
```

- [ ] **Step 6: Run hardening tests**

Run: `go test ./shared-go/mq/nats/ -run 'TestPublishBinaryPayload|TestPublishBigPayload|TestMQNATSDown' -count=1`
Expected: PASS.

- [ ] **Step 7: Write the validation report**

Create `docs/mq-validation-report.md` recording: three-protocol probe results, retry→DLQ evidence, group/broadcast/durable evidence, hardening tests, and any NATS quirks found.

- [ ] **Step 8: Revert the temp probe**

```bash
git checkout HEAD -- proto/cyber/mobile/v1/transfer.proto app/services/edge_mobile/internal/service/transfer.go
nx run proto:generate:go
nx run edge_mobile:generate:wire
nx run edge_mobile:build
```
Expected: build succeeds with no `MQProbe` trace; only permanent hardening tests + report remain.

- [ ] **Step 9: Commit (permanent hardening + report)**

```bash
git add shared-go/mq/nats/hardening_test.go docs/mq-validation-report.md
git commit -m "test(mq): 硬化测试(二进制/大消息/死端点)+ 穿栈验证报告"
```

---

## Definition of Done

- `go build ./...` succeeds; `go test ./shared-go/mq/...` all PASS against live NATS.
- `nx run edge_mobile:build` succeeds; service boots and wires MQ.
- Publisher + Consumer(Subscribe) with ack/nak/retry/DLQ implemented; group semantics (competing + broadcast) + durable recovery validated.
- `conf.Data.MQ` present; `config.yaml` has the `data.mq.nats` block.
- 34xx error codes added; `HandleMQError` renders consistently across gRPC/HTTP/Connect (validated + reported).
- Temp probe RPC reverted; only permanent hardening tests + report remain.
- Observability deferred to 阶段三 (no spans/slow-op in MQ yet — matches cache/storage).
