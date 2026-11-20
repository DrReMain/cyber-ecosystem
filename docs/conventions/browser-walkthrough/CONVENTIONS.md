# Browser Walkthrough Conventions

**Status:** ACTIVE

> The craft law for driving a live browser as a verification instrument
> (chrome-devtools-mcp class tooling): how to walk a surface without lying
> to yourself. This is the browser loop named by tanstack CONVENTIONS §11
> Close — the finishing pass over surfaces already chain-green and
> spec-green. Playwright e2e law lives in the stack CONVENTIONS; this doc
> stays tool-craft. Business-neutral by ruling: phenomena, mechanisms,
> prescriptions — business detail belongs to the owning campaign's
> ledger, never here.
>
> **Scope:** browser platform only — a future platform (mini-program, app)
> grows its walkthrough craft in its own stack area, born with that
> platform's first real code; its pits never land here. At a second
> platform's birth the platform-agnostic judgements lift into a shared
> walkthrough home, built at that trigger, not before. Component libraries
> are project choices; library-keyed operations live in LIBRARIES.md
> beside this law.

## Read map

| About to | Read |
|---|---|
| click / fill / type on a live page | §1 |
| judge a transient or one-shot UI | §2 |
| trust what the tooling shows you | §3 |
| run multiple personas or sessions | §4 |
| capture a screenshot as evidence | §5 |
| claim "it persisted" | §6 |
| blame the page / the app | §7 |
| dig into framework internals | §8 |
| drive an antd / react-aria-components surface | LIBRARIES.md |

Entries are judgements, not narrative: **phenomenon** — mechanism;
prescription. A pit caught in the wild is pinned here the day it is
caught, in this form.

---

## 1) Interaction triggering

- **A pre-hydration interaction is a lost event.** The dev server serves
  the client graph asynchronously and hydration resets controlled values
  — a click or fill fired before hydration never lands, and the page then
  looks mysteriously unresponsive. The two observable hydration facts
  (the input value tracker; the loader RPC having fired) are stack law,
  owned by tanstack CONVENTIONS §11 and not repeated here — walking
  obeys the same: wait on one before interacting.
- **The interaction ladder — trusted input first, every descent named.**
  The interaction tools (click / fill / fill_form / hover / press_key /
  type_text / drag) ride the CDP Input domain: browser-generated,
  `isTrusted: true`, through the real hit-test and focus pipeline — the
  human-like baseline (Playwright's own channel). Per-key faces (typeahead,
  autocomplete) take `type_text` over `fill` (one-shot insert). Descend
  only with a reason, named in the report:
  - *programmatic* (`evaluate`): `el.click()` reaches a control a mask,
    backdrop, or watermark genuinely covers — a human is blocked too;
    the bypass is the point, not a dead-button verdict. A controlled
    input takes the native value setter, then `input` + `change` —
    React's change tracking ignores a direct `.value` write (raw
    `evaluate` is not a fill helper).
  - *refusal*: some surfaces (engines, canvases) reject automation
    entirely — per-surface, never generalized. The natural timeout is the
    recorded evidence; an attempt that burned server budget (a quota, a
    scored run) is a fact reported, never silently retried.
  `event.isTrusted`, read by a listener installed beforehand, separates
  "the page swallowed it" from "I sent synthetic". Touch has no tool —
  taps approximate as clicks, declared as such.
- **Identify the surface's component library before interacting** — one
  glance at package.json or the imports. Library-keyed operations
  (LIBRARIES.md) apply only where the library is present.

## 2) Transient observation

- **One-shot UI is observed with a script installed before the event.**
  Toasts, notices, and animations that unmount within seconds cannot be
  polled after the fact — inject one self-contained observer
  (`MutationObserver` watching from document start, or an event tap on the
  trigger) before firing the action, and read its record out afterwards.
  After-the-fact polling proves nothing about what flashed by.
- **A concurrent text-wait still races the toast.** Toast lifetimes and
  tool round-trips sit in the same few seconds — one toast survives into
  the next snapshot, the next dies inside it, so a text-wait miss proves
  nothing and a catch proves only luck. The observer is the only
  zero-round-trip instrument: in its callback, harvest the toast-family
  nodes (`role=status`, `role=alert`, the toast library's container) into
  a deduped list, and read that list after disconnecting.

## 3) Reading the tooling

- **Page titles and tab lists lag reality.** The automation UI's page
  inventory updates late — never gate a judgment on what the tab is
  called; gate on the page's own DOM.
- **A hidden tab renders stale or blank frames.** Background tabs get
  throttled painting — bring the tab to front and let a frame settle
  before treating its pixels or screenshot as evidence.
- **The browser console is not the whole truth.** Dev-server terminal
  output carries transform and SSR errors the page console never shows;
  when a page misbehaves, read the server task log too.

## 4) Sessions and personas

- **One persona per isolated context.** Tabs of one profile share a
  cookie jar — walking two users in one profile clobbers their sessions.
  A new context per persona, cookies installed explicitly.
- **The locale cookie outranks negotiation and outlives the session.**
  `Accept-Language` loses to a persisted locale cookie, and that cookie
  survives across walkthrough sessions — set it explicitly before judging
  any locale-sensitive face, or yesterday's locale answers for today.
- **Multi-plane session cookies are plane-keyed.** An app with more than
  one session plane reads a distinct cookie slot per plane — installing
  one plane's cookie and expecting another plane's face to authenticate
  is a walkthrough artifact, not an app bug.

## 5) Screenshots

- **Evidence screenshots: front-most tab, after paint.** Bring the tab
  forward, let the frame settle (a `requestAnimationFrame` or a visible
  settle wait), then capture — a capture from a throttled background tab
  is a phantom.
- **Blob URLs die with their page.** A screenshot or copy that references
  a blob URL rots the moment the page navigates — fetch the bytes and
  copy them out; never link what the page owns.

## 6) Server-truth cross-verification

- **The happy paint is not the receipt.** A walkthrough claim of
  persistence ("it saved", "it unbound", "it filed") closes against the
  server's own stores — the database row, the wire response — read
  where it lives. The face shows the claim; the store proves it. When a
  write is fire-and-forget on the client, poll the store for
  eventual-on-the-wire, and say so.

## 7) Environment state

- **Readiness is the server's line, not the port's silence.** A port
  accepting connections does not mean the app is up — grep the dev
  server's task output for its ready line before walking.
- **Stale optimized deps poison the page.** After a dependency change,
  the dev server may serve a stale pre-bundled cache — hard-reload, or
  clear the cache, before blaming the app for missing exports.
- **Environment spoofing is a walkthrough dial.** Branch-by-environment
  faces (a browser-sniffed flow) are exercised by spoofing the signal at
  the context level — the UA at context creation, the cookie before
  navigation — never by editing the live page.

## 8) Framework probes

- **Check the error vocabulary before diagnosing the face.** A wrong-looking
  toast word is more often the error domain's own table than a rendering
  bug — read the vocabulary the word comes from first; six misdiagnosis
  rounds were once spent skipping this.
- **React inspection tools are discovered per page load.** If the
  inspection session attached after the load, reload once before calling
  the wiring broken; find the component uid first and inspect one —
  dumping hooks wholesale buries the signal. Uids survive re-renders, not
  reloads.
- **Framework state replays from its own serializer.** The router's
  in-page state object can be serialized and re-injected for precise
  re-entry into a navigation bug — cheaper than reconstructing the click
  path by hand.
- **DOM truth and component truth are two instruments.** chrome-devtools-mcp
  reads the DOM (portals, overlays); react-devtools-cdt-mcp reads component
  state — props, hooks, a form library's store. When a library-built
  surface's DOM and behavior disagree, read the store through the inspector
  before blaming the DOM (find the uid, inspect one — above).
