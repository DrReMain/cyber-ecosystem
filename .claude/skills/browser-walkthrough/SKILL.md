---
name: browser-walkthrough
description: Simulated human-like walkthrough of web clients via chrome-devtools-mcp — pre-flight checks, trap defenses, and session bootstrap. Use for simulated walkthroughs, live-proof passes, render-coverage sweeps, or any chrome-mcp driven UI testing or debugging of this repo's web clients. Trigger phrases include "simulated walkthrough", "live proof", "render coverage", "chrome mcp", "browser verification".
---

# Browser walkthrough (chrome-devtools-mcp)

The traps live in `docs/conventions/browser-walkthrough/CONVENTIONS.md` — read it before the first interaction of the session if not already loaded. This skill is the operating sequence around it, and it is a skeleton-side asset: no business names, ports, or flow specifics — those live in the owning campaign's docs.

## 1. Pre-flight

1. **Conventions**: if the law card's file is not yet read this session, read it now.
2. **Stack health**: bring the stack per the campaign plan (dev targets via `./nx run <project>:dev` as background tasks; never start a task that is already running). TCP-poll each dev port, then an HTTP 200 chain over `/healthz` and the page root — targets and ports live in each project's `project.json`/configs; readiness lines live in each background task's `.output`.
3. **Tool loading**: chrome-mcp tools may be deferred — `ToolSearch "select:"` for click / fill / fill_form / evaluate_script / take_snapshot / take_screenshot / navigate_page / new_page / emulate / wait_for / list_pages / select_page.
4. **Observation sanity**: an empty `list_network_requests` is not "no requests" — the panel under-records; `performance.getEntriesByType("resource")` is the fallback (§3.1).

## 2. Walk discipline

- Controlled inputs: native setter + `dispatchEvent("input")` (§1.1). Clicks: CDP click first, programmatic `.click()` fallback (§1.2).
- Transient feedback: one evaluate_script — trigger → sleep → query — reading `[data-sonner-toast]` whole, no invented wordlists (§2.1). "Silent" verdicts only after enumerating every feedback surface (§2.2).
- Wire captures: initScript fetch hook (§3.2). Broken page: read the vite task log first (§3.3).
- Session-bound probes: use the page's own UI or curl with the real cookie; HttpOnly sessions require the real login flow (§4).
- In-memory handoffs: same-mount continuity for sms state; dev dynamic-import seam for module singletons (§6.2).
- Every simulated action closes with a server-truth cross-check when one exists — redis key, SQL row, or service log line (§6).

## 3. Session bootstrap (mock sms)

1. Fill the phone with the native setter, check the agreement boxes, click send.
2. Read the live code from the system auth service's redis key: `auth:web:sms-code:<phone>`.
3. Fill the code the same way, submit; the landing face confirms the session.

## 4. Close

- Record live evidence into the campaign's gate-ledger rows with the case date.
- Write any newly-diagnosed trap back into the conventions file verdict-style — the file is the memory, the session is not.
- Stack teardown per the campaign plan (TaskStop the background tasks; never `kill`).
