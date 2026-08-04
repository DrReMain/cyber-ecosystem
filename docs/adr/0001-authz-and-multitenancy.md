# ADR-0001: Auth, RBAC, ABAC, Datascope & Multi-tenancy

- **Status:** Proposed — root decisions locked. **M1 (identity) + M2 (tenant) + M3 (selector) + M4 (token full loop) landed and e2e-verified; M5–M8 deferred pending business/service/multi-instance thickness.** Implementation details still evolving (see *Open questions*).
- **Date:** 2026-08-01 (last revised 2026-08-05)

## Context

cyber-ecosystem is a generic, business-agnostic platform base. Target workloads span **management apps** (e.g. a CRM — single-tenant) and **IoT / robotics platforms** (e.g. a power-grid robot-inspection system — multi-tenant with org hierarchy: 省/市/县/变电所). Earlier thinking split the world into "pure admin (single-tenant)" vs "SaaS (multi-tenant)" and designed a separate auth system for each. This ADR replaces that with **one model that spans both**: the migration cost of promoting a single-tenant app to SaaS under the two-system split is prohibitive, and the boundary is blurred in practice.

Starting point: cyber's auth is a **green field** — no auth/rbac/casbin/tenant/datascope code, no `proto/ext` access annotations, only `item` in the system schema. Two foundations already exist: ent's `intercept` feature is enabled (`generate.go`), and a kratos-style `Matcher` (operation → middleware chain) lives in `shared-go/kratos/transport/connect/internal/matcher`.

Reference `smart-park` (Kratos v2, lab repo) was studied and **rejected as a reference implementation** — its auth is disconnected scaffolding (consumer-side real JWT vs admin-side random UUID; two JWT middlewares; a tenant resolver→ctx→ent-hook chain) **never registered to any server**, so no endpoint was actually protected and no tenant isolation took effect. It is kept only as a **vocabulary and an anti-pattern checklist**.

## Decision principle

**One system, layered.** Concepts every management app must touch go in a mandatory **Core**; everything else is **opt-in extensions**. This is neither "one monolithic system forced on all apps" nor "two specialized systems" — the line is drawn at *what is universal to management apps*, not at *single vs multi tenant*.

## Core (mandatory) vs Extensions (opt-in)

| | Concept | Decision |
|---|---|---|
| Core | Authentication (opaque token) | D5 |
| Core | RBAC (casbin, interface-level) | D6 |
| Core | Datascope (flat: mine / my-dept / all) | D4 |
| Core | Org unit (flat `dept`) | D3 |
| Core | Tenant (always present, transparent) | D2 |
| Core | Service-to-service identity | D10 |
| Ext | Multi-tenant management plane | |
| Ext | Org tree (unlocks "my-dept **and sub**") | D3 |
| Ext | UI permission granularity (menu/button) | |
| Ext | ABAC (dynamic context: time/location/device) | D6 |
| Ext | Device identity / registry (IoT) | separate from user auth |

## Decisions

### D1 — One layered system
Not a monolith (forces multi-tenant tax on single-tenant apps), not two systems (forces a rewrite on single→multi). Layering gives reuse where universal (auth/RBAC/tenant isolation) and freedom where business-specific (org hierarchy, UI perms).

### D2 — Tenant is "degenerate core": always present, transparent
Every core table carries `tenant_id`; single-tenant deploy = one default tenant; an ent interceptor injects `WHERE tenant_id = ?` from ctx **transparently — business code never touches tenant**. Chosen over "pure extension (no tenant_id in core)" because the single-tenant cost is low (one column + one default seed + transparent injection) and it makes CRM→SaaS a **zero-schema-migration** evolution.

**Convention:** every unique index MUST be composite with `tenant_id` (e.g. `(tenant_id, username)`), so multi-tenant promotion never rewrites indexes.

### D3 — Org unit (flat `dept`) is Core; tree hierarchy is Extension
Datascope needs an anchor (which unit a user belongs to), so a flat `dept` is Core. Tree hierarchy (`parent_id` + materialized path) is an opt-in mixin that unlocks the "my-dept and subordinates" datascope tier.

### D4 — Datascope concept is Core; tier implementation is layered
Core datascope = flat three tiers **mine / my-dept / all** (a single-tenant CRM needs it too — sales rep sees own accounts, team lead sees the team's). "My-dept **and subordinates**" requires the org-tree extension.

### D5 — Authentication: opaque token (redis)
Management apps need **instant revocation** (kick user, apply permission change immediately, ban). JWT can't revoke instantly without a denylist (which defeats statelessness). Optional local LRU cache (short TTL ~5s) for high-frequency paths: cache hit ≈ stateless performance, TTL bounds revocation latency.

IoT/robotics high-frequency traffic is **device telemetry**, not user API — it flows over NATS/MQTT with connection-level auth (cert/mTLS), **off the token hot path**. Device identity is a separate extension, not user auth.

### D6 — Interface-level: casbin (RBAC with domains) + ABAC two-layer
casbin is a general engine (RBAC / RBAC-with-domains / ABAC); RBAC-with-domains maps naturally to multi-tenancy (domain = tenant). ABAC dynamic context (time/location/device) is **not** shoved into casbin policy — it's a separate condition evaluator (middleware) that runs **after** casbin passes, AND-ed with context attributes. casbin owns the role/permission skeleton; dynamic conditions stay in code. ABAC is an extension; Core's casbin integration must leave an **extension point** (matcher can take context, ctx carries env attributes).

### D7 — Data-level: ent mixin + privacy/intercept
ent's `intercept` feature is already enabled and **already in production use** — `SoftDeleteMixin.Interceptors()` is a live row-level-filter template (intercept + ctx key + `q.WhereP`), auto-wired into every query by codegen; D2/D11 copy it 1:1 (swap the predicate). The `privacy` feature (declarative `policy.QueryRule` / `DecisionContext`, rule signatures natively carry `context.Context`) is a **one-line opt-in** (`--feature privacy`) and composes cleanly with ctx propagation (D8) — but is **not a hard prerequisite**: MVP can ship on pure intercept, then upgrade to privacy sugar with no ctx-model change. Mixins carry `tenant_id` / `org_id` / `created_by`. Lesson from SortMixin: row-level filtering uses **intercept/privacy, not hooks**.

### D8 — Request-level and data-level are bridged by ctx
Middleware resolves identity (tenant/user/datascope/ABAC context) into ctx → biz → ent interceptor reads ctx → injects filters. Same ctx-propagation pattern as `InTx` handle-in-ctx and SortMixin tenant-in-ctx.

### D9 — Interface chain driven by proto `access` annotation + kratos selector
A `proto/ext` access annotation (`ACCESS_PUBLIC` / `ACCESS_ADMIN` / … — already defined at `proto/ext/v1/access.proto`, `E_Access` generated) selects which auth middleware chain each RPC walks. The chain-selection MUST use **kratos's transport-agnostic `middleware/selector`** (`selector.Server().Match(fn).Build()` → a plain middleware) — NOT the in-repo `transport/connect/internal/matcher`, which is **connect-only** (grpc server is upstream kratos and can't mount it; using it would break grpc/connect symmetry). Deny-by-default; public endpoints opt out. A small bridge (~30 lines) reads `E_Access` from proto descriptors at startup, builds `operation → Access`, feeds `selector.Match`. The in-repo matcher stays reserved for connect-stream-specific cases.

### D10 — Service-to-service identity: application-layer internal token (mesh-agnostic)
cyber must run across **docker compose → k3s → customer-existing k8s**, and a service mesh may or may not be present in any of them — so service identity **cannot depend on mTLS/mesh**. It is an **application-layer internal token** (signed by a service key), carrying two fields:
- `caller` — the calling service's identity (always present; verified by callee → casbin service permission).
- `for` — the proxied user token (optional; present = proxy mode, absent = autonomous/service-account mode).

One mechanism covers both call patterns (user-driven proxy, and service-autonomous jobs/timers). User-identity propagation (token relay) rides in `for`, so datascope survives cross-service hops. mTLS is **not** a replacement: when a mesh lands later for large self-operated apps, it layers on top (transport identity) and the internal token stays as app-level auth.

### D11 — Datascope: role-scoped, ent-privacy row filter
A role carries a `data_scope` enum (`mine` / `my-dept` / `all`; `my-dept-and-sub` is the org-tree extension). On login, middleware resolves the user's roles' scopes into ctx (widest wins: all > and-sub > dept > mine). Business tables carry mixins: `tenant_id` (D2), `created_by`, `org_id`. An ent privacy rule reads scope from ctx and injects:
- `mine` → `WHERE created_by = current_user`
- `my-dept` → `WHERE org_id = current_user.org_id`
- `all` → no filter (still tenant-filtered)
- `my-dept-and-sub` (ext) → `WHERE org_path LIKE current_user.org_path || '%'`

The privacy rule ANDs with the tenant rule. **Scope is global** (one user, one scope across all resources), not per-resource. **"mine" defaults to `created_by`** but the mixin field name is configurable per table (`owner_id` / `assignee_id`).

### D12 — casbin policy storage & sync: PG + ent adapter + cache pub-sub reload
casbin is an **in-memory model**: each service instance loads the full policy into an in-process enforcer at startup; `enforce` runs in memory (microsecond), never hitting DB per request — so there is no per-request cache problem.

- **Storage:** PostgreSQL via a **self-implemented ent adapter** (no official ent adapter exists; the policy table is managed by ent like any business table, for consistency). Roles/permissions are dynamic, so policy is DB-backed, not a `.csv` file.
- **Multi-instance reload:** when any instance writes policy to DB, it publishes a `policy.changed` signal; all instances subscribe and call `LoadPolicy()`. The signal is a **best-effort cache-invalidation broadcast** — a lost message is harmless (subscriber keeps stale policy; truth is in DB, corrected on next reload/restart).
- **Channel:** the broadcast goes over the **cache (redis) abstraction's pub-sub** (`shared-go/cache.PubSub`: `Publish`/`Subscribe`/`PSubscribe`, redis backend fully wired) — **zero extensions needed, the API already exists in v3**. Not `mq`: semantically this is instance-coordination (cache invalidation), not a business event; `mq` is reserved for business messages (IoT/order/notification).
- Enforce-result caching: not needed (in-memory enforce is fast enough).

### D13 — Token & refresh lifecycle: opaque, rotated refresh, revoke-via-broadcast
- **Access token:** opaque random string, **short TTL (15min, configurable)**, redis `token → {userId, tenantId, roles, scope, sessionId}` — **carries permissions** so per-request authz doesn't re-query roles.
- **Refresh token:** opaque, **long TTL (7d, configurable)**, redis `refresh → sessionId`, used only to mint a new access. **Rotation + reuse detection** (OAuth2 best practice): each refresh issues a new refresh and invalidates the old; if an old refresh is used twice (suspected theft), the whole session is revoked.
- **Session:** redis is the source of truth; a PG session table is optional (audit/history only).
- **Permission-change effective immediately (option A):** access carries permissions, so a change broadcasts revocation of that user's access tokens (delete from redis + cache pub-sub to all instances, same channel as D12) → next request 401 → the frontend refresh-interceptor mints a new access with the new permissions. **User feels nothing; permissions take effect immediately.** This is what D5's opaque choice promised; it pairs with the reactive 401-interceptor contract (`auth-client-refresh-contract`).
- **Emergency ban / kick:** revoke the whole session (access + refresh deleted from redis), immediate, forces re-login.

## Request-level vs Data-level (key boundary)

| Concern | Layer | Mechanism |
|---|---|---|
| Interface RBAC (who can call which API) | request | casbin middleware |
| ABAC dynamic context (time/location/device) | request | condition-evaluator middleware |
| tenant isolation (row) | data | ent interceptor |
| datascope (mine/dept/all, row) | data | ent privacy/intercept |
| resource-attribute ABAC (e.g. secret-level, row) | data | ent privacy rule |
| owner/creator | data | mixin (`created_by`) + interceptor |

## Request flow

```
RPC → [selector: access annotation picks chain]
    → [middleware: authn(opaque→ctx) → casbin(RBAC) → ABAC context eval → ctx]
    → biz
    → [data: ent privacy/interceptor reads ctx → row-level filter (tenant/datascope)]
```

## Consequences

- **+** Single reusable model; CRM→SaaS evolution is zero-schema-migration.
- **+** Declarative row-level access control (ent privacy); business code unaware of tenant/datascope.
- **+** Instant revocation (opaque token).
- **−** Tenant cognitive tax on every developer (mitigated by injection transparency).
- **−** Recommended: enable ent `privacy` feature (+ regenerate); MVP can defer and ship on pure intercept.
- **Risk** Middlewares must actually be wired — smart-park's core failure was defining middlewares but never registering them. Needs an integration check that endpoints are really protected. **Mitigated in M1**: e2e verified login → token → protected RPC (200) and 401 without token, across connect/grpc/http.

## Open questions (to decide next)

1. **Device identity extension (IoT)** — device registry, certs, online state (separate ADR when IoT lands).
2. **Token at-rest hardening (deferred to real-business phase)** — opaque tokens currently key redis by plaintext (`auth:token:<token>`); a redis dump/RDB leak exposes usable tokens. Plan: store SHA-256 of the token and key redis by the hash (token is 256-bit high-entropy → fast SHA-256 suffices; no bcrypt/salt needed, unlike low-entropy passwords). Deferred deliberately — not a skeleton-phase concern; recorded so M4–M8 implementers know the gap.

**Resolved in this revision:** service-to-service identity (D10), datascope tiers (D11), casbin storage & sync (D12), token & refresh lifecycle (D13).

## Implementation Roadmap

> This section is **execution guidance, not a contract**. The order follows "**trunk first, thickening follows**": identity (who is logging in) is the trunk of auth; tenant / RBAC / datascope are thickening that grows around identity. Auth's core must reach a basic working state first, then thicken incrementally. Steps may be adjusted during implementation, provided the **paradigm rules** below are kept. Each sub-task is a **minimum closed loop** — compile/test passes, you can stop and verify there. **Transitional code is allowed** but must be replaced by a later M and marked.

### Execution discipline — paradigm rules

Verified against the ad73adb layered paradigm; evolved through M1 landing (see *Implementation paradigm*). Every step MUST keep:

1. **biz is the dependency center for domain ports** (`UserRP` / `TokenRP` / …); data implements them; **biz depends on no shared business interface**. Cross-cutting contracts consumed by shared middleware (`Authenticator`, future `Authorizer`) are defined in the shared middleware package and implemented by biz via `wire.Bind` — DIP, detailed below.
2. **Service-specific middleware belongs to the service** (`internal/server/middleware/`); **cross-cutting, reuse-intended middleware** (BearerAuth / RBAC / datascope) lives in the shared layer (`shared-go/kratos/security/`) via dependency inversion — it imports its own package's contract, never biz. `shared-go` ships matchers / guards / cross-cutting middleware, never a concrete service's business middleware.
3. **`shared-go/kratos/security/` holds only multi-service pure primitives + cross-cutting contracts** (ctx plumbing `Subject` / `WithSubject`; matchers; guards; `Authenticator`); token *storage* / RBAC / datascope *domain* ports belong to **each service's biz** (they depend on that service's domain types).
4. **Minimum closed loop**: each sub-task compiles / is verifiable, no dead code, no dangling deps; transitional code must be replaced by a later M and marked.

### Paradigm open points (discuss before coding that M)

- **M5 (RBAC)** — casbin enforcer lifecycle ownership: biz adapter vs platform handle.
- **M6 (datascope)** — datascope scope source: baked into token payload vs queried from role live.
- **M7 (service identity)** — service-identity middleware ownership: shared reuse vs each service internal.

### Milestones

> **Roadmap re-sequenced (2026-08-05).** The auth **trunk closes first** (identity → tenant → selector → token full loop): these need no business thickness and are curl-verifiable. **RBAC / datascope are deferred** — they need real role/permission scenarios and data distributions to validate, which only land with actual business logic (same YAGNI as multi-tenant deepening). Service identity needs a 2nd service; cross-instance broadcast needs multi-instance deploy. The old linear M1–M7 was renumbered so **milestone order = execution order**.

| M | Phase | ADR | Goal | Minimum closed loop |
|---|---|---|---|---|
| **M1** ✅ identity (single-tenant) | done | D5+D8+D13p | user entity CRUD + login + opaque-token verify; identity (Subject) into ctx | create→get user; login → token → protected RPC → Subject in ctx — **landed** |
| **M2** ✅ tenant | done | D2+D8 | user gains `tenant_id`; login carries tenant; ent transparent filter | tenant always present, transparent — **landed** |
| **M3** ✅ selector | done | D9 | access annotation + selector; public/admin/unspecified walk different chains | public RPC token-free, admin requires token, unspecified denied — **landed with M1** |
| **M4** ✅ token full loop | done | D13 | refresh token + rotation + single-instance revoke (reuse detection deferred) | login → refresh → new pair; old refresh 401; logout → access 401 — **landed** |
| **M5** RBAC (casbin) | deferred — needs business thickness | D6 | role/permission schema + casbin + interface-authz middleware | unauthorized role rejected |
| **M6** datascope | deferred — needs business thickness + RBAC | D11 | DataScopeMixin intercept (mine/dept/all) | different-scope users see different rows |
| **M7** service identity | deferred — needs 2nd service | D10 | internal token (caller+for) outbound + downstream verify | sample→system carries internal token, downstream verifies |
| **M8** cross-instance broadcast | deferred — needs multi-instance | D13 | cache PubSub revocation broadcast / policy-reload | revocation effective across instances immediately |

**Relation to D2:** M1 single-tenant is the *development-time* evolutionary start (no `tenant_id` on the table); `tenant_id` + transparent injection land only in M2 — at which point D2's "tenant always present, transparent" becomes the end-state. D2's "zero-schema-migration" promise targets **production-time CRM→SaaS evolution** (tenant already present), not development-time M1→M2.

> The selector's PUBLIC tier shipped with M1's Login (Login must be token-free); M3 (selector) is fully landed. **M4 (token full loop) closes the auth trunk**: refresh + rotation + single-instance revoke (`KV.Del` is nearly free under the opaque token). M8 adds cross-instance broadcast (cache PubSub) — only when multi-instance deploy lands.

### M1 small steps — landed

M1 shipped as **vertical slices** (split by entity/capability full-stack, not by layer; depended-on capability first: user → login → verify). **Done and e2e-verified.** The original step grid planned `TokenService{Issue,Verify}` / `Session` / `capability/auth` paths, which drifted on landing — the authoritative shape lives in *Implementation paradigm (landed)* below. Notable deltas vs the plan:

- domain token port is `TokenRP{Set,Get}` (biz defines, data implements), not a `TokenService{Issue,Verify}`.
- identity aggregate is `security.Subject` (root `identity.go`), not a per-auth `Session`.
- verify runs in the shared `security/auth.BearerAuth` + DIP `Authenticator` (`wire.Bind` → `*biz.AuthUC`), not a service-internal middleware file.

Closed loop: create→get user; login → opaque token → BearerAuth → `Subject` in ctx; 401 without token. M2 layered `tenant_id` + transparent filtering on top.

### M2–M8 small steps (refined on arrival, same method)

- **M2 tenant** ✅ landed — `Subject.TenantID` (no separate `WithTenant`; Subject carries all identity fields) → `TenantMixin` (local, intercept-based transparent `WHERE tenant_id`) → user gains `tenant_id` + composite unique `(tenant_id, email)` → login carries tenant → `issueToken` marshals TenantID → interceptor reads `SubjectFromCtx().TenantID`. Single-tenant via `defaultTenant`. **Multi-tenant deepening (Login tenant field + `FindByEmail(tenant,email)` + cross-tenant e2e) deferred** — no real 2nd tenant to validate against; M2's hooks make it a pure additive change later.
- **M3 selector** ✅ landed with M1 — `security.MatchAccess` + `DefaultGuard`, deny-by-default, PUBLIC/ADMIN/UNSPECIFIED three chains (connect/grpc/http).
- **M4 token full loop** ✅ landed — M4.1 session 引擎(`TokenRP{Set,Get,Del}` + `auth:access/refresh/session:` prefix 规约) → M4.2 login 双 token(access+refresh) + `Authenticate` session 存活校验(D1 即时吊销) → M4.3 `Refresh` rotation → M4.4 `Logout`(Del session)。e2e 全通过:login→refresh→新对、旧 refresh 401、logout→access 401。
  - **复用检测 deferred(策略 A)**: rotation 删旧 refresh 已让旧 token 无法再 refresh(核心达成);复用检测(`CurrentRefresh` 指针比对)额外价值仅"泄漏可见 + 竞争防护",代价是误报(网络重试→合法用户旧 refresh 被用→误吊销)——管理后台非必需。需高安全时再加(指针比对 + 不删旧 refresh)。
  - **M4.5 session policy(deferred,零重写)**: `SINGLE_GLOBAL`(单会话)/`SINGLE_PER_CLIENT`(单客户端)并发登录控制。机制:`auth:user:<uid>(:<client>)`→sid 索引 + login 读索引踢旧 + 写新索引;策略由 config 选预定义枚举;`client_type` 加 `LoginRequest`。`Session` struct 不改,Login 线性结构使踢旧/写索引插首尾即可。
- **M5 RBAC** *(deferred — needs business thickness)* — decide enforcer ownership (open point) → role/permission schema → casbin → Authorizer interface → middleware.
- **M6 datascope** *(deferred — needs business thickness + RBAC)* — decide scope source (open point) → DataScopeMixin.
- **M7 service identity** *(deferred — needs 2nd service)* — decide middleware ownership (open point) → internal token (caller+for).
- **M8 cross-instance broadcast** *(deferred — needs multi-instance)* — cache PubSub revocation broadcast + policy-reload.

## Implementation paradigm (landed in M1)

M1 落地后确立/调整的实现范式(对上方 paradigm rules 的细化与演进):

### 包结构 — security 顶层分层
- `shared-go/kratos/security/`(根):身份聚合 + access 选链基建 —— `identity.go`(`Subject`/`WithSubject`/`SubjectFromCtx`,跨 auth/authz/datascope 共用,故置于根而非 auth 子包)、`matcher.go`(`MatchAccess`/`AccessOf` 读 proto 注解)、`default_guard.go`(`DefaultGuard`,UNSPECIFIED 业务 RPC 兜底拒)。
- `shared-go/kratos/security/auth/`:认证 —— `authenticator.go`(`Authenticator` 契约 / `ErrInvalidToken`)、`bearer.go`(`BearerAuth`/`ErrMissingToken`)、`token.go`(`GenerateToken`)。
- 后续 `security/authz/`(RBAC/ABAC)、`security/datascope/` 平级子包,各 import `security`(选链)+ `security/auth`(身份)。
- 横切 middleware 放共享层(不塞服务 internal),命名按职责、不带 MW 前缀(子包已表达)。

### paradigm rule 演进 — 横切 middleware 的依赖倒置
原 rule①② 假设 middleware 属服务、port 在 biz。但**横切、需跨服务复用**的 middleware(BearerAuth/RBAC/datascope)按 DIP 调整:
- **契约定义在共享 middleware 包**(`security/auth.Authenticator`),**biz 实现**(`AuthUC` implements),**wire.Bind 显式注入**(`wire.Bind(new(krauth.Authenticator), new(*biz.AuthUC))`)。
- 共享 middleware 依赖自己包的契约,**不 import biz** → 可复用;biz import 契约实现,逻辑留 biz;data 保持原子(cache 只 `Get`/`Set`,被 biz 调)。
- (ad73adbc 的 `pkg/security` 同款:契约在共享层,实现在服务。)
- **domain port**(如 `biz.TokenRP`,只本服务用)仍在 biz、data 实现 —— 与横切契约并存,不混。

### 原子可组合 middleware
每个横切件单一职责(`BearerAuth`=认证、`Authorize`=RBAC、`InjectScope`=datascope),`selector.Server(MW...).Match(MatchAccess(level))` 按 audience 组合链;audience 由 proto access 注解决定,不由 MW 名表达。

### middleware 边界 mask(安全)
middleware 是对外边界,**不透传 UC 的 err**(可能裸 error / 带细节):
- 已知安全(`ErrInvalidToken`)→ 收口成它(401);其余 → mask 成兜底(原 err 只进 cause 给 log,对外只 reason)。
- 中间件内置 kratos error var(`ErrMissingToken`/`ErrInvalidToken`/`ErrMissingANNOTATION`),server `init()` 覆盖成服务 errorspb(空 message + 原 var 作 cause)—— validator 模式。
- **biz 只产 proto 错误枚举(errorspb)或原始 err,不碰 middleware 内置 var**。

### data 原子
data 只做原子能力(cache `Get`/`Set` + `HandleXxxError`),无业务逻辑;unmarshal / 判断 / mask 在 biz/middleware。middleware/biz 都不直碰 cache 句柄。

### access 选链 + deny-by-default
proto `cyber.ext.v1.access` 注解 + `security.MatchAccess` 驱动 selector 多链(PUBLIC→公共链、ADMIN→BearerAuth 链、UNSPECIFIED→DefaultGuard 拒)。UNSPECIFIED **不兜底**任何 audience(不假设 ADMIN);default guard 按 namespace 放行框架内置(grpc health/reflection 经 interceptor 走链,非 `cyber.*` 放行)。

## Alternatives considered

- **Two systems (pure single-tenant + pure multi-tenant).** Pro: single-tenant apps minimally customized. Con: single→multi is an auth rewrite — exactly the cost this ADR exists to avoid.
- **Tenant as pure extension (no tenant_id in core).** Pro: core tables cleanest. Con: single→multi requires adding columns + migration, and core code needs conditional tenant branches.
- **JWT (stateless).** Pro: no redis lookup per request. Con: no instant revocation; management apps need it.
- **casbin owns all ABAC.** Con: dynamic context (time/location) is awkward in policy tables.
- **ent hooks for row-level filtering.** Rejected — fragile (SortMixin lesson); use intercept/privacy.
