---
name: kratos-new-service
description: New service derivation — use when creating a new service under app/services/* (copy-set from system, ent two-pass generation, wire cold start, IAM consumption, catalog federation, deploy plane).
---

# Deriving a new service

Read first: `docs/conventions/kratos/CONVENTIONS.md` §5a, closing subsection "Non-system services (IAM consumers)" — cookie forwarding, the single remote Decide (CheckGrants, fail-closed), client middleware order, grant-aware hydration.

Each service owns its middleware chain assembly (per-service policy, never shared-go); the components come from shared-go — session forwarding via `kratos/transport/connect` (`WithSessionToken` + `SessionCookieInterceptor`), mapping helpers from `shared-go/utils`, pagination from `shared-go/helper` — re-implementing a component per service is a duplicate (§5).

Procedure: §7 ("Checklist") → the "Deriving a new service" block — the eight steps exactly. Traps live inside them: the ent two-pass stub dance (step 3 — entc also refuses an empty schema package, so the first aggregate's schema lands in the derive slice), wire cold start + interface-bind pruning (step 4), the atlas dev-database reset before any squash (step 5).

Deploy wiring: `docs/conventions/deploy/CONVENTIONS.md` §7.

Gates (all exit 0): `./nx run <service>:build` and `:test`; after the deploy-plane step, the compose merge validation. Green means the Nx "Successfully ran" banner or a checked exit code — a bare log tail is not a gate (a red build once hid behind `tail`'s last lines).
