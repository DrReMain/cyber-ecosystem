---
name: kratos-selfcheck
description: Kratos convention gate — use when creating or editing any .go file under app/services/* or shared-go/* (load the dispatch row before writing) and again before reporting done (emit the §7 gate table). Trigger phrases include Go service, module, biz, repo, capability, serialization.
---

# Kratos convention gate

Before writing: `docs/conventions/kratos/CONVENTIONS.md` §0 execution contract — load the dispatch row for the artifact being touched; those sections are the law for the edit.

Before done: §7 "Pre-completion gate" — run every row over the combined diff and emit the report table (`item | verdict | evidence`) verbatim in the final message (no gate table = not done, §0.2). Final line: the root `AGENTS.md` §6.1 declaration — change-kind + counted evidence.

Machine rows already enforced by tests (still report them): `TestNoDeclarationComments`, `TestReasonPredicates`, `TestSerializationViaUtils`, `TestModuleDependencyDAG`. For `shared-go` changes emit only K-ERR and K-DEP-1 (capability self-containment, root §4).

Editor diagnostics lag codegen — trust `go build` / the Nx targets, never the LSP panel (`docs/conventions/kratos/RUNBOOKS.md` §5).
