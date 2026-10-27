# Kratos law card

- Before writing: read the read map at the top of `docs/conventions/kratos/CONVENTIONS.md` — it names the sections that govern the artifact.
- Read-first by surface: §1 (layering & cross-module law), §3 (four-file shape & ordering), §4 (comments), §6 (errors).
- Trap: serialization goes through `utils` only — never `encoding/json` directly (§1).
- Trap: editor diagnostics lag codegen — trust `go build` and the Nx targets, never the LSP panel.
- Verify: owning chain + live RPC probe incl. the refused path (§7).
