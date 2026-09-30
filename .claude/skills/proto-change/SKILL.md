---
name: proto-change
description: Contract change workflow — use when editing any .proto file or adding RPCs, messages, or error enums (regenerate all ends, three-locale error copy, breaking check, affected builds).
---

# Proto contract change

Procedure: `docs/conventions/proto/CONVENTIONS.md §8` — the changing-contracts checklist; read it and follow it exactly.

Rules in force: `docs/conventions/proto/CONVENTIONS.md §3` (optionality & the null/zero contract), §4 (error enums & i18n keys), §5 (access annotation is mandatory on every business RPC).

Known trap: `docs/conventions/proto/CONVENTIONS.md §9` — buf breaking cannot see protovalidate rule removals/weakenings; review validation rules manually.

Done = per the checklist's closing line: exit codes checked, breaking green (or weakenings consciously reviewed), gen diff clean, error copy present in all three locales.
