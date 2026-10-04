# Cyber Ecosystem

[English](./README.md) · [简体中文](./README.zh-CN.md)

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev) [![Node.js](https://img.shields.io/badge/Node.js-%E2%89%A524-339933?logo=nodedotjs&logoColor=white)](https://nodejs.org) [![pnpm](https://img.shields.io/badge/pnpm-11-F69220?logo=pnpm&logoColor=white)](https://pnpm.io) [![Nx](https://img.shields.io/badge/Nx-monorepo-143055?logo=nx&logoColor=white)](https://nx.dev) [![License: MIT](https://img.shields.io/github/license/DrReMain/cyber-ecosystem?color=blue)](./LICENSE) [![Last Commit](https://img.shields.io/github/last-commit/DrReMain/cyber-ecosystem)](https://github.com/DrReMain/cyber-ecosystem/commits)

> A contract-first, AI-native-oriented full-stack skeleton for long-lived products: a Go/Kratos base service, a second-service reference implementation, Connect-RPC contracts, realtime and media infrastructure, and a TanStack Start web client.

## What it is

Cyber Ecosystem is a **product-development skeleton**, not a vertical SaaS product. It exists so a team can fork long-lived applications from one governed base instead of repeatedly assembling — and then diverging — identity, authorization, transport, file, realtime, and admin patterns.

The repository currently contains three coordinated pieces:

- **`system`** — the mandatory base service. It owns authentication and sessions, users, organization structure, roles, the RBAC / ABAC / data-scope authorization engine, the operation catalog, files, audit, and the cross-service identity/authorization surface.
- **`agent`** — the second-service reference implementation. It owns per-user agent configuration and chat aggregates, consumes identity and authorization from `system` through introspection, and demonstrates how to add an application-facing AI workflow without duplicating IAM.
- **`admin`** — the TanStack Start management/workbench client, including cookie-custody SSR, permission-derived routing, typed errors, localization, file upload, and streaming chat consumption.

The skeleton supplies the shared product base; business behavior belongs in product services.

## Console preview

| Sign-in | Agent initial | Policies (Arabic, RTL) |
| :---: | :---: | :---: |
| ![Sign-in](docs/screens/login.png) | ![Agent initial](docs/screens/agents-chat-init.png) | ![Policies](docs/screens/policies-rtl.png) |

| Agent conversation | Audit logs | Files & preferences |
| :---: | :---: | :---: |
| ![Agent conversation](docs/screens/agents-chat-session.png) | ![Audit logs](docs/screens/audit.png) | ![Files and preferences](docs/screens/files.png) |

Sign-in, Agent initial, and Policies merge their light and dark renderings along the same diagonal. Agent conversation, Audit logs, and Files & preferences are single live-interface screenshots.

## Architecture

### Contract-first control plane

`proto/` is the single source of truth. One contract derives:

- Go gRPC and HTTP bindings;
- Connect server registration;
- the Connect TypeScript client;
- an OpenAPI document;
- per-package operation manifests for the authorization catalog;
- the operation names consumed by backend middleware and frontend routing.

Method options also declare access audience, builtin status, and datascope applicability. A forgotten access annotation is rejected by a deny-by-default guard rather than silently exposed.

This turns the contract into a control plane: adding a service operation also makes it visible to role management, runtime authorization, diagnostics, and the admin route model.

### Clean architecture and DDD

Every Go service follows the same application boundary:

```text
proto / Connect / gRPC / HTTP
        ↓
service.go    — transport adapter; proto ↔ domain mapping
biz.go        — use cases, domain objects, and ports
data.go       — repository adapters
platform      — service-owned infrastructure facade
```

More precisely:

- **Transport is an adapter.** `service.go` keeps RPC handlers thin and maps generated messages to domain objects.
- **Use cases depend on ports.** A module depends on narrow interfaces such as `TokenRP`, `UserRP`, or `AuthzRP`; it does not depend directly on Redis, Ent, S3, or another service client.
- **Repositories are adapters.** `data.go` implements the ports with Ent and platform capabilities.
- **Infrastructure belongs to the service boundary.** Wire composes the platform, providers, cleanup functions, and servers without leaking infrastructure concerns into use cases.
- **Boundaries follow DDD.** `system` and `agent` are separate bounded contexts. `system.User` is the authentication identity; an application service owns its own application aggregates and refers to users by ID. It does not copy or extend the system user table.
- **Cross-service reads use remote adapters.** A business service consumes identity and authorization through the `system` RPC surface; it does not query another service's database.

The result is a deliberate dependency rule:

```text
transport → use case → port ← adapter
                       ↓
                  platform capability
```

Business modules can be tested with port fakes, replaced behind interfaces, and eventually extracted without rewriting their use cases.

### Authorization model

Roles carry grants: an operation pattern, a data-scope kind, and optional constraint policies.

- **Operation patterns** support exact RPC names, service wildcards, and the global `/*` administrator pattern.
- **Data scopes** support all, self, and department-tree narrowing.
- **Constraint policies** are evaluated with AND semantics; time-window and calendar policies are currently implemented.
- **Evaluation fails closed** when a policy kind is unknown, attribute resolution fails, or policy evaluation errors.
- **Decisions are inspectable.** The diagnostics view explains the roles, grants, scopes, and policy states behind an allow or deny.
- **Changes are versioned.** Authorization tables compile into an in-memory snapshot; version bumps and notifications rebuild replicas, with periodic reconciliation as a safety net.

### AI-native direction

AI-native has two layers in this skeleton:

- **Development time.** Repository rules, area conventions, generated artifacts, and Nx targets give humans and coding agents the same executable boundaries. Fast AI-generated changes should still surface contract, permission, generation, and documentation drift before runtime.
- **Runtime.** The current `agent` service is a real vertical slice: per-user OpenAI-compatible providers, model listing, chat sessions, and server-streamed responses. The platform direction is to model agents, credentials, tools, tasks, and retrieval as first-class principals and IAM operations rather than treating an agent as a chat-only feature. Those capabilities are directional and are not yet current platform guarantees.

### Capability families

`shared-go/capability` packages infrastructure as self-contained families:

- `cache`: KV, hash, list, set, sorted set, counter, lock, rate limiter, pub/sub, and session interfaces, with a Redis backend.
- `storage`: object, list, presign, multipart, bucket, per-bucket views, and operational limits, with an S3-compatible backend.
- `mq`: durable at-least-once producer/consumer semantics, retry and DLQ behavior, with NATS and PostgreSQL backends.

A capability family depends only on the standard library, its third-party providers, and its own interface root. It never imports service code, so a family can be copied or extracted as a unit.

### Web client

The `admin` client is feature-sliced along a one-way dependency chain:

```text
libs → stores → domains → services → features → routes
```

It provides:

- HttpOnly cookie custody on the server;
- SSR loaders and server functions;
- permission-derived menus and route guards;
- typed generated Connect clients;
- unified error classification and feedback;
- light/dark theming, five locales, and RTL support;
- keep-alive tabs, breadcrumbs, and a workbench layout;
- server-streaming chat consumption and resumable direct uploads.

## Why these technologies

| Technology | Advantage in this skeleton |
|---|---|
| **Protobuf / Buf** | One typed contract across Go, TypeScript, OpenAPI, authorization, and generated tooling. Reduces hand-synchronized surface area and contract drift. |
| **Go / Kratos** | Compiled performance, lightweight concurrency, and composable transport/middleware architecture. The service layout stays stable as modules grow. |
| **Connect-RPC** | A modern RPC boundary across browser, server, gRPC, and HTTP while preserving protobuf typing and efficient serialization. |
| **ent / Atlas** | Typed schema and query generation with versioned migrations, reducing raw-SQL drift and making data ownership explicit. |
| **TanStack Start** | SSR, nested routing, typed routing context, and server functions provide a strong application shell without abandoning SPA-style navigation. |
| **React Query / Connect Query** | Request de-duplication, cache lifetimes, mutation state, and streaming integration keep data flow predictable. |
| **Ant Design + Tailwind** | Dense administrative UI components plus token-based styling and dark/RTL variants. |
| **Redis / S3-compatible storage / NATS or PG-MQ** | The capability seams support local development, self-hosting, and managed-provider substitution. |
| **OpenTelemetry / SigNoz** | Traces, metrics, logs, and slow-query hooks are built into service wiring rather than bolted on later. |
| **Nx** | Declared workflows keep generation, migration, tests, builds, and deployment composable and reproducible. |

## Performance profile

This is an architectural profile, not a benchmark claim. Actual results depend on schemas, indexes, providers, deployment, and workload.

### Backend

- **Go and Kratos** provide a compiled, concurrently efficient service runtime with relatively low per-request overhead.
- **Connect with protobuf** avoids hand-parsed JSON on internal RPC paths and preserves compact binary encoding where supported.
- **Authorization uses an in-memory compiled snapshot**, so normal decisions do not query the role, permission, policy, and binding tables on every request.
- **Ent generates typed SQL**, while connection pooling, explicit indexes, and Atlas migrations keep the database access path inspectable.
- **Presigned uploads and downloads** move large object bytes directly between the browser and S3-compatible storage; the app server coordinates metadata and confirmation instead of proxying every byte.
- **Audit publishing is asynchronous**, decoupling the request path from MQ persistence.
- **Streaming chat** emits deltas as the model produces them instead of waiting for the full answer.

Likely bottlenecks are external and workload-specific: model providers, database queries, object storage, cross-service introspection, and policy/data volume. Those are also the surfaces with explicit seams for measurement and optimization.

### Frontend

- **SSR** improves first meaningful render and keeps session cookies server-custodied.
- **Route-level code splitting** limits the JavaScript and feature modules loaded for the current page.
- **React Query caching and de-duplication** reduce repeated requests and make mutation invalidation explicit.
- **Generated Connect clients** avoid per-call hand serialization and preserve compile-time operation typing.
- **Direct file transfer** avoids pushing large files through the admin server.
- **Streaming UI state** presents model and transfer progress without waiting for a final response.

The admin stack therefore favors predictable interaction under I/O-bound work: loading data, uploading files, and consuming streams. Rendering-heavy grids should still be paginated, indexed, and selectively virtualized by the product.

## Repository layout

```text
proto/             Protobuf contracts, extensions, generation scripts, and catalog generator
gen/               Generated Go, Connect TypeScript, OpenAPI, and operation catalogs
app/
  services/system  Mandatory base service: IAM, authz, resource catalog, files, audit, introspection
  services/agent   Second-service reference: per-user agent config and streaming chat
  clients/admin    TanStack Start management/workbench client
shared-go/         Reusable Go packages: capability families, orm, kratos, codegen, helpers
shared-ts/         Shared TypeScript packages: error, antd, theme, store, cookie, progress
deploy/            Compose profiles, Traefik edge, and migration wiring
tools/             Repository bootstrap and Go tooling targets
docs/              Engineering conventions and screenshots
```

## Quickstart

Everything with a declared workflow runs through Nx.

### Prerequisites

- Go 1.26+
- Node.js 24.15+
- pnpm 11+
- Docker with Compose v2

### 1. Initialize tooling and dependencies

```bash
./nx run tools:init
```

### 2. Generate derived code

```bash
./nx run system:generate
./nx run agent:generate
```

Review the `gen/` diff. Generated files are never edited directly.

### 3. Start local infrastructure

```bash
./nx run deploy:start
./nx run deploy:status
```

The default profile raises PostgreSQL, Redis, SeaweedFS, and NATS. The Postgres container creates the `system`, `agent`, `mq`, and `atlas_dev` databases.

### 4. Apply migrations

```bash
./nx run system:migrate:apply
./nx run agent:migrate:apply
```

### 5. Run the services

Use one shell per process:

```bash
./nx run system:dev
./nx run agent:dev
./nx admin:dev
```

The development proxy expects `system` on `localhost:13001` and `agent` on `localhost:13002`. A user can configure an OpenAI-compatible provider from the admin profile; the API key is encrypted server-side and never returned by an API.

The checked-in configuration contains development-only defaults, including seeded administrators and the agent master key. Change them before any shared or internet-reachable deployment.

### Common tasks

- Proto lint / generation: `./nx run proto:lint` · `proto:generate`
- Go lint / tests / format: `./nx run tools:go:lint` · `tools:go:test` · `tools:go:format`
- Ent generation: `./nx run system:generate:ent` or `agent:generate:ent`
- Migration diff: `NAME=add_x ./nx run system:migrate:diff` or `agent:migrate:diff`
- Optional stacks: `deploy:realtime:start` · `deploy:media:start` · `deploy:observability:start`
- Edge stack: `deploy:pre:start` or `deploy:pre:full:start`
- Teardown: `deploy:stop` · `deploy:reset`

## Using it as a product base

The intended workflow is:

1. Start from a released skeleton snapshot or tag.
2. Keep a private product repository or isolated product branch.
3. Add services under `app/services/<name>` and contracts under `proto/cyber/<name>/v1`.
4. Keep business aggregates out of `system`.
5. Consume identity and authorization through the `system` surface.
6. Periodically merge or cherry-pick a released skeleton tag.
7. Do not merge product-specific behavior back into the skeleton. Promote a capability only when it has a second real consumer and carries tests, observability, and a copyable pattern.

## Open-source model

This GitHub repository is a **curated public snapshot** of a privately maintained skeleton. Development history, active roadmap, and implementation fronts remain private.

Public issues are welcome. Public pull requests are not the primary development path; materially useful changes are applied to the private upstream and released in a later snapshot.

## Status

Active development; the foundation is usable but not finished.

Currently implemented:

- `system`: authentication, sessions, users, organization, roles, authorization, resource catalog, files, audit, and cross-service introspection.
- `agent`: per-user provider configuration, model listing, chat sessions, and server-streamed responses.
- `admin`: permission-derived management/workbench UI, localization, file workflows, and streaming chat.

Future platform work includes runtime agent principals, API-key credentials, tool authorization, task orchestration, retrieval integration, and additional vertical slices.

## Documentation

Area conventions live in:

- [`docs/conventions/proto/CONVENTIONS.md`](./docs/conventions/proto/CONVENTIONS.md)
- [`docs/conventions/kratos/CONVENTIONS.md`](./docs/conventions/kratos/CONVENTIONS.md)
- [`docs/conventions/tanstack/CONVENTIONS.md`](./docs/conventions/tanstack/CONVENTIONS.md)
- [`docs/conventions/deploy/CONVENTIONS.md`](./docs/conventions/deploy/CONVENTIONS.md)

## License

[MIT](./LICENSE)
