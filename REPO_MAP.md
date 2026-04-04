# Repository Map (Cyber Ecosystem)

## 1) Top-Level Structure

```text
apps/
  app_1/
    api/                         # App-level proto API definitions
    gen/                         # Generated Go stubs and OpenAPI
    services/service_1/          # Kratos service: blog + author
    services/service_2/          # Kratos service: reading
contracts/                       # Shared proto contracts across apps
shared-go/                       # Shared Go platform libraries
infra/docker/                    # Local infra and observability stack
tools/                           # Dev/debug utility commands
```

---

## 2) API / Biz / Data Core Locations

### API (Protocol Definitions)

App API files:
- `apps/app_1/api/v1/author.proto`
- `apps/app_1/api/v1/blog.proto`
- `apps/app_1/api/v1/reading.proto`
- `apps/app_1/api/v1/error_reason.proto`

Shared contract files:
- `contracts/auth/author.proto`
- `contracts/common/page.proto`
- `contracts/errors/errors.proto`

Generated API code:
- `apps/app_1/gen/go/v1/**`
- `apps/app_1/gen/go/v1/app1V1connect/**`
- `apps/app_1/gen/oas/openapi.yaml`

### Biz (Business Core)

service_1:
- `apps/app_1/services/service_1/internal/biz/biz.go`
- `apps/app_1/services/service_1/internal/biz/blog.go`
- `apps/app_1/services/service_1/internal/biz/author.go`

service_2:
- `apps/app_1/services/service_2/internal/biz/biz.go`
- `apps/app_1/services/service_2/internal/biz/reading.go`

Biz owns:
- domain entities
- UC orchestration
- repository interfaces
- transaction abstraction (`Transaction`)

### Data (Persistence + Integration)

service_1:
- `apps/app_1/services/service_1/internal/data/data.go`
- `apps/app_1/services/service_1/internal/data/ent.go`
- `apps/app_1/services/service_1/internal/data/cache.go`
- `apps/app_1/services/service_1/internal/data/blog.go`
- `apps/app_1/services/service_1/internal/data/author.go`

service_2:
- `apps/app_1/services/service_2/internal/data/data.go`
- `apps/app_1/services/service_2/internal/data/ent.go`
- `apps/app_1/services/service_2/internal/data/cache.go`
- `apps/app_1/services/service_2/internal/data/reading.go`
- `apps/app_1/services/service_2/internal/data/grpc_service_1.go`
- `apps/app_1/services/service_2/internal/data/connect_service_1.go`

Ent schema source:
- `apps/app_1/services/service_1/internal/data/ent/schema/blog.go`
- `apps/app_1/services/service_1/internal/data/ent/schema/author.go`
- `apps/app_1/services/service_2/internal/data/ent/schema/reading.go`

Generated Ent code:
- `apps/app_1/services/service_1/internal/data/ent/**`
- `apps/app_1/services/service_2/internal/data/ent/**`

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

Cross-service call path (reading flow):
```text
service_2 ReadingRP
  -> gRPC client to service_1 BlogService.GetBlog
  -> Connect client to service_1 BlogService.QueryBlog
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
  - `ReadingService`: API -> service -> biz -> data complete (implemented in `service_2`)

Generated artifacts to expect during normal workflow:
- `contracts/go/**`
- `apps/app_1/gen/**`
- `apps/app_1/services/service_1/internal/conf/conf.pb.go`
- `apps/app_1/services/service_1/cmd/app/wire_gen.go`
- `apps/app_1/services/service_1/internal/data/ent/**`
- `apps/app_1/services/service_1/internal/i18n/translations/v1.*.yaml`
- `apps/app_1/services/service_2/internal/conf/conf.pb.go`
- `apps/app_1/services/service_2/cmd/app/wire_gen.go`
- `apps/app_1/services/service_2/internal/data/ent/**`
- `apps/app_1/services/service_2/internal/i18n/translations/v1.*.yaml`

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
  - And/or `./nx run app_1_service_2:proto:conf` based on touched service
  - Then run affected generate target(s) if DI/types are impacted
- Change in `internal/data/ent/schema/**`:
  - Run affected generate target (`app_1_service_1:generate` and/or `app_1_service_2:generate`)
- Change in `cmd/app/wire.go` or provider sets:
  - Run affected generate target (`app_1_service_1:generate` and/or `app_1_service_2:generate`)
