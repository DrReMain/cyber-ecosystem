# Cyber Ecosystem

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev) [![Node.js](https://img.shields.io/badge/Node.js-%E2%89%A524-339933?logo=nodedotjs&logoColor=white)](https://nodejs.org) [![pnpm](https://img.shields.io/badge/pnpm-11-F69220?logo=pnpm&logoColor=white)](https://pnpm.io) [![Nx](https://img.shields.io/badge/Nx-monorepo-143055?logo=nx&logoColor=white)](https://nx.dev) [![License: MIT](https://img.shields.io/github/license/DrReMain/cyber-ecosystem?color=blue)](./LICENSE) [![Last Commit](https://img.shields.io/github/last-commit/DrReMain/cyber-ecosystem)](https://github.com/DrReMain/cyber-ecosystem/commits)

> A contract-first full-stack monorepo for networked applications: a Go (Kratos v3) backend, Connect-RPC contracts, real-time & media infrastructure, and a TanStack Start web client.

## What it is

Cyber Ecosystem is an open-source full-stack monorepo for building networked applications. It is intended to cover a broad range of scenarios — CRUD, realtime data (live streaming, media, online meetings), and IoT — rather than being scoped to a single one.

Protobuf contracts in `proto/` are the single source of truth; the Go server, TypeScript client, and shared libraries are derived from them.

## How errors are caught early

The codebase leans on compile-time checks and a fixed toolchain so that mistakes — including those introduced by AI-assisted editing — surface at build or generation time rather than at runtime:

- **Strict types end to end** — ent schemas on the server; TanStack Start's strict TypeScript consuming Connect-RPC types generated from proto on the client.
- **Protobuf as the contract** — definitions drive the Go server, TypeScript client, and OpenAPI; mismatched fields fail the build.
- **Locked toolchain** — Nx runs all generation / build / test steps; generated code in `gen/` is not hand-edited.

## Highlights

- **Broad scope.** CRUD through realtime (streaming, media, meetings) and IoT are all in scope.
- **Type-safe end to end.** Contracts flow from proto through Connect-RPC into strict TypeScript; types hold on both server and client.
- **Multi-transport.** Each service is served over gRPC, HTTP, and Connect from one contract — choose per client.
- **Multi-client.** The web client (TanStack Start, SSR) is the current one; React Native, Flutter, and native iOS/Android clients are intended to consume the same contracts.
- **Capability packs.** Infrastructure concerns are reusable Go packages in `shared-go/` (`cache`, `orm`, `storage`, `mq`, `kratos`, …), shared across services.
- **Self-hosted infra.** SeaweedFS (object storage), Redis (cache), PostgreSQL, and OpenTelemetry observability (→ SigNoz) run in-cluster; managed equivalents can be substituted.
- **Current-generation stack.** Go 1.25, TypeScript 7, Nx 23, Kratos v3, Connect-RPC, Atlas, TanStack Start.
- **Realtime & media.** Centrifugo (realtime) and LiveKit (media) sit on top of Kratos v3's native streaming; NATS is the IoT / event transport.
- **Kubernetes deployment.** Kustomize base + environment overlays behind a Traefik edge (host routing + TLS).
- **Single toolchain.** Nx orchestrates build, generation, test, and lint across Go and TypeScript.

## Tech stack

| Layer | Stack |
|---|---|
| Backend | Go · Kratos v3 · Connect-RPC · ent / Atlas migrations |
| Realtime / Media | Centrifugo · LiveKit · NATS |
| Data & Infra | PostgreSQL · Redis · SeaweedFS |
| Observability | OpenTelemetry · SigNoz |
| Frontend | TypeScript · TanStack Start (SSR) · Ant Design · Tailwind CSS · Connect-RPC (web) · Paraglide i18n |
| Tooling | Nx monorepo · Buf · pnpm · Biome |

## Repository layout

```
proto/          Protobuf contracts (source of truth) + errors + generation scripts
gen/            Generated Go / TypeScript — derived, do not edit
app/
  services/core   Go core service (Kratos v3)
  clients/admin   Web client — TanStack Start + Ant Design (admin shell today; mobile/native clients planned)
shared-go/      Reusable Go capability packs: cache · orm · storage · mq · kratos · helper · utils
shared-ts/      Shared TypeScript packages (Ant Design layer)
deploy/         Kubernetes manifests (kustomize base + overlays) + Postgres bootstrap
tools/          Dev tooling — env init, Go lint/test/format (Nx targets)
docs/           Per-area conventions (e.g. kratos/)
```

## Quickstart

> Everything runs through **Nx**; each target is declared in the owning project's `project.json`. Cluster provisioning is out of scope (see the note below).

### Prerequisites

- **Go** 1.25+, **Node.js** 24+, **pnpm** 11+
- **[Atlas](https://atlasgo.io)** CLI — for database migrations
- A **Kubernetes / k3s cluster** (k3s is the primary target; non-k8s setups are possible but unsupported — see the note below)
- Infra reachable at the addresses in `app/services/core/configs/config.yaml` (defaults: Postgres `localhost:5432`, Redis `localhost:6379`; override via env or adjust the config to match your cluster)

### 1. Initialize the environment

Installs the toolchain (grpcurl, buf) and language dependencies (Go modules, pnpm):

```bash
./nx run tools:init
```

### 2. Generate code from contracts

```bash
./nx run proto:generate    # Go + TypeScript + OpenAPI from proto
./nx run core:generate     # ent + wire + go mod tidy
```

### 3. Start the infra this repo manages

Applies the Postgres / Redis / SeaweedFS / NATS manifests to your cluster (observability / Centrifugo / LiveKit are opt-in):

```bash
./nx run deploy:start
./nx run deploy:status     # verify pods are ready
```

### 4. Create the databases (the only manual step)

Migrations create tables, not databases — create them once:

```bash
psql -h localhost -U postgres -c "CREATE DATABASE core;"
psql -h localhost -U postgres -c "CREATE DATABASE mq;"   # only if using the Postgres MQ backend (see config.yaml)
```

### 5. Apply migrations

```bash
./nx run core:migrate:apply
```

### 6. Run it

```bash
./nx run core:dev                      # Go backend (Kratos v3)
./nx run @cyber-ecosystem/admin:dev    # Web client (Vite)
```

### Common tasks

- Lint / test / format Go → `./nx run tools:go:lint` · `tools:go:test` · `tools:go:format`.
- Lint proto → `./nx run proto:lint`.
- Edit contracts → `./nx run proto:generate` (review the `gen/` diff).
- Change ent schemas → `./nx run core:generate:ent`.
- Create a migration → `NAME=add_x ./nx run core:migrate:diff`.
- Tear the managed infra down → `./nx run deploy:stop`.

<details>
<summary><strong>Cluster environments (k3s first; others unsupported)</strong></summary>

The `deploy/` manifests target **k3s / Kubernetes**; `deploy:start` applies them. Provisioning the cluster itself is out of scope — bring your own (k3d for local dev, a self-hosted k3s node, or any managed k8s).

Non-k3s/k8s environments (plain Docker, other orchestrators, or running infra directly on a host) can also be used for development or deployment, but this repo provides no manifests or tooling for them. In that case you stand up the equivalent services (Postgres, Redis, SeaweedFS, NATS, …) yourself and make them reachable at the addresses configured in `app/services/core/configs/config.yaml`.

</details>

## Status

Active development on `main` (Kratos v3 line). The foundation is in place: 7 capability packs in `shared-go/`, 6 infra stacks (db / storage / mq / realtime / media / observability) deployable to a cluster, full Kratos layering, and contract-first generation across Go and TypeScript.

The business layer is a sample skeleton — `User` / `Resource` / `Transfer` services over a single domain entity — included as a reference for where domain code goes. It is not a finished product.

## License

[MIT](./LICENSE).
