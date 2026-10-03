---
name: proto-change
description: Contract change workflow — use when editing any .proto file or adding RPCs, messages, or error enums (regenerate all ends, locale-complete error copy, breaking check, affected builds).
---

# Proto contract change

Procedure: `docs/conventions/proto/CONVENTIONS.md §8` — the changing-contracts checklist; read it and follow it exactly.

Rules in force: §3 (optionality & the null/zero contract), §4 (error enums & i18n keys), §5 (access annotation is mandatory on every business RPC).

Blind spot: §9 — protovalidate rule changes are invisible to `buf breaking`; read it before merging.

Done = per the checklist's closing line.
