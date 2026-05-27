# Kratos + Go Service Stack Guide

Guide for coding agents working on Go microservices built with the Kratos framework in this monorepo.

---

## 1. Architecture

### Layered Structure

Every service follows a strict layered architecture with unidirectional dependencies:

```
server → service → biz ← data → platform
                      ↑         ↑
                   proto      ent
```

**Dependency rule:** arrows point inward. Inner layers MUST NOT import outer layers.

| Layer | Directory | Responsibility |
|-------|-----------|----------------|
| server | `internal/server/` | Transport setup (HTTP, gRPC, Connect), middleware registration |
| service | `internal/service/` | Request/response mapping, proto handler implementation |
| biz | `internal/biz/` | Domain models, use cases, port interfaces (RP), business logic |
| data | `internal/data/` | Repository implementations (RP), external data access |
| platform | `internal/platform/` | Infrastructure container (DB, cache, storage), error handling |
| ent | `internal/ent/` | Ent ORM schemas and generated code |
| i18n | `internal/i18n/` | Translation files (YAML) and generated bundle |

**Special rule:** biz layer MAY depend on proto (contracts) for error codes. This is the only allowed biz → outer dependency.

### Wire DI Assembly

`cmd/app/wire.go` is the composition root — the ONLY file that knows concrete bindings:

```go
panic(wire.Build(
    server.ProviderSet,
    service.ProviderSet,
    biz.ProviderSet,
    data.ProviderSet,
    i18n.ProviderSet,
    platform.ProviderSet,
    wire.Bind(new(biz.Transaction), new(*platform.Platform)),
    newApp,
))
```

Each package exposes a `ProviderSet` via `wire.NewSet(...)`. Interface-to-implementation bindings (`wire.Bind`) live exclusively in `wire.go`.

### Transaction Interface

Defined in `biz/biz.go`, implemented by `platform.Platform`:

```go
type Transaction interface {
    InTx(ctx context.Context, fn func(context.Context) error) error
}
```

Use cases that modify data wrap their logic in `uc.tm.InTx(ctx, func(ctx) error { ... })`.

### Bootstrap Flow

`cmd/app/main.go` orchestrates startup:

1. Load config from directory via `file.NewSource`
2. Initialize logger (Zap — console/file/OTLP modes)
3. Initialize metrics (OTel MeterProvider)
4. Initialize tracing (OTel TracerProvider)
5. Initialize Sentry (error reporting)
6. Call `wireApp(...)` to assemble the full DI graph
7. Run the Kratos app; defer all cleanups

Each telemetry component returns a cleanup function. `main()` defers all cleanups in reverse order.

### Service Capabilities

Every Kratos service shares the same layered architecture. Capabilities are enabled by including or omitting specific files in the Platform layer:

| Capability | Platform files | Config key |
|------------|---------------|------------|
| Database (Ent) | `platform_ent.go`, `platform_ent_handler.go` | `Data.database` |
| gRPC Client (remote calls) | `platform_grpc.go` | `Data.<service_name>` |
| Cache | `platform_cache.go`, `platform_cache_handler.go` | `Data.cache` |
| Storage (S3) | `platform_storage.go`, `platform_storage_handler.go` | `Data.storage` |

Common configurations in this repo:
- **Base service** (owns domain logic, DB access): DB + Cache + gRPC server + Ops
- **BFF service** (calls other services, exposes HTTP/Connect): gRPC client + Cache + HTTP + Connect + Ops
- **Monolith**: all capabilities enabled

The `Platform` struct adapts to the enabled capabilities:

```go
// With database:
type Platform struct {
    cache            *cache.Cache
    handleCacheError CacheErrorHandler
    db               *ent.Client
    handleEntError   EntErrorHandler
}

// With gRPC client (no database):
type Platform struct {
    cache            *cache.Cache
    handleCacheError CacheErrorHandler
    articleClient    appV1.ArticleServiceClient
    resourceClient   appV1.ResourceServiceClient
}
```

`InTx` also adapts: with DB it opens a transaction; without DB it calls `fn(ctx)` directly.

---

## 2. File Naming Convention

Pattern: `{aggregate}_{type}.go` — suffix-based, same-aggregate files group alphabetically.

### biz layer

| Suffix | Purpose | Example |
|--------|---------|---------|
| (none) | Aggregate root model | `message.go` |
| `_uc` | Use case struct + RP port interface + methods | `message_uc.go` |
| `_fsm` | FSM / state machine behavior | `message_fsm.go` |
| `_ds` | Domain service (cross-aggregate logic) | `notification_ds.go` |
| `_oc` | Orchestration / saga (cross-UC coordination) | `order_oc.go` |
| `_event` | Domain event definitions | `message_event.go` |
| `_handler` | Event handler / subscriber | `message_handler.go` |

Base types (`UC` struct, `Transaction` interface) live in `biz.go`.

### data layer

| Suffix | Purpose | Example |
|--------|---------|---------|
| `_rp` | Repository implementation | `message_rp.go` |

Base types (`RP` struct, shared helpers) live in `data.go`.

### service layer

No suffix — files named by aggregate: `message.go`, `file.go`, `resource.go`.
Base types (`Registrar` interface, `ProviderSet`) live in `service.go`.

### platform layer

Split by concern — each infrastructure type has two files:

| Pattern | Purpose |
|---------|---------|
| `platform_{type}.go` | Initialization (client setup, connection) |
| `platform_{type}_handler.go` | Error handler for that infrastructure |

Example: `platform_cache.go` + `platform_cache_handler.go`.

Base types (`Platform` struct, handler types) live in `platform.go` and `interface.go`.

---

## 3. Code Organization (Region Annotations)

All Go and proto files use `// region[rgba(...)] EMOJI Label` / `// endregion` pairs for code block organization.

These render as colored backgrounds in VSCode (Colored Regions extension) and as foldable blocks in JetBrains IDEs.

### Color Scheme (Scheme A)

| Section | RGBA | Emoji |
|---------|------|-------|
| Model | `rgba(239,83,80,0.15)` | 🔴 |
| Port | `rgba(66,165,245,0.15)` | 🔵 |
| UC | `rgba(102,187,106,0.15)` | 🟢 |
| Method | `rgba(186,104,200,0.15)` | 🟣 |
| Struct | `rgba(236,64,122,0.15)` | 🩷 |
| Handler | `rgba(255,167,38,0.15)` | 🟠 |
| Private | `rgba(144,164,174,0.10)` | ⚪ |
| Repo | `rgba(0,188,212,0.12)` | 🩵 |
| FSM States | `rgba(255,238,88,0.12)` | 🟡 |
| FSM | `rgba(255,167,38,0.15)` | 🟠 |
| Domain Method | `rgba(186,104,200,0.15)` | 🟣 |

Example:

```go
// region[rgba(239,83,80,0.15)] 🔴 Model

type Message struct {
    ID        *string
    Title     *string
    Status    *string
}

// endregion

// region[rgba(66,165,245,0.15)] 🔵 Port

type MessageRP interface {
    Get(ctx context.Context, id string) (*Message, error)
}

// endregion
```

### Section Order Within Files

**biz `{aggregate}_uc.go`:** Port → UC → Method
**service `{aggregate}.go`:** Struct → Handler → Private
**data `{aggregate}_rp.go`:** Repo → Private
**biz `{aggregate}.go`:** Model only

---

## 4. Proto Workflow

### Source-First

1. Edit `.proto` files in `api/v1/`
2. Run `./nx run <project>:generate` (or individual targets)
3. Generated Go code lands in `gen/go/v1/`
4. Generated TypeScript clients land in `clients/<client>/src/services/connect/` (Connect) and `clients/<client>/src/services/openapi/` (OpenAPI)

### Error Handling Chain

Define error enums in `error_reason.proto`:

```proto
enum ErrorReason {
  ERROR_REASON_ENT_NOT_FOUND = 10 [(errors.code) = 404];
  ERROR_REASON_MESSAGE_INVALID_STATUS_TRANSITION = 44 [(errors.code) = 400];
}
```

Generated code produces `ErrorErrorReasonXxx(cause)` functions that create Kratos `*errors.Error` with correct HTTP status codes.

### Validation

- Field-level: `(buf.validate.field)` annotations (required, min_len, in, etc.)
- Cross-field: `(buf.validate.message).cel` for complex constraints
- Runtime: `validate.ProtoValidate()` middleware auto-validates all requests

---

## 5. Error Handling Pattern

Error contracts are centralized in `contracts/go/errors` (not project-local). The platform layer maps infrastructure errors to these contract errors.

### Ent (Database)

```go
// platform/platform_ent_handler.go
var defaultEntError = &entutil.DefaultError{
    NotFound:    errorspb.ErrorInfraErrorDbNotFound(""),
    Validation:  errorspb.ErrorInfraErrorDbValidation(""),
    NotSingular: errorspb.ErrorInfraErrorDbNotSingular(""),
    NotLoaded:   errorspb.ErrorInfraErrorDbNotLoaded(""),
    Constraint:  errorspb.ErrorInfraErrorDbConstraint(""),
    Internal:    errorspb.ErrorInfraErrorDbInternal(""),
}
```

### Storage (S3)

```go
// platform/platform_storage_handler.go — maps S3 errors to errorspb codes
```

### Cache (Redis)

```go
// platform/platform_cache_handler.go — maps Redis errors to errorspb codes
```

### Built-in Error Overrides

`server/server.go` uses an `init()` function to replace Kratos built-in errors with project-specific ones:

```go
func init() {
    recovery.ErrUnknownRequest = errorspb.ErrorGeneralErrorUnspecified("").WithCause(recovery.ErrUnknownRequest)
    ratelimit.ErrLimitExceed = errorspb.ErrorFlowErrorRateLimited("").WithCause(ratelimit.ErrLimitExceed)
    validate.ErrVALIDATOR = errorspb.ErrorGeneralErrorValidationFailed("").WithCause(validate.ErrVALIDATOR)
}
```

### Usage

```go
// Data layer calls via Platform:
return nil, rp.platform.HandleEntError(err)
return nil, rp.platform.HandleCacheError(err)
return nil, rp.platform.HandleStorageError(err)
```

**Rule:** error handlers live in `platform/`, data layer calls them via `rp.platform.HandleXxxError(err)`. The original error is preserved via `.WithCause(err)`.

---

## 6. i18n

Translation files live in `internal/i18n/locales/` as YAML, one per locale:

- `v1.en-US.yaml`
- `v1.zh-CN.yaml`
- `v1.ar-SA.yaml`

Locale files are embedded via `//go:embed locales/*.yaml`. The bundle is created with a default language (`zh-CN`).

Error enum entries are auto-generated as stubs by `geni18n`. The `generate.go` file contains the `go:generate` directive with an `i18n.protos` file listing which proto files to extract error keys from.

Fill in translations:

```yaml
ERROR_REASON_MESSAGE_INVALID_STATUS_TRANSITION: "Invalid status transition."
```

The i18n middleware reads `Accept-Language` header and translates error messages automatically.

Regenerate stubs: `./nx run <project>:generate:i18n`

---

## 7. Ent ORM

### Schema Definition

Schemas in `internal/ent/schema/`, one file per entity:

```go
type Message struct {
    ent.Schema
}

func (Message) Fields() []ent.Field {
    return []ent.Field{
        field.String("title").NotEmpty().MaxLen(64),
        field.String("status").Default("draft").MaxLen(10),
    }
}

func (Message) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixins.IDStringMixin{},       // xid-based 20-char string ID
        mixins.CreatedUpdatedMixin{}, // created_at, updated_at
        local_mixins.SortMixin{SoftDelete: true},
        local_mixins.SoftDeleteMixin{},
    }
}
```

### Available Mixins

| Mixin | Package | Provides |
|-------|---------|----------|
| `IDStringMixin` | `shared-go/orm/ent/mixins` | 20-char xid string primary key |
| `CreatedUpdatedMixin` | `shared-go/orm/ent/mixins` | `created_at`, `updated_at` with indexes |
| `SortMixin` | `local_mixins` | `sort` field with fractional indexing, optional partial unique index for soft-delete |
| `SoftDeleteMixin` | `local_mixins` | `deleted_at` field, query filtering |

### Generation

```bash
./nx run <project>:generate:ent
```

### Query Patterns

```go
// Pagination
total, offset, limit, err := entutil.ApplyPagination(ctx, query, pageReq, config, ce)

// Conditional where (only applies if ptr is non-nil)
entutil.WherePtr(query, filter.Title, func(v string) predicate.Message {
    return message.TitleContains(v)
})

// Order by (maps field names to ent order functions via FOMapping)
entutil.ApplyOrderBy(orderBy, ascFunc, descFunc, fieldMapping)

// Fields mask update (partial update by field name)
utils.Handler{
    "title":   {Condition: m.Title != nil, OnTrue: func() { updater.SetTitle(*m.Title) }, OnFalse: func() {}},
    "content": {Condition: m.Content != nil, OnTrue: func() { updater.SetContent(*m.Content) }, OnFalse: func() { updater.SetContent("") }},
}.Emit(fieldsMask)
```

---

## 8. shared-go Utilities

### Pointer Operations (`utils`)

```go
utils.Ptr(value)              // *T from T
utils.Deref(ptr, default)     // T from *T, with default
utils.PtrApply(ptr, fn)       // apply fn if ptr non-nil
utils.ConvNum[R](intPtr)      // numeric pointer type conversion
```

### Protobuf Conversion (`utils`)

```go
utils.Wrap(ptr, utils.StringW)    // *string → *wrapperspb.StringValue
utils.Unwrap(wrapper)             // *wrapperspb.StringValue → *string
utils.StringW(value)              // string → *wrapperspb.StringValue
utils.Ptr(value)                  // T → *T (e.g., string → *string for request fields)
utils.ToTimestamp(timePtr)        // *time.Time → *timestamppb.Timestamp
utils.FromTimestamp(ts)           // *timestamppb.Timestamp → *time.Time
utils.EnsurePageRequest(req)      // nil-safe page request
```

### Slice Operations (`utils`)

```go
utils.SliceMap(slice, fn)         // []T → []R
utils.SliceMapErr(slice, fn)      // []T → ([]R, error)
```

### Sorting (`utils`)

```go
utils.ParseOrderBy([]string)      // ["field:asc"] → []*OrderBy
utils.StringifyOrderBy([]*OrderBy) // reverse
```

### Pagination (`entutil`)

```go
entutil.ApplyPagination(ctx, query, req, config, ce)  // returns total, offset, limit, err
entutil.BuildPageResponse(total, offset, limit)        // → *common.PageResponse
```

---

## 9. Server and Middleware

### Transport Servers

Each service exposes up to four servers:

| Server | Port (default) | Purpose |
|--------|----------------|---------|
| HTTP | 11000 | REST API (Kratos HTTP) |
| gRPC | 12000 | gRPC API |
| Connect | 13000 | Connect RPC API |
| Ops | 14000 | Prometheus metrics + configurable pprof |

The Ops server is a separate HTTP server for operational concerns, independent of the business API servers.

### Middleware Chain

All transport servers share the same middleware chain, applied in this order:

```
i18n → recovery → ratelimit → metrics → tracing → connect_span* → metadata → logging → validate → error_report
```

- `i18n.Server(bundle)` — translates error messages based on `Accept-Language`
- `recovery.Recovery()` — panic recovery
- `ratelimit.Server()` — rate limiting
- `metrics.Server(...)` — request count and latency histograms
- `tracing.Server(...)` — OTel tracing (conditional on TracerProvider)
- `connect_span.Server()` — Connect-specific span enrichment (Connect server only)
- `metadata.Server()` — metadata propagation
- `logging.Server(logger)` — request logging
- `validate.ProtoValidate()` — proto validation
- `error_report.Server()` — Sentry error reporting

CORS is configured with wildcard origins on all servers.

### `Registrar` Pattern

`service/service.go` defines a `Registrar` interface:

```go
type Registrar interface {
    RegisterGRPC(*grpc.Server)
    RegisterHTTP(*http.Server)
    RegisterConnect(*connect.Server)
}
```

`NewRegistrarList` collects all `Registrar` implementations into a `[]Registrar` slice that feeds the server constructors. Each service aggregate implements `Registrar`.

### Connect Registration

Services with HTTP annotations use generated Connect handlers:

```go
import (
    appV1connect "cyber-ecosystem/apps/<app>/gen/go/v1/<app>V1connect"
)

func (s *ArticleService) RegisterConnect(srv *connect.Server) {
    srv.Register(appV1connect.NewArticleServiceHandler(s, srv.HandlerOptions()...))
}
```

Connect supports three wire formats simultaneously: Connect, gRPC, and gRPC-Web. Clients can use `grpcurl` (gRPC), `curl` with JSON (HTTP), or Connect-native clients.

If a proto service has NO `google.api.http` annotations, the generated code will NOT include `RegisterXxxHTTPServer` or Connect handler. `RegisterHTTP` and `RegisterConnect` must be no-ops.

### BFF HTTP Path Prefix Convention

When an app has multiple BFF services (admin, mobile), each BFF must use a distinct HTTP path prefix to avoid OpenAPI generation collisions:

| BFF | Path Prefix |
|-----|-------------|
| Admin BFF | `/api/v1/admin/` |
| Mobile BFF | `/api/v1/mobile/` |

Without the prefix, two BFFs sharing `GET /api/v1/article` would cause the OpenAPI generator to produce only one set of client functions (the second overwrites the first). The prefix is set in the proto `google.api.http` annotations.

### gRPC Client (Calling Other Services)

Services that call other services use client-side middleware. Create `platform_grpc.go`:

```go
func dialGRPC(c *conf.Data, logger, tp, _metricRequests, _metricSeconds) (*grpc.ClientConn, func(), error) {
    // Client middleware — use Client() variants, not Server()
    var middlewares []middleware.Middleware
    middlewares = append(middlewares, recovery.Recovery())
    middlewares = append(middlewares, circuitbreaker.Client())
    middlewares = append(middlewares, metrics.Client(...))
    if tp != nil { middlewares = append(middlewares, tracing.Client(...)) }
    middlewares = append(middlewares, metadata.Client())
    middlewares = append(middlewares, logging.Client(logger))
    middlewares = append(middlewares, grpc_status.Client()) // innermost — maps gRPC status to Kratos errors

    conn, err := kgrpc.DialInsecure(context.Background(),
        kgrpc.WithEndpoint(c.BaseService.Addr),
        kgrpc.WithTimeout(c.BaseService.Timeout.AsDuration()),
        kgrpc.WithMiddleware(middlewares...),
    )
    // ...
}
```

Key points:
- Use `metrics.Client()`, `tracing.Client()`, `logging.Client()` — client-side variants, not `Server()`
- `grpc_status.Client()` MUST be the innermost middleware (last appended). It maps gRPC transport errors (`Unavailable`, `DeadlineExceeded`) to Kratos errors with proper reason codes, so all outer middlewares (logging, tracing, metrics, circuit breaker) see structured errors with reasons. Without it, gRPC transport failures return empty reason and raw message, breaking i18n translation and error reporting.
- Each remote service gets its own `NewGRPCXxxClient` constructor, all sharing `dialGRPC`
- Remote service config lives in `conf.Data`, named per service (e.g., `Data.BaseService`)
- `kgrpc.DialInsecure` returns `*grpc.ClientConn` (standard gRPC), not a Kratos type

Remote service config in `conf.proto`:

```protobuf
message Data {
  message BaseService {
    string addr = 1;
    google.protobuf.Duration timeout = 2;
  }
  Cache cache = 1;
  BaseService base_service = 2;  // named per remote service
}
```

### `newApp` Server Selection

All services create gRPC, HTTP, Connect, and Ops servers for structural consistency. `newApp` selectively starts them — the signature is identical across all services, only the `srv = append(...)` lines differ:

```go
func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, cs *connect.Server, os *server.OpsServer) *kratos.App {
    var srv []transport.Server
    // Uncomment the servers this service exposes:
    // srv = append(srv, gs)  // gRPC (for services with DB)
    if hs != nil { srv = append(srv, hs) }
    if cs != nil { srv = append(srv, cs) }
    if os != nil { srv = append(srv, os) }
    return kratos.New(kratos.Server(srv...), ...)
}
```

This allows quick migration: toggle which servers are started by commenting/uncommenting `srv = append(srv, gs)`.

---

## 10. File Service Pattern

File operations (upload/download/delete) use raw HTTP handlers registered via `srv.Route(prefix).Handle(method, path, handler)` with `ctx.Middleware()` for middleware chain integration.

Key patterns:
- Upload: `http.MaxBytesReader` for size limit → S3 upload → DB metadata with status `"attached"`. On DB failure, the S3 object is deleted (compensating transaction to prevent orphans).
- Download: DB lookup → S3 download → stream response with correct headers
- Delete: S3 delete first → then DB metadata delete (prevents orphans)
- No transactions needed for delete (single DB op, S3 is non-transactional)

---

## 11. Rich Domain Model Pattern

Use `looplab/fsm` for state machines. Define in `biz/{aggregate}_fsm.go`:

```go
func newMessageFSM(current string, m *Message) *fsm.FSM {
    return fsm.NewFSM(current,
        []fsm.EventDesc{
            {Name: "published", Src: []string{"draft"}, Dst: "published"},
            {Name: "archived",  Src: []string{"draft", "published"}, Dst: "archived"},
            {Name: "draft",     Src: []string{"archived"}, Dst: "draft"},
        },
        map[string]fsm.Callback{
            "after_published": func(_ context.Context, _ *fsm.Event) { *m.Status = "published" },
            // ...
        },
    )
}

func (m *Message) TransitionTo(ctx context.Context, target string) error {
    m.Status = utils.Ptr(utils.Deref(m.Status, "draft"))
    f := newMessageFSM(*m.Status, m)
    if err := f.Event(ctx, target); err != nil {
        return appV1.ErrorErrorReasonXxx("").WithCause(err)
    }
    return nil
}
```

UC layer calls: load entity → domain method → save:

```go
func (uc *MessageUC) UpdateStatus(ctx context.Context, id string, target string) (out *Message, err error) {
    err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
        m, e := uc.messageRP.Get(ctx, id)
        if e != nil { return e }
        if e = m.TransitionTo(ctx, target); e != nil { return e }
        out, e = uc.messageRP.Update(ctx, []string{"status"}, m)
        return
    })
    return
}
```

---

## 12. Resource Introspection

The `ResourceService` is a special aggregate that provides proto service introspection — it reads proto file descriptors from `protoregistry.GlobalFiles` at runtime to enumerate services, methods, HTTP annotations, and comments.

This is NOT a database-backed entity. It has no ent schema. The repository reads from the proto registry directly.

---

## 13. Nx Targets

Common targets for Go services (check `project.json` for availability):

```bash
./nx run <project>:build              # Compile binary
./nx run <project>:generate           # Full generation chain (ent + wire + proto + i18n)
./nx run <project>:generate:ent       # Regenerate ent ORM code
./nx run <project>:generate:wire      # Regenerate Wire DI
./nx run <project>:generate:proto     # Regenerate proto stubs (if separate target)
./nx run <project>:generate:i18n      # Regenerate i18n translation stubs
```

Always use `./nx run` — never run `go build`, `wire`, `buf generate`, etc. directly.

---

## 14. Project Structure Reference

```
apps/<app>/
  api/v1/                            # Proto source files (app-level)
  gen/go/v1/                         # Generated proto code (app-level)
  services/<service>/
    internal/
      ent/schema/                    # [DB] Ent schema definitions
        local_mixins/                # Service-specific mixins
      biz/                           # Domain layer
        biz.go                       # UC, Transaction base types
        {aggregate}.go               # Aggregate root models
        {aggregate}_uc.go            # Use cases + RP ports
        {aggregate}_fsm.go           # FSM / state machine
      data/                          # Data access layer
        data.go                      # RP base, shared helpers
        {aggregate}_rp.go            # Repository implementations
      service/                       # Transport handler layer
        service.go                   # Registrar, ProviderSet, NewRegistrarList
        {aggregate}.go               # Proto service handlers
      server/                        # Transport setup
        server.go                    # init() error overrides, ProviderSet
        http.go                      # HTTP server
        grpc.go                      # gRPC server
        connect.go                   # Connect server
        ops.go                       # Ops server (Prometheus + pprof)
      platform/                      # Infrastructure container
        platform.go                  # Platform struct, InTx, ProviderSet
        interface.go                 # Handler type definitions
        platform_ent.go              # [DB] Ent client init (slow query logging)
        platform_ent_handler.go      # [DB] Ent error mapping config
        platform_grpc.go             # [gRPC Client] Remote service client creation with middleware
        platform_cache.go            # Cache init (Redis or in-memory)
        platform_cache_handler.go    # Cache error mapping
        platform_storage.go          # [Storage] S3 storage init
        platform_storage_handler.go  # [Storage] Storage error mapping
      i18n/                          # i18n bundle
        generate.go                  # go:generate directive
        i18n.go                      # Bundle init with embed
        locales/                     # Translation YAML files
      conf/                          # Config proto definitions
    cmd/app/                         # Entry point
      main.go                        # Bootstrap: config → logger → metrics → tracing → sentry → wire
      logger.go                      # Zap logger setup (console/file/OTLP)
      tracing.go                     # OTel TracerProvider
      metrics.go                     # OTel MeterProvider
      sentry.go                      # Sentry init
      resource.go                    # OTel resource with service metadata
      wire.go                        # Wire injection (build tag)
      wire_gen.go                    # Wire generated (DO NOT EDIT)
    configs/config.yaml              # Service configuration
```

`[DB]`, `[gRPC Client]`, `[Storage]` markers indicate capability-specific files — include only the files needed for the service's capabilities.
