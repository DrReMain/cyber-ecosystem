---
name: deploy-change
description: Deploy-plane change workflow — use when touching deploy/, Dockerfiles, compose files, edge/traefik config, migrations, or wiring a new application/service into a stack, and before reporting done on any deploy change. Trigger phrases include compose, stack wiring, edge route, migration wiring, service set dial.
---

# Deploy-plane change

Read first: `docs/conventions/deploy/CONVENTIONS.md` — §6 (config mirrors & the env plane), §7 (adding an application), §10 (Atlas migrations), §11 (the service set as a deployment dial).

- Secrets and topology are deploy-plane facts: no key, DNS name, or ip:port lives in code or project.json dev targets beyond a public dev dummy.
- Config changes mirror across `configs/` and `container/` per service (§6); a service missing its mirror half is the classic drift.
- An application joins the stack through §7's steps exactly — compose profile, edge wiring, migration attach — never an ad-hoc compose edit.

Gates (exit 0): `./nx run deploy:validate` (compose config validation over every compose file) plus the owning service's build; after app wiring, the stack actually comes up with the new member.

Trap: a compose file that parses locally but references an image target that was never built fails only at `up` — build the image target in the same change when the compose file names it.
