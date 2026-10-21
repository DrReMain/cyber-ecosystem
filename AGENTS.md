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
| `shared-ts/` | Cross-client TS capabilities (error, antd, theme, store, cookie, router-progress, transport, storybook) |
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

**DO NOT use npm or yarn — pnpm is the only package manager.** `pnpm-lock.yaml` is the sole lockfile; a `package-lock.json` or `yarn.lock` appearing anywhere means some command bypassed pnpm — delete the artifact and fix the command, never add an ignore rule to hide it.

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
5. Any skipped step or pre-existing failure is called out in the §6.1 declaration line.
6. Paradigm or mechanism changes: the owning `docs/conventions/<area>/CONVENTIONS.md` is updated and affected skill pointers are verified — `./nx run docs:check` is green.

### 6.1) Verification posture — declare, then close

Every change closes with a one-line declaration: change-kind + the verification actually executed, with counted evidence ("proven by: 2/2 refused-path probes; chain green"). The declaration closes the slice's final report and rides into the commit message. A closed change without its evidence line is not done — missing verification shows up as silence, so the positive statement is mandatory, not the exemption.

Kind set (closed; classify by substance, not the commit verb — a "refactor" that changes contracts is a contract change; a full-stack vertical slice takes the posture of its highest-risk face, aggregate > logic > surface):
`contract · aggregate · logic · surface · capability · refactor · chore · docs`

Default posture (the owning conventions instantiate each cell with concrete commands):

| kind | gates | runtime verification | tests | review |
|---|---|---|---|---|
| contract | drift gate + consumer chains | counterexample e2e, both sides | the drift gate is the test | full, second reviewer |
| aggregate | chain + migration applied | live RPC probe incl. the refused path | none by default | full |
| logic | chain | live probe, counterexamples at the same bar as tests | silent-regression guards where a regression would be invisible at runtime | normal |
| surface | client chain | browser closed loop (viewports, locales, zero console) | none | visual + mechanical checklist |
| capability | chain + declared conformance | probe / smoke | probe suites + silent-regression guards | multi-expert adversarial, two rounds |
| refactor | chain, cache bypassed | counterexample replay — pin the baseline before the change | existing only | diff shape |
| chore / docs | owning chain / docs gate | none | none | skim |

Test posture is a function of the review regime, not a constant: under strict human review the human carries the verification (defaults above); when output is agent-heavy and review is light, the burden moves into executable form — test-first becomes the default for behavior-bearing kinds, and the reviewer reviews the spec rather than the implementation. The active regime is declared in the active working lane's collaboration section (lane law: docs/README.md §1), not here. Runtime verification escalates with blast radius: single-surface smoke → protocol matrix → full-flow e2e → cross-commit integration e2e.

Trigger tiers: **slice close** (the table above) · **absorb** (integrating a reviewed slice: the mechanical checklist runs — deleted exports grepped for remaining consumers, new magic values traced to their source — and the integration diff equals exactly the absorbed slice's files) · **freeze** (pre-milestone: multi-stage e2e campaign with counterexamples, fresh-cache gate matrix, live-stack health; campaign scripts are throwaway, never repo assets).

Ordering: self-check walks its own diff before gates run · e2e precedes doc codification · a slice stops for review before commit · under strict review, review precedes test-writing; under test-first, the spec is what review reads first.

---

## 7) Documentation — Three Lanes

Knowledge lives in three lanes; each kind of knowledge has exactly one home — durable rules (`docs/conventions/<area>/CONVENTIONS.md`, EN), procedural skill dispatchers (`.claude/skills/<name>/SKILL.md`, EN; trigger + pointers, never content), and a temporary, fully isolated working lane for active fronts (defined only in `docs/README.md`). Full governance law (admission, promotion, backfill, pointer discipline, instance-state, drift checks): `docs/README.md`.

Lane rules:

- `MUST` read the relevant `CONVENTIONS.md` before writing code in an area, and follow it exactly (structure, naming, ordering, comments, dependency rules).
- `MUST NOT` duplicate content across lanes or into skills — pointers only. Rule text and step text live in docs; a skill body holds a trigger, pointers, completion pointers, one-line trap annotations (a warning naming the trap and its home — never the rule or the steps), and its gates line — the owning chain's concrete commands, instantiated from the conventions, which stay the source.
- `MUST NOT` reference roadmap content from here, `CLAUDE.md`, skills, or conventions — the working lane is temporary and fully isolated (law: `docs/README.md` §2.7); active fronts are introduced by the user, not by root docs.
- Skills auto-trigger by description (Claude Code); any agent or human can read `.claude/skills/*/SKILL.md` as plain markdown.
- Path-scoped rule shells (`.claude/rules/*.md`) inject their law card automatically when a matching file is read or edited (delivery law: `docs/README.md` §2.9).

Area index — read before writing code in the area:

| Area | Doc | Read when |
|---|---|---|
| proto (contracts) | `docs/conventions/proto/CONVENTIONS.md` | touching any `.proto`, or code that consumes `gen/` |
| kratos (Go services) | `docs/conventions/kratos/CONVENTIONS.md` | writing Go under `app/services/*` or `shared-go/` — start at §0 |
| tanstack (web clients) | `docs/conventions/tanstack/CONVENTIONS.md` | writing code under `app/clients/*` |
| deploy (containers & pipelines) | `docs/conventions/deploy/CONVENTIONS.md` | touching `deploy/`, Dockerfiles, compose, edge, migrations |

Adding a new tech stack: create `docs/conventions/<area>/CONVENTIONS.md` **with the first real code of that stack** (not before), and add it to this index in the same change. Areas are tech stacks only — topic or domain rules live in the owning stack doc or here (§8), never in their own area directory.

---

## 8) Multi-app composition

One shared base, many business applications:

- **`system` is the mandatory base service** every application deploys: authentication, user identity, org (dept), roles, authz engine, resource catalog. Each application adds services under `app/services/<name>` with contracts under `proto/cyber/<name>/v1`, owning its aggregates end-to-end (schema, migrations, biz, RPC).
- Cross-service reads at runtime go through the owning service's RPCs — from clients, or behind a `<Remote>RP` ACL adapter in the consuming service. Business services `MUST NOT` grow identity/auth concerns; those live in `system` and are shared by all apps.
- **User split**: `system.User` is the authentication identity core only (credentials, org membership, role bindings). App-specific person data is the app's own domain aggregate keyed by `user_id` (plain id field), named by its domain concept (`Delegate`, `Operator`) — never `UserExt`-style field bags; own proto, own schema, logical reference only. A field enters `system.User` only when the declared scenario set needs it in the shared identity core — never one app's business shape. Reads hydrate identity via `system` RPCs — a missing profile degrades rendering, never breaks the row; each aggregate keeps a single write entrance. `system` cascades only what it owns; business profiles tolerate orphaning; cross-service invalidation events appear when a declared scenario requires them.
- **Pluggable system surface**, cheapest tier first: permission gating by default (hide-by-default — an ungranted operation never surfaces) → config-level module registry when the API itself must be absent → code fork only when isolation physically forces it, and then as a delivery action, not a development structure.
- **Composition & branching**: applications coexist as directories in one trunk; an application is a composition unit — deploy profiles select the service set, the client composes menu and skin per app. Per-app long-lived branches `SHOULD NOT` (every base fix would need cherry-picking across them); fork per app only at isolation-forced delivery time, as late as possible.
- `shared-go` / `shared-ts` carry capabilities, never business aggregates. **Skeleton completeness law**: the base implements the owner's declared scenario set at first touch — multi-app composition, C-end member planes, SaaS multi-tenancy, ToB editions are stated requirements of the base, not future speculation. A base surface that cannot hold a declared scenario is a mechanism defect — fix the model, never patch or defer. Inside a business app, concrete-first still holds: extract on a second real consumer.
