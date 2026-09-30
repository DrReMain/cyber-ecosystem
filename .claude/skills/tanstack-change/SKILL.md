---
name: tanstack-change
description: TanStack client change & verification — use when creating or editing pages, routes, or i18n directories under app/clients/*, or when verifying any client change.
---

# TanStack client change

Page creation: `docs/conventions/tanstack/CONVENTIONS.md §3` — route tiers and the page read-set contract (`staticData.operations` = every query the page issues, whole-set judged); §5 — one component per file, naming grammar; §6 — i18n ownership (directory registration in `project.inlang/settings.json`, key prefix = directory name, deliberate key order).

Verification: `docs/conventions/tanstack/CONVENTIONS.md §11` — build FIRST (regenerates routeTree + paraglide), then typecheck, then check; all `--skip-nx-cache` with exit codes checked; delete `*.tsbuildinfo` after any regeneration (incremental cache reports stale greens).

Reachable-surface rule: `docs/conventions/tanstack/CONVENTIONS.md §11` — login-reachable surfaces get e2e before done; auth-only surfaces carry an explicit user-verification list in the change report.
