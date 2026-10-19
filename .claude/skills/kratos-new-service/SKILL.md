---
name: kratos-new-service
description: New service derivation — use when creating a new service under app/services/* (copy-set from system, ent two-pass generation, wire cold start, IAM consumption, catalog federation, deploy plane).
---

# Deriving a new service

Read first: `docs/conventions/kratos/CONVENTIONS.md` §5a "Non-system services (IAM consumers)" — cookie forwarding, the single remote Decide (CheckGrants, fail-closed), client middleware order, grant-aware hydration.

Procedure: §7 "Deriving a new service" — the eight steps exactly. Traps live inside them: the ent two-pass stub dance (step 3), wire cold start + interface-bind pruning (step 4), the atlas dev-database reset before any squash (step 5).

Deploy wiring: `docs/conventions/deploy/CONVENTIONS.md` §7.

Gates (all exit 0): `./nx run <service>:build` and `:test`; after the deploy-plane step, the compose merge validation.
