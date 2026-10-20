---
name: dev-environment
description: Recommended (optional) agent tooling — use when onboarding a new developer or machine, when asked about environment setup or which MCP/plugin tools this repo recommends, when a recommended tool is detected missing from the session, or before live verification when no browser automation tool is connected. Trigger phrases include environment setup, onboarding, recommended tools, missing mcp, missing plugin.
---

# Dev environment (recommend only)

Single source: `docs/environment.md` — what each recommended tool serves, its identity and upstream, its per-clone state, and what degrades without it.

Flow: detect the gap (codegraph / memsearch / browser-automation tools absent from the session) → point the developer at the document → stop.

Trap: installation is never executed by the agent — setup is the developer's manual action per the upstream docs (docs/environment.md header rule).
