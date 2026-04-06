# Cyber Ecosystem (CLAUDE.md)

## 1) Role

Platform Engineer + Service Architect. Nx-driven, strict layering (`server → service → biz → data`), proto/schema/wire as source-of-truth. Favor reusable patterns over one-off shortcuts.

`MUST`/`MUST NOT` = hard gate. `SHOULD` = strong default. `MAY` = optional.

## 2) Stack

Go 1.25.x | Kratos v2 | gRPC + HTTP + ConnectRPC | PostgreSQL + Ent ORM | Redis | OTel + Prometheus + Jaeger + Loki + Grafana | Nx (`./nx`) | Buf | Wire | xid (20-char IDs) | Zap + Loki | JWT (HS256) | go-i18n/v2

## 3) Commands

`MUST` use `./nx run <project>:<target>`. `MUST NOT` bypass with ad-hoc commands.

### Platform & Proto

| Target | Purpose |
|---|---|
| `tools:go:init` | Install proto toolchain |
| `tools:buf:dep/format/lint` | Buf operations |
| `tools:gci:check/format` | Import ordering |
| `tools:debug:jwt` | JWT debug |
| `contracts:proto` | → `contracts/go/**` |
| `<app>_api:proto:api` | → `apps/<app>/gen/go/**` + `openapi.yaml` |
| `<app>_<service>:proto:conf` | → `internal/conf/conf.pb.go` |

### Service Targets

| Target | Purpose |
|---|---|
| `<app>_<service>:generate` | Wire + Ent + i18n + `go mod tidy` |
| `<app>_<service>:build` | Build to `./bin/` |
| `<app>_<service>:dev` | Dev server |
| `<app>_<service>:ent:new` | Scaffold Ent entity |
| `<app>_<service>:generate:i18n` | i18n templates |

### Infrastructure

`infra:docker:up/down/clean/ps/logs/restart` | Individual: `postgres/redis/jaeger/prometheus/grafana/minio/tusd` | Groups: `monitoring`, `oss`

## 4) Architecture

### Layers & Access

| Layer | Dir | Role | Can access |
|---|---|---|---|
| Server | `internal/server/` | Transport, middleware, error mapping | service, pkg, conf |
| Service | `internal/service/` | Proto ↔ entity mapping | biz, pkg |
| Biz | `internal/biz/` | Use cases, models, port interfaces | contracts, shared-go |
| Pkg | `internal/pkg/` | Cross-cutting domain services | contracts, shared-go |
| Data | `internal/data/` | Repo impl, Ent/cache, cross-service clients | biz + pkg interfaces, shared-go |

### Dependency Rules

- Direction: `server → service → biz → data` and `server/service/biz → pkg → data`
- `MUST NOT`: service→data, server→biz, server→data
- Pkg `MUST NOT` import biz, service, server, or data
- Pkg defines minimal interfaces; Data implements; Wire binds

### Pkg (Domain Service) Pattern

Pkg provides cross-cutting domain capabilities (auth, casbin, etc.) accessible by all upper layers:

1. Pkg defines a minimal interface for its infrastructure needs (e.g., `auth.Store{GetCache()}`, `casbin.PolicyRepo{Load/Create/Delete}`)
2. Data layer implements the interface (via Store methods or dedicated RP)
3. Wire binds the interface to the implementation
4. All upper layers can directly use pkg types (e.g., `*auth.Manager`, `*casbin.Manager`)

### Security Capability Pattern

Platform-level security concerns `MUST` be exposed through pkg contracts, not direct `server -> biz` wiring:

1. `internal/pkg/security` defines capability contracts and shared request context (`SessionValidator`, `Authorizer`, `ConditionChecker`, `Mode`, request metadata helpers)
2. Server layer consumes only pkg contracts + config modes
3. Biz layer adapts concrete UCs/RPs into those contracts
4. Data permission filters must read operation metadata from shared security context, not directly from transport state

### Provider Sets & Wire

- Every layer `MUST` export `var ProviderSet = wire.NewSet(...)`
- Data layer binds interfaces: `wire.Bind(new(biz.Transaction), new(*Store))`, `wire.Bind(new(auth.Store), new(*Store))`
- New service: (1) add to `service.ProviderSet`, (2) add to `NewRegistrarList`, (3) wire all 3 transports

## 5) Naming Conventions

### Files

| Concept | Pattern | Example |
|---|---|---|
| Service | `<entity>.go` | `account_auth.go` |
| Use case | `uc_<entity>.go` | `uc_account.go` |
| Repository | `rp_<entity>.go` | `rp_user.go` |
| Domain service (pkg) | `<name>.go` | `manager.go` |
| Ent schema | `<entity>.go` | `system_user.go` |
| Mixin | `<descriptive>.go` | `soft_delete.go` |

### Structs

| Concept | Pattern | Example |
|---|---|---|
| Service | `<Entity>Service` | `AccountAuthService` |
| Use case | `<Entity>UC` | `AccountUC` |
| Repository | `<Entity>RP` | `UserRP` |
| Domain service | `<Name>Manager` or `<Name>` | `auth.Manager` |
| Entity | `<Entity>` or `<Entity>Entity` | `User` |
| Query I/O | `<Entity>QueryIn`/`Out` | `UserQueryIn` |

### Constructors

- `New<StructName>`, first/second param `log.Logger`
- Repo returns interface: `func NewBlogRP(...) biz.BlogRP`
- Pkg accepts interface: `func NewManager(repo PolicyRepo) (*Manager, error)`
- Logger module: `"<layer>/<filename>"` (e.g., `"data/rp_blog"`)

### Embedding

- Biz: `UC{log, tm Transaction}` → concrete UCs
- Data: `RP{log, store}` → concrete RPs

### Imports (gci enforced)

Stdlib → Third-party → Kratos → Shared internal (`cyber-ecosystem/shared-go`, `cyber-ecosystem/contracts`) → App-local (`cyber-ecosystem/apps/<app>`)

## 6) Layer Patterns

### Service

Map proto → entity, call UC, error passthrough (`nil, err`), map result → proto.

Utilities: `utils.ToTimestamp`, `utils.FromTimestamp`, `utils.Wrap`, `utils.StringW`, `utils.Unwrap`, `utils.SliceMap`, `utils.EnsurePageRequest`, `utils.ParseOrderBy`

### Biz (`uc_*.go` three-section layout)

1. **Model**: Entity, `QueryIn` (embeds `*common.PageRequest`), `QueryOut` (embeds `*common.PageResponse`)
2. **Port**: `<Entity>RP` interface
3. **UC**: Struct embedding `UC`, constructor, methods

Write ops: `uc.tm.InTx(ctx, func(ctx context.Context) error { ... })`. Reads: passthrough.

### Data

- **Store**: `{cache, db}` → `NewStore` returns `(*Store, cleanup, error)`. Implements `biz.Transaction` and pkg interfaces.
- Errors: `HandleError(err)` → `entutil.HandleEntError()`
- fields_mask: `utils.Handler{...}.Emit(fieldsMask)`
- Query: `GetClient → Query → WherePtr → ApplyOrderBy → ApplyPagination → All → map`

## 7) Proto Conventions

- Package: `api.<app>.v1`, Go alias: abbreviated (`singletonV1`)
- Request/Response: `<Method><Entity>Request`/`Response`
- Nullable: request use `optional`, response use `StringValue`/`Timestamp` wrappers
- `fields_mask`: `repeated string fields_mask = 100`
- `order_by`: `repeated string order_by = 100` with CEL
- Annotations: `google.api.http`, `(auth.public_access)`, `(desc.service_comment)`, `(buf.validate)`, `(errors.code)`
- Error enum: `ErrorReason` with `ERROR_REASON_*` values + `(errors.code)`
- Pagination: `PageRequest` (`page_no`, `page_size`, `all`, `*_a`=GTE / `*_z`=LTE dates), `PageResponse` (`page_no`, `page_size`, `total`, `more`)

## 8) Ent Conventions

- Schema: embed `ent.Schema`, mixins in order: ID → Timestamps → domain
- Mixins: `IDStringMixin` (xid, 20 chars), `CreatedUpdatedMixin`, `SortMixin`
- Soft delete: app-local, `deleted_at` + auto-filter interceptor + delete-to-update hook
- Indexes: `entsql.IndexWhere("deleted_at IS NULL")` for partial indexes
- Defaults: `var XxxDefault<Field> = func() T { ... }`

## 9) Server & Middleware

Middleware chain (all transports): i18n → recovery → ratelimit → metrics → tracing → metadata → logging → JWT auth (selector) → security capabilities → validate. HTTP/Connect add CORS.

Security capabilities are controlled by `security.<capability>.mode`:
- `disabled`: capability does not participate
- `observe`: capability runs but logs failures instead of blocking
- `enforce`: capability runs and fails closed

Rules:
- Server `MUST NOT` import biz directly for security decisions
- Enabled security capabilities `MUST` fail closed on runtime/config/context errors
- Public access remains explicit via proto annotations, not implicit middleware fallthrough
- Data scope filtering `MUST` rely on shared request context, so non-transport callers can opt in explicitly

Error mapping in `init()`: framework errors → proto error reasons.

## 10) Generation & Source of Truth

Change first, then regenerate:
- Proto: `contracts/**`, `apps/<app>/api/**`, `internal/conf/conf.proto`
- Schema: `internal/data/ent/schema/**`
- Wire: `cmd/app/wire.go` + provider sets

Flow: update source → regenerate (`contracts` → `<app>_api` → service) → implement → build → test → commit.

When Ent `schema/local_mixins` imports generated `ent` types, use two-stage generation:
1. Temporarily detach mixin usage from schema files and generate base `ent`
2. Restore mixins and generate again for final output

## 11) File Policy

### Allowed

- `apps/<app>/api/**/*.proto`, `contracts/**/*.proto`
- `apps/<app>/services/<service>/internal/**` (non-generated)
- `configs/config.yaml`, `shared-go/**` (platform-level, backward compatible)

### Generated Only (via generators)

`gen/**`, `contracts/go/**`, `conf.pb.go`, `data/ent/**` (except `schema/`), `wire_gen.go`, `i18n/translations/**`

### No-Go

No Nx bypass. No generated code patching. No proto naming breakage. No hardcoded secrets.

### Cross-layer Edits

Touches service + biz + data together: stabilize interfaces first → update implementations → generate/build.

## 12) Shared Modules

| Path | Key exports |
|---|---|
| `shared-go/cache/` | Cache facade (KV, Counter, Session, SortedSet, RateLimiter). Redis + memory |
| `shared-go/kratos/logging/zap/` | Zap logger + Loki |
| `shared-go/kratos/middleware/` | auth, i18n, validate |
| `shared-go/kratos/port/` | Transaction interface |
| `shared-go/kratos/transport/connect/` | ConnectRPC server/client/codec/health |
| `shared-go/orm/ent/entutil/` | InTx, GetClientFromTx, ApplyPagination, ApplyOrderBy, WherePtr, HandleEntError |
| `shared-go/orm/ent/mixins/` | IDStringMixin, CreatedUpdatedMixin, SortMixin |
| `shared-go/utils/` | `conv_*` (Ptr, Wrap/Unwrap, ToTimestamp/FromTimestamp, SliceMap, ConvPtr, Deref, EnsurePageRequest), proto_reflect, masks (Handler), orderby, encrypt |

## 13) Environment

Ports: `1100x` (HTTP), `1200x` (gRPC), `1300x` (Connect), `1400x` (Ops). DB: `cyber_ecosystem_<app>_<service>`.

Init: `infra:docker:up` → `contracts:proto` → `<app>_api:proto:api` → `proto:conf` → `generate` → `dev`

## 14) Validation

Test: `go test ./...` (no unified Nx target).

**DoD**: (1) Correct Nx generation, (2) Build passes, (3) Tests on touched packages, (4) No unintended generated changes, (5) Layering intact. Skip = document reason.

Failure handling: `generate` fails → fix source defs; `build` fails → fix hand-written code; `test` fails in untouched packages → record as pre-existing.

**Caveat**: `tools/go-jwt` vet warning (redundant newline in Println). Pre-existing, ignore unless task targets it.

## 15) Error Handling

### Principles

Errors are Kratos `*errors.Error`（proto-defined）at creation time. Upper layers passthrough or remap at domain boundary.

### Core Rules

| # | Rule | Violation | Correct |
|---|---|---|---|
| 1 | 定型于出错层 | `fmt.Errorf("%w", HandleError(err))` | `HandleError(err)` |
| 2 | 禁止二次包装 | `fmt.Errorf("ctx: %w", kratosErr)` | 直接返回 Kratos error |
| 3 | 禁止嵌套 | `Unauthorized("").WithCause(Unauthorized(""))` | 收到 Kratos error 直接 passthrough |
| 4 | 禁止吞写操作错误 | `_ = repo.WriteOp(...)` | 返回 error |
| 5 | 裸 error 仅限 cause | `return nil, rawCacheErr` | `Sentinel.WithCause(rawCacheErr)` |

### Layer Responsibilities

- **Data**: Ent → `HandleError()`；Cache → `singletonV1.ErrorReasonXxx("").WithCause(err)`。`HandleError()` 结果不得再用 `fmt.Errorf` 包装。
- **Pkg**: 返回 Kratos error。`MAY` 定义 `var ErrXxx` sentinel 提升可读性。初始化阶段（NewXxx）`MAY` 用 `fmt.Errorf`。
- **Biz**: 默认 passthrough。领域重映射用 `singletonV1.IsXxx()` 检查后创建新 Kratos error，原 error 作为 `.WithCause()`。
- **Service**: 纯 passthrough（`return nil, err`）。`MUST NOT` 创建新 Kratos error。
- **Server**: `init()` 映射框架中间件错误。

### Exception

初始化/启动阶段（NewClient、NewManager、Schema.Create）`MAY` 用 `fmt.Errorf`，因错误不经过业务 middleware。
