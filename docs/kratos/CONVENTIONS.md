# Kratos Service Conventions

**Scope:** Rules for kratos-based Go services in this repo (`app/services/*`). Prescriptive — `MUST` / `SHOULD` / `MAY`.

This document states **constraints, rationale, and anti-patterns** — the things code alone can't enforce or explain. It deliberately does not restate what the code already shows; look at any service under `app/services/*` for the shape. For proto contracts see `docs/proto/CONVENTIONS.md`.

---

## 1. Layering & dependencies

Each service is a **modular monolith**: one self-contained module per aggregate root under `module/`, shared infra at the top. Dependencies point **inward** (outer → inner, never reverse), and **across modules only via port read / domain event** (see §1b).

```
cmd/app          wire DI assembly (wireApp + newApp + ...; //go:generate wire)
server           transports (HTTP/gRPC/Connect) + middleware chain + Registrar port
module/<d>/      one domain module per aggregate root, self-contained:
                   biz.go     UC + DO + repo PORT (<DO>RP / <Remote>RP)
                   data.go    local repo adapter (ent-backed); omit if a pure ACL facade
                   service.go RPC handlers (thin); implements server.Registrar
                   module.go  wire ProviderSet for this module
shared/kernel.go cross-module kernel: UC base, RP base, Transaction, DefaultTenant
client           outbound adapter: remote repo ports (ACL) — present only if the service calls out
platform         capability facade (cache / db / storage / mq)
ent              generated persistence (shared by all modules in the service)
conf             config (proto-defined, generated)
```

- `MUST` keep business logic in `biz`; `service` is thin; `data`/`client` are adapters only.
- `MUST NOT` let `biz` import `data`/`client`/`service`/`server`/`platform`. Direction is `data → biz` and `client → biz` (both import biz to satisfy ports), never reverse. biz depends on its ports + `gen/` + `shared-go/`, and MAY take `*conf.X` in its constructor (same pattern as `server`/`platform`).
- **A UC depends on repo ports, never on another UC.** Cross-aggregate *read* → the other aggregate's port; cross-aggregate *side-effect* → a **domain event**, not a direct UC call.
- Two aggregates constantly reaching into each other are probably **one** aggregate — merge or redraw the boundary.
- **Entity relations follow aggregate boundaries.** *Intra-aggregate* composition (same root, same lifecycle, never a separate service — e.g. Order→OrderLine) uses an **ent edge**; *inter-aggregate* references (independent aggregates, any cross-service possibility — e.g. User→Dept) use a **plain `<x>_id` field** resolved via a repo port (local or `<Remote>RP`). Rationale: edges lock a relation to one DB and can't span services / DTM-Temporal sagas; plain IDs migrate to a remote RP with zero refactor. Edges also clash with soft-delete cascade and add nothing over the intercept-based (tenant/datascope) column filtering already in use. Default to plain FK unless the relation is definitively intra-aggregate-and-intra-service-forever.
- **Module dependency graph MUST be acyclic (DAG).** Each module is its own package, so Go catches cross-module import cycles at compile time; keep inter-module dependencies a DAG (consumer → provider via repo port).
- Events + outbox are `MAY` until a real cross-aggregate side-effect appears. Do not scaffold speculatively; `shared-go/mq` is there when needed.

---

## 1b. Module-to-module dependency paradigm

Modules are aggregate-root boundaries. Cross-module access is constrained so the graph stays a sparse DAG, never a mesh:

- **Read (query)** — a module MAY import another module's repo port for a read (e.g. `auth` uses `user.UserRP.FindByEmail`). Read-only, no side-effect.
- **Write (side-effect)** — cross-module writes are **forbidden** as direct calls; use a **domain event** (the other module subscribes). Events + outbox are `MAY` until a real cross-aggregate write exists.
- **Keep the graph sparse** — every cross-module import is a review point. Go only blocks cycles, not meshes, so default to no cross-module dependency and add one only for a real read need.
- **Kernel is per-service** — `shared/kernel.go` (UC/RP base) lives in each service's `internal/shared`, because `RP` embeds the service-specific `*platform.Platform`. It is NOT in `shared-go`.

---

## 2. Naming

| Concept | Rule | Example |
|---|---|---|
| Package | one per **domain module** (`module/<d>/` is package `<d>`, holding biz+data+service+module.go); `shared`/`server`/`client`/`platform`/`conf` are top-level | `module/user` |
| biz file | one per module: `module/<d>/biz.go` | `module/user/biz.go` |
| DO (aggregate root) | `<Domain>`, singular, no redundant prefix | `User` |
| Repo port — local (interface, in biz) | `<DO>RP` | `UserRP` |
| Repo port — remote service (interface, in biz, ACL) | `<Remote>RP` | `<Remote>RP` |
| client file | one per remote: `<remote>.go` | `<remote>.go` |
| client adapter struct | `<remote>Client` (dials a remote — a client, not a repo) | `<remote>Client` |
| client adapter constructor | `New<Remote>Client` → biz `<Remote>RP` | `New<Remote>Client` |
| provider→biz mapping fn | `map<ProviderDO>` | `mapItem` |
| Use case | `<Name>UC` + `New<Name>UC` | `UserUC` |
| data repo impl | `<do>RP` + `New<DO>RP` → `<DO>RP` | `userRP` |
| ent↔biz mapping fn | `map<DO>` | `mapUser` |
| service struct | `<Name>Service` | `UserService` |
| domain-dependency field | lowercase-first of its type | `userRP UserRP` |

- biz holds **one** port family — repo ports (`<X>RP`). A local repo (`UserRP`, backed by `data`) and a remote repo (`<Remote>RP`, backed by `client`) are the same shape: biz injects an RP and is indifferent to where the bytes come from. `data` and `client` are just two kinds of adapter.
- A remote RP is named for the **provider** (the provider *is* the "table"); its payloads use a **neutral ACL value type** (e.g. `Item`), not a consumer DO, so several consumers share one RP.
- Two views, two names for one object: the **port** is a repo (`<Remote>RP`, biz's view — "a store I call"); the **adapter** implementing it is a client (`<remote>Client`, the client layer's view — "a thing that dials a remote").
- DO names carry no service prefix (`User`, not `SystemUser`) — the package gives context. Infra fields keep conventional short names (`log`); the lowercase-type rule is for domain deps only.

---

## 3. File structure & ordering

**One module per domain (aggregate root): `module/<d>/{biz,data,service,module}.go`.** Each file is one concern within the module: `biz.go` = DO + port + UC + aux (FSM, entity behavior); `data.go` = local repo; `service.go` = handlers; `module.go` = ProviderSet. `MUST NOT` split a concern across extra files (`<d>_uc.go` etc.). If `biz.go` becomes unwieldy, the **aggregate boundary** is wrong — split into two modules, don't fragment one. Cross-module infra (`UC`/`RP` base, `Transaction`, `DefaultTenant`) lives in `shared/kernel.go`, not in any module.

**Section order** (each preceded by a divider, §4; omit empty):

| Layer | Sections |
|---|---|
| biz | `DO` → `Port` → `UC` → `Method` → `Private` |
| data | `Repo` → `Private` |
| service | `Struct` → `Handler` → `Private` |
| client | `Adapter` → `Private` |

- biz `DO` = data shape only. `Private` = trailing catch-all (FSM, entity-behavior, helpers). Always last.
- client: `Adapter` = `<remote>Client` + `New<Remote>Client` + port-method impls; `Private` = `map<ProviderDO>`. `client.go` is the layer's infra file (shared `standardMiddleware` + `ProviderSet`) — no dividers.

**Client files** — `client.go` (shared `standardMiddleware` + `ProviderSet`) plus `<remote>.go` per remote. One remote = one repo port = one adapter = one file. Adding an RPC adds a method, **never a new file**; only a new remote opens one.

**Method order (by frequency)** — uniform across proto RPCs, biz Port, biz UC methods, data repo methods, service handlers:
1. `Create` 2. `Update` (+variants) 3. `Delete` 4. `List` → `Get` → `GetByXxx`/`ExistsXxx` 5. Other (`Sort`, …)

**ProviderSet** — constructors alphabetical.

**Imports** — five groups, blank-line separated: (1) stdlib (2) third-party (3) `cyber-ecosystem/shared-go/...` (4) `cyber-ecosystem/gen/...` (5) `cyber-ecosystem/app/services/<svc>/internal/...`.

**Proto import alias** — MUST be `<scope>pb`, never bare `pb`, so multiple proto imports never collide: `commonpb` (`cyber.shared.common.v1`), `errorspb` (`cyber.shared.errors.v1`), `extpb` (`cyber.ext.v1`), and per-service `<service>pb` (e.g. a `foo` service → `foopb`).

**Struct fields** — embedded base first, then `log`, then dependencies. Base `shared.UC`/`shared.RP` use **exported** fields (`Log`/`Tm`/`Platform`) so they stay visible across packages when embedded; service structs keep the conventional unexported `log`.

---

## 3a. proto ↔ DO pointer paradigm

Align proto optional fields, DO fields, and the service/data mapping around pointers so the layers stay free of nil-check boilerplate:

- **proto** body fields `optional` (`*T`) + `buf.validate` controls required/format (see proto CONVENTIONS §3); **path-bound `{id}` plain `string`**.
- **DO**: optional / proto-bound fields are `*T` (align with proto); mandatory system fields (`ID`, `CreatedAt`, `UpdatedAt`, `TenantID`) stay non-pointer.
- **service**: proto `*T` → DO `*T` **passed straight through — no deref, no nil-check**.
- **data**: optional fields `SetNillableX(*T)` / `ClearX`; mandatory NOT-NULL fields `SetX(*p)` (validate guarantees `*p` non-nil).
- **ent schema**: constraints by need — mandatory `.NotEmpty()` (NOT NULL), optional `.Optional().Nillable()`. Do **not** force everything Nillable: DB integrity (NOT NULL on mandatory fields) outweighs saving one deref.
- **biz**: nil-semantics (e.g. update-leave-unchanged) live in the UC.

---

## 4. Comments

- **Section dividers are the ONLY top-level markers.** `MUST NOT` add godoc-style declaration comments on symbols within a section (no `// <Type> describes ...` above a type).
- **Divider format** — `// <Name>` + `-` to ~110 chars (`biz`/`data`) or ~120 (`service`). Copy a neighbor's dash run.
- **Inline comments** — only non-obvious *why* (trade-off, gotcha, concurrency). `MUST NOT` restate the code.

---

## 5. Layer specifics

**biz** — domain errors via `errorspb.ErrorXxx("").WithCause(errors.New(<reason>))` (generated from proto error enums); **the message MUST stay empty (`""`) — it is serialized back to the client, so the real reason goes in `WithCause` (server-log only, stripped by outbound sanitize). Never write `ErrorXxx("some text")` — that leaks internals.** Infra errors come from `data`/`client` already mapped (see §6). Multi-op atomicity via `uc.Tm.InTx`. An aggregate FSM lives in `Private` (`looplab/fsm`; `TransitionTo(ctx, target)` returns a domain error on illegal transition).

**data** — repo `<do>RP` embeds `shared.RP{Log, Platform}`; on ent error `return ..., rp.Platform.HandleEntError(err)` (maps to `InfraError`). **No business rules.**

**service** — `<Name>Service` embeds the proto `Unimplemented<X>Server`, implements `Registrar` (`RegisterGRPC/HTTP/Connect`). Thin: `in.GetXxx()` → UC → map to proto. No direct repo access.

**client** — outbound adapter for a remote repo, the counterpart of `data`. `<remote>Client` dials one connection (`grpc.NewClient` / `connect.DialInsecure`, with `standardMiddleware`), implements `<Remote>RP`. **`MUST NOT` let biz import the provider's proto** — only the client layer maps provider types → the port's neutral ACL type. Outbound middleware assembly lives in `client.go` per-service (not `shared-go`), mirroring the server: `shared-go` supplies components, the service owns the chain.

**Transports are symmetric to the app.** gRPC and Connect clients both surface errors as **kratos `*errors.Error`**, not transport-specific types — so biz/service/middleware handle one error shape regardless of transport. (gRPC does this natively; Connect normalizes at its boundary.) `web/JS` clients are out of scope here — they use `@connectrpc/connect` directly against the Connect *server*.

---

## 6. Errors & observability

**Two-layer error model** (the security boundary):
- **To the outside (downstream client):** only predefined error enums (`errorspb`) — never leak internals. Outbound errors from a remote provider are remapped into this service's own error space: `shared-go/kratos/middleware/sanitize` `Client()` collapses any provider error (transport failure or provider business error) by HTTP code to a clean `GeneralError`/`InfraError`, regenerating reason/message — no upstream reason or connection detail crosses back. `Server()` masks any non-kratos inbound error the same way.
- **To the inside (own logs):** the **original**, detailed error — the logging middleware records it with `trace_id`. The erroring service's own log holds the detail; `trace_id` chains per-service logs across the call chain.

So a UC never returns a raw upstream error — the client middleware both logs the detail and returns a sanitized enum.

**Anti-patterns:** don't `return nil, err` from a client with the provider's error untouched; don't expose provider `reason` strings; don't skip the logging middleware (you lose the detail + trace correlation).

**Observability** — tracing/metrics/log middleware are wired on both server (inbound) and client (outbound) chains; trace context propagates across services over both gRPC and Connect. Outbound metrics are optional (server-side covers the callee view).

---

## 7. Checklist

**Adding a new domain module (aggregate root):**
1. `module/<d>/biz.go`: `DO` → `Port` → `UC` → `Method` (→ `Private`); UC embeds `shared.UC` (`uc.Tm` / `uc.Log`).
2. ent `schema/<do>.go`: fields + mixins; `./nx run <service>:generate:ent`, `:migrate:diff`, `:migrate:apply` (persisted only).
3. `module/<d>/data.go`: `<do>RP` + `New<DO>RP` + `map<DO>` (embed `shared.RP`, use `rp.Platform`). **Omit** this file for a pure ACL facade (no persistence).
4. `module/<d>/service.go`: `<Name>Service` + handlers + Registrar. `module/<d>/module.go`: `ProviderSet{New<Name>UC, New<DO>RP, New<Name>Service}`.
5. Wire it in: add the module's `ProviderSet` to `wireApp`; add its `*<Name>Service` param to `server.NewRegistrarList`.
6. `./nx run <service>:generate:wire` then `:build`. If it imports another module, keep it a **read** (port query) — cross-module writes go via domain event. Verify dependency direction (UC→port, no UC→UC).

**Adding a remote-service dependency (treat as a repo):**
1. biz: declare `<Remote>RP` with neutral ACL payloads; inject into the UC like any repo.
2. client `<remote>.go`: `<remote>Client` + `New<Remote>Client` + `map<ProviderDO>` over one connection. Register `New<Remote>Client` in `client.ProviderSet`. One file per remote.
3. conf `conf.proto` `Remote`: add endpoint field; `configs/config.yaml`: set it. `./nx run <service>:proto:conf`.
4. `./nx run <service>:generate:wire` then `:build`. Verify biz never imports the provider's proto.

**Pure ACL facade** (no own persistence): the module has **no `data.go`**; its UC depends on a `<Remote>RP` satisfied by the `client` layer. Its `module.go` `ProviderSet` lists only `New<Name>UC` + `New<Name>Service`.
