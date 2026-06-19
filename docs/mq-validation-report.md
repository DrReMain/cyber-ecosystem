# MQ Capability Test Report

Full test coverage for the MQ capability (`shared-go/mq` + NATS JetStream +
`edge_mobile` Platform wiring). All tests run against a **real NATS JetStream**
instance (k3d, `localhost:4222`) — no mocks.

- **Capability commits:** `c26caba5` (impl) · `dbd6c4ff` (hardening + this report).
- **Test totals:** 17 automated tests — 3 pure-logic (`shared-go/mq`) + 14
  NATS-backed (`shared-go/mq/nats`) — plus a temporary 3-transport stack probe
  (reverted after passing).
- **Status:** all green, including `go test ./shared-go/mq/... -count=10 -race`
  (no data races, no failures). `./nx run tools:go:test` + `tools:go:lint` green.

## Capability → test coverage matrix

| Capability | Behavior verified | Test(s) | Result |
|---|---|---|---|
| **Publish** | message stored; ack sequence returned as id | `TestPublishSubscribeRoundTrip`, hardening tests | ✅ |
| **Publish** | empty/over-length topic rejected | `TestValidateTopicGroup`, `TestNewClientConfigValidation` | ✅ |
| **Subscribe** | single consumer receives a published message | `TestPublishSubscribeRoundTrip` | ✅ |
| **Subscribe** | empty/over-length group rejected | `TestValidateTopicGroup` | ✅ |
| **Competing group** | same group, N msgs → each delivered exactly once | `TestGroupCompetingWorkQueue`, `TestHighVolumeNoLoss`, `TestConcurrentPubSub` | ✅ |
| **Broadcast** | different groups → every group gets every message | `TestGroupBroadcast` | ✅ |
| **Durable resume** | ack+close, later subscriber resumes and skips acked msgs | `TestDurableResume` | ✅ |
| **Retry** | handler error → `Nak` redelivery up to `MaxRetries` | `TestConsumerRetryThenDLQ` | ✅ |
| **Dead-letter (DLQ)** | at `MaxRetries` → dead-lettered + acked off original | `TestConsumerRetryThenDLQ`, `TestDLQHeaderFidelity` | ✅ |
| **Payload fidelity** | binary (all 256 byte values: null/0xFF/control) byte-exact | `TestBinaryPayloadRoundTrip` | ✅ |
| **Payload fidelity** | 1 MiB large payload round-trips intact | `TestLargePayloadRoundTrip` | ✅ |
| **Header fidelity** | ASCII keys carry CJK/emoji/special-char **values** + empty value | `TestHeaderFidelity` | ✅ |
| **Header fidelity** | DLQ metadata (`mq-original-topic`/`mq-delivered`/`mq-error`/`mq-orig-*`) | `TestDLQHeaderFidelity` | ✅ |
| **High volume** | 500 messages, no loss/duplicate | `TestHighVolumeNoLoss` | ✅ |
| **Concurrency** | 4 publishers × 50 msgs under `-race`, exact-once, no race | `TestConcurrentPubSub` | ✅ |
| **Error mapping** | `ErrInvalidArgument`→3400, `Unavailable`→3401, `Timeout`→3402 | `TestHandleMQError`, `TestValidateMQDefaultError` | ✅ |
| **Config validation** | nil config / empty endpoint → `ErrInvalidArgument` | `TestNewClientConfigValidation` | ✅ |
| **Fault tolerance** | dead endpoint → `ErrUnavailable` | `TestNewClientFaultUnavailable` | ✅ |
| **Transport stack** | round-trip + 3400 error across gRPC/HTTP/Connect | probe (temp, reverted) | ✅ |

## Boundary values exercised

- **Payload size:** 0 bytes (empty), 256 bytes (every byte value), 1 MiB.
- **Volume:** 500 messages to one consumer.
- **Concurrency:** 4 goroutine publishers × 50 messages, race detector on.
- **Headers:** 4 headers incl. CJK / emoji / special chars / empty value;
  non-ASCII **key** (documented as dropped).
- **Retry/DLQ:** `MaxRetries=3` (tests) / `5` (default config); `AckWait=2s` (tests).
- **Fault:** unroutable endpoint (`127.0.0.1:1`).

## Error-model coverage (`shared-go/mq/error.go`, 34xx)

| Sentinel | Maps to | HTTP | Test |
|---|---|---|---|
| `ErrInvalidArgument` | `INFRA_ERROR_MQ_INVALID_ARGUMENT` | 400 | `TestHandleMQError`, `TestNewClientConfigValidation` |
| `ErrUnavailable` | `INFRA_ERROR_MQ_UNAVAILABLE` | 503 | `TestHandleMQError`, `TestNewClientFaultUnavailable` |
| `ErrTimeout` | `INFRA_ERROR_MQ_TIMEOUT` | 504 | `TestHandleMQError` (probe also exercises the path) |

## Transport stack probe (temp `MQProbe`, reverted)

A temporary `MQProbe` RPC on `MobileTransferService` (Platform-injected) did a
publish→subscribe round-trip through `Platform.GetMQ()` and mapped every error via
`Platform.HandleMQError`. Hit over gRPC / HTTP / Connect, happy path + error path
(empty topic → `ValidateTopic` → 3400). All passed:

| Transport | Round-trip | Error path (empty topic) |
|---|---|---|
| gRPC `:12002` | ✓ `receivedPayload`=base64(hello mq), headers kept | ✓ `InvalidArgument` + `reason: INFRA_ERROR_MQ_INVALID_ARGUMENT` |
| HTTP `:11002` | ✓ snake-case JSON | ✓ `400 {"code":400,"reason":"INFRA_ERROR_MQ_INVALID_ARGUMENT"}` |
| Connect `:13002` | ✓ camelCase JSON | ✓ `400 {"code":"invalid_argument", details[…3400]}` |

## Bugs found & fixed during validation

1. **Startup panic** — `data.mq.nats.max_age: 168h` was an invalid protobuf
   Duration (proto wants seconds); service panicked on boot. Fixed to `604800s`
   (7d). The hardening tests never load `config.yaml` (they use `testConfig()`
   with `MaxAge=0`→default), so only running the real service exposed it.
2. **Flaky test signal** (pre-commit) — `TestPublishSubscribeRoundTrip` failed
   under `-count`. Root cause (found via timestamped instrumentation: handler ran
   at +13ms yet the select timed out) was a test-side race: non-blocking send on
   an *unbuffered* signal channel can lose against a multi-case select under load.
   The consumer was correct. Fix: buffered channel. Applied to round-trip +
   durable-resume tests.

## Known limits (documented, by design or backend)

- **NATS header keys are ASCII-only** (HTTP-token chars); non-ASCII keys are
  silently dropped by the backend. `headerToMap` faithfully maps what arrives.
  Codified in `TestNonASCIIHeaderKeyDropped`. → Use ASCII keys, put unicode in
  the value.
- **At-least-once delivery.** Idempotent handling of duplicates is the
  application's responsibility (standard for JetStream). Exactly-once is out of
  scope.
- **MQ error messages render empty** — `defaultMQError` sentinels are built with
  an empty format string; the reason code carries semantics, the original error
  is attached via `WithCause`. Consistent with cache/storage.
- **Stream-size upper bound (1 GiB) not stress-tested** — payloads only up to 1
  MiB; hitting the cap is a configuration concern, not a code path.
- **Delayed/scheduled delivery is out of scope** — belongs to a scheduler
  capability, not MQ (decided during design).
- **PG backend** is designed-for-later but not implemented; only NATS is tested.

## How to run

```bash
./nx run tools:go:test                       # full workspace (incl. mq), green
go test ./shared-go/mq/... -count=10 -race   # MQ suite stress, green
go test ./shared-go/mq/nats/ -v              # per-test detail
```

NATS must be reachable at `NATS_ENDPOINT` (default `nats://localhost:4222`);
tests `t.Skip` if it is down.
