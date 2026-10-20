---
name: proto-change
description: Contract change workflow — use when editing any .proto under proto/ or adding RPCs, messages, error enums, or validation rules, and before reporting done on any contract change. Never hand-edit gen/ — regenerate instead. Trigger phrases include proto contract, RPC, error reason, breaking, gen drift.
---

# Proto contract change

Read first: `docs/conventions/proto/CONVENTIONS.md` §3 (optionality & the null/zero contract), §4 (error enums, reserved law, locale copy), §5 (access annotation is mandatory on every business RPC).

Procedure: §8 "Checklist — changing contracts" — follow it exactly.

Gates (all exit 0, in order): `./nx run proto:generate` → `./nx run proto:breaking` → `./nx run <service>:build`. Client surface changed → also each consuming client's chain under `app/clients/*` (law: tanstack §11).

Trap: §9 — protovalidate rule removals/weakenings are invisible to `buf breaking`; review validation rules explicitly before merging.

Done = per the §8 closing line: every exit code checked, gen diff clean, error copy present in every registered locale. Final line: the root `AGENTS.md` §6.1 declaration — change-kind + counted evidence.
