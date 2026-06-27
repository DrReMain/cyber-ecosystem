# Kratos Service Conventions

**Scope:** Rules for kratos-based Go services in this repo (`app/services/*`). Prescriptive — `MUST` / `SHOULD` / `MAY`. For the *why*/design rationale, see `ARCHITECTURE.md` (TODO). For proto contract rules, see `docs/proto/CONVENTIONS.md` (TODO).

---

## 1. Layering & dependencies

Layered; dependencies point **inward** (outer → inner, never reverse):

```
cmd/app   wire DI assembly
server    transports (HTTP / gRPC / Connect) + middleware chain
service   RPC handlers — thin
biz       use cases (UC) + entities + repo PORTS (interfaces)
data      repo impls of biz ports (ent-backed)
platform  capability facade (cache / db / storage / mq)
conf      config (proto-defined, generated)
```

- `MUST` keep business logic in `biz`; `service` is thin; `data` is persistence only.
- `MUST NOT` let `biz` import `data`/`service`/`server`/`platform`. `biz` depends on its ports + `gen/` + `shared-go/`, and MAY take its config (`*conf.X`) directly in a constructor — same injection pattern as `server`/`platform` (a config block is a `wireApp` param, consumed by the provider that needs it; no adapter functions). Direction is `data → biz` (data imports biz to satisfy ports), never the reverse.
- **A UC depends on repo ports, never on another UC.** Cross-aggregate *read* → the other aggregate's port. Cross-aggregate *side-effect* → a **domain event** (publisher doesn't know subscribers), not a direct UC call.
- Two aggregates constantly reaching into each other are probably **one** aggregate — merge or redraw the boundary.
- **Events + outbox** are `MAY` until a real cross-aggregate side-effect appears (e.g. "on register, enqueue welcome email"). Do not scaffold an event bus/outbox speculatively; `shared-go/mq` is there when needed.

---

## 2. Naming

| Concept | Rule | Example |
|---|---|---|
| Package | one per layer, flat | `biz`, `data`, `service` |
| biz file | one per aggregate: `<domain>.go` | `user.go`, `resource.go` |
| biz shared file | `biz.go` (cross-aggregate infra) | `biz.go` |
| Entity (aggregate root) | `<Domain>`, singular, **no redundant prefix** | `User` |
| Repo port (interface, in biz) | `<Entity>RP` | `UserRP` |
| Use case | `<Name>UC` + `New<Name>UC` | `UserUC`, `NewUserUC` |
| UC method | verb, business semantics | `Create`, `Update` |
| data repo impl | `<entity>RP` + `New<Entity>RP` → `<Entity>RP` | `userRP`, `NewUserRP` |
| ent↔biz mapping fn | `map<Entity>` | `mapUser` |
| service struct | `<Name>Service` | `UserService` |
| domain-dependency field | lowercase-first of its type | `userRP UserRP`, `userUC *biz.UserUC` |

- Entity names carry no redundant layer/service prefix (`User`, not `CoreUser`) — the package already gives context.
- A coordinator UC that spans aggregates (when one is warranted) is named by **capability** — e.g. a hypothetical `AuthUC` for an auth flow — not by an entity it doesn't own.
- Infra fields keep conventional short names (`log *slog.Logger`); the lowercase-type rule is for domain deps only.

---

## 3. File structure & ordering

**One file per aggregate — absolute, never split by concern.** `<domain>.go` bundles the aggregate's entity + port + UC + auxiliary code together. `MUST NOT` split one aggregate across files (no `<domain>_uc.go` / `<domain>_fsm.go` / `<domain>_repo.go`). Consistency overrides file size — 30 lines or 1000, it is one file; if it becomes unwieldy, the **aggregate boundary** is wrong (split into two aggregates), do not fragment one. `biz.go` holds only cross-aggregate infra (`Transaction`, `UC` base, `ProviderSet`). Tests go in `<domain>_test.go`.

**Section order** — each section preceded by a divider (§4); omit any section with no content:

| Layer | Sections (in order) |
|---|---|
| biz | `Entity` → `Port` → `UC` → `Method` → `Private` |
| data | `Repo` → `Private` |
| service | `Struct` → `Handler` → `Private` |

- biz `Entity` = **data shape only** (struct + value objects + list in/out types).
- **`Private` is the trailing catch-all for everything outside the fixed sections** — domain logic (FSM, entity-behavior methods like `TransitionTo`), helpers, constants. It is always last. (This is how non-fixed code — FSM and anything similar — is categorized: it is auxiliary, it goes in `Private`.)

**Imports** — five groups, blank-line separated, in order: (1) stdlib (2) third-party (3) `cyber-ecosystem/shared-go/...` (4) `cyber-ecosystem/gen/...` (5) `cyber-ecosystem/app/services/<svc>/internal/...`.

**Struct fields** — embedded base first (`biz.UC`), then `log`, then dependencies.

**Method order (by usage frequency)** — applies **uniformly** to proto RPCs, biz Port interface, biz UC methods, data repo methods, and service handlers (all five in the same order):
1. `Create`
2. `Update` (+ variants, e.g. `UpdateStatus`)
3. `Delete`
4. Read: `List` (general query) → `Get` (by id) → `GetByXxx` / `ExistsXxx` (by other conditions)
5. Other (`Sort`, and other capability methods)

**ProviderSet** — constructors in **alphabetical** order.

---

## 4. Comments

**Section dividers are the ONLY top-level markers.** `MUST NOT` add godoc-style declaration comments on individual symbols within a section — e.g. no `// <Type> describes ...` above a `type` declaration; the divider already groups and identifies the block, the name already says what it is.

**Divider format** — `// <Name>` followed by `-` to a consistent column (~110 chars for `biz`/`data`, ~120 for `service`). Copy the dash run from a neighbor file in the same package.

**Inline comments** — only for non-obvious *why* (a trade-off, a gotcha, a concurrency note). `MUST NOT` restate what the code already says; no `// create user` above `func Create`, no leftover scaffolding.

---

## 5. Layer specifics

**biz** — multi-op atomicity via `uc.tm.InTx(ctx, func(ctx context.Context) error { ... })`. Domain errors via `errorspb.ErrorXxx("")` (generated from proto error enums); infra errors propagated from `data` (already mapped).

**data** — repo `<entity>RP` embeds `RP{log, platform}`; `New<Entity>RP(logger, *platform.Platform) <Entity>RP`. Each method: build ent query/mutation → `Save/All/Exec` → on error `return ..., rp.platform.HandleEntError(err)`. Map ent↔biz via `map<Entity>(d *ent.<Ent>) *<Entity>` (ent entity name may differ from biz entity — the DB schema name is a separate concern). No business rules here.

**service** — `<Name>Service` embeds `pb.Unimplemented<X>Server`, holds `log` + the UC(s) it drives; implements `Registrar` (`RegisterGRPC/HTTP/Connect`). Handler signature `(ctx context.Context, in *pb.XxxRequest)` — the request parameter is always `in` (streaming handlers' request param too). Thin: `in.GetXxx()` → call UC → map to proto (`toProtoXxx`). No business logic, no direct repo access.

---

## 6. Checklist — adding a new aggregate / UC

1. biz `<domain>.go`: `Entity` → `Port` → `UC` → `Method` (→ `Private` if there's auxiliary code). Register `New<Name>UC` in `biz.ProviderSet`.
2. data `<entity>_rp.go`: `<entity>RP` impl + `New<Entity>RP` + `map<Entity>`. Register in `data.ProviderSet`.
3. service `<name>.go`: `<Name>Service` + handlers + Registrar. Register in `service.ProviderSet` + `NewRegistrarList`.
4. `./nx run <service>:generate:wire` then `./nx run <service>:build` (replace `<service>` with the project name). Verify dependency direction is UC→port only (no UC→UC).
