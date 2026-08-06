# ADR-0001: Auth, RBAC, ABAC, Datascope & Multi-tenancy

- **Status:** M1–M4 (identity / tenant / selector / token full loop) + dept (flat + tree) **landed and e2e-verified — code is the source of truth; this ADR no longer tracks their implementation.** Remaining work is re-sequenced in *Remaining work*. Root design decisions (D1–D13) are stable. **Current focus (2026-08-07): pivoted to completing auth entities (user full CRUD) + admin web UI before resuming backend authorization — see *Remaining work › Current focus*.**
- **Date:** 2026-08-01 (last revised 2026-08-07)

## Context

cyber-ecosystem is a generic, business-agnostic platform base. Target workloads span **management apps** (single-tenant CRM) and **IoT / robotics platforms** (multi-tenant with org hierarchy: 省/市/县/变电所). This ADR replaces a two-system split ("pure admin single-tenant" vs "SaaS multi-tenant" auth) with **one layered model** — the migration cost of promoting single-tenant → SaaS under the split is prohibitive, and the boundary is blurred in practice.

Reference `smart-park` (Kratos v2, lab repo) was studied and **rejected as a reference implementation** — disconnected scaffolding (consumer JWT vs admin UUID; two JWT middlewares; tenant resolver→ctx→ent-hook chain) never registered to any server, so no endpoint was protected and no tenant isolation took effect. Kept only as vocabulary + anti-pattern checklist.

## Decision principle

**One system, layered.** Concepts every management app must touch go in a mandatory **Core**; everything else is **opt-in extensions**. The line is drawn at *what is universal to management apps*, not at *single vs multi tenant*.

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
Every core table carries `tenant_id`; single-tenant deploy = one default tenant; an ent interceptor injects `WHERE tenant_id = ?` from ctx **transparently — business code never touches tenant**. Chosen over "pure extension" because the single-tenant cost is low (one column + one default seed + transparent injection) and it makes CRM→SaaS a **zero-schema-migration** evolution.

**Convention:** every unique index MUST be composite with `tenant_id` (e.g. `(tenant_id, username)`), so multi-tenant promotion never rewrites indexes.

### D3 — Org unit (flat `dept`) is Core; tree hierarchy is Extension
Datascope needs an anchor (which unit a user belongs to), so a flat `dept` is Core. Tree hierarchy is an opt-in mixin that unlocks the "my-dept and subordinates" datascope tier. **Landed as adjacency list** (`parent_id` + recursive `IsAncestor` traversal + cycle/delete guards), not materialized path — simpler to maintain, sufficient for shallow org trees. Whether a `path` column / closure table is needed for the my-dept-and-sub *query* is an open implementation point at M6 (see D11).

### D4 — Datascope concept is Core; tier implementation is layered
Core datascope = flat three tiers **mine / my-dept / all**. "My-dept **and subordinates**" requires the org-tree extension.

### D5 — Authentication: opaque token (redis)
Management apps need **instant revocation**. JWT can't revoke instantly without a denylist (defeats statelessness). Optional local LRU cache (short TTL ~5s) for high-frequency paths (see *Optional*).

IoT/robotics high-frequency traffic is **device telemetry**, not user API — flows over NATS/MQTT with connection-level auth, off the token hot path. Device identity is a separate extension.

### D6 — Interface-level: casbin (RBAC with domains) + ABAC two-layer
casbin is a general engine; RBAC-with-domains maps to multi-tenancy (domain = tenant). ABAC dynamic context (time/location/device) is a **separate condition evaluator** running after casbin, AND-ed with context attributes — not shoved into casbin policy. casbin owns role/permission skeleton; dynamic conditions stay in code. ABAC is an extension; Core's casbin integration must leave an extension point.

### D7 — Data-level: ent mixin + privacy/intercept
ent `intercept` is enabled and in use (`SoftDeleteMixin`, `TenantMixin` — intercept + ctx + `q.WhereP`, auto-wired by codegen). The `privacy` feature (declarative `policy.QueryRule`) is a one-line opt-in composing with ctx, but **not a hard prerequisite** — can ship on pure intercept, upgrade later. Mixins carry `tenant_id` / `org_id` / `created_by`. Lesson: row-level filtering uses intercept/privacy, not hooks.

### D8 — Request-level and data-level bridged by ctx
Middleware resolves identity (tenant/user/datascope/ABAC context) into ctx → biz → ent interceptor reads ctx → injects filters.

### D9 — Interface chain driven by proto `access` annotation + kratos selector
A `proto/ext` access annotation selects which auth middleware chain each RPC walks. Chain-selection MUST use kratos's transport-agnostic `middleware/selector`, NOT the in-repo connect-only `matcher` (breaks grpc/connect symmetry). Deny-by-default; public endpoints opt out.

### D10 — Service-to-service identity + multi-service user-token verification
cyber runs across **docker compose → k3s → customer k8s**, and a service mesh may or may not be present — so service identity **cannot depend on mTLS/mesh**. It is an **application-layer internal token** (signed by a service key):
- `caller` — the calling service's identity (always present; verified by callee → casbin service permission).
- `for` — the proxied user token (optional; present = proxy mode, absent = autonomous/service-account).

One mechanism covers both call patterns; user-identity propagation rides in `for`, so datascope survives cross-service hops. mTLS layers on top later (transport identity); the internal token stays app-level.

**⚠ Open — multi-service user-token verification.** D10 relays a user token across services but does not define how the *callee verifies* it. The landed single-service auth (`BearerAuth` + `Authenticator` bound to system's redis/biz) cannot be reused by sample. Must decide before M7 (2nd service): shared token-store redis / callback-to-system verify / other. (See *Open questions*.)

### D11 — Datascope: role-scoped, ent-privacy row filter
A role carries a `data_scope` enum (`mine` / `my-dept` / `all`; `my-dept-and-sub` = org-tree ext). On login, middleware resolves the user's scopes into ctx (widest wins). Business tables carry mixins: `tenant_id`, `created_by`, `org_id`. An ent privacy/intercept rule reads scope from ctx:
- `mine` → `WHERE created_by = current_user`
- `my-dept` → `WHERE org_id = current_user.org_id`
- `all` → no filter (still tenant-filtered)
- `my-dept-and-sub` (ext) → subtree membership of the user's dept (adjacency-list traversal today; a `path`/closure-table optimization is an open point at M6)

The rule ANDs with the tenant rule. Scope is global (one user, one scope). "mine" defaults to `created_by`, field name configurable per table.

### D12 — casbin policy storage & sync: PG + ent adapter + cache pub-sub
casbin is in-memory per instance (`enforce` microsecond, no per-request DB). Storage: PostgreSQL via a **self-implemented ent adapter** (policy table managed by ent). Roles/permissions dynamic → DB-backed, not `.csv`. Multi-instance reload: on policy write, publish `policy.changed`; subscribers `LoadPolicy()` — best-effort broadcast over **cache (redis) pub-sub** (already wired), not `mq` (reserved for business events). Enforce-result caching not needed.

### D13 — Token & refresh lifecycle
- **Access token:** opaque, short TTL (15min), redis `token → {userId, tenantId, sessionId}`. *(roles/scope appended when RBAC lands — see M5.)*
- **Refresh token:** opaque, long TTL (7d), redis `refresh → sessionId`. **Rotation**: each refresh issues a new one and deletes the old.
- **Session:** redis source of truth; PG session table optional (audit only).
- **Revocation:** delete the session → all bound access tokens fail liveness check immediately.
- **Permission-change broadcast:** when RBAC lands, a change broadcasts revocation of that user's access tokens → next request 401 → frontend refresh mints new access with new perms.

> **Landed as strategy A (M4):** rotation deletes the old refresh (it can't refresh again). **Reuse detection deferred** — its only extra value is "leak visibility + race protection", at the cost of false positives (network retry → a legitimate old refresh reused → session revoked). Add a `CurrentRefresh`-pointer check only if a high-security deployment needs it.

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
    → [data: ent privacy/intercept reads ctx → row-level filter (tenant/datascope)]
```

## Consequences

- **+** Single reusable model; CRM→SaaS is zero-schema-migration.
- **+** Declarative row-level access control; business code unaware of tenant/datascope.
- **+** Instant revocation (opaque token).
- **−** Tenant cognitive tax on every developer (mitigated by injection transparency).
- **−** Recommended: enable ent `privacy` feature; can ship on pure intercept meanwhile.
- **Risk** Middlewares must actually be wired — smart-park's failure was defining middlewares but never registering them. *(M1–M4 verified wiring across connect/grpc/http.)*

## Implementation paradigm

Cross-M code-layering rules — followed by M1–M4, required for M5–M8:

### 范式规则
1. **biz 是依赖中心,定义所有 domain port**(`UserRP`/`TokenRP`…);data 实现;**biz 不依赖共享业务接口**。横切契约(`Authenticator`,未来 `Authorizer`)定义在共享 middleware 包,biz 实现 + `wire.Bind` 注入(DIP)。
2. **服务特有 middleware 在服务 internal**;**横切、跨服务复用**的 middleware(BearerAuth/RBAC/datascope)在共享层(`shared-go/kratos/security/`),import 自己包的契约、不 import biz。
3. **`shared-go/kratos/security/` 只放多服务纯原语 + 横切契约**(`Subject`/`WithSubject`;matchers;guards;`Authenticator`);token 存储/RBAC/datascope 的 *domain* port 在各服务 biz。
4. **biz private 辅助优先纯函数(不接 port)**;port 调用在 UC method 编排;一次性计算内联 method,复用的纯计算才提 port-free 纯函数。

### 包结构 — security 顶层分层
- `shared-go/kratos/security/`(根):`identity.go`(`Subject`/`WithSubject`/`SubjectFromCtx`,跨 auth/authz/datascope 共用)、`matcher.go`(`MatchAccess`)、`default_guard.go`(`DefaultGuard`)。
- `shared-go/kratos/security/auth/`:`authenticator.go`(`Authenticator` 契约)、`bearer.go`(`BearerAuth` 两阶段 + 3 个 middleware 错误集中)、`token.go`(`GenerateToken`)。
- 后续 `security/authz/`、`security/datascope/` 平级子包。

### middleware 错误范式
- middleware 内置 kratos error var(`ErrMissingToken`/`ErrTokenExpired`/`ErrInvalidToken`),server `init()` 覆盖成服务 errorspb(validator 模式)——middleware 不直接返回 proto,解耦。
- reason 由 middleware **调用结构**决定(BearerAuth 两阶段:Authenticate 失败→`ErrTokenExpired`;CheckSession 失败→`ErrInvalidToken`),biz 返回裸 err,不碰 middleware var。
- biz 只产 proto 错误枚举(errorspb)或原始 err。

### access 选链 + deny-by-default
proto `cyber.ext.v1.access` 注解 + `security.MatchAccess` 驱动 selector 多链(PUBLIC/ADMIN/UNSPECIFIED→DefaultGuard 拒)。UNSPECIFIED 不兜底任何 audience。

## Remaining work (re-sequenced)

> M1–M4 landed — code is the source of truth. Below is only what's **not yet built**, regrouped by what unblocks it.

### Current focus (2026-08-07)

Backend authorization (M5 RBAC / M6 datascope / M7 service identity / M8 broadcast) is **paused**. Pivoting to: **complete auth entities (user full CRUD) + admin web UI first, then resume RBAC.**

Rationale — RBAC depends on what isn't built yet:
- `UserRole` (assign roles to users) needs a **complete user entity** — currently only `create/get`, no `list`.
- RBAC admin UI (role/permission pages) needs the **admin web framework**.

Building these first lands RBAC on a real base rather than abstract scaffolding. Product-driven (usable admin console early) over engine-driven (authz engine + curl only).

Sequence:
1. ✅ casbin groundwork deferred (no-dependency foundation; role/permission schema re-added when M5 resumes).
2. **user full CRUD** (backend: list/update/delete) — dept already complete.
3. **admin web UI framework** (login + layout + user/dept management pages).
4. resume **M5 RBAC** (+ role/permission UI + permission control) → security closure.
5. M6 datascope / ABAC after.

### Milestones (pending)

| M | Unblocks / depends | ADR | Scope |
|---|---|---|---|
| **M4.5** session policy | —(零重写) | D13 | `SINGLE_GLOBAL`/`SINGLE_PER_CLIENT` 并发登录;`auth:user:<uid>(:<client>)` 索引 + login 踢旧;config 选策略枚举 |
| **M5** RBAC | 业务厚度 | D6+D12 | casbin + `Authorizer` + **ent adapter**;access token 补 roles/scope |
| **M6** datascope | dept + M5 | D11 | DataScopeMixin(mine/dept/all);intercept(或 privacy) |
| **M7** service identity | 第 2 个服务 | D10 | internal token(caller+for)+ **多服务 user-token 验证(见 Open)** |
| **M8** cross-instance broadcast | 多实例 | D13 | revocation + permission-change + policy-reload(cache PubSub) |

### Extensions (deferred, business-driven — no milestone)
- **ABAC** dynamic context (time/location/device) — D6 Ext
- **Org tree** (adjacency-list structure **landed**; the my-dept-and-sub datascope tier still pending M6) — D3 Ext
- **UI permission** granularity (menu/button)
- **Multi-tenant management plane** (Login tenant field + tenant CRUD + cross-tenant e2e)
- **Device identity / registry** (IoT) — separate ADR when IoT lands

### Optional (not scheduled)
- **LRU token cache** (D5, short TTL ~5s) — only if token-verify QPS demands
- **ent `privacy` feature** (D7) — M6 can ship on pure intercept; privacy is sugar
- **Refresh reuse detection** (D13) — strategy A landed; add only for high-security deploys

## Open questions

1. **Multi-service user-token verification (D10 gap)** — single-service `BearerAuth`+`Authenticator` is bound to system's redis/biz; sample can't verify a user token. Options: shared token-store redis / callback-to-system verify / other. **Must decide before M7** (2nd service).
2. **Device identity (IoT)** — separate ADR when IoT lands.
3. **Token at-rest hardening (SHA-256)** — opaque tokens key redis by plaintext; a redis dump leak exposes usable tokens. Deferred to real-business phase.

## Alternatives considered

- **Two systems (pure single-tenant + pure multi-tenant).** Con: single→multi is an auth rewrite — exactly the cost this ADR avoids.
- **Tenant as pure extension (no tenant_id in core).** Con: single→multi needs columns + migration + conditional branches.
- **JWT (stateless).** Con: no instant revocation; management apps need it.
- **casbin owns all ABAC.** Con: dynamic context (time/location) is awkward in policy tables.
- **ent hooks for row-level filtering.** Rejected — fragile (SortMixin lesson); use intercept/privacy.
