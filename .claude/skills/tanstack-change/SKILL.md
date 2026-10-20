---
name: tanstack-change
description: TanStack client change & verification — use when creating or editing anything under app/clients/* (pages, routes, components, stores, domains, i18n copy), and before reporting done on any client change. Trigger phrases include page, route, component, store, i18n, locale, paraglide, tabbar.
---

# TanStack client change

Read first: the §0 dispatch row for your artifact — `docs/conventions/tanstack/CONVENTIONS.md` §0 (by surface: §1 layering & transport shapes, §3 route tiers & read-set, §5 component grammar, §6 i18n ownership).

Gates (order is law, all exit 0, all `--skip-nx-cache`) — run for each client under `app/clients/*` affected by the change: `./nx run <client>:build` FIRST (regenerates routeTree + paraglide) → `./nx run <client>:typecheck` → `./nx run <client>:check` (§11).

Traps: §11 reachable-surface rule — changes on login pages, guard redirects, or the SSR entry need e2e verification before done. Library-seam changes (dnd, scroll physics, menus, toasts) get the §11 behavioral pairing review. No browser-automation MCP connected when live verification is due → recommend docs/environment.md (never auto-install); a manual developer walkthrough is the degraded form.

Finish with the §12 pre-completion gate table (one row per item, emitted verbatim) before reporting done. Final line: the root `AGENTS.md` §6.1 declaration — change-kind + counted evidence.
