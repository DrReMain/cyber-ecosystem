# Cyber Ecosystem

## 1) Scope

Root-level rules only. Applies to the entire monorepo.

`MUST` / `MUST NOT` = hard requirement. `SHOULD` = strong default. `MAY` = optional.

Stack-specific and service-specific details belong in local documents under the owning directory, not here.

---

## 2) Repository Overview

Monorepo for the Cyber Ecosystem platform.

Protobuf contracts (`proto/`) are the source of truth. Generated code (`gen/`) is derived output — never edit directly.

---

## 3) Running Commands — MUST Use Nx

`MUST` use Nx for all workflows that have declared targets:

```bash
./nx run <project>:<target>
```

**How to find targets:** Read the project's `project.json` — every valid target is declared there. Do not assume a target exists.

Recurring build, test, lint, generation, dev, and automation workflows `SHOULD` be exposed through Nx targets. If a needed workflow has no target yet, add one to `project.json` instead of running ad-hoc commands.

---

## 4) Hard Rules

**DO NOT bypass Nx** with direct toolchain commands (e.g. `buf generate`, `go generate`, `go build`) for workflows that have Nx targets.

**DO NOT manually edit generated files.** Fix the source or generator, then regenerate via the owning Nx target.

**DO NOT introduce cross-service dependencies.** Move shared capability into `shared-go/` instead.

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

