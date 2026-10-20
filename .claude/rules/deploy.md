---
paths: ["deploy/**", "**/Dockerfile", "**/ent/migrate/migrations/**"]
---

# Deploy law card

- Read first: docs/conventions/deploy/CONVENTIONS.md — compose organization by profile, image self-containment, placeholder parameterization, migration wiring.
- Operations are Nx targets — check the declared targets before running compose or atlas commands by hand.
