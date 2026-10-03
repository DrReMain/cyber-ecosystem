# Cyber Ecosystem

## 1) Scope

Root-level rules only. Applies to the entire monorepo and to every contributor — human or coding agent.

`MUST` / `MUST NOT` = hard requirement. `SHOULD` = strong default. `MAY` = optional.

Stack-specific and service-specific details belong in local documents under the owning directory, not here.

This file is the canonical root-rules source; `CLAUDE.md` is a Claude Code import shell pointing here.

---

## 2) Repository Overview

Monorepo for the Cyber Ecosystem platform — a **generic development skeleton**, not a vertical product: one shared base, business apps fork and grow on top. Protobuf contracts (`proto/`) are the source of truth; generated code (`gen/`) is derived output — never edit directly.

| Path | Holds |
|---|---|
| `proto/` | Protobuf contracts — the source of truth |
| `gen/` | Generated code (Go, connect-ts, OpenAPI) — derived, never hand-edited |
| `app/services/<name>/` | Kratos Go services; `system` = the mandatory base service (IAM) |
| `app/clients/<name>/` | TanStack Start web clients (`admin`) |
| `shared-go/` | Cross-service Go capabilities (transport, security, capability facades) |
| `shared-ts/` | Cross-client TS capabilities (error, antd, theme, store, cookie, router-progress, storybook) |
| `deploy/` | Compose stacks, traefik edge, migration wiring |
| `tools/` | Repo bootstrap & tooling Nx targets |
| `docs/` | Documentation — three lanes (§7) |

---

## 3) Running Commands — MUST Use Nx

`MUST` use Nx for all workflows that have declared targets:

```bash
./nx run <project>:<target>
```

**How to find targets:** Read the project's `project.json` — every valid target is declared there. Do not assume a target exists. For JS/TS packages, Nx also picks up **every script in `package.json` as a runnable target** (`./nx run <pkg>:<script>`) — check `package.json` scripts before concluding a workflow has no target.

Recurring build, test, lint, generation, dev, and automation workflows `SHOULD` be exposed through Nx targets. If a needed workflow has no target yet, add one to `project.json` instead of running ad-hoc commands.

---

## 4) Hard Rules

**DO NOT bypass Nx** with direct toolchain commands (e.g. `buf generate`, `go generate`, `go build`) for workflows that have Nx targets.

**DO NOT manually edit generated files.** Fix the source or generator, then regenerate via the owning Nx target.

**DO NOT introduce cross-service dependencies.** Move shared capability into `shared-go/` instead.

**DO NOT couple `shared-go/capability` to the rest of the repo.** A capability package depends only on stdlib, its third-party providers, and its own capability family (the parent interface package — `mq/pg` → `mq`); never on `utils`, `kratos`, `orm`, or any other shared-go package — a family copies out as a unit. In-package substitutes: `encoding/json` directly; explicit bounds guards (clamp above `math.MaxInt32`, then convert) instead of `ConvNum`.

**DO NOT hardcode secrets or environment-specific credentials.**

---

## 5) Source-First Workflow

1. Edit the source definition (`.proto`, generator config, etc.)
2. Run the relevant Nx generation target
3. Review the generated diff — exclude unintended churn
4. If output changes unexpectedly, fix the source or generation flow

For cross-project changes: stabilize shared contracts first → update implementations → regenerate.

---

## 6) Validation — Definition of Done

Before closing any change:

1. Relevant Nx generation targets were run when required.
2. Touched projects were validated with declared Nx targets.
3. If a needed validation step has no Nx target yet, note the gap explicitly.
4. Generated outputs were reviewed; unrelated churn was excluded.
5. Any skipped step or pre-existing failure is called out.
6. Paradigm or mechanism changes: the owning `docs/conventions/<area>/CONVENTIONS.md` is updated and affected skill pointers are verified — `./nx run docs:check` is green.

---

## 7) Documentation — Three Lanes

Knowledge lives in three lanes; each kind of knowledge has exactly one home — durable rules (`docs/conventions/<area>/CONVENTIONS.md`, EN), procedural skill dispatchers (`.claude/skills/<name>/SKILL.md`, EN; trigger + pointers, never content), and a temporary, fully isolated working lane for active fronts (defined only in `docs/README.md`). Full governance law (admission, promotion, backfill, pointer discipline, instance-state, drift checks): `docs/README.md`.

Lane rules:

- `MUST` read the relevant `CONVENTIONS.md` before writing code in an area, and follow it exactly (structure, naming, ordering, comments, dependency rules).
- `MUST NOT` duplicate content across lanes or into skills — pointers only. Rule text and step text live in docs; a skill body holds a trigger, pointers, and completion pointers.
- `MUST NOT` reference roadmap content from here, `CLAUDE.md`, skills, or conventions — the working lane is temporary and fully isolated (law: `docs/README.md` §2.7); active fronts are introduced by the user, not by root docs.
- Skills auto-trigger by description (Claude Code); any agent or human can read `.claude/skills/*/SKILL.md` as plain markdown.

Area index — read before writing code in the area:

| Area | Doc | Read when |
|---|---|---|
| proto (contracts) | `docs/conventions/proto/CONVENTIONS.md` | touching any `.proto`, or code that consumes `gen/` |
| kratos (Go services) | `docs/conventions/kratos/CONVENTIONS.md` | writing Go under `app/services/*` or `shared-go/` |
| tanstack (web clients) | `docs/conventions/tanstack/CONVENTIONS.md` | writing code under `app/clients/*` |
| deploy (containers & pipelines) | `docs/conventions/deploy/CONVENTIONS.md` | touching `deploy/`, Dockerfiles, compose, edge, migrations |

Adding a new tech stack: create `docs/conventions/<area>/CONVENTIONS.md` **with the first real code of that stack** (not before), and add it to this index in the same change. Areas are tech stacks only — topic or domain rules live in the owning stack doc or here (§8), never in their own area directory.

---

## 8) Multi-app composition

One shared base, many business applications:

- **`system` is the mandatory base service** every application deploys: authentication, user identity, org (dept), roles, authz engine, resource catalog. Each application adds services under `app/services/<name>` with contracts under `proto/cyber/<name>/v1`, owning its aggregates end-to-end (schema, migrations, biz, RPC).
- Cross-service reads at runtime go through the owning service's RPCs — from clients, or behind a `<Remote>RP` ACL adapter in the consuming service. Business services `MUST NOT` grow identity/auth concerns; those live in `system` and are shared by all apps.
- **User split**: `system.User` is the authentication identity core only (credentials, org membership, role bindings). App-specific person data is the app's own domain aggregate keyed by `user_id` (plain id field), named by its domain concept (`Delegate`, `Operator`) — never `UserExt`-style field bags; own proto, own schema, logical reference only. A field enters `system.User` only when ≥2 applications genuinely need it. Reads hydrate identity via `system` RPCs — a missing profile degrades rendering, never breaks the row; each aggregate keeps a single write entrance. `system` cascades only what it owns; business profiles tolerate orphaning; cross-service invalidation events appear only on a second real consumer.
- **Pluggable system surface**, cheapest tier first: permission gating by default (hide-by-default — an ungranted operation never surfaces) → config-level module registry when the API itself must be absent → code fork only when isolation physically forces it, and then as a delivery action, not a development structure.
- **Composition & branching**: applications coexist as directories in one trunk; an application is a composition unit — deploy profiles select the service set, the client composes menu and skin per app. Per-app long-lived branches `SHOULD NOT` (every base fix would need cherry-picking across them); fork per app only at isolation-forced delivery time, as late as possible.
- `shared-go` / `shared-ts` carry capabilities, never business aggregates; a pattern is promoted to shared only on its second real consumer.
