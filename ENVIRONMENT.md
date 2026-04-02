# Environment and Runtime Guide

## 1) Prerequisites

- OS: macOS or Linux
- Go: `1.25.x` (follow `go.mod`)
- Bun: `>=1.3.0` (see `package.json`)
- Nx: use repository-local `./nx`
- Docker + Docker Compose

Optional toolchain bootstrap:
```bash
./nx run tools:go:init
```

---

## 2) Recommended Initialization Sequence

1. Start infra stack:
```bash
./nx run infra:docker:up
```

2. Generate shared contract Go code:
```bash
./nx run contracts:proto
```

3. Generate app API artifacts (Go + HTTP + gRPC + Connect + OpenAPI):
```bash
./nx run app_1_api:proto:api
```

4. Generate service config proto:
```bash
./nx run app_1_service_1:proto:conf
```

5. Run service generation chain (Wire + Ent + i18n + tidy):
```bash
./nx run app_1_service_1:generate
```

6. Launch service in development:
```bash
./nx run app_1_service_1:dev
```

---

## 2.1 Execution Profiles

### Quick profile (API-only or docs-only changes)
1. Run only required generation target(s).
2. Run `./nx run app_1_service_1:build` if service code is touched.

### Standard profile (default)
1. Run full generation path for affected modules.
2. Run `./nx run app_1_service_1:build`.
3. Run focused tests on touched packages.

### Full profile (release-sensitive or cross-cutting changes)
1. Run `contracts -> app_1_api -> service_1` generation chain.
2. Run `./nx run app_1_service_1:build`.
3. Run `go test ./...` and document known failures.

---

## 3) Command Cookbook

### Buf
```bash
./nx run tools:buf:dep
./nx run tools:buf:format
./nx run tools:buf:lint
```

### Build
```bash
./nx run app_1_service_1:build
```

### Generate new Ent entity schema scaffold
```bash
./nx run app_1_service_1:ent:new --args="Entity=<EntityName>"
```

### Common infra controls
```bash
./nx run infra:docker:ps
./nx run infra:docker:logs
./nx run infra:docker:down
./nx run infra:docker:clean
```

### Targeted infra start
```bash
./nx run infra:docker:postgres
./nx run infra:docker:redis
./nx run infra:docker:monitoring
./nx run infra:docker:minio
```

---

## 4) Test and Validation Loop

There is no unified Nx `test` target currently. Use:
```bash
go test ./...
```

Audit status on 2026-04-04:
- `go test ./...` fails at `tools/go-jwt` because of a vet finding:
  - `tools/go-jwt/main.go:162:4: fmt.Println arg list ends with redundant newline`

Recommended minimum validation for feature changes:
1. `./nx run app_1_service_1:generate` (if source definitions changed)
2. `./nx run app_1_service_1:build`
3. Focused `go test` on touched packages

---

## 4.1 Failure Handling Rules

- If `generate` fails, fix source definitions first (proto/schema/wire), do not patch generated output manually.
- If `build` fails after generation, fix hand-written source files before rerunning generation.
- If tests fail in untouched legacy packages, record them as pre-existing only after verifying your change did not widen the failure surface.

---

## 5) Logs and Observability

### Service logs
- Dev logs are streamed in the terminal from `./nx run app_1_service_1:dev`.
- File logging path (when enabled): `apps/app_1/services/service_1/logs/service_1.log`.

### Infra logs
```bash
./nx run infra:docker:logs
```

### Default endpoints (from current `configs/config.yaml`)
- HTTP API: `http://localhost:11000`
- gRPC: `localhost:12000`
- Connect API: `http://localhost:13000`
- Ops server: `http://localhost:14000`
- Metrics: `http://localhost:14000/metrics`
- pprof index: `http://localhost:14000/debug/pprof/`
- Jaeger UI: `http://localhost:16686`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (`admin/admin`)
- Loki: `http://localhost:3100`

---

## 6) Agent Execution Rules

- Always inspect the owning `project.json` before running commands.
- Prefer Nx targets over ad-hoc direct commands.
- After proto/schema/wire changes, regenerate before coding around compile errors.
- Keep an execution trail in PR/task notes: generation, build, test, runtime checks.
- For every task handoff, include: command list, pass/fail status, and unresolved blockers.
