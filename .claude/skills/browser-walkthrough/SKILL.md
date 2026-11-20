---
name: browser-walkthrough
description: Live browser verification loop — use when a walkthrough of web client surfaces is due (closing a client change per the reachable-surface rule, or a campaign verification session), and when driving a live browser (chrome-devtools-mcp class tooling) to observe or interact with running pages. Trigger phrases include walkthrough, 走查, verify live, browser loop, observe the running app.
---

# Browser walkthrough (the live verification loop)

Read first: the read map in `docs/conventions/browser-walkthrough/CONVENTIONS.md` — the judgement corpus this loop walks under.

## Pre-flight

1. Stack reality: every service and dev server the face needs is up; readiness is the server's own ready line in its task output, never port silence.
2. One persona per isolated browser context; session cookies installed explicitly — credentials come from the owning campaign's harness notes, never from repo docs. Set the locale cookie before the first navigation.
3. Identify the surface's component library (package.json, imports) — library-keyed operations live in `docs/conventions/browser-walkthrough/LIBRARIES.md`.
4. If the walkthrough writes (fixtures, binds, state changes), plan the restore BEFORE walking — note the exact rows/keys touched and the restore statements, so drift returns to zero.

## The loop

4. Navigate, then wait on hydration before interacting (input value tracker / loader RPC fired — the stack's verification law names the two shapes).
5. Transient UI: install the self-contained observer script before firing the event; never poll after the fact.
6. Interact through the input tools first — CDP-trusted, isTrusted, the human-like baseline (per-key faces take `type_text` over `fill`); programmatic driving and refusals are descents, named in the report (§1).
7. Cross-verify every persistence claim against server truth — the database row or the wire response, read where it lives; a fire-and-forget write polls the store until eventually-on-the-wire.
8. Evidence: screenshots from the front-most tab after paint; blob bytes copied out, never linked. Declaring a page clean reads both the browser console AND the dev-server task log.

## Teardown and close

9. Run the planned restores; re-read what was touched to prove zero residue; discard the persona contexts.
10. The report declares: faces walked, interactions observed, server-truth receipts, zero-residue proof, console and task-log cleanliness. A pit caught mid-walk is pinned into the CONVENTIONS the same day, judgement form, business-neutral.
