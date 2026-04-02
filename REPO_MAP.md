# Repository Map (Cyber Ecosystem)

## 1) Top-Level Structure

```text
apps/
  app_1/
    api/                         # App-level proto API definitions
    gen/                         # Generated Go stubs and OpenAPI
    services/service_1/          # Kratos service implementation
contracts/                       # Shared proto contracts across apps
shared-go/                       # Shared Go platform libraries
infra/docker/                    # Local infra and observability stack
tools/                           # Dev/debug utility commands
```

---

## 2) API / Biz / Data Core Locations

### API (Protocol Definitions)

App API files:
- `apps/app_1/api/v1/auth.proto`
- `apps/app_1/api/v1/blog.proto`
- `apps/app_1/api/v1/reading.proto`
- `apps/app_1/api/v1/error_reason.proto`

Shared contract files:
- `contracts/auth/auth.proto`
- `contracts/common/page.proto`
- `contracts/errors/errors.proto`

Generated API code:
- `apps/app_1/gen/go/v1/**`
- `apps/app_1/gen/go/v1/app1V1connect/**`
- `apps/app_1/gen/oas/openapi.yaml`

### Biz (Business Core)

- `apps/app_1/services/service_1/internal/biz/biz.go`
- `apps/app_1/services/service_1/internal/biz/blog.go`
- `apps/app_1/services/service_1/internal/biz/author.go`

Biz owns:
- domain entities
- UC orchestration
- repository interfaces
- transaction abstraction (`Transaction`)

### Data (Persistence + Integration)

- `apps/app_1/services/service_1/internal/data/data.go`
- `apps/app_1/services/service_1/internal/data/ent.go`
- `apps/app_1/services/service_1/internal/data/cache.go`
- `apps/app_1/services/service_1/internal/data/blog.go`
- `apps/app_1/services/service_1/internal/data/author.go`

Ent schema source:
- `apps/app_1/services/service_1/internal/data/ent/schema/blog.go`
- `apps/app_1/services/service_1/internal/data/ent/schema/author.go`

Generated Ent code:
- `apps/app_1/services/service_1/internal/data/ent/**`

---

## 3) Runtime Composition and Call Path

```text
cmd/app/main.go
  -> wireApp (cmd/app/wire.go)
    -> server providers (internal/server)
      -> service registrars (internal/service)
        -> use cases (internal/biz)
          -> repositories (internal/data)
            -> Ent client / cache adapters
```

Transport constructors:
- `internal/server/grpc.go`
- `internal/server/http.go`
- `internal/server/connect.go`
- `internal/server/ops.go`

Service registration hub:
- `internal/service/service.go`

---

## 3.1 Dependency Boundary Matrix

- `internal/server` MAY depend on: `internal/service`, `internal/conf`, `shared-go/*`.
- `internal/service` MAY depend on: `internal/biz`, generated API packages, `shared-go/*`.
- `internal/biz` MAY depend on: contracts/common utilities and interfaces only.
- `internal/data` MAY depend on: `internal/biz` interfaces, Ent generated code, `shared-go/*`.

Forbidden:
- `internal/biz` MUST NOT depend on `internal/server` or transport packages.
- `internal/service` MUST NOT use Ent client directly.
- `internal/server` MUST NOT implement business rules.

---

## 4) Shared Platform Modules (shared-go)

Key reusable modules:
- `shared-go/kratos/middleware/**` (auth, i18n, traceheader, validate)
- `shared-go/kratos/transport/connect/**` (Connect server/client helpers)
- `shared-go/kratos/logging/zap/**` (Zap + Loki logging integration)
- `shared-go/cache/**` (memory/redis adapters and wrappers)
- `shared-go/orm/ent/**` (Ent client, transaction, pagination, error helpers)
- `shared-go/utils/**` (order-by parsing, masks, value helpers)

---

## 5) Current Implementation Status (Audit: 2026-04-04)

- Fully implemented service chains:
  - `BlogService`: API -> service -> biz -> data complete
  - `AuthorService`: API -> service -> biz -> data complete
- Partially implemented:
  - `ReadingService` proto exists, but no matching service/biz/data implementation yet.

Generated artifacts to expect during normal workflow:
- `contracts/go/**`
- `apps/app_1/gen/**`
- `apps/app_1/services/service_1/internal/conf/conf.pb.go`
- `apps/app_1/services/service_1/cmd/app/wire_gen.go`
- `apps/app_1/services/service_1/internal/data/ent/**`
- `apps/app_1/services/service_1/internal/i18n/translations/v1.*.yaml`

---

## 6) Change Impact Guide

Use this quick map to choose regeneration/build scope:
- Change in `contracts/**/*.proto`:
  - Run `./nx run contracts:proto`
  - Then run `./nx run app_1_api:proto:api` if app API imports are affected
- Change in `apps/app_1/api/**/*.proto`:
  - Run `./nx run app_1_api:proto:api`
  - Rebuild service consumers
- Change in `internal/conf/conf.proto`:
  - Run `./nx run app_1_service_1:proto:conf`
  - Then `./nx run app_1_service_1:generate` if DI/types are impacted
- Change in `internal/data/ent/schema/**`:
  - Run `./nx run app_1_service_1:generate`
- Change in `cmd/app/wire.go` or provider sets:
  - Run `./nx run app_1_service_1:generate`
