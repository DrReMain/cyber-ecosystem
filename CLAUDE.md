# Cyber Ecosystem (CLAUDE.md)

## 1) Scope

Root-level rules only. Applies to the entire monorepo.

`MUST` / `MUST NOT` = hard requirement. `SHOULD` = strong default. `MAY` = optional.

Stack-specific, service-specific, and client-specific details belong in local documents
under the owning directory, not here.

---

## 2) Repository Overview

This is a large-scale development platform monorepo. It hosts multiple independent
business apps, shared libraries, contracts, infrastructure, and tooling.

**Apps** (`apps/`) may contain multiple backend services, clients, and technology
stacks within a single business unit.

**Shared contracts** (`contracts/`) use **Protobuf / gRPC**. Generated code is derived
output — never edit it directly.

**Shared Go libraries** (`shared-go/`) contain reusable Go-specific utilities and
middleware. Promote code here only when it is clearly reusable across more than
one Go consumer.

**Infrastructure** (`infra/`) uses **Docker Compose** for local environment support.

---

## 3) Running Commands

`MUST` use Nx for repository workflows that already have declared targets:

```bash
./nx run <project>:<target>
```

**How to find available targets for a project:**
Check the project's `project.json` — every valid target is declared there.
Do not assume a target exists; read `project.json` first.

```bash
# Run a specific declared target
./nx run <project>:<target>

# Run a target across all affected projects (only if the target is declared in each)
./nx affected --target=<target>
```

Recurring build, test, lint, generation, dev, and automation workflows `SHOULD`
be exposed through Nx targets.

If a needed recurring workflow has no Nx target yet, add one to `project.json`
instead of normalizing an ad-hoc command.

---

## 4) Hard Rules — Never Do This

These are the mistakes most likely to cause silent breakage in this repo:

**DO NOT bypass Nx with direct toolchain commands for workflows that have Nx targets:**
```bash
# WRONG — use ./nx run <project>:<target> instead
go build ./...
```

**DO NOT manually edit generated files.** Protobuf-generated code and any other
derived output must be regenerated via the owning Nx target.
If generated output looks wrong, fix the source or generator — not the output.

**DO NOT introduce cross-app dependencies.** If `apps/foo` needs something from
`apps/bar`, move the shared capability into an appropriate repository-level shared
module owned for reuse — not imported directly across app boundaries.

**DO NOT hardcode secrets, tokens, or environment-specific credentials** anywhere
in the repository.

---

## 5) Source-First Workflow

Always change sources before regenerating derived outputs:

1. Edit the source definition (`.proto` file, generator config, etc.)
2. Run the relevant Nx generation target
3. Review the generated diff — exclude unintended churn
4. If output changes unexpectedly, fix the source or generation flow

For cross-project changes: stabilize shared contracts or interfaces first →
update implementations → regenerate affected outputs.

---

## 6) Repository Layout

| Path | Purpose |
|---|---|
| `apps/` | Product code. Each app owns its services, clients, and local conventions. |
| `contracts/` | Protobuf schemas and other shared cross-project source-of-truth definitions. |
| `shared-go/` | Shared Go utilities and middleware. Default to backward-compatible changes. |
| `infra/` | Docker Compose and local environment support. |
| `tools/` | Generators, tooling, and developer automation. |

---

## 7) Validation — Definition of Done

Before closing any change:

1. Relevant Nx generation targets were run when required.
2. Touched projects were validated with the Nx targets they actually declare
   (for example `build`, `test`, `lint`, `proto`, `generate`, `dev:check`, etc.).
3. If a needed validation step has no Nx target yet, use the owning project's
   established workflow and explicitly note the gap.
4. Generated outputs were reviewed; unrelated churn was excluded.
5. Any skipped step or pre-existing failure is called out explicitly.

If generation fails → fix source definitions or generator inputs.
If declared validation fails in touched code → fix it before closing.
If a failure is clearly pre-existing and out of scope → record it explicitly.

---

## 8) Shared Go Libraries (`shared-go/`)

- Changes here affect all Go consumers across the monorepo — default to
  backward-compatible behavior.
- Coordinated breaking changes must be explicit and staged: update the library →
  update all consumers → verify.
- Do not add app-specific logic to shared libraries.

---

## 9) Documentation Policy

- This root `CLAUDE.md` covers only repository-wide rules.
- Stack-specific, service-specific, client-specific, or tool-specific guidance
  belongs in local `CLAUDE.md` files near the owning code.
- Local files may be stricter than this root file but must not conflict with it.
- Until local files exist, keep this root file minimal rather than embedding
  deep stack-specific instructions here.

---

## 10) Tech Stack Guides

When working on a specific stack, read the corresponding guide first:

| Stack | Guide | Applies to |
|-------|-------|------------|
| Kratos + Go | `docs/stacks/kratos-go.md` | Go microservices using Kratos framework |
| TanStack + React | `docs/stacks/tanstack-react.md` | React clients using TanStack Start, Router, and Query |
| Expo + React Native | `docs/stacks/expo-react-native.md` | Mobile clients using Expo SDK and expo-router |
| Observability | `docs/stacks/observability.md` | Local dev observability (SigNoz, GlitchTip, OTel) |

---

## 11) Skills Index

Skills live in `.claude/commands/` and provide step-by-step guidance for common workflows. Invoke them with `/skill-name`.

| Skill | File | When to Use |
|-------|------|-------------|
| `/proto-design` | `.claude/commands/proto-design.md` | Creating or modifying proto files, setting up buf generation |
| `/kratos-scaffold` | `.claude/commands/kratos-scaffold.md` | Scaffolding Go microservices, adding aggregates, implementing business logic |
| `/tanstack-client` | `.claude/commands/tanstack-client.md` | Creating React web clients, adding pages, connecting to BFF APIs |
| `/expo-react-native` | `.claude/commands/expo-react-native.md` | Creating mobile clients, adding screens, Connect RPC, i18n, navigation |
| `/observability` | `.claude/commands/observability.md` | Setting up error reporting, metrics scraping, traces for new services |

Each skill is self-contained with inline code examples. For deep architecture details,
see the corresponding stack guide in `docs/stacks/`.

---

## 12) Commit & Branch Conventions

### Commit Messages

```
type(scope): description
```

| Field | Values |
|-------|--------|
| type | `feat`, `fix`, `refactor`, `docs`, `chore`, `test` |
| scope | App or component name (e.g., `genesis`, `contracts`, `shared-go`) |
| description | Concise summary in English, lowercase, no period |

### Branch Naming

Use descriptive kebab-case: `feat/genesis-comment-aggregate`, `fix/base-ent-migration`.

### Pull Requests

Squash merge to `main`. PR title follows the commit format.

---

## 13) Testing Strategy

- **Go services**: `go test` (standard library `testing`; `testify` optional). Tests co-located as `{file}_test.go`.
  Run per-service: `cd apps/<app>/services/<service> && go test ./...`
- **React clients**: recommended `vitest` + `@testing-library`, tests co-located as `{file}.test.tsx` (set up per client; no global `test` target yet).
  Run via: `./nx run <project>:test` (if declared) or `pnpm vitest`.
- **Shared libraries**: Co-located test files. `shared-go/` uses standard `go test`.
- New code SHOULD include tests for business logic (UC layer) and critical mappings (service layer).

Stack-specific testing patterns: see `docs/stacks/<stack>.md` Testing section.

---

## 14) Generation Recovery

When code generation fails, follow these recovery steps:

| Failure | Cause | Recovery |
|---------|-------|----------|
| Wire "no provider found" | Missing constructor in ProviderSet | Add to the appropriate `ProviderSet`, then `./nx run <project>:generate:wire` |
| Ent "undefined" for local_mixins | Chicken-and-egg: mixins need generated code | Three-stage bootstrap (see `/kratos-scaffold`) |
| buf import not found | Import path doesn't resolve from module root | Check path resolves relative to `contracts/` or `apps/`, then `./nx run <app>_api:proto:api` |
| i18n empty stubs | Wrong relative path in `i18n.protos` | Fix path from `internal/i18n/` to proto file, then `./nx run <project>:generate:i18n` |
| `go mod tidy` removes deps | Imports look unused before generation | Run generation first; `generate` target runs `go mod tidy` as last step |

Always run generation BEFORE `go mod tidy`. The `generate` target handles ordering automatically.
