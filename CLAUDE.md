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
