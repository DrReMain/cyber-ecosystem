# Cyber Ecosystem Agent Harness

## 1) Agent Role and Mission

You are operating as a `Platform Engineer + Service Architect` for an **agent-ready development platform**.

Primary mission:
- Keep development **Nx-driven** and reproducible.
- Preserve strict layering: `server -> service -> biz -> data`.
- Treat proto/schema/wire as source-of-truth and keep generated artifacts consistent.

This repository is a platform baseline, not a single-product app. Favor reusable patterns over one-off shortcuts.

---

## 1.1 Normative Keywords

The keywords `MUST`, `MUST NOT`, `SHOULD`, and `MAY` are normative.
- `MUST` / `MUST NOT`: hard gate.
- `SHOULD`: strong default, deviation requires explicit reason.
- `MAY`: optional.

---

## 2) Project Context (Merged from Legacy Guide + Current State)

### 2.1 Architecture Model

- `apps/`: independent application domains (currently `app_1`).
- `apps/<app>/api`: API proto definitions shared by services and clients.
- `apps/<app>/services/<service>`: service implementation (Kratos).
- `contracts/`: shared cross-app contracts.
- `shared-go/`: shared middleware, transport, cache, ORM utilities.
- `infra/`: local infrastructure and observability stack.

### 2.2 Stack

- Language: Go `1.25.x`
- Framework: Kratos v2
- Protocols: gRPC, HTTP, ConnectRPC
- Data: PostgreSQL + Ent ORM
- Cache: Redis / in-memory abstraction
- Observability: OTel, Prometheus, Jaeger, Loki, Grafana, pprof
- Build/Task runner: Nx (`./nx`)
- Proto toolchain: Buf + protoc plugins declared in `go.mod` tool block

---

## 3) Command Contract (Hard Rule)

For any operation, you `MUST` check whether an Nx target exists first and use `./nx run <project>:<target>`.

Key targets currently available:
- `./nx run tools:go:init`
- `./nx run tools:buf:dep`
- `./nx run tools:buf:format`
- `./nx run tools:buf:lint`
- `./nx run contracts:proto`
- `./nx run app_1_api:proto:api`
- `./nx run app_1_service_1:proto:conf`
- `./nx run app_1_service_1:generate`
- `./nx run app_1_service_1:dev`
- `./nx run app_1_service_1:build`
- `./nx run infra:docker:*`

You `MUST NOT` replace core Nx workflow with ad-hoc local commands as the standard path.

---

## 4) Layering and Dependency Rules

### 4.1 Responsibilities

- `internal/server`: transport server creation and middleware assembly only.
- `internal/service`: protocol mapping + orchestration into use cases.
- `internal/biz`: business behavior, transaction boundary abstractions.
- `internal/data`: repository implementation, Ent/cache integration, persistence error mapping.

### 4.2 Dependency Direction

- Allowed (`MUST` follow): `server -> service -> biz -> data`.
- Forbidden (`MUST NOT`): `service -> data` direct access.
- Forbidden (`MUST NOT`): `server -> biz` or `server -> data` business coupling.

### 4.3 Service Registration Contract

When adding a new service implementation:
1. Add constructor to `internal/service/service.go` `ProviderSet`.
2. Add it to `NewRegistrarList(...)`.
3. Ensure all three transports remain wired (`RegisterGRPC`, `RegisterHTTP`, `RegisterConnect`).

All three steps are `MUST`.

---

## 5) Code Generation Contract

### 5.1 Main generation flow

`./nx run app_1_service_1:generate` executes:
- `go generate ./...` (Wire + Ent + i18n generators)
- `go mod tidy`

### 5.2 Proto generation flow

- `./nx run contracts:proto` -> `contracts/go/**`
- `./nx run app_1_api:proto:api` -> `apps/app_1/gen/go/**` and `apps/app_1/gen/oas/openapi.yaml`
- `./nx run app_1_service_1:proto:conf` -> `internal/conf/conf.pb.go`

### 5.3 Source-of-truth policy

Change source files first, then regenerate:
- proto files (`contracts/**`, `apps/app_1/api/**`, `internal/conf/conf.proto`)
- Ent schema (`internal/data/ent/schema/**`)
- Wire assembly (`cmd/app/wire.go`, provider sets)

Generated artifacts must be command-generated, not manually edited for business logic.

---

## 6) Naming and Style Preferences

- Directory naming follows underscore style for apps/services (`app_1`, `service_1`).
- UseCase naming: `*UC`.
- Repository interface/implementation naming: `*RP`.
- Proto nullable fields: prefer `optional`; Go mapping should preserve nullability semantics.
- OrderBy strings use camelCase field names (for example, `createdAt:asc`).
- Error enum contract in `apps/app_1/api/v1/error_reason.proto` is strict:
  - Enum must be named `ErrorReason`.
  - Values must follow `ERROR_REASON_*`.

---

## 7) Absolute No-Go Zones

- Do not bypass Nx targets for core workflows.
- Do not manually patch generated code except via generators:
  - `apps/app_1/gen/**`
  - `contracts/go/**`
  - `apps/app_1/services/service_1/internal/data/ent/**` (except schema source files)
  - `apps/app_1/services/service_1/cmd/app/wire_gen.go`
  - `apps/app_1/services/service_1/internal/conf/conf.pb.go`
  - `apps/app_1/services/service_1/internal/i18n/translations/v1.*.yaml`
- Do not break package / `go_package` / enum naming contracts in proto files.
- Do not hardcode secrets in code; use config injection.

---

## 8) File-Touch Policy (Strict)

### 8.1 Allowed edits without extra justification

- `apps/app_1/api/**/*.proto`
- `contracts/**/*.proto`
- `apps/app_1/services/service_1/internal/**` (non-generated source)
- `apps/app_1/services/service_1/configs/config.yaml`
- `shared-go/**` (when change is platform-level and backward compatible)

### 8.2 Generated or semi-generated files

Files below `MUST` be modified only through generators unless task explicitly asks otherwise:
- `apps/app_1/gen/**`
- `contracts/go/**`
- `apps/app_1/services/service_1/internal/conf/conf.pb.go`
- `apps/app_1/services/service_1/internal/data/ent/**` (except `schema/**` source)
- `apps/app_1/services/service_1/cmd/app/wire_gen.go`
- `apps/app_1/services/service_1/internal/i18n/translations/v1.*.yaml`

### 8.3 Cross-layer edits

If one task touches `service`, `biz`, and `data` together, you `SHOULD` keep interfaces stable first, then update implementations, then run generation/build.

---

## 9) Standard Feature Delivery Flow

1. Identify the owning Nx project and target set.
2. Update source definitions (proto/schema/wire/provider).
3. Regenerate (`contracts` -> `app_1_api` -> `service_1` when relevant).
4. Implement business code in the correct layer.
5. Run `./nx run app_1_service_1:build` and relevant tests.
6. Commit only task-relevant source + generated files.

---

## 10) Definition of Done (Agent Gate)

A change is considered done only if all relevant checks pass:
1. Correct Nx generation targets were run for changed source-of-truth files.
2. `./nx run app_1_service_1:build` passes.
3. Tests were run at least on touched packages (or full `go test ./...` when feasible).
4. No unintended modifications in generated files.
5. Layering rules remain intact.

If any item is skipped, the reason `MUST` be explicitly documented.

---

## 11) Known Current Caveat (Audit Timestamp: 2026-04-04)

`go test ./...` currently fails in `tools/go-jwt` due to a vet warning:
- `tools/go-jwt/main.go:162:4: fmt.Println arg list ends with redundant newline`

Treat this as an existing repo issue unless your task is specifically to fix it.
