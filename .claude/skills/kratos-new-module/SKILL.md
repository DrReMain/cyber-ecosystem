---
name: kratos-new-module
description: Domain module creation — use when creating a kratos module/<d>/ four-file set or a new ent schema in app/services/*.
---

# New domain module (aggregate root)

Procedure: `docs/conventions/kratos/CONVENTIONS.md §7` "Adding a new domain module" — follow it exactly through schema, migrations, wiring, and build.

Rules in force: `docs/conventions/kratos/CONVENTIONS.md §1` (layering & module-dependency paradigm), §2 (naming), §3 (file structure & ordering), §4 (comment discipline), §5 (layer specifics).

Known trap: `docs/conventions/kratos/CONVENTIONS.md §7` step 6 — wire prunes unconsumed providers; a startup-shaped component must hang off a live dependency chain (`newApp` param / `BeforeStart` hook) or it silently never instantiates.

Finish with `docs/conventions/kratos/CONVENTIONS.md §7` "Pre-completion self-check" before reporting done.
