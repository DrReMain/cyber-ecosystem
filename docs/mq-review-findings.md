# MQ Capability — Multi-Expert Code Review Findings

Review of commit `a91febea` (the MQ capability). Method: 3 expert lenses
(distributed-architect / Go / NATS-JetStream) × 2 cross rounds. Round 1 =
independent findings; Round 2 = each expert cross-adjudicated all R1 findings
against the actual code + nats.go **v1.52.0** source, settling severity debates
and refuting non-issues. This document is the synthesized final verdict.

Severity = my orchestration call after weighing the 3 R2 verdicts + evidence.
`file:line` cited throughout.

## Resolution (production-readiness pass)

Addressed in the fix pass (`fix(mq): production-readiness`):

- **P0-1 (Creds→Token)** — `Creds` now loads via `nats.UserCredentials(path)`.
- **P0-2 (DLQ-fail-then-Ack loss)** — decision extracted to `decideAck`; DLQ-write failure → `NakWithDelay` (retain), success → `Term`. Never acks-and-loses. Unit-tested.
- **P0-3 (MaxAckPending storm / false-DLQ)** — `MaxAckPending` added (default 256, configurable) on the consumer; `AckWait` must exceed handler p99 (documented). Note: full slow-handler protection needs an `InProgress` heartbeat, which would change the handler signature — documented as a known limit.
- **P0-4 (subscription lifecycle)** — backend `handle` tracks subscriptions; the client cleanup drains all (`Drain`+`Closed`) before `nc.Drain()`. `subscription.Close()` uses the same graceful drain.
- **P1-1 (Close waits for callback)** — `Drain()` + `<-Closed()` (bounded) in `Close`/`drain`.
- **P1-2 (ErrTimeout dead arm)** — `nats.ErrTimeout` + `*jetstream.APIError`(504) → `ErrTimeout`. Unit-tested.
- **P1-3 (Message.ID portability)** — documented `ID` as opaque/backend-local; idempotency via `Headers`.
- **P1-7 (DLQ eviction)** — `mq-dlq` gets dedicated longer retention (`DLQMaxAge`/`DLQMaxBytes`, default 30d).
- **Design-doc divergences fixed** — `NakWithDelay` (linear backoff, `NakBackoffStep`), durable name `<group>-<topic>`, `Term` for terminal poison, `meta==nil` → Term (no infinite loop), `ValidateTopic`/`ValidateGroup` reject non-token chars.

Explicitly **deferred** (not correctness): H (stream/consumer caching, perf); P1-4 (`Nats-Msg-Id` publish-dedup); per-topic DLQ streams; metrics/logging + reconnect tests (observability phase); NEW-R/NEW-S operational edges. P1-5 (Term+native-advisory) adopted partially (`Term` used; native advisory stream not added — app-DLQ retained for replay value). P1-6 (contract portability) handled via docs, not contract rework (PG backend deferred).

New tests: `TestMapErrorTimeout`, `TestConsumerMaxAckPendingSet`, `TestSubscriptionCloseStopsCallbacks`, `TestDecideAck`, expanded `TestValidateTopicGroup`. All green incl. `-count=10 -race`; service boots with the expanded config.

---


## P0 — fix before any real MQ traffic

### P0-1. `Creds` is wired as `nats.Token()`, not `UserCredentials()` — auth always fails on a secured NATS
`shared-go/mq/nats/client.go:27-29`. **3/3 CONFIRMED.** nats.go: `Token()`
(`nats.go:1355`) sets a bearer auth_token sent verbatim in CONNECT;
`UserCredentials()` (`nats.go:1380`) is the JWT+NKey/seed challenge-response
path. The field is named `Creds` and the config doc implies a creds/JWT file.
Passing a `.creds` file path/content through `Token()` → auth always fails AND
ships the creds blob in the CONNECT handshake (exposure). Uncaught only because
dev runs unauthenticated. **Fix:** dispatch on intent — `UserCredentials(path)`
for a file, or rename the field to `Token`/`AuthToken` and document it.

### P0-2. DLQ-fail-then-Ack silently drops poison messages
`shared-go/mq/nats/consumer.go:65-68`. **3/3 confirmed real** (severity P0
architect / P1 Go+MQ — I rank P0: silent loss of the worst messages, cheap fix).
```go
_ = c.dlq(...)   // error discarded
_ = msg.Ack()    // unconditional
```
If the DLQ publish fails (transient NATS blip, stream not yet created, cancelled
ctx), the message is acked off the source AND never written to DLQ → **permanent
silent loss**. Symmetric: DLQ ok but Ack fails → redelivered → duplicate DLQ.
**Fix:** on DLQ-publish failure, `Nak`/`Term` (retain in source), never Ack;
only Ack after a confirmed DLQ write; make the DLQ write idempotent (Nats-Msg-Id
from stream sequence).

### P0-3. No `MaxAckPending` + slow handler → redelivery storm + false-positive DLQ
`shared-go/mq/nats/consumer.go:41-48` (AckWait default 30s, MaxDeliver=-1, no
MaxAckPending). **REFRAMED in R2** (important): the R1 "in-process concurrent
self-compete" framing was **REFUTED** — nats.go dispatches the `Consume` callback
serially on one goroutine (`pull.go:287`), so there's no in-process race. The
real, confirmed hazard: a handler slower than `AckWait` → unacked messages
redeliver (bounded by the `MaxAckPending` server default **1000**, MQ-expert via
`consumer_config.go:188`) → `NumDelivered` inflates → the `>= maxRetries` gate
DLQs **slow-but-healthy** messages as if they were poison. **Fix:** set
`MaxAckPending` explicitly (e.g. 10–50), raise `AckWait` to comfortably exceed
handler p99, support handler-side `InProgress()` heartbeat; document the contract.

### P0-4. No subscription lifecycle on shutdown
Architect (R2 NEW). `Platform` (`platform.go`) holds `*mq.MQ` but tracks no
subscriptions; nothing in the `kratos.App` shutdown path calls
`subscription.Close()`. On shutdown `nc.Drain()` runs while consumer goroutines
are still active → in-flight `Ack`/`Nak`/`dlq` fail silently. No "stop consuming
→ finish in-flight → drain conn" ordering. Latent today (no business `Subscribe`
caller yet — the probe was reverted) but must be solved before adoption.
**Fix:** Platform-level subscription registry + ordered shutdown.

---

## P1

### P1-1. `subscription.Close()` doesn't wait for the in-flight callback
`consumer.go:22-28`. **3/3 CONFIRMED.** `cc.Stop()` is async (nats.go flips a
flag, returns immediately — `pull.go:768`); `cancel()` fires at once, so a handler
already running sees a cancelled `cctx` and its `dlq`/`Ack`/DB-write aborts
mid-flight. **Fix:** `cc.Stop()` then wait on `cc.Closed()` (bounded) before
`cancel()`.

### P1-2. `ErrTimeout` arm is dead; timeouts misclassified as 503
`shared-go/mq/nats/error.go:15-29`. **3/3 CONFIRMED.** Only
`context.DeadlineExceeded` maps to `ErrTimeout`; JetStream timeouts are
`jetstream.ErrTimeout`/`nats.ErrTimeout` (typed, not wrapping ctx.Deadline) and
the impl never sets its own deadline → all real timeouts fall through to `default`
→ `Unavailable` (503). The 504 path + the `Timeout` sentinel slot are unreachable.
**Fix:** add `errors.Is(err, natsclient.ErrTimeout)` and `errors.As(*jetstream.APIError)`
branches; consider per-op `context.WithTimeout`.

### P1-3. `Message.ID` = stream sequence — unstable idempotency key
`consumer.go:59`. **3/3 CONFIRMED.** Resets to 1 on stream purge/delete/recreate;
unique only within one stream. Handlers deduping on `ID` (expected under
at-least-once) dedup incorrectly after any stream lifecycle event. **Fix:**
document `ID` as opaque/unstable; recommend a publisher-supplied idempotency key
in `Headers` (round-trips faithfully).

### P1-4. No `Nats-Msg-Id` → publish retries silently duplicate
`publisher.go:31`. MQ-expert. `js.PublishMsg` with no MsgId and the stream's
`DuplicateWindow` default 0 (off) → a Publish retry (timeout-then-success) creates
a downstream duplicate. **Fix:** set `Nats-Msg-Id` from a producer-supplied key.

### P1-5. Poison handling should use `Term()`, not app-Nak-to-DLQ
MQ-expert (design). JetStream's idiomatic poison path is `msg.Term()` (never
redeliver) + the native `MaxDeliver`-exceeded advisory stream. The code instead
Naks until an app-side counter trips and re-publishes to a second stream —
non-idiomatic, emits no advisories, gives operators two places to monitor.
**Fix (design-level):** prefer `Term()` for unrecoverable errors + native
advisory; reserve the app-DLQ stream for transformation/replay.

### P1-6. The "backend-agnostic" contract leaks JetStream semantics
Architect (R2 NEW-3). `Message.ID`=stream-seq (P1-3) and `NumDelivered`-based
retry gating (`consumer.go:65`) have **no PG analog** (PG `SKIP LOCKED` has no
delivery counter). A faithful PG backend needs its own `delivery_count` column +
its own DLQ table + a different ID scheme. Any business building on `Message.ID`
or JetStream retry semantics will not port. **Fix:** narrow the contract —
document `ID` as opaque/unstable, move retry-count semantics out of the portable
interface (or accept the PG backend implements a defined superset).

### P1-7. Shared `mq-dlq` stream couples topics + evicts poison under load
`stream.go:13,32-41`. Architect + MQ. One `mq-dlq` stream, subjects `dlq.<topic>`,
single `MaxAge`/`MaxBytes` inherited from live streams → one tenant's poison storm
fills the shared cap and **evicts every other tenant's poison messages** (silent
loss of exactly the messages DLQ exists to preserve). **Fix:** per-topic (or
per-tenant) DLQ stream, or longer/limitless retention for DLQ.

---

## P2 (condensed)

- **H** `ensureStream` + `CreateOrUpdateConsumer` on every Publish/Subscribe → per-op control-plane RTT (3/3). Cache stream/consumer handles after first use.
- **F** `MaxRetries` off-by-one/naming: NumDelivered is 1-based, `>= maxRetries` DLQs on the Nth delivery (= N-1 retries); self-consistent but the name/comment mislead (3/3). Pin + document semantics.
- **M** `Nak` has no backoff (`consumer.go:70`) → tight retry loop on fast-failing handler; **diverges from the design doc** (`design.md:141` specifies `Nak(delay)`). Use `NakWithDelay`.
- **I** `meta==nil` skips the DLQ gate (`consumer.go:65`) → infinite Nak loop. Guard: treat nil-meta as terminal Ack/Term.
- **J (reframed)** durable name is **group-only** (`consumer.go:42`), not `<group>-<topic>` as the design doc says → two topics sharing a group collide on one durable. (R1's "config-drift ping-pong" was REFUTED — identical config literal = no-op update.)
- **ValidateTopic/Group length-only** (`validate.go`) — accepts spaces, `.`, `>`, `*`, control chars → NATS subject injection + `dlq.<topic>` confusion (2/3). Add token-char validation.
- **Per-stream 1 GiB, no global guard** — N topics × 1 GiB = unbounded disk on a shared cluster; no account quota.
- **`mq-<topic>` collision / no tenant scoping** — topic literally named `dlq` → stream `mq-dlq` collides with the shared DLQ; multi-tenant subject squatting.
- **Zero metrics/logging** across the capability (publish/consume/dlq/nak/ack-fail silent) — breaks the repo's OTel mandate.
- **Unknown errors → 503** (`error.go` default) masks non-retryable bugs (e.g. permission) as retryable.
- **`Publish(ctx, topic, nil)` panics** (`publisher.go:23` — no nil guard).
- **Slice-aliasing hazard** — handler receives `msg.Data()` backing array; async handlers retaining `Payload` past return alias recycled buffer. Document "copy if retained."
- **No reconnect handler / no reconnect-mid-op test** — `MaxReconnects(-1)` set but no `DisconnectedErrCB`/`ReconnectedCB`; partition→redelivery path untested.
- **NEW-R** `CreateOrUpdateStream` silently shrinks an existing stream if `MaxBytes` is lowered → silent eviction on a config "improvement."
- **NEW-S** `DeliverAllPolicy` + a purged durable → full-stream replay storm on re-subscribe.
- **`headerToMap` drops multi-value headers** (keeps `v[0]`) (3/3) — document the single-value contract.

---

## Refuted / downgraded in Round 2 (transparency)

- **`nc.Drain()` hang (R1 P0)** — **REFUTED** by Go-expert w/ nats.go source: `Drain()` is non-blocking (`nats.go:6182`), internally bounded by `DefaultDrainTimeout=30s` then force-close (`nats.go:6126-6172`). Not a hang; minor ≤30s shutdown + swallowed `ErrDrainTimeout`.
- **C "in-process concurrent self-compete" (R1 P0)** — **REFUTED** (nats.go serializes the callback, `pull.go:287`). Reframed as the MaxAckPending redelivery storm (P0-3) — same bad *outcome*, different (correct) mechanism.
- **J "config-drift ping-pong" (R1 P1)** — **REFUTED** (identical config literal = no-op upsert). Reframed as the durable-naming bug (P2).
- **Non-issues dropped:** `HandleMQError` per-call re-validate (cheap, harmless); zero-value==default config helpers (intentional, matches cache/storage); `group_test` AckWait-boundary flakiness (buffered-channel fix already resolved it).

---

## Recommended fix order

1. **P0-1 (Creds/Token)** — one-line branch; unblocks any secured NATS deploy. Trivial.
2. **P0-2 (DLQ-fail-then-Ack)** — rework the DLQ error path (Nak/Term on DLQ failure). Small, high-value.
3. **P0-3 (MaxAckPending + AckWait + contract doc)** — config + consumer-config + docs. Medium.
4. **P1-1 (Close waits for callback)** — `cc.Stop()` + `<-cc.Closed()`. Small.
5. **P1-2 (ErrTimeout mapping)** — add nats.ErrTimeout / `*jetstream.APIError` branches. Small.
6. **P1-3 / P1-4 / P1-6 (ID + Nats-Msg-Id + contract portability)** — decide the idempotency story together; touches the portable contract + publisher. Design call.
7. **P0-4 + P1-7 (subscription lifecycle + per-topic DLQ)** — Platform registry + ordered shutdown; DLQ isolation. Architectural, do before business adoption.
8. P1-5 (Term vs app-DLQ), P2s — hardening pass.

CI is currently green because the tests exercise only fast happy-paths with short
AckWait and no slow handlers, metadata failures, auth, shutdown races, or
reconnect — every P0/P1 above is live but untested.
