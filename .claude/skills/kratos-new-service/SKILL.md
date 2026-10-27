---
name: kratos-new-service
description: New service derivation — use when creating a new service under app/services/* (copy-set from system, ent two-pass generation, wire cold start, IAM consumption, catalog federation, deploy plane). Trigger phrases include derive, scaffold, new service skeleton.
---

# Deriving a new service

Read first: the read map in `docs/conventions/kratos/CONVENTIONS.md`; §5a "Non-system services (IAM consumers)" is the identity law for the derived service; deploy wiring is `docs/conventions/deploy/CONVENTIONS.md` §7.

Each service owns its middleware chain assembly (per-service policy, never shared-go); components come from shared-go (session forwarding via `kratos/transport/connect`, mapping helpers from `shared-go/utils`, pagination from `shared-go/helper`) — re-implementing a component per service is a duplicate.

Procedure: `docs/conventions/kratos/RUNBOOKS.md` §4 "Deriving a new service" — the steps exactly; traps live inside them: the ent two-pass stub dance (entc refuses an empty schema package — the first aggregate's schema lands in the derive slice), wire cold start + interface-bind pruning, the atlas dev-database reset before any squash.

Verify per kratos §7 and the deploy closer: builds green, migration applied on the dev stack, one live probe; counted declaration.
