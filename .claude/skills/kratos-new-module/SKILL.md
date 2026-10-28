---
name: kratos-new-module
description: Domain module creation — use when creating a kratos module/<d>/ four-file set or a new ent schema in app/services/*. Trigger phrases include aggregate root, new table, new entity.
---

# New domain module (aggregate root)

Read first: the read map in `docs/conventions/kratos/CONVENTIONS.md` — the `module/<d>/**` and `ent/schema/**` rows.

Procedure: `docs/conventions/kratos/RUNBOOKS.md` §1 "New domain module" — the steps exactly. Trap inside the wire step: wire prunes unconsumed providers — anything startup-shaped must hang off a live dependency chain (a `newApp` param or HookRegistry self-registration).

Verify per kratos CONVENTIONS §7: the generation chain green, the migration applied, one live RPC probe including the refused path; close with the counted declaration.
