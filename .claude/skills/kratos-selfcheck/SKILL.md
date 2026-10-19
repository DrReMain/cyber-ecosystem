---
name: kratos-selfcheck
description: Pre-completion conventions pass — use before reporting done on any Go change under app/services/* or shared-go (compile-green is not convention-green).
---

# Kratos pre-completion self-check

Procedure: `docs/conventions/kratos/CONVENTIONS.md` §7 "Pre-completion self-check" — walk every hunk of the diff against its five bullets (§2–§6).

Scope: full walk for `app/services/*`; for `shared-go` only the §6-errors, dependency-direction, and capability self-containment bullets apply.

Done = every bullet answered for every hunk. Editor diagnostics lag codegen — trust `go build` / the Nx targets, never the LSP panel (§7).
