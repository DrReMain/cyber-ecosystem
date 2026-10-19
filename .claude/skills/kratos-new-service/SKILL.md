---
name: kratos-new-service
description: New service derivation — use when creating a new service under app/services/* (copy-set from system, ent two-pass generation, wire cold start, IAM consumption, catalog federation, deploy plane).
---

# Deriving a new service

Read first: `docs/conventions/kratos/CONVENTIONS.md` §0 (execution contract) and §5a "Non-system services (IAM consumers)" — cookie forwarding, the single remote Decide (CheckGrants, fail-closed), client middleware order, grant-aware hydration, member-plane composition reads.

Each service owns its middleware chain assembly (per-service policy, never shared-go); the components come from shared-go — session forwarding via `kratos/transport/connect` (`WithSessionToken` + `SessionCookieInterceptor`), mapping helpers from `shared-go/utils`, pagination from `shared-go/helper` — re-implementing a component per service is a duplicate (§5).

Procedure: `docs/conventions/kratos/RUNBOOKS.md` §4 "Deriving a new service" — the five steps exactly. Traps live inside them: the ent two-pass stub dance (step 3 — entc also refuses an empty schema package, so the first aggregate's schema lands in the derive slice), wire cold start + interface-bind pruning (step 4), the atlas dev-database reset before any squash (step 5).

Deploy wiring: `docs/conventions/deploy/CONVENTIONS.md` §7.

Gates (all exit 0): `./nx run <service>:build` and `:test`; after the deploy-plane step, the compose merge validation. Green means the Nx "Successfully ran" banner or a checked exit code — a bare log tail is not a gate. Finish with the §7 gate table.
