---
name: kratos-selfcheck
description: Kratos convention gate — use when creating or editing any .go file under app/services/* (load the dispatch row before writing) and again before reporting done (emit the gate table).
---

# Kratos convention gate

Before writing: `docs/conventions/kratos/CONVENTIONS.md` §0 execution contract — load the dispatch row for the artifact being touched; those sections are the law for the edit.

Before done: §7 "Pre-completion gate" — run every row over the combined diff and emit the report table (`item | verdict | evidence`) verbatim in the final message. No gate table = not done, whatever the build says.

Machine rows already enforced by tests (still report them): `TestNoDeclarationComments`, `TestReasonPredicates`, `TestSerializationViaUtils`, `TestModuleDependencyDAG`. For `shared-go` changes emit only K-ERR, K-DEP, and the capability self-containment rule (root §4).

Editor diagnostics lag codegen — trust `go build` / the Nx targets, never the LSP panel (`docs/conventions/kratos/RUNBOOKS.md` §5).
