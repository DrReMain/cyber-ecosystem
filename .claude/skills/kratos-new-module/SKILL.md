---
name: kratos-new-module
description: Domain module creation — use when creating a kratos module/<d>/ four-file set or a new ent schema in app/services/*.
---

# New domain module (aggregate root)

Procedure: `docs/conventions/kratos/CONVENTIONS.md §7` "Adding a new domain module" — follow it exactly through schema, migrations, wiring, and build. Step 6 carries the wire-pruning trap; read it before wiring anything startup-shaped.

Rules in force: §1 (layering & module-dependency paradigm), §2 (naming), §3 (file structure & ordering), §4 (comment discipline), §5 (layer specifics).

Finish with §7 "Pre-completion self-check" before reporting done.
