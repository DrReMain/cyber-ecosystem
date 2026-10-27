---
name: proto-change
description: Contract change workflow — use when editing any .proto under proto/ or adding RPCs, messages, error enums, or validation rules, and before reporting done on any contract change. Never hand-edit gen/ — regenerate instead. Trigger phrases include proto contract, RPC, error reason, breaking, gen drift.
---

# Proto contract change

Read first: the read map in `docs/conventions/proto/CONVENTIONS.md`; §3 (optionality & the null/zero contract), §4 (error enums & locale copy), §5 (the access annotation is mandatory on every business RPC).

Procedure: §8 "Checklist — changing contracts" — follow it exactly. Trap: §9 — protovalidate rule removals and weakenings are invisible to `buf breaking`; review validation rules explicitly.

Done per the §8 closing line: every exit code checked, gen diff clean, error copy present in every registered locale of every client whose error union includes the package; one counterexample live probe on the producer and a consumer; counted declaration.
