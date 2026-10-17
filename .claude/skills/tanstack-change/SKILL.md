---
name: tanstack-change
description: TanStack client change & verification — use when creating or editing anything under app/clients/* (pages, routes, components, stores, domains, i18n), and before reporting done on any client change.
---

# TanStack client change

Read first (by surface): `docs/conventions/tanstack/CONVENTIONS.md` §1 (layering, transport JSON shapes, stream consumption), §3 (route tiers & the read-set contract — before any route edit), §5 (component & naming grammar), §6 (i18n ownership — before any message edit).

Gates (order is law, all exit 0): `./nx run admin:build` FIRST (regenerates routeTree + paraglide) → `./nx run admin:typecheck` → `./nx run admin:check`. After a regeneration, delete `*.tsbuildinfo` before the typecheck (§11 — stale cache reports green).

Traps: §11 reachable-surface rule — changes on login pages, guard redirects, or the SSR entry need e2e verification before done. Library-seam changes (dnd, scroll physics, menus, toasts) get the §11 behavioral pairing review.
