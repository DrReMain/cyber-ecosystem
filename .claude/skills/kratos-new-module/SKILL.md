---
name: kratos-new-module
description: Domain module creation — use when creating a kratos module/<d>/ four-file set or a new ent schema in app/services/*. Trigger phrases include aggregate root, new table, new entity.
---

# New domain module (aggregate root)

Read first: `docs/conventions/kratos/CONVENTIONS.md` §0 (execution contract), §1 (layering & module-dependency law), §2 (naming table), §3 (file structure & ordering), §4 (comment law).

Procedure: `docs/conventions/kratos/RUNBOOKS.md` §1 "New domain module" — steps 1–6 exactly. Trap inside step 6: wire prunes unconsumed providers — anything startup-shaped must hang off a live dependency chain (a `newApp` param or HookRegistry self-registration).

Gates (all exit 0): `./nx run <service>:generate:ent` → `:migrate:diff` → `:migrate:apply` → `:generate:wire` → `:build`, plus `:test` (TestMixinOrder guards the mixin order; the conventions guard must stay green).

Finish with the §7 gate table before reporting done; final line: the root `AGENTS.md` §6.1 declaration — change-kind + counted evidence.
