---
name: tanstack-change
description: TanStack client change & verification — use when creating or editing pages, routes, or i18n directories under app/clients/*, or when verifying any client change.
---

# TanStack client change

Page creation: `docs/conventions/tanstack/CONVENTIONS.md §3` (route tiers & the page read-set contract), §5 (component & naming grammar), §6 (i18n ownership).

Verification: §11 — order and cache rules live there (build-first, `*.tsbuildinfo`); follow exactly.

Reachable-surface rule: §11 (login pages, guard redirects, SSR entry) — applies before reporting done.
