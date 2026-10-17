---
name: kratos-new-module
description: Domain module creation — use when creating a kratos module/<d>/ four-file set or a new ent schema in app/services/*.
---

# New domain module (aggregate root)

Read first: `docs/conventions/kratos/CONVENTIONS.md` §1 (layering & module-dependency law), §2 (naming table), §3 (file structure & ordering), §4 (comment discipline).

Procedure: §7 "Adding a new domain module" — steps 1–6 exactly. Trap before step 6: wire prunes unconsumed providers — anything startup-shaped must hang off a live dependency chain (a `newApp` param or HookRegistry self-registration).

Gates (all exit 0): `./nx run <service>:generate:ent` → `:migrate:diff` → `:migrate:apply` → `:generate:wire` → `:build`, plus `:test` (TestMixinOrder guards the mixin order).

Finish with §7 "Pre-completion self-check" before reporting done.
