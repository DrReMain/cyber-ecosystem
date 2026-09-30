---
name: kratos-selfcheck
description: Pre-completion conventions pass — use before reporting done on any Go change in app/services/* (compile-green is not convention-green).
---

# Kratos pre-completion self-check

Procedure: `docs/conventions/kratos/CONVENTIONS.md §7` "Pre-completion self-check" — walk every hunk of the diff against the five bullets: §2 naming, §3 structure & ordering, §4 comment discipline, §5 layer boundaries, §6 error construction.

Done = every bullet answered for every hunk. For `shared-go` changes only the §6-errors and dependency-direction bullets apply.
