# Browser Walkthrough Conventions

**Scope:** human-like walkthrough of the repo's `app/clients/*` via **chrome-devtools-mcp** (install guidance: the `tool-recommendations` skill). Tool-bound: these are chrome-devtools-mcp shapes — never extrapolate to another driver; a different driver earns its own page. Skeleton-side asset, business-neutral by law: cases keep their date and generic stack shape; business detail lives in the owning campaign's ledger. Prescriptive — MUST / SHOULD / MAY.

**Maintenance law:** a trap diagnosed mid-walkthrough gets written back here the moment it is understood, verdict-style (symptom / mechanism / prescription / case date). Same law as conventions-drift: the file is the memory, the session is not.

## Read map

| About to | Read first |
|---|---|
| any simulated walkthrough / live-proof pass / browser verification | all sections |

## §1 Triggering interactions

### 1.1 Controlled inputs — the fill tool is not trustworthy

- **Symptom:** `fill` sets the DOM value, the snapshot shows it — but the React state never receives it (no `onChange`), so validation and submit run against an empty value. The page looks filled while the component thinks it is not. (Case 2026-11-06: a login phone field.)
- **Mechanism:** the tool writes the property directly, bypassing React's value tracker; React only honors the native setter paired with an `input` event.
- **Prescription — MUST** drive React-controlled inputs with the native setter:

```js
const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value").set;
input.focus();
setter.call(input, "");
input.dispatchEvent(new Event("input", { bubbles: true }));
setter.call(input, "13800000000");
input.dispatchEvent(new Event("input", { bubbles: true }));
input.blur();
```

- Checkboxes/toggles: programmatic `.click()` on the input switches them correctly (the React onChange listens to click at the input level).
- **Trap — checkbox choreography:** index picks and label-text clicks are unreliable (indexes drift with noise checkboxes and stale state; labels without `for`/wrap association toggle nothing). Find the row whose `textContent` includes a known phrase, `querySelector` its `input`, `.click()` the input body. (Case 2026-11-06: three label-click misses before the input-body click landed.)

### 1.2 Click order — real events first, programmatic fallback

- **Symptom:** mcp `click` by uid fails with "element did not become interactive" (overlay/mask swallowing the hit), or lands with no effect.
- **Prescription:** take a fresh snapshot and use mcp `click` first (real CDP input events — React synthetic handlers honor them). Only on failure or visible no-op fall back to a programmatic click, which bubbles and triggers React `onClick` just the same:

```js
[...document.querySelectorAll("button")].find(b => b.textContent?.includes("<the button's label>"))?.click();
```

- **Trap:** icon-only buttons carry an `aria-label` and no text — a text finder never matches them. When a text regex finds nothing, enumerate buttons with their `aria-label`s before concluding the control is absent. (Case 2026-11-06: an overlay HUD's pause control.)

### 1.3 Visibility changes — tab switching is a dead end; patch `document.hidden`

- **Symptom:** simulating "user leaves the page" by switching to another tab (mcp `select_page`/`new_page`) fires no `visibilitychange` — the page's own visibility log stays empty. (Case 2026-11-06: three failed attempts before the patch landed.)
- **Mechanism:** the tool drives target switching without the browser tab-lifecycle the real interaction goes through; the event never reaches the page.
- **Prescription — MUST** synthesize the state: patch the property, then dispatch manually. Timing requirement: runtimes with a pre-run noise guard (`!started` drops early flips) need the dispatch after an observable start proof (e.g. a ticking run timer), held past the server's budget threshold before unpatching:

```js
Object.defineProperty(document, "hidden", { configurable: true, get: () => true });
document.dispatchEvent(new Event("visibilitychange"));
// …hold past the threshold…
Reflect.deleteProperty(document, "hidden"); // restore
```

## §2 Observing transient feedback

### 2.1 One script: trigger → wait → observe (toast law)

- **Symptom:** toasts (sonner, ~3s auto-dismiss), countdowns, and pulse states observed across separate tool calls come back empty — and an invented copy wordlist guesses wording the real copy never uses, producing a false "silent hang" verdict. (Case 2026-11-06: six wasted rounds before the toast was found.)
- **Prescription — MUST** complete trigger, wait, and observation inside a single `evaluate_script`:

```js
btn.click();
await new Promise(r => setTimeout(r, 500));
const toasts = [...document.querySelectorAll("[data-sonner-toast]")].map(t => t.textContent);
```

- Read the toast container selector whole. **Never** filter candidate feedback by an invented wordlist — the wordlist is the bug.

### 2.2 The "silent" verdict discipline

- **Prescription:** a UI judged "no reaction" is presumed an observation false-negative until every feedback surface has been enumerated empty: toast container, inline error rows (form `p`/`span`), control state flips (`disabled`, countdown text), button label changes. Only then may "silent" be recorded — ideally with a no-console-error check beside it.

## §3 Observation-plane fallbacks

### 3.1 The mcp network/console panels may under-record

- **Symptom:** `list_network_requests` shows zero entries while a fetch in `evaluate_script` demonstrably returns (the panel's CDP attachment is not guaranteed).
- **Prescription:** trust the page itself over the panel:

```js
performance.getEntriesByType("resource")
  .filter(r => r.initiatorType === "fetch")
  .map(r => ({ n: r.name, s: r.responseStatus, d: Math.round(r.duration) }))
```

### 3.2 initScript fetch hook for wire captures

- **Symptom:** the raw wire body of a request (error JSON, headers) is needed and the network panel is unusable.
- **Prescription:** inject a hook on the next navigation (`navigate_page` `initScript`), then read the accumulator:

```js
window.__probe = [];
const orig = window.fetch;
window.fetch = async (...a) => {
  const r = await orig(...a);
  try { const b = await r.clone().text(); window.__probe.push({ u: String(a[0]).slice(0, 120), s: r.status, b: b.slice(0, 500) }); } catch {}
  return r;
};
```

### 3.3 The dev-server task log is a gold mine

- **Symptom:** the page is stuck (progress loop, boundary churn) and in-page observation is a fight.
- **Mechanism:** vite dev forwards browser `console.error` / `console.warn` into its task output — full error objects included, which the page may never visibly render. (Case 2026-11-06: the decisive error-envelope capture came from the vite log, not the DOM.)
- **Prescription — SHOULD** read the dev task log (background task `.output` file) before dissecting a broken page.

### 3.4 The assumed container may be empty — capture from body

- **Symptom:** a `main`-scoped (or any assumed-container) query returns empty text/DOM on one page while the same query works on another — layout families render different container shapes.
- **Prescription:** scope selectors to tighten output, never as a structural assumption; when a scoped capture comes back unexpectedly empty, re-run from `document.body` before concluding the element is absent. (Case 2026-11-06: twice in one day — a modal-layout page and a full-screen HUD timer.)

## §4 Sessions and cookies

### 4.1 evaluate-script fetch may not carry the session

- **Symptom:** logged in on the page, yet a hand-rolled `fetch` inside `evaluate_script` returns 401 — the cookie does not always ride along.
- **Prescription:** prefer the page's own UI for anything session-bound; for direct API probes use `curl` with the cookie taken from the dev flow. HttpOnly cookies cannot be injected manually — the real login flow is the only road (see §6 for the mock-sms seam).
- Multi-account sessions: open pages with `isolatedContext` so cookies and storage never trample each other.

## §5 Screenshots

### 5.1 bringToFront against phantom frames; blob copy fidelity

- **Prescription:** bring the page to the front before capturing (background tabs yield stale/phantom frames); when capturing generated assets, copy blob URLs promptly — they expire with the page. (Case 2026-10.)

## §6 Server-truth cross-checks

### 6.1 Browser observation × server truth

- **Law:** a simulated walk proves an interaction only when the browser says "done" AND the server truth agrees. Keep both sides in the record.
- **Prescription:** read the backend directly for the counterpart fact — redis keys (e.g. the system auth service's mock-sms login code under `auth:web:sms-code:<phone>`), SQL fixtures to plant states, `sessionStorage` seams the client already exposes (e.g. parked third-party identity profiles, one-shot handoff payloads) for faces the normal flow cannot reach offline. (Case 2026-11-06: sms code read from redis seconds after the toast.)

### 6.2 In-memory handoffs — same-mount continuity and the dev-import seam

- **Symptom (shape A — per-mount state):** an sms flow (send → submit) driven with a navigation in between shows the page's GENERIC failure word instead of the specific reason — the remount destroyed the per-mount `state` minted by send, the submit's validator threw, and the catch rendered the fallback. A "no specific copy" verdict here is a walkthrough artifact, not a product fact. (Case 2026-11-06: a password-reset leg misread as missing its not-found copy.)
- **Prescription (A) — MUST** keep send and submit inside the same mounted page: navigate once, then run send and submit as separate `evaluate_script` calls **without navigation** between them (the component stays mounted across tool calls; only navigating remounts it).
- **Symptom (shape B — module-singleton handoff):** a one-shot handoff held in module scope (not sessionStorage, not URL — e.g. a payload that cannot be re-derived) is unreachable by URL gardening or storage injection.
- **Prescription (B):** in dev, `await import("/src/path/to/module.ts")` from the page returns the **live module instance** — call its setter, then navigate SPA-side (the router global or an in-app link); a hard reload resets module state (which itself proves the honest-bounce path). (Case 2026-11-06: a module-singleton handoff's faces, cross-validated against its real producer.)

## §7 Environment-state traps

### 7.1 Stack bring-up checks

- Ports first, then HTTP: TCP-listen polling for the dev stack, then a 200-chain over healthz/page roots. Readiness words live in each task's `.output` ("ready in", listening lines). Service logs: `tail` the task file, never guess from silence.

## §8 Framework probe handles

### 8.1 `window.__TSR_ROUTER__` — replicate before accusing

- **Symptom:** a navigate call inside a component seems to misresolve (doubled path segments in router warnings).
- **Prescription:** TanStack Router's dev global exposes the live router; replicate the exact call in the console (`router.navigate({ to, search })`) and watch URL/state/matches. A healthy manual replication clears the API and pins the fault on timing or render-cycle context. (Case 2026-11-06.)
