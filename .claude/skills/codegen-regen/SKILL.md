---
name: codegen-regen
description: Full regeneration pass after a codegen or generator upgrade — use when bumping buf/protoc, connect-es, hey-api/openapi-typescript, ent, wire, nx graph edges, or any generator dependency, and before reporting done on such an upgrade. Trigger phrases include generator upgrade, regenerate everything, codegen bump, gen drift after upgrade.
---

# Full regeneration after a generator upgrade

Read first: `docs/conventions/proto/CONVENTIONS.md` §7 (gen pipeline) and the kratos read map's generate rows.

- An upgrade that touches any generator means **every** generate target re-runs — not only the project under change: `proto:generate` (all three output families), each service's `generate:ent` and `wire`. Partial regeneration leaves stale shapes that compile against the old runtime.
- The terminal check is delete-and-regenerate on one output family: a regen that cannot reproduce its own tree means the pipeline, not the tree, drifted.
- Re-run the drift and breaking gates after regeneration (`proto:check-drift` with gen staged; `proto:breaking`), then the touched services' build/test.

Traps: ent with mixin packages that depend on generated code needs the two-pass generation dance; paraglide may report "skipped — inputs unchanged" after an upgrade that changed hashing — the emitted file names are the truth, trust `typecheck`, not the skip line; generate targets race when run in parallel — the declared dependsOn edges (catalog←go, oapi-ts←connect-ts) are the order, not a suggestion.
