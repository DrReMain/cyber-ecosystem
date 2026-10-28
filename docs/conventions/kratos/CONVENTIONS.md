# Kratos Service Conventions

**Scope:** Rules for kratos-based Go services in this repo (`app/services/*`). Prescriptive — `MUST` / `SHOULD` / `MAY`.

This document states **constraints, rationale, and anti-patterns** — the things code alone can't enforce or explain. It deliberately does not restate what the code already shows; look at any service under `app/services/*` for the shape. Rules here win over precedent in code. For proto contracts see `docs/conventions/proto/CONVENTIONS.md`; procedures (new module, remote dependency, service derivation, operating facts) live in `RUNBOOKS.md` beside this file. Every change closes with its chain green and its behavior probed live (§7); the human reviewer carries what no machine checks.

## Read map

| About to touch | Read first |
|---|---|
| `module/<d>/biz.go` | §1, §2, §3 biz row, §3a, §4 |
| `module/<d>/data.go` | §2 repo rows, §3 data row, §5 data |
| `module/<d>/service.go` | §2, §3 service row, §5 service |
| `client/<remote>.go` | §2 client rows, §5 client, §5a |
| `internal/{bootstrap,shared,server}/**` | §1 composition bullets, §6 |
| `ent/schema/**`, `local_mixins/**` | §3a, §5 schema block |
| `shared-go/**` | §1 capability law, §6 |
| any other `.go` under `app/services/*` | §4, §6, the touched layer's §5 block |

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
platform         capability facade (the service's assembled module set — §5)
ent              generated persistence (shared by all modules in the service)
conf             config (proto-defined, generated)
```

- `MUST` keep business logic in `biz`; `service` is thin; `data`/`client` are adapters only.
- `MUST NOT` let `biz` import `data`/`client`/`service`/`server`/`platform`. Direction is `data → biz` and `client → biz` (both import biz to satisfy ports), never reverse. biz depends on its ports + `gen/` + `shared-go/`, and MAY take `*conf.X` in its constructor (same pattern as `server`/`platform`).
- **A UC depends on repo ports, never on another UC — injection is repo-port-only.** Cross-module *read* → the other module's repo port (read-only, no side-effect). Cross-module *write* has exactly two lawful shapes: **(1) a domain event** — the writer publishes the fact (the event type lives in the service's `internal/shared` as a neutral cross-module protocol; the in-process registry sits beside the lifecycle machinery, synchronous and best-effort — the publisher's own transaction is the truth), and the owning module subscribes and writes; **(2) an in-transaction ledger grant** — when a settled ruling demands same-database atomicity, the consumer may call the owning module's port *inside its own transaction*, provided the write is idempotent insert-only (conflict → no-op), because the transaction, not an event, is the consistency boundary — one sanctioned move on an existing row: a lapsed single-row grant may be re-granted forward-only (a conditional update that lengthens, never shortens), so an expired row cannot strand its owner. Everything else — a direct call outside a transaction, one UC invoking another — stays forbidden. A cross-module business-rule need (not a query) is a design smell — raise it, don't wire UC→UC.
- The four files are a required-by-role set, not an exclusive list: `data.go` may be
  omitted (pure ACL facade), `service.go` may be omitted (module exposes no RPCs), and a
  module MAY add one extra file per cohesive non-layer concern when its content belongs
  to neither biz orchestration nor data lifecycle (e.g. authz's `casbin.go`,
  the engine compile).
- Two aggregates constantly reaching into each other are probably **one** aggregate — merge or redraw the boundary.
- **One table, two modules, split by lifecycle operation.** A channel module (e.g. `fileproxy`, the upload proxy) and a management module (e.g. `file`) may share one table when the channel is a different concern, not a different aggregate: every write has exactly one owner — create in the channel module's own data layer, update/delete in the management module's. The channel port is named for the channel (`FileProxyRP`, not `<DO>RP`); the channel module imports the owning module for DO/constants/mapper types only; binding a stored reference (e.g. a user's avatar file id) validates through the owning module's port.
- **Entity relations follow aggregate boundaries.** *Intra-aggregate* composition (same root, same lifecycle, never a separate service — e.g. Order→OrderLine) uses an **ent edge**; *inter-aggregate* references (independent aggregates, any cross-service possibility — e.g. User→Dept) use a **plain `<x>_id` field** resolved via a repo port. Edges lock a relation to one DB (no cross-service or saga span) and clash with soft-delete cascade; plain IDs migrate to a remote RP with zero refactor. Default to a plain FK unless the relation is definitively intra-aggregate-and-intra-service-forever.
- **Module dependency graph MUST be acyclic (DAG) and sparse.** Each module is its own package, so Go catches cycles at compile time — but Go blocks cycles, not meshes: every cross-module import is a review point. Default to no cross-module dependency; add one only for a real read need (consumer → provider via repo port). UC-level composition is legal only toward foundation modules (`user`, `authz`, `file` — base-service capabilities).
- **Operator faces are strippable modules.** A service's operator face (list-everyone views that may not ship in every deployment) is its own module and own proto file, depending on the business module (reads its UC + ports); the business module MUST NOT import it back (e.g. `module/agentconfig` never imports `module/agentconfigadmin`) — deletion is then a closed operation on the graph: remove the proto file + module + one wire/registrar line.
- **Composition layering is fixed: `cmd/app → bootstrap → {platform, shared} ← module/*`.** Hand-written `cmd/app` assembly (`main.go`, `newApp`) imports `conf` + `bootstrap` + kernel/server types, never a business module — modules enter exclusively through wire provider sets (`wire.go`/`wire_gen.go`). `internal/bootstrap` is process-boot machinery below business — the `Lifecycle` registry and the idempotent seed job; it imports kernel/platform, never a module. `internal/shared` is a contracts-only kernel: module-facing bases (`UC`/`RP`, `Paginate`), cross-module neutral protocols, and bind-injected interfaces (`Transaction`, `HookRegistry` — both `wire.Bind`ed from their implementations); ent-free, no providers. Admission test for both packages: "do modules import this?" — no → bootstrap (or it does not exist).
- **A module's outward surface is its biz layer.** Bind-injected shared contracts (`Transaction`, `HookRegistry`) land in UC constructors only; the UC implements the shared-go transport interfaces (`Authenticator`/`Authorizer`/`audit.Sink`); phase registration happens in the UC too — a repo's own loops are exposed as port methods on the RP and registered by its UC. `data`/`client` never receive app-level injected contracts.
- **Lifecycle pattern: phase machinery registers; preconditions and degraded acquisition stay constructional.** Background loops and boot sweeps register on `shared.HookRegistry` in constructors and run in app phases; a dependency-graph precondition (authz's initial compile — every request reads the snapshot) stays in the constructor so failure fails construction; resource acquisition with designed degradation (audit's MQ subscribe — audit availability yields to service availability) stays constructional with a wire cleanup.
- Events + outbox are `MAY` until a real cross-aggregate side-effect appears. Do not scaffold speculatively; `shared-go/capability/mq` is there when needed.
- **Service code serializes through `utils`, never `encoding/json` directly** — `utils` is the single library-replacement seam (swap the JSON library in one place): `utils.Marshal` (error-returning) / `utils.MustMarshal` / `utils.Unmarshal[T]`; if the helper you need is missing, add it. Exemptions: `utils` itself, `jsoncodec`, generated code, `protojson`; capability packages are the other way round by the self-containment law below. Numeric narrowing conversions go through `utils.ConvNum[R](x)` the same way — gosec G115-class findings are fixed at the seam, never silenced: a safety *rationale* comment may stay, `#nosec` directives may not — with one exception class: a protocol-mandated primitive (the weak algorithm is the counterparty's wire contract, not a security choice) carries `#nosec <rule>` inline with the protocol named, its implementation pinned by a golden-vector test.
- **Secrets pick their algorithm by recovery semantics.** `utils.Hash` (argon2id, OWASP baseline, one-way; legacy bcrypt digests verify but never re-write — login rehashes them lazily) for credentials that only need verification; `utils.Seal` (AES-256-GCM over a SHA-256-derived master key, base64(nonce||ct)) for secrets the service must present again verbatim — provider API keys a harness will dial out with. Never hash a dialing secret (functional death: the plaintext is unreachable); never bother sealing a password (verification never needs it back). The master key comes from `conf.Crypto`, validated at construction — an empty key fails the boot, not the first write.
- **Capability packages are self-contained units.** A package under `shared-go/capability` depends only on stdlib, its third-party providers, and its own capability family (the parent interface package — `mq/pg` → `mq`); never on `utils`, `kratos`, `orm`, or any other shared-go package — a family copies out as a unit. In-package substitutes: `encoding/json` directly; explicit bounds guards (clamp above `math.MaxInt32`, then convert) instead of `ConvNum`.

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
| log tag | UC = `module/<d>`; repo = `module/<d>_rp`; service = `module/<d>_service` — the module is the unit, never the instance: a module holding several UCs or RPs still uses exactly these three forms, distinguishing instances with log fields, not tag suffixes | `module/role`, `module/role_rp`, `module/role_service` |

- biz holds **one** port family — repo ports (`<X>RP`). A local repo (`UserRP`, backed by `data`), a remote repo (`<Remote>RP`, backed by `client`), and a provider repo (backed by a `shared-go` capability — e.g. a payment-channel port) are the same shape: biz injects an RP and is indifferent to where the bytes come from. The adapter set is open — a port abstracts one external system; the list of adapter kinds is not itself law.
- A remote RP is named for the **provider** (the provider *is* the "table"); its payloads use a **neutral ACL value type**, not a consumer DO, so several consumers share one RP.
- Two views, two names for one object: the **port** is a repo (`<Remote>RP`, biz's view — "a store I call"); the **adapter** implementing it is a client (`<remote>Client`, the client layer's view — "a thing that dials a remote").
- DO names carry no service prefix (`User`, not `SystemUser`) — the package gives context. Infra fields keep conventional short names (`log`); the lowercase-type rule is for domain deps only.
- A module package exports **types, `New*` constructors, and methods only** — never a package-level exported function. A helper needed by several modules of one service lives in `internal/shared` (single source — scattered magic values drift); a module-internal helper stays private.

---

## 3. File structure & ordering

**One module per domain (aggregate root): `module/<d>/{biz,data,service,module}.go`.** Each file is one concern within the module: `biz.go` = DO + port + UC + aux (FSM, entity behavior); `data.go` = local repo; `service.go` = handlers; `module.go` = ProviderSet. `MUST NOT` split one concern across extra files (`<d>_uc.go` etc.) — the ban's unit is the concern, not the file: a distinct concern takes its own aux file (below), mechanical fragments of one concern stay banned. If `biz.go` becomes unwieldy with no distinct concern left to carve out, the **aggregate boundary** is wrong — split into two modules, don't fragment one. **Multiple UCs/RPs in one file** — `biz.go`/`data.go` are each layer's home, not one type's: a module MAY hold several UCs (several RPs) in its single biz/data file when they serve the same aggregate (a member-facing UC beside an operator UC, an export UC beside the list UC); UCs that share nothing are the wrong-aggregate signal above — split the module, not the file. Cross-module infra (`UC`/`RP` base, `Transaction`) lives in `shared/kernel.go`, not in any module — the kernel is per-service (`RP` embeds the service-specific `*platform.Platform`), hence not in `shared-go`. Tenant resolution (`DefaultTenant`, `TenantFromCtx`) lives in `shared-go/kratos/security/tenant.go` (sibling of `SubjectFromCtx` — subject-derived policy, not mixin logic); the ent-dependent mixin shells stay in `local_mixins`. The ent tree (schema, local_mixins) MUST NOT import the service's `internal/shared` (shared → platform → ent/runtime → ent/schema closes a cycle); business code never assigns tenants (the TenantMixin hook back-fills on create).

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
- **Companion files** — a module MAY hold extra lowercase files beside the four, in two kinds. **Domain files** carry self-sufficient pure domain logic: zero ent imports and no UC/RP receivers — own types with their own methods are the norm (game engines, ABAC plugins and their attribute sources, provider adapters, the casbin compile, guarded-HTTP transports); one file = one concept (one game, one plugin, one provider), the carve-out unit being the concept, not biz.go's size. **Layer companions** carry one theme of a layer's concern that outgrew the layer file (authz's `scopes.go` holds RP-side dept hydration): the layer's receiver and that layer's discipline apply. No layer dividers inside either; internal order constants → types → functions, each declared before its first use.
- **Transport-fixture modules** (e.g. `transfer`) are exempt from the four-file shape: `service.go` + `module.go` only, no aggregate, no biz/data. Recorded so nobody "completes" them later.
- **Face modules** — when the contract splits a second service for a different access audience over one aggregate (a member face plus an operator face), the second module is a face: `service.go` + `module.go` only, injecting the sibling module's UCs — thin proto↔domain mapping, zero business rules (the called UC owns those). Unlike a transport fixture it is a real RPC face, but it owns no aggregate and never grows biz/data; nothing imports a face (the module DAG forbids it), so deleting the face plane stays a closed operation.

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
- **ent schema**: constraints by need — mandatory `.NotEmpty()` (NOT NULL), optional `.Optional().Nillable()`. Do **not** force everything Nillable: DB integrity (NOT NULL on mandatory fields) outweighs saving one deref. Field semantics a DB reader cannot infer from the name go in `.Comment("...")` — English, and the schema's `entsql.WithComments(true)` flows them into DDL column comments — never in Go line comments on the field chain. Self-evident fields (name, account, remark…) get none.
- **biz**: nil-semantics (e.g. update-leave-unchanged) live in the UC.

---

## 4. Comments

Two comment kinds are legal under `app/services/*`; a comment that is neither is a violation.

| Kind | Shape | Where |
|---|---|---|
| Divider | `// <Name>` + `-` run (~110 chars `biz`/`data`, ~120 `service`; equal ±5 to a same-file neighbor's run) | section boundaries only (§3 table) |
| Inline | states non-obvious *why* only (trade-off, gotcha, concurrency); `MUST NOT` restate the code | inside a function, beside the line it explains |

**Review signature — declaration comment**: a `//` line whose next non-blank line opens a declaration (`type` / `func` / `const` / `var`) and whose text names or describes that symbol. **Review signature — file-header comment**: a `//` line after the package clause and before the first import or declaration — neither legal kind lives in that position, and the shape otherwise escapes the declaration signature (the next non-blank line opens `import`). Both are what a reviewer greps added lines for; §7 carries them.

Carve-outs: ent schema field semantics go in `.Comment("…")` (flows into the DDL), never Go line comments (§3a); `shared-go/` keeps godoc conventions — this section binds `app/services/*` only.

---

## 5. Layer specifics

**platform**

- Assembles **exactly the capabilities its modules consume** — the module set is the capability matrix, stated once per service. An unneeded capability is not assembled: no file, no `ProviderSet` entry, no getter, no conf field.
- A module calling a missing capability fails the **build** (never a runtime nil); an assembled capability without its conf block fails the **boot** with a clean wiring error. Deriving a service trims the platform package to the module set; adding a capability restores that same set.
- The storage object plane has no context-free getter: `GetStorageFor(ctx)` resolves the tenancy domain (`security.WithTenant` over the subject, the tenant mixin's precedence) — platform tenant → configured default bucket; every other registry domain → default name + domain suffix, provisioned lazily once per process.
- Rows and buckets resolve one axis; a domain pass never crosses tenants.

**biz**

- Domain errors: bare factory from the proto error enums (`errorspb.ErrorXxx("")`; full construction rules §6).
- Infra errors arrive from `data`/`client` already classified at the choke points (§6).
- Multi-op atomicity: `uc.Tm.InTx`.
- Aggregate FSM lives in `Private` (`looplab/fsm`; `TransitionTo(ctx, target)` → domain error on illegal transition).

**conf & env keys**

- A `${VAR:default}` placeholder references a config **key**; env enters the key space only through `env.NewSource("APP_")` (registered after the file source in every `main.go`) — prefix stripped, remainder becomes the key.
- A bare exported `VAR` never resolves — the placeholder silently keeps its default; an endpoint override then dials the production default with no error anywhere. Harnesses/deployments export `APP_VAR` (agent master key and wechat endpoint override are the standing examples).

**data**

- `<do>RP` embeds `shared.RP{Log, Platform}`; ent error → `return ..., rp.Platform.HandleEntError(err)` (maps to `InfraError`).
- **No business rules.**

**schema / local_mixins**

- Behavioral mixin declaration order is load-bearing: `ID/CU → Tenant → Datascope → SoftDelete` (SoftDeleteMixin's delete→update rewrite relies on DatascopeMixin guarding `OpUpdate`).
- `DatascopeMixin` declares per-entity dimension fields (`UserField` for `self`, `DeptField` for `dept_tree`); an absent field = kind not expressible → compiles to the empty set. Narrowing never degrades to unfiltered.
- Narrowing comes only from the ctx `Decision`; paths without one (seed, login, engine compilation) skip filtering, mirroring TenantMixin's no-subject skip.
- `local_mixins.Unscoped(ctx)` lifts narrowing for one enclosed read — reference-plane shape only (authorization rides the referring row: a visible user's avatar → file id → key), review-guarded.
- Never hand-write scope filters in `data`; declare dimensions on the schema, let the mixin compile them.

**service**

- Embeds `Unimplemented<X>Server`, implements `Registrar` (`RegisterGRPC/HTTP/Connect`).
- Thin: `in.GetXxx()` → UC → map to proto. No direct repo access.
- Mapping helpers come from `shared-go/utils` first — `Wrap`/`Unwrap`/`Ptr`/`Deref`/`PtrApply`/`ConvPtr`/`ConvNum`, the `*W` wrapper constructors, `ToTimestamp`/`FromTimestamp`, `StringWEmptyNil`, `SliceMap`. A hand-rolled nil-guard wrapper around a utils call is a duplicate — call utils directly; a missing generic primitive gets added to utils, never re-implemented per service.

**client**

- `<remote>Client` dials one connection (`grpc.NewClient` / `connect.DialInsecure`, `standardMiddleware`) and implements `<Remote>RP` — the outbound counterpart of `data`.
- **`MUST NOT` let biz import the provider's proto**; only the client layer maps provider types → the port's neutral ACL type.
- Outbound middleware assembly lives in per-service `client.go`, not `shared-go` — components from `shared-go`, chain owned by the service (session forwarding is a component: `transport/connect`'s `WithSessionToken` + `SessionCookieInterceptor`).
- Before deleting an apparently-unreferenced RPC, message, or error value, replay the remote-consumer scenario — a sibling service may reach it through a `<Remote>RP` adapter invisible to this service's imports.

**LLM provider seam (eino)**

- Model access via `eino-ext/components/model/openai`, constructed **per request** in the owning UC — base_url comes from per-user sealed config; a pooled ChatModel would cross users.
- `ChatModelConfig.BaseURL` must end in `/v1` (the component appends `/chat/completions` directly).
- `schema.Message` is dual-channel: `Content` and `ReasoningContent` both stream; a reasoning-only tick is an empty-content delta, not a skipped frame.
- The decrypted api key exists only inside the eino config, never in a log line.
- Model listing is outside eino: plain `GET <baseURL>/models` with an explicit timeout.
- **Outbound dials toward user-configured endpoints carry the dial-time guard**: reserved address classes only — loopback, link-local, multicast, unspecified — checked on the address actually connected (DNS rebinding is structurally irrelevant, IP-literal targets need no DNS). Deployment topology never enters the code: which LAN, which IP, DNS or not is the deploy plane's knowledge, and private unicast is a first-class target (a fully-private deployment is a declared scenario). Redirects are refused and response bodies are capped; one guard instance injects everywhere the SDK dials (eino's `HTTPClient` seam included).

**Transports**

- gRPC and Connect clients surface errors as **kratos `*errors.Error`** — one error shape in biz/service/middleware regardless of transport (gRPC natively; Connect normalizes at its boundary).
- `web/JS` clients are out of scope: direct `@connectrpc/connect` against the Connect *server*.

---

## 5a. Authz & audit engine assembly (IAM)

The IAM engine is a **mechanism every service replicates**, not a system-service dependency: interfaces + middleware live in `shared-go` (`kratos/security`, `kratos/audit`); the first implementation and the data ownership live in `system`. Three sibling interfaces assemble identically — `Authenticator` (`SessionAuth`), `Authorizer` (`OperationAuthz`), `audit.Sink` (`Audit`): the owning module's UC implements the shared-go interface, `wire.Bind` satisfies it, `server` injects it and never imports the module. Any "special structure" the authz side grows beyond the auth pattern should be challenged by this symmetry.

- **Middleware position is a contract.**
  - ADMIN block order: `SessionAuth → Audit → OperationAuthz` — only this slot sees both the Subject attribution and the deny attribution (ctx propagates inward, errors bubble outward).
  - The deny dimension classifies by the error identity `OperationAuthz` already translated (`PERMISSION_DENIED` vs `AUTHZ_POLICY_DENIED`), never by reading the Decision.
  - PUBLIC and 401 faces are unaudited (v1); streaming faces carry no audit middleware.
- **Collection semantics.**
  - Allow face collects only `ACTION_WRITE`; deny face collects everything (reads included — probing is the audit value).
  - Action ladder: `desc.action` (authoritative) → method-name vocabulary → `WRITE` floor. Fail-safe over-collection: missed events are invisible, over-collection is visible and fixable; floor hits carry a counter.
- **ABAC plugins.**
  - One file per kind (`policy_<kind>.go`) with `init()` self-registration; a duplicate kind panics at startup — the file list IS the registry; a new kind touches zero existing files.
  - A proto oneof member MUST have a registered plugin — an unregistered member would degrade "unknown kind" from a direct-DB-write-only state to reachable-by-normal-API.
  - The evaluation kernel (`survivingGrants` family) is frozen.
  - `Evaluate` is a pure function (attrs in, verdict out; the engine feeds time, the shared-go suite table-drives it).
  - Attribute sources are lazy: no linked policy → zero resolution; `Attrs()` union selects sources; at most one resolve per source per request.
  - Evaluation errors, unregistered kinds, and dirty params fail closed.
- **Module dependency rulings (IAM DAG).**
  - `module/authz` reads ent tables directly, never imports `role`/`policy`.
  - `role → policy` is a read-only port; `auth → user` and `auth → authz` are one-way.
  - `role` never imports `user` (member faces show principal_id only).
  - Cross-side protocols (the policy-change wake channel, the session epoch marker) live in `internal/shared` — writer and reader both reach the neutral layer without importing each other. The version truth is not here: it is the `authz_stamps` row (below).
  - An interface injected to invert a semantically-cyclic dependency is the invisibility cloak of a cycle — banned even when the import graph stays acyclic.
- **Version-stamp discipline.**
  - The version truth is a single `authz_stamps` row, bumped inside the change transaction by `AuthzStampMixin` on the tables feeding the snapshot — a bump that is a property of the tables cannot be forgotten by a writer (seed, admin paths, future code all covered). User rows bump only when a mutation assigns `enabled` or sets the soft-delete marker; profile edits never move the stamp (a password change must not recompile every replica).
  - Writers wake after commit (`NotifyPolicyChanged` → pubsub publish only): a failed publish costs timeliness, never correctness.
  - Replicas rebuild compiled snapshots on watch — and every subscription-establishment point (first subscribe, reconnect) re-checks the stamp once: pubsub has no backlog, so a change published before the subscription existed is otherwise lost. Plus a low-frequency reconcile backstop.
  - Reconcile compares `!=` and rebuilds on any divergence: the stamp is monotonic by construction, but a rewind or a restored backup must heal, never pin a stale replica; a forced rebuild every 24h bounds any unknown staleness class.
  - No app-layer decision cache: the compiled in-memory snapshot IS the cache; ABAC constraints evaluate live on every request.
- **Seeds are service-level.**
  - The seed job lives in `internal/bootstrap`, mounts `kratos.BeforeStart` explicitly after `lc.Start` (its completion publishes a fire-and-forget change notification, so the subscriptions it notifies must already exist).
  - Idempotent ensure (query → create; unique-index conflict → re-query), multi-replica safe without advisory locks.
  - The superadmin triplet is conf-driven (`Seed{super_admins, initial_password, populations}`) — under deny-by-default, a fresh environment could otherwise never create its first user.

### Non-system services (IAM consumers)

Every non-system service consumes IAM remotely and grows no identity of its own:

- **Identity travels by user credential, never by claim.** The `client` layer forwards the inbound session cookie upstream (`IntrospectService/VerifySession`) and the engine decides grants (`IntrospectService/CheckOperation`); system re-authenticates on its own chain. Identity always derives from the credential, never from a downstream-asserted parameter — that law is what keeps the introspection faces safe to expose at all. Decide is a single remote `CheckGrants` through the `AuthzRP` port; the remote port fails closed (an engine error is a 5xx, never an allow).
- **Member-plane identities are read through the population composition face.** The seat-oriented identity faces (`GetUser`/`ListUsers`) exclude member-plane seats by design — a member-view composition (roster, member list/detail) reads identities only via `ListUsersByPopulation` with the caller's forwarded credential, and derives its not-found errors from the local result.
- **The client middleware order is load-bearing**: `tracing → circuitbreaker → metadata → sanitize.Client → logging.Client`. The chain wraps the raw transport call; `logging.Client` (innermost) sees the provider's original error first (kratos-rendered, so an upstream connect error logs with its message even when the code renders as a bare 500); `sanitize.Client` re-mints it into this service's catalog (provider reason dropped, detail only in cause); `circuitbreaker` (outermost of the three) judges the translated code — semantic 401/403 count as success, only 500/503/504 accumulate. `ConnectToError` sits at the chain exit: an already-translated kratos error passes through untouched, anything raw gets a final conversion. A call site therefore never predicates on a provider-specific upstream reason — the re-mint drops it before any predicate sees it (`ConnectToError` cannot recover it); shared-classification reasons (401/403/404…) re-mint to the same shared name and stay predicate-safe. Derive domain errors from local facts. Reorder any two and either the logs lose fidelity or the breaker arms on denied logins.
- **Degraded hydration is grant-aware and graded.** An ungranted viewer is a normal audience state, so it must never be *reached* through the denial channel: the adapter decides eligibility once per request via `CheckOperation` (payload decision — an allow-side read that leaves no audit trail; the operation is named by the generated `<Service><Method>Procedure` constant, never a hand-written string — a drifted literal reads as permanently ungranted and silently blanks hydration) and skips the reads entirely; an ungranted viewer reaching the read face would land one audit denial — probe signal — per id per page view, polluting exactly the signal audit exists to collect. When granted, hydration reuses the normal user-facing RPC with the caller's own credential (never a privileged batch-lookup face — that would be a directory-enumeration surface); per-id failures then grade debug (denied — an ABAC flip between check and read) vs warn (any other failure); the row always renders with absent fields.
- **Service credentials are a seam, not a feature.** Per-request identity rides the ctx-key → interceptor pattern (cookie header today; the same slot forwards Bearer when app clients land). A future service identity lands as a sibling interceptor on the same `WithInterceptors` list plus a credentials block in `conf.Remote` — do not pre-scaffold it; the two identities (on-behalf-of vs service-principal) never merge.

---

## 6. Errors & observability

**Construction — one identity layer over classifications over one factual layer.**

- Identity only from catalog factories (`errorspb.ErrorXxx`); never mint kratos `errors.New`/`Newf` as identity.
- Factory message stays `""` (it serializes to the client verbatim; client copy = i18n keyed by reason). Non-empty message is a reviewed exception behind a named constant.
- Domain-rule violation → bare factory, no cause: the reason is the whole story.
- `WithCause` carries the classification, via one of two channels:
  - **factory**: fuzzy identity over a specific internal factory — `ErrorGeneralErrorUnauthenticated("").WithCause(ErrorGeneralErrorNotFound(""))`;
  - **authored prose**: `fmt.Errorf("user not found: %s", account)` — always `fmt.Errorf` (no `errors.New`), lowercase start, no trailing punctuation, `%w` to chain an underlying error. `SCREAMING_SNAKE` is reserved for reason alone; splits finer than reason go to structured fields/metrics labels, never string prefixes.
- Cause never crosses the wire (`ErrorToConnect`/`GRPCStatus` emit message/reason/metadata only) — the fuzzy-out/specific-collected split is safe by construction.
- Infra is classified at choke points, not call sites: `HandleEntError` (ent; classified kratos errors pass through), `HandleCacheError`/`HandleStorageError`/`HandleMQError` (capabilities), `MapUpstreamErr` (remotes, behind `sanitize.Client`); each attaches the original as cause. A UC never returns a raw upstream error.
- **Postgres transaction poisoning (25P02).** A failed statement poisons every later statement in its transaction (`current transaction is aborted`, SQLSTATE 25P02) — Postgres has no statement-level rollback. Ensure-semantics (an insert that may already exist) uses `ON CONFLICT DO NOTHING` (ent upsert, the `sql/upsert` feature) — no error fires, so nothing can poison; a tolerated failure is the fallback shape only where the statement is not a plain insert: hoist the existence guard before the transaction, or return from the body at the tolerance (the race loser skips the rest). An in-process fake repo cannot surface this class — a fake does not poison shared state; verification is live probing against a real database (§7).
- Identity checks on errors that passed a choke point use the **generated reason predicates**, never kratos code checks: `errorspb.IsInfraErrorDbNotFound(err)` (reason identity), not `errors.IsNotFound(err)` — the latter matches on HTTP code and only works while the mapped code happens to be 404.
- Intra-transaction control flow (retry vs abort on a classified failure) travels as a closure-scoped flag set at the failing statement — never as a circulating sentinel error; a reason predicate alone cannot say which statement in the transaction failed.
- shared-go middleware sentinels declare internal-code reasons (`errors.Unauthorized("MISSING_SESSION", "")`) and are rebound onto catalog factories at service assembly (`server.go` init, original kept as cause): shared-go holds no generated-code knowledge, vocabulary drift = compile error at the rebind.

**Wire boundary.** Only predefined enums leave the service. `sanitize.Server()` (chain outermost) masks any non-kratos outbound error to a `GeneralError`; `sanitize.Client()`/`MapUpstreamErr` remap any provider error by HTTP code into this service's space — no upstream reason or transport detail crosses back.

**Observability — stock middleware composed, never forked.** Chain (outermost → innermost), identical across all three servers:

```
sanitize → tracing → metrics → logging(RenderErrors(logger)) → ErrorTelemetryServer → recovery → ratelimit → metadata → guard → validator
```

- The `guard` slot is layered: `DefaultGuard` (matched on `ACCESS_UNSPECIFIED` — no annotation fails closed) plus one selector layer per credential family, e.g. `krauth.SessionAuth(authn, krauth.SessionCookie(...))` matched on `ACCESS_ADMIN`. `authn` arrives as the `krauth.Authenticator` interface — the module's UC satisfies it via `wire.Bind`, so `server` never imports the module.
- Plane gates ride the credential rungs: the ADMIN rung carries `security.BackendPlaneGuard()` ahead of authz — a wrong-plane session is rejected as wrong-plane, never as a missing grant — and a rung whose audience is one world carries that world's gate (a C-end BUILTIN rung takes `security.MemberPlaneGuard()`; a rung serving both worlds — the session verifier, presign, transfer — stays session-only). The plane always derives from the verified seat, never from request input.
- `RenderErrors(logger)` wraps the logger handed to the stock logging middleware (an slog handler):
  - the `error` attr becomes the factual root;
  - `internal` and a non-empty `message` become their own attrs;
  - the duplicated `stack` attr is dropped;
  - 4xx failures are demoted to WARN.
- `ErrorTelemetryServer` sits immediately inside logging and inside sanitize — it observes the **pre-mask** error. Three duties, none of which the stock middleware provides:
  - span attrs `error.reason`/`error.internal` (the aggregation keys);
  - counter `kratos.server.errors{operation,reason,code}`;
  - outbound re-rendering via `renderFlat` — `Error()` becomes one line `REASON [INTERNAL > PATH] (message): root` while `errors.As`/`Unwrap`/`FromError` keep resolving to the original chain, so the stock tracing Exception event and the wire translation stay semantically unchanged.
- The Exception event itself stays with the stock tracing middleware — recording a second one would split every failure into two error groups.
- Framework middleware errors are pre-mapped in `internal/server/server.go` `init()`: recovery → `GENERAL_ERROR_UNSPECIFIED`, ratelimit → `FLOW_ERROR_RATE_LIMITED`, validator → `GENERAL_ERROR_VALIDATION_FAILED`, missing access annotation → `GENERAL_ERROR_INTERNAL`.
- biz does not log errors — the observability outlet is complete by construction; `Log` on UC/RP carries flow info only.
  - Defensible exception — **swallow paths whose log line is the only outlet**: (a) fail-closed engine/RP swallows (policy evaluation error, unregistered kind, dirty params, attribute resolution) log ERROR/WARN at the swallow site — the dropped input's only telemetry; (b) best-effort follow-ups in a UC after the primary action succeeded (revoke-sessions after disable/delete, orphan cleanup after a committed write, sweep/generation loop skips) log WARN at the swallow site — same only-outlet rationale.
- Aggregation keys: wire & counter = `reason`; spans = `error.reason` + `error.internal`; log attrs = `reason` + `internal`.

**Client side (when a remote RP client lands):** mirror the server — an `ErrorTelemetryClient` inside `sanitize.Client`, observing the pre-map upstream error (counter `kratos.client.errors`); `RenderErrors` already covers `logging.Client`.

**Anti-patterns:** don't `return nil, err` from a client with the provider's error untouched; don't expose provider `reason` strings; don't skip the logging middleware (you lose detail + trace correlation); don't author cause prose with a SCREAMING prefix or `errors.New` (prose = `fmt.Errorf`; reason owns SCREAMING); don't nest a second identity under a factory message; don't classify ent/capability errors at call sites instead of the choke points.

---

## 7. Verification

Every change under `app/services/*` or `shared-go/**` closes with its chain green and its behavior probed live; the report states what ran, with counted evidence.

**Chain** — all exit 0, cache bypassed (`--skip-nx-cache`):

- schema or wiring touched → `<service>:generate:ent` → `:migrate:diff` → `:migrate:apply` → `:generate:wire` → `:build`; otherwise → `<service>:build`
- always → `tools:go:lint`, `tools:go:format`
- a service whose `_test.go` files the change touches or adds → `<service>:test` (e2e tests skip themselves without `ENT_E2E=1`; unit tests and guards run bare)
- `shared-go/**` touched → `shared-go:test` (the kept suites are the capability conformance face)
- proto touched → the proto chain (proto CONVENTIONS §8)

**Tests have three homes, all opt-in per slice** (business code is not test-first by default; a slice adds tests only as regression anchors it wants to keep): live end-to-end tests live ONLY in `<svc>/internal/e2e/` — a package of nothing but `_test.go` files so `go build` skips it and production can never import it — each env-gated on `ENT_E2E=1` and self-skipping when the dev stack is absent, fixtures self-contained and prefix-swept; unit tests target ONLY the biz layer — `biz.go` itself and its domain companion files (§3: zero ent imports, no UC/RP receivers) — each in `<file>_test.go` beside its code (pure logic — no DB, redis, or network; RP ports appear as interfaces only); invariant guards live ONLY in `<svc>/internal/guards/` — the same only-test-files package shape as e2e — each guarding one invariant, never rehearsing a business flow: a guard qualifies when its red means a law was broken, structural (module DAG, comment and serialization conventions, mixin ordering, seed vocabulary) or behavioral (fail-closed predicates, tenant-scoped storage routing), and stays DB-free (AST walks, file scans, in-memory fakes) so it runs under `<service>:test` with the rest. No `_test.go` files anywhere else under `app/services/*`: layer companions (ent-touching, layer-receiver files) and the data/service/platform layers carry no unit tests — their verification is the live probe above plus the e2e face.

**Runtime verification = live RPC probe.** grpcurl against the dev stack (migrated, seeded, up): the happy path plus the **refused path** — every permission gate gets one denied call, every idempotent write gets one replay, every narrowing claim gets one cross-tenant read refused. Probe scripts are throwaway, never repo assets.

**Review carries what no machine checks.** The reviewer walks the diff against this document's named signatures: declaration/file-header comments (§4), reversed layer direction and UC→UC wiring (§1), the four-file shape and section order (§3), mixin declaration order (§5), hand-rolled duplicates of utils helpers (§5), identity outside catalog factories and choke-point classification (§6), the middleware contracts (§5a). A difference between siblings of one family is either a bug or a distinction worth a comment.

**Declaration.** The change report closes with one line of counted evidence — what was run and what was observed ("proven by: 2/2 refused-path probes; chain green").
