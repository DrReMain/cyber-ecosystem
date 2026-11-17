# Kratos law card

- Before writing: read the read map at the top of `docs/conventions/kratos/CONVENTIONS.md` — it names the sections that govern the artifact.
- Read-first by surface: §1 (layering & dependencies), §3 (file structure & ordering), §4 (comments), §6 (errors).
- Trap: serialization goes through `utils` only — never `encoding/json` directly (§1).
- Trap: editor diagnostics lag codegen — trust `go build` and the Nx targets, never the LSP panel.
- Spec first: the case table (scenario|given|action|expect|counterexample?) lands beside `biz.go` before the UC body fills in — code aligns to the table, never the reverse (§7).
- Verify: owning chain + spec green (table cases, counterexamples included) + live RPC probe incl. the refused path (§7).
