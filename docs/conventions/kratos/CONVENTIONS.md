# Kratos Service Conventions

**Status:** ACTIVE

**Scope:** Rules for kratos-based Go services in this repo (`app/services/*`). Prescriptive — `MUST` / `SHOULD` / `MAY`.

This document states **constraints, rationale, and anti-patterns** — the things code alone can't enforce or explain. It deliberately does not restate what the code already shows; look at any service under `app/services/*` for the shape. For proto contracts see `docs/conventions/proto/CONVENTIONS.md`.

---

## 1. Layering & dependencies

Each service is a **modular monolith**: one self-contained module per aggregate root under `module/`, shared infra at the top. Dependencies point **inward** (outer → inner, never reverse), and **across modules only via port read / domain event** (the module-dependency bullets below).

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
- **A UC depends on repo ports, never on another UC — injection is repo-port-only.** Cross-module *read* → the other module's repo port (e.g. `auth` uses `user.UserRP.FindByEmail`; `user` uses `dept.DeptRP.Get` to reject dangling dept references; read-only, no side-effect — binding a UC behind an `*RP`-shaped dependency blurs the layering). Cross-module *write* is forbidden as a direct call → a **domain event** instead (the other module subscribes). A cross-module business-rule need (not a query) is a design smell — raise it, don't wire UC→UC.
- The four files are a required-by-role set, not an exclusive list: `data.go` may be
  omitted (pure ACL facade), `service.go` may be omitted (module exposes no RPCs), and a
  module MAY add one extra file per cohesive non-layer concern when its content belongs
  to neither biz orchestration nor data lifecycle (e.g. authz's `builtin.go`,
  a static proto-reflection collector). A package-level var that reflects protos MUST
  blank-import the gen package it reflects: imported packages fully initialize first,
  sibling init order does not.
- Two aggregates constantly reaching into each other are probably **one** aggregate — merge or redraw the boundary.
- **One table, two modules, split by lifecycle operation.** A channel module (e.g. `fileproxy`, the upload proxy) and a management module (e.g. `file`) may share one table when the channel is a different concern, not a different aggregate: every write operation keeps exactly one owner — create in the channel module's own data layer, update/delete in the management module's — so no port borrowing. The channel port is named for the channel (`FileProxyRP`, not `<DO>RP`), and the channel module imports the owning module for DO/constants/mapper types only. Binding a stored reference (e.g. a user's avatar file id) validates through the owning module's port like any other cross-module read.
- **Entity relations follow aggregate boundaries.** *Intra-aggregate* composition (same root, same lifecycle, never a separate service — e.g. Order→OrderLine) uses an **ent edge**; *inter-aggregate* references (independent aggregates, any cross-service possibility — e.g. User→Dept) use a **plain `<x>_id` field** resolved via a repo port (local or `<Remote>RP`). Rationale: edges lock a relation to one DB and can't span services / DTM-Temporal sagas; plain IDs migrate to a remote RP with zero refactor. Edges also clash with soft-delete cascade and add nothing over the intercept-based (tenant/datascope) column filtering already in use. Default to plain FK unless the relation is definitively intra-aggregate-and-intra-service-forever.
- **Module dependency graph MUST be acyclic (DAG) and sparse.** Each module is its own package, so Go catches cycles at compile time — but Go blocks cycles, not meshes: every cross-module import is a review point. Default to no cross-module dependency; add one only for a real read need (consumer → provider via repo port). UC-level composition is legal only toward foundation modules (`user`, `authz`, `file` — base-service capabilities).
- **Operator faces are strippable modules.** A service's operator face (list-everyone views that may not ship in every deployment) is its own module and own proto file, depending on the business module (reads its UC + ports); the business module MUST NOT import it back — deletion is then a closed operation on the graph: remove the proto file + module + one wire/registrar line. Enforced by the per-service DAG guard (e.g. `module/agentconfig` bans `module/agentconfigadmin`).
- **Composition layering is fixed: `cmd/app → bootstrap → {platform, shared} ← module/*`, enforced by `TestCompositionBoundaries` in `internal/shared`.** Hand-written `cmd/app` assembly (`main.go`, `newApp`) imports `conf` + `bootstrap` + kernel/server types, never a business module — modules enter exclusively through wire provider sets (`wire.go`/`wire_gen.go`). `internal/bootstrap` is process-boot machinery below business — the `Lifecycle` registry and the idempotent seed job; it imports kernel/platform, never a module. `internal/shared` is a contracts-only kernel: module-facing bases (`UC`/`RP`, `Paginate`), cross-module neutral protocols, and bind-injected interfaces (`Transaction`, `HookRegistry` — both `wire.Bind`ed from their implementations); ent-free, no providers. Admission test for both packages: "do modules import this?" — no → bootstrap (or it does not exist).
- **A module's outward surface is its biz layer.** Bind-injected shared contracts (`Transaction`, `HookRegistry`) land in UC constructors only; the UC implements the shared-go transport interfaces (`Authenticator`/`Authorizer`/`audit.Sink`); phase registration happens in the UC too — a repo's own loops are exposed as port methods on the RP and registered by its UC. `data`/`client` never receive app-level injected contracts.
- **Lifecycle pattern: phase machinery registers; preconditions and degraded acquisition stay constructional.** Background loops and boot sweeps register on `shared.HookRegistry` in constructors and run in app phases; a dependency-graph precondition (authz's initial compile — every request reads the snapshot) stays in the constructor so failure fails construction; resource acquisition with designed degradation (audit's MQ subscribe — audit availability yields to service availability) stays constructional with a wire cleanup.
- Events + outbox are `MAY` until a real cross-aggregate side-effect appears. Do not scaffold speculatively; `shared-go/capability/mq` is there when needed.
- **Service code serializes through `utils`, never `encoding/json` directly** — `utils` is the single library-replacement seam (swap the JSON library in one place): `utils.Marshal` (error-returning) / `utils.MustMarshal` / `utils.Unmarshal[T]`; if the helper you need is missing, add it. Exemptions: `utils` itself, `jsoncodec`, generated code, `protojson`; capability packages are the other way round by the root self-containment rule (`AGENTS.md` §4). Numeric narrowing conversions go through `utils.ConvNum[R](x)` the same way — gosec G115-class findings are fixed at the seam, never silenced: a safety *rationale* comment may stay, `#nosec` directives may not.
- **Secrets pick their algorithm by recovery semantics.** `utils.Hash` (bcrypt, one-way) for credentials that only need verification; `utils.Seal` (AES-256-GCM over a SHA-256-derived master key, base64(nonce||ct)) for secrets the service must present again verbatim — provider API keys a harness will dial out with. Never hash a dialing secret (functional death: the plaintext is unreachable); never bother sealing a password (verification never needs it back). The master key comes from `conf.Crypto`, validated at construction — an empty key fails the boot, not the first write.

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
| provider→biz mapping fn | `map<ProviderDO>` | `mapDept` |
| Use case | `<Name>UC` + `New<Name>UC` | `UserUC` |
| data repo impl | `<do>RP` + `New<DO>RP` → `<DO>RP` | `userRP` |
| ent↔biz mapping fn | `map<DO>` | `mapUser` |
| service struct | `<Name>Service` | `UserService` |
| log tag | UC = `module/<d>`; repo = `module/<d>_rp`; service = `module/<d>_service` | `module/role`, `module/role_rp`, `module/role_service` |

- biz holds **one** port family — repo ports (`<X>RP`). A local repo (`UserRP`, backed by `data`) and a remote repo (`<Remote>RP`, backed by `client`) are the same shape: biz injects an RP and is indifferent to where the bytes come from. `data` and `client` are just two kinds of adapter.
- A remote RP is named for the **provider** (the provider *is* the "table"); its payloads use a **neutral ACL value type**, not a consumer DO, so several consumers share one RP.
- Two views, two names for one object: the **port** is a repo (`<Remote>RP`, biz's view — "a store I call"); the **adapter** implementing it is a client (`<remote>Client`, the client layer's view — "a thing that dials a remote").
- DO names carry no service prefix (`User`, not `SystemUser`) — the package gives context. Infra fields keep conventional short names (`log`); the lowercase-type rule is for domain deps only.

---

## 3. File structure & ordering

**One module per domain (aggregate root): `module/<d>/{biz,data,service,module}.go`.** Each file is one concern within the module: `biz.go` = DO + port + UC + aux (FSM, entity behavior); `data.go` = local repo; `service.go` = handlers; `module.go` = ProviderSet. `MUST NOT` split a concern across extra files (`<d>_uc.go` etc.). If `biz.go` becomes unwieldy, the **aggregate boundary** is wrong — split into two modules, don't fragment one. Cross-module infra (`UC`/`RP` base, `Transaction`) lives in `shared/kernel.go`, not in any module — the kernel is per-service (`RP` embeds the service-specific `*platform.Platform`), hence not in `shared-go`. Tenant resolution (`DefaultTenant`, `TenantFromCtx`) lives in `shared-go/kratos/security/tenant.go` (sibling of `SubjectFromCtx` — subject-derived policy, not mixin logic); the ent-dependent mixin shells stay in `local_mixins`. The ent tree (schema, local_mixins) MUST NOT import the service's `internal/shared` (shared → platform → ent/runtime → ent/schema closes a cycle); business code never assigns tenants (the TenantMixin hook back-fills on create).

**Section order** (each preceded by a divider, §4; omit empty):

| Layer | Sections |
|---|---|
| biz | `DO` → `Port` → `UC` → `Method` → `Private` |
| data | `Repo` → `Method` → `Private` |
| service | `Struct` → `Handler` → `Private` |
| client | `Adapter` → `Private` |

- biz `DO` = data shape only. `Private` = trailing catch-all (FSM, entity-behavior, helpers). Always last. Inside it: **private methods first, private functions after** — no sub-divider (the receiver in the signature is the marker); methods may touch ports, functions are stateless pure computation (see the private-func-vs-method paradigm).
- data `Repo` = the repo struct + its constructor only; `Method` = the port-method implementations (the section mirrors biz's UC `Method` — structural declaration and behavior are separated by a divider, never one section to the floor).
- UC methods default to **exported** — the `Method` section is the module's capability surface, keep it public even with a single in-module caller; unexported helpers belong in `Private` (with its own divider).
- client: `Adapter` = `<remote>Client` + `New<Remote>Client` + port-method impls; `Private` = `map<ProviderDO>`. `client.go` is the layer's infra file (shared `standardMiddleware` + `ProviderSet`) — no dividers.
- **Value types live in `DO`; `Port` holds interfaces only.** A port's payload vocabulary (`GrantView`, `RoleMember`, `ListIn/Out`) is module data shape — declare it in `DO` even when only the port mentions it, so one section owns all types.
- **Cross-section vocabulary (aux consts/types a layer file needs but no section owns)** — declared between the imports and the first divider; reads top-down. Declarations that ARE a section's content (DO types, the port interface, UC/repo/service structs) live inside their section. The `Private` section holds private methods then private functions only — never interleaved type/const declarations.
- **Aux/engine files** — a module MAY hold extra lowercase files for a distinct non-layer concern (e.g. authz's `casbin.go` engine compile, `builtin.go` baseline collection); no layer dividers inside. Internal order: constants → types → functions, each declared before its first use.
- **Transport-fixture modules** (e.g. `transfer`, the e2e regression base) are exempt from the four-file shape: `service.go` + `module.go` only, no aggregate, no biz/data. Recorded so nobody "completes" them later.

**Client files** — `client.go` (shared `standardMiddleware` + `ProviderSet`) plus `<remote>.go` per remote. One remote = one repo port = one adapter = one file. Adding an RPC adds a method, **never a new file**; only a new remote opens one.

**Method order (by frequency, aspect-blocked)** — uniform across proto RPCs, biz Port, biz UC methods, data repo methods, service handlers:
1. `Create` 2. `Update` (+variants) 3. `Delete` 4. `List` → `Get` → `GetByXxx`/`FindByXxx`/`ExistsXxx` 5. Other (`Sort`, …)

UC and repo CRUD methods take **bare verbs** (`Create`, `UpdateStatus`, `List`) — the package name carries the noun (Go anti-stutter; `userUC.Create`, never `userUC.CreateUser`). Module-specific capabilities keep descriptive names (`Login`, `ExplainOperation`).

Methods over the same attached object form one contiguous **aspect block** — a shared noun suffix (`…Grants`, `…Members`, `…UserRoles`, `…Bindings`, `…ByPrincipal`) reached by a CRUD-verb prefix (`Create/Update/Delete/Replace/Remove/Add/Set/Load/List/Get/Find/Exists/Count`). Blocks order: core aggregate methods (no suffix) first, then aspect blocks by noun alphabet, `NotifyChanged`-style infra hooks last; inside a block, frequency order applies. Diagnostics (`Explain*`, `Preview*`, `DryRun*`) take non-CRUD verbs and therefore land in Other, never inside an aspect block.

**ProviderSet** — layered order: `New<Name>UC` → `New<DO>RP` → `New<Name>Service` → `wire.Bind` entries last (assembly mirrors the layering; within a layer, alphabetical).

**Imports** — five groups, blank-line separated: (1) stdlib (2) third-party (3) `cyber-ecosystem/shared-go/...` (4) `cyber-ecosystem/gen/...` (5) `cyber-ecosystem/app/services/<svc>/internal/...`.

**Proto import alias** — MUST be `<scope>pb`, never bare `pb`, so multiple proto imports never collide: `commonpb` (`cyber.shared.common.v1`), `errorspb` (`cyber.shared.errors.v1`), `extv1` (`cyber.ext.v1` — version-scoped, not `<scope>pb`), and per-service `<service>pb` (e.g. a `foo` service → `foopb`).

**Struct fields** — embedded base first, then `log`, then dependencies. Base `shared.UC`/`shared.RP` use **exported** fields (`Log`/`Tm`/`Platform`) so they stay visible across packages when embedded; service structs keep the conventional unexported `log`.

---

## 3a. proto ↔ DO pointer paradigm

Align proto optional fields, DO fields, and the service/data mapping around pointers so the layers stay free of nil-check boilerplate:

- **proto** body fields `optional` (`*T`) + `buf.validate` controls required/format (see proto CONVENTIONS §3); **path-bound `{id}` plain `string`**.
- **DO**: optional / proto-bound fields are `*T` (align with proto); mandatory system fields (`ID`, `CreatedAt`, `UpdatedAt`, `TenantID`) stay non-pointer.
- **service**: proto `*T` → DO `*T` **passed straight through — no deref, no nil-check**.
- **data**: optional fields `SetNillableX(*T)` / `ClearX`; mandatory NOT-NULL fields `SetX(*p)` (validate guarantees `*p` non-nil).
- **ent schema**: constraints by need — mandatory `.NotEmpty()` (NOT NULL), optional `.Optional().Nillable()`. Do **not** force everything Nillable: DB integrity (NOT NULL on mandatory fields) outweighs saving one deref. Field semantics a DB reader cannot infer from the name go in `.Comment("...")` — English, and the schema's `entsql.WithComments(true)` flows them into DDL column comments — never in Go line comments on the field chain. Self-evident fields (name, email, remark…) get none.
- **biz**: nil-semantics (e.g. update-leave-unchanged) live in the UC.

---

## 4. Comments

- **Section dividers are the ONLY top-level markers.** `MUST NOT` add godoc-style declaration comments on symbols within a section (no `// <Type> describes ...` above a type).
- **Divider format** — `// <Name>` + `-` to ~110 chars (`biz`/`data`) or ~120 (`service`). Copy a neighbor's dash run.
- **Inline comments** — only non-obvious *why* (trade-off, gotcha, concurrency). `MUST NOT` restate the code.

---

## 5. Layer specifics

**biz** — domain errors return the bare factory from the proto error enums (`errorspb.ErrorXxx("")`; full construction rules in §6). Infra errors arrive from `data`/`client` already classified at the choke points (§6). Multi-op atomicity via `uc.Tm.InTx`. An aggregate FSM lives in `Private` (`looplab/fsm`; `TransitionTo(ctx, target)` returns a domain error on illegal transition).

**data** — repo `<do>RP` embeds `shared.RP{Log, Platform}`; on ent error `return ..., rp.Platform.HandleEntError(err)` (maps to `InfraError`). **No business rules.**

**schema / local_mixins** — declaration order of the behavioral mixins is load-bearing: `ID/CU → Tenant → Datascope → SoftDelete` — hooks and interceptors compose in declaration order, and SoftDeleteMixin's delete→update rewrite relies on DatascopeMixin guarding `OpUpdate`. Guarded by `TestMixinOrder` in the schema package (`./nx run <service>:test`). `DatascopeMixin` declares per-entity dimension fields (`UserField` for `self`, `DeptField` for `dept_tree`); an absent field means the kind is not expressible on that entity and its scopes pass through unfiltered. Narrowing comes only from the ctx `Decision` — paths without one (seed, login, engine compilation, builtin baseline) skip filtering, mirroring TenantMixin's no-subject skip. `local_mixins.Unscoped(ctx)` lifts narrowing for one enclosed read: the reference-plane shape, where authorization rides the referring row (a visible user's avatar → file id → key) rather than the target's own owner scope — reserved for that shape and review-guarded. Never hand-write scope filters in `data`; declare dimensions on the schema and let the mixin compile them.

**service** — `<Name>Service` embeds the proto `Unimplemented<X>Server`, implements `Registrar` (`RegisterGRPC/HTTP/Connect`). Thin: `in.GetXxx()` → UC → map to proto. No direct repo access.

**client** — outbound adapter for a remote repo, the counterpart of `data`. `<remote>Client` dials one connection (`grpc.NewClient` / `connect.DialInsecure`, with `standardMiddleware`), implements `<Remote>RP`. **`MUST NOT` let biz import the provider's proto** — only the client layer maps provider types → the port's neutral ACL type. Outbound middleware assembly lives in `client.go` per-service (not `shared-go`), mirroring the server: `shared-go` supplies components, the service owns the chain.

**Transports are symmetric to the app.** gRPC and Connect clients both surface errors as **kratos `*errors.Error`**, not transport-specific types — so biz/service/middleware handle one error shape regardless of transport. (gRPC does this natively; Connect normalizes at its boundary.) `web/JS` clients are out of scope here — they use `@connectrpc/connect` directly against the Connect *server*.

---

## 5a. Authz & audit engine assembly (IAM)

The IAM engine is a **mechanism every service replicates**, not a system-service dependency: interfaces + middleware live in `shared-go` (`kratos/security`, `kratos/audit`); the first implementation and the data ownership live in `system`. Three sibling interfaces assemble identically — `Authenticator` (`SessionAuth`), `Authorizer` (`OperationAuthz`), `audit.Sink` (`Audit`): the owning module's UC implements the shared-go interface, `wire.Bind` satisfies it, `server` injects it and never imports the module. Any "special structure" the authz side grows beyond the auth pattern should be challenged by this symmetry.

- **Middleware position is a contract.** ADMIN block order: `SessionAuth → Audit → OperationAuthz`. The kratos onion propagates ctx inward and bubbles errors outward — only this slot sees both the Subject attribution and the deny attribution. The deny dimension classifies by the error identity `OperationAuthz` already translated (`PERMISSION_DENIED` vs `AUTHZ_POLICY_DENIED`), never by reading the Decision. PUBLIC and 401 faces are unaudited (v1); streaming faces carry no audit middleware.
- **Collection semantics.** Allow face collects only `ACTION_WRITE`; deny face collects everything (reads included — probing is the audit value). Action ladder: `desc.action` (authoritative) → method-name vocabulary → `WRITE` floor (fail-safe over-collection: missed events are invisible, over-collection is visible and fixable; floor hits carry a counter).
- **ABAC plugins.** One file per kind (`policy_<kind>.go`) with `init()` self-registration; a duplicate kind panics at startup — the file list IS the registry, a new kind touches zero existing files. A proto oneof member MUST have a registered plugin: an unregistered member would degrade "unknown kind" from a direct-DB-write-only state to reachable-by-normal-API. The evaluation kernel (`survivingGrants` family) is frozen; `Evaluate` is a pure function (attrs in, verdict out — the engine feeds time, tests table-drive it); attribute sources are lazy (no linked policy → zero resolution; `Attrs()` union selects sources; at most one resolve per source per request); evaluation errors, unregistered kinds, and dirty params fail closed.
- **Module dependency rulings (IAM DAG).** `module/authz` reads ent tables directly and never imports `role`/`policy`; `role → policy` is a read-only port; `auth → user` and `auth → authz` are one-way; `role` never imports `user` (member faces show principal_id only). Cross-side protocols (policy version stamp, session epoch marker) live in `internal/shared` so writer and reader both reach the neutral layer without importing each other. An interface injected to invert a dependency that is semantically cyclic is the invisibility cloak of a cycle — banned even when the import graph stays acyclic.
- **Version-stamp discipline.** Every write that can change a decision (role/policy/grant/user-status mutations) notifies `NotifyChanged` after commit; replicas rebuild compiled snapshots on change (watch — every subscription-establishment point (first subscribe, reconnect) re-checks the version once, because pubsub has no backlog and a change published before the subscription existed (the seed racing startup) is otherwise lost — plus low-frequency reconcile; a single replica self-notifies on the same path — no code branch). There is no app-layer decision cache: the compiled in-memory snapshot IS the cache, and ABAC constraints evaluate live on every request.
- **Seeds are service-level.** The seed job lives in `internal/bootstrap` and mounts `kratos.BeforeStart` (explicitly after `lc.Start` — its completion publishes a fire-and-forget change notification, so the subscriptions it notifies must already exist) as an idempotent ensure (query → create; unique-index conflict → re-query), multi-replica safe without advisory locks. The superadmin triplet is conf-driven (`Authz{super_admins, initial_password}`) — under deny-by-default, a fresh environment could otherwise never create its first user.

### Non-system services (IAM consumers)

Every non-system service consumes IAM remotely and grows no identity of its own:

- **Identity travels by user credential, never by claim.** The `client` layer forwards the inbound session cookie upstream (`IntrospectService/VerifySession`) and the engine decides grants (`IntrospectService/CheckOperation`); system re-authenticates on its own chain. Identity always derives from the credential, never from a downstream-asserted parameter — that law is what keeps the introspection faces safe to expose at all. Decide is two-stage: own-proto builtin set first (`kauthz.BuiltinOperations("cyber/<svc>/v1/")` — local short-circuit, no remote hop), else remote `CheckGrants` through the `AuthzRP` port; the remote port fails closed (an engine error is a 5xx, never an allow).
- **The client middleware order is load-bearing**: `tracing → circuitbreaker → metadata → sanitize.Client → logging.Client`. The chain wraps the raw transport call; `logging.Client` (innermost) sees the provider's original error first (kratos-rendered, so an upstream connect error logs with its message even when the code renders as a bare 500); `sanitize.Client` re-mints it into this service's catalog (provider reason dropped, detail only in cause); `circuitbreaker` (outermost of the three) judges the translated code — semantic 401/403 count as success, only 500/503/504 accumulate. `ConnectToError` sits at the chain exit: an already-translated kratos error passes through untouched, anything raw gets a final conversion. Reorder any two and either the logs lose fidelity or the breaker arms on denied logins.
- **Degraded hydration is grant-aware and graded.** An ungranted viewer is a normal audience state, so it must never be *reached* through the denial channel: the adapter decides eligibility once per request via `CheckOperation` (payload decision — an allow-side read that leaves no audit trail; the operation is named by the generated `<Service><Method>Procedure` constant, never a hand-written string — a drifted literal reads as permanently ungranted and silently blanks hydration) and skips the reads entirely; an ungranted viewer reaching the read face would land one audit denial — probe signal — per id per page view, polluting exactly the signal audit exists to collect. When granted, hydration reuses the normal user-facing RPC with the caller's own credential (never a privileged batch-lookup face — that would be a directory-enumeration surface); per-id failures then grade debug (denied — an ABAC flip between check and read) vs warn (any other failure); the row always renders with absent fields.
- **Service credentials are a seam, not a feature.** Per-request identity rides the ctx-key → interceptor pattern (cookie header today; the same slot forwards Bearer when app clients land). A future service identity lands as a sibling interceptor on the same `WithInterceptors` list plus a credentials block in `conf.Remote` — do not pre-scaffold it; the two identities (on-behalf-of vs service-principal) never merge (see root §8 user-split law and the capability-token reservation in the platform vision).

---

## 6. Errors & observability

**Construction — one identity layer over classifications over one factual layer.**

- Identity only from catalog factories (`errorspb.ErrorXxx`); never mint kratos `errors.New`/`Newf` as identity.
- Factory message stays `""` (it serializes to the client verbatim; client copy = i18n keyed by reason). Non-empty message is a reviewed exception behind a named constant.
- Domain-rule violation → bare factory, no cause: the reason is the whole story.
- `WithCause` carries the classification, via one of two channels:
  - **factory**: fuzzy identity over a specific internal factory — `ErrorGeneralErrorUnauthenticated("").WithCause(ErrorGeneralErrorNotFound(""))`;
  - **authored prose**: `fmt.Errorf("user not found: %s", email)` — always `fmt.Errorf` (no `errors.New`), lowercase start, no trailing punctuation, `%w` to chain an underlying error. `SCREAMING_SNAKE` is reserved for reason alone; splits finer than reason go to structured fields/metrics labels, never string prefixes.
- Cause never crosses the wire (`ErrorToConnect`/`GRPCStatus` emit message/reason/metadata only) — the fuzzy-out/specific-collected split is safe by construction.
- Infra is classified at choke points, not call sites: `HandleEntError` (ent; classified kratos errors pass through), `HandleCacheError`/`HandleStorageError`/`HandleMQError` (capabilities), `MapUpstreamErr` (remotes, behind `sanitize.Client`); each attaches the original as cause. A UC never returns a raw upstream error.
- Identity checks on errors that passed a choke point use the **generated reason predicates**, never kratos code checks: `errorspb.IsInfraErrorDbNotFound(err)` (reason identity), not `errors.IsNotFound(err)` — the latter matches on HTTP code and only works while the mapped code happens to be 404.
- shared-go middleware sentinels declare internal-code reasons (`errors.Unauthorized("MISSING_SESSION", "")`) and are rebound onto catalog factories at service assembly (`server.go` init, original kept as cause): shared-go holds no generated-code knowledge, vocabulary drift = compile error at the rebind.

**Wire boundary.** Only predefined enums leave the service. `sanitize.Server()` (chain outermost) masks any non-kratos outbound error to a `GeneralError`; `sanitize.Client()`/`MapUpstreamErr` remap any provider error by HTTP code into this service's space — no upstream reason or transport detail crosses back.

**Observability — stock middleware composed, never forked.** Chain (outermost → innermost), identical across all three servers:

```
sanitize → tracing → metrics → logging(RenderErrors(logger)) → ErrorTelemetryServer → recovery → ratelimit → metadata → guard → validator
```

- The `guard` slot is layered: `DefaultGuard` (matched on `ACCESS_UNSPECIFIED` — no annotation fails closed) plus one selector layer per credential family, e.g. `krauth.SessionAuth(authn, krauth.SessionCookie(...))` matched on `ACCESS_ADMIN`. `authn` arrives as the `krauth.Authenticator` interface — the module's UC satisfies it via `wire.Bind`, so `server` never imports the module.
- `RenderErrors(logger)` wraps the logger handed to the stock logging middleware (an slog handler): the `error` attr becomes the factual root, `internal` and a non-empty `message` become their own attrs, the duplicated `stack` attr is dropped, 4xx failures are demoted to WARN.
- `ErrorTelemetryServer` sits immediately inside logging and inside sanitize — it observes the pre-mask error. Three duties, none of which the stock middleware provides: span attrs `error.reason`/`error.internal` (the aggregation keys), counter `kratos.server.errors{operation,reason,code}`, and an outbound re-rendering via `renderFlat` — `Error()` becomes one line `REASON [INTERNAL > PATH] (message): root` while `errors.As`/`Unwrap`/`FromError` keep resolving to the original chain, so the stock tracing middleware's Exception event and the wire translation are semantically unchanged and clean in rendering.
- The Exception event itself stays with the stock tracing middleware — recording a second one would split every failure into two error groups.
- Framework middleware errors are pre-mapped in `internal/server/server.go` `init()`: recovery → `GENERAL_ERROR_UNSPECIFIED`, ratelimit → `FLOW_ERROR_RATE_LIMITED`, validator → `GENERAL_ERROR_VALIDATION_FAILED`, missing access annotation → `GENERAL_ERROR_INTERNAL`.
- biz does not log errors — the observability outlet is complete by construction; `Log` on UC/RP carries flow info only. Defensible exception: **fail-closed swallow paths** — where an engine must keep serving past an internal failure (policy evaluation error, unregistered kind, dirty params, attribute resolution), the engine/RP layer logs ERROR/WARN at the swallow site; for a dropped decision input, that line is its only telemetry outlet.
- Aggregation keys: wire & counter = `reason`; spans = `error.reason` + `error.internal`; log attrs = `reason` + `internal`.

**Client side (when a remote RP client lands):** mirror the server — an `ErrorTelemetryClient` inside `sanitize.Client`, observing the pre-map upstream error (counter `kratos.client.errors`); `RenderErrors` already covers `logging.Client`.

**Anti-patterns:** don't `return nil, err` from a client with the provider's error untouched; don't expose provider `reason` strings; don't skip the logging middleware (you lose detail + trace correlation); don't author cause prose with a SCREAMING prefix or `errors.New` (prose = `fmt.Errorf`; reason owns SCREAMING); don't nest a second identity under a factory message; don't classify ent/capability errors at call sites instead of the choke points.

---

## 7. Checklist

### Pre-completion self-check (any Go change)

Before reporting done on any Go change under `app/services/*`, walk the diff against §2–§6 — compile-green is not convention-green:

- **§2 naming** — every new symbol matches the table: `DO` unprefixed, `<DO>RP`/`<Remote>RP` ports, `<Name>UC`, `<do>RP` repo, `map<DO>` mapping; CRUD UC/repo methods take bare verbs; aspect blocks ordered (core methods first, aspect blocks by noun alphabet, infra hooks last).
- **§3 structure** — four-file shape intact (no `<d>_uc.go` fragments; extra files only per the aux-file rule); section order + dividers per the table; imports in the five groups; proto import aliases `<scope>pb`, never bare `pb`; struct fields embedded base → `log` → dependencies.
- **§4 comments** — no declaration-level comments on symbols; dividers are the only top-level markers; inline comments state non-obvious *why* only. (Service-internal rule — `shared-go` keeps its godoc conventions.)
- **§5 layers** — business logic in biz; `data`/`client` adapters only; no business rules in data; no repo access in service; direction never reversed (biz imports neither `data`/`client` nor `service`/`server`/`platform`); cross-module access is a read via repo port.
- **§6 errors** — identity only from catalog factories (`errorspb.ErrorXxx("")`); cause via factory-chain or authored `fmt.Errorf` prose; infra classified at choke points, never at call sites; a UC never returns a raw upstream error; biz does not log errors.

Done = every bullet answered for every hunk in the diff. For `shared-go` changes only the §6-errors and dependency-direction bullets apply, plus the capability self-containment rule (`AGENTS.md` §4).

**Adding a new domain module (aggregate root):**
1. `module/<d>/biz.go`: `DO` → `Port` → `UC` → `Method` (→ `Private`); UC embeds `shared.UC` (`uc.Tm` / `uc.Log`).
2. ent `schema/<do>.go`: fields + mixins; `./nx run <service>:generate:ent`, `:migrate:diff`, `:migrate:apply` (persisted only).
3. `module/<d>/data.go`: `<do>RP` + `New<DO>RP` + `map<DO>` (embed `shared.RP`, use `rp.Platform`). **Omit** this file for a pure ACL facade (no persistence).
4. `module/<d>/service.go`: `<Name>Service` + handlers + Registrar. `module/<d>/module.go`: `ProviderSet{New<Name>UC, New<DO>RP, New<Name>Service}`.
5. Wire it in: add the module's `ProviderSet` to `wireApp`; add its `*<Name>Service` param to `server.NewRegistrarList`.
6. `./nx run <service>:generate:wire` then `:build`. If it imports another module, keep it a **read** (port query) — cross-module writes go via domain event. Verify dependency direction (UC→port; UC→UC only toward foundation modules — user/authz/file). **Wire prunes unconsumed providers**: a startup-shaped component (seed job, engine watch loop) must hang off a live dependency chain — a `newApp` param or self-registration on `shared.HookRegistry` in a UC constructor — or it silently never instantiates.

**Adding a remote-service dependency (treat as a repo):**
1. biz: declare `<Remote>RP` with neutral ACL payloads; inject into the UC like any repo.
2. client `<remote>.go`: `<remote>Client` + `New<Remote>Client` + `map<ProviderDO>` over one connection. Register `New<Remote>Client` in `client.ProviderSet`. One file per remote.
3. conf `conf.proto` `Remote`: add endpoint field; `configs/config.yaml`: set it. `./nx run <service>:proto:conf`.
4. `./nx run <service>:generate:wire` then `:build`. Verify biz never imports the provider's proto.

**Pure ACL facade** (no own persistence): the module has **no `data.go`**; its UC depends on a `<Remote>RP` satisfied by the `client` layer. Its `module.go` `ProviderSet` lists only `New<Name>UC` + `New<Name>Service`.

**Deriving a new service** (from `system`, the mandatory base and reference skeleton):

1. **Copy-set**: `project.json` / `Dockerfile` / both `config.yamls` / `conf.proto`(+buf.gen.conf) / `cmd/app` / `bootstrap` (lifecycle only, no seed) / `shared` (kernel trimmed + `dependency_guard_test` rewritten) / `platform` (verbatim) / `server` (five files) / `client` (once the service calls out) / `ent` schema + `local_mixins`. Deliberately NOT copied: `seed.go`, system business modules, the authz engine internals (casbin/attrs/policies), audit.
2. **Dials**: next service takes the next port triple (+1000 on http/grpc/connect); redis db index isolated per service (both configs carry the why-comment).
3. **ent first generation is a two-pass stub dance**: `local_mixins` import the generated tree (chicken-and-egg), so stub the mixins (nil `Hooks`/`Interceptors` over entgo-level types) → generate → restore the real mixins → regenerate. After removing an entity, nuke the whole generated tree first — the loader reads the stale tree.
4. **wire cold start**: the `go:generate` directive lives in `wire_gen.go` itself — a fresh service runs `go run github.com/google/wire/cmd/wire` manually; wire rejects interface bindings with no consumer in the graph (comment the bind out until its consumer exists).
5. **First migration**: `NAME=init <svc>:migrate:diff` → hand-prepend `CREATE EXTENSION` lines the atlas scratch already has (a service database owns its extensions) → `:migrate:hash` → `:migrate:apply`. Recreate the service database rather than fighting a drifted init.
6. **IAM**: consume per §5a "Non-system services"; blank-import the service's gen in `shared-go/kratos/security/access_annotation_test.go` and point `BuiltinOperations` at the new prefix.
7. **Catalog**: nothing to wire — the grant tree picks the new service up when `proto:generate:catalog` emits its manifest; system live-loads it (no restart, no import).
8. **Deploy plane**: compose migrate job + service block (network alias `<svc>.api`, healthz, no host ports), traefik package-prefix routers on the admin/api hosts, `02-databases.sql` shell (existing volumes need a one-time manual `CREATE DATABASE`), `deploy/project.json` target lists, vite proxy key + dev env entry.
