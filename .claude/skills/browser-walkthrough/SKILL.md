---
name: browser-walkthrough
description: Simulated live-proof walkthrough of this repo's web clients via chrome-devtools-mcp — pre-flight checks, trap defenses, session bootstrap. Triggers: simulated walkthrough, live proof, render coverage, chrome mcp, browser verification.
---

# Browser walkthrough (chrome-devtools-mcp)

Traps live in `docs/conventions/browser-walkthrough/CONVENTIONS.md` — read it before the session's first interaction if not already loaded. Skeleton-side asset: no business names, ports, or flow specifics; those live in the owning campaign's docs.

## 1. Pre-flight

1. Conventions: read via the law card if not yet loaded this session.
2. Stack health: dev targets (`./nx run <project>:dev`) as background tasks, never restarted while alive; TCP-poll each dev port, then HTTP 200 over `/healthz` and the page root. Targets and ports live in each project's `project.json`; readiness lines in each task's `.output`.
3. Tool loading: chrome-mcp tools may be deferred — `ToolSearch "select:"` for click / fill / evaluate_script / take_snapshot / navigate_page / new_page / wait_for / list_pages / select_page.
4. Empty `list_network_requests` ≠ "no requests" — the panel under-records; `performance.getEntriesByType("resource")` is the fallback (§3.1).

## 2. Walk discipline

- Inputs: native setter + `dispatchEvent("input")`; CDP click first, `.click()` fallback (§1.1–1.2). Toasts/countdowns: one evaluate_script, container whole, no wordlists (§2.1); "silent" only after every feedback surface is enumerated empty (§2.2).
- Broken page: vite task log first (§3.3). Session-bound probes: the page's own UI, or curl with the real cookie (§4).
- In-memory handoffs: same mount for send+submit; dev `import()` + SPA nav for module singletons (§6.2). Close every simulated action with a server-truth cross-check when one exists (§6).

## 3. Session bootstrap (mock sms)

1. Native-setter fill the phone, check the agreement boxes, click send.
2. Read the live code from redis: `auth:web:sms-code:<phone>`.
3. Fill the code the same way, submit; the landing face confirms the session.

## 4. Close

- Record live evidence into the campaign's gate ledger with the case date.
- Write newly-diagnosed traps back into the conventions file verdict-style.
- Stack teardown per the campaign plan (TaskStop the background tasks; never `kill`).
