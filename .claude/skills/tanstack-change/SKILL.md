---
name: tanstack-change
description: TanStack client change & verification — use when creating or editing anything under app/clients/* or shared-ts/* (pages, routes, components, stores, domains, i18n copy, transport, theme, antd, cookie, store capability packages), and before reporting done on any client or shared-ts change. Trigger phrases include page, route, component, store, i18n, locale, paraglide, tabbar, capability package.
---

# TanStack client change

Read first: the read map in `docs/conventions/tanstack/CONVENTIONS.md` — it routes skin-preset paths to `docs/conventions/skin/CONVENTIONS.md`.

Chain (order is law, all exit 0, `--skip-nx-cache`), per affected client: `<client>:build` FIRST (regenerates routeTree + paraglide) → `:typecheck` → `:check`. A `shared-ts` edit also runs the touched package's own chain (`:typecheck`, `:check`, `:test` where the package has it), and the consuming clients' chain still runs.

Framework behavior the change turns on (an API's real shape, option semantics, version drift) is checked against the primary docs, not memory: `tanstack search-docs "<query>" --json` locates the page, `tanstack doc <library> <path> --json` fetches it — availability and full command shapes in `docs/environment/README.md` (tanstack-cli).

Traps: §11 reachable-surface rule — changes on login pages, guard redirects, or the SSR entry need live browser verification before done; library-seam changes get the §11 behavioral pairing review. No browser-automation tool connected when live verification is due → recommend `docs/environment/README.md` (never auto-install); a manual developer walkthrough is the degraded form. Close per §11 Close.
