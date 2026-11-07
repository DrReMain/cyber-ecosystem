# Browser walkthrough law card

- Before any simulated walkthrough / live-proof pass / browser verification: read `docs/conventions/browser-walkthrough/CONVENTIONS.md` in full — tool-bound to chrome-devtools-mcp.
- Traps: `fill` doesn't fire React onChange — native value setter + `dispatchEvent("input")`; CDP click first, programmatic `.click()` only as fallback; icon-only controls match by `aria-label`, never text; checkboxes — `.click()` the input body, found via its enclosing row's text (§1.1–1.2).
- Traps: toasts/countdowns observed inside one evaluate_script (trigger → sleep → query), container selector whole, never an invented wordlist (§2.1); a "silent" verdict requires every feedback surface enumerated empty first (§2.2).
- Traps: broken page — read the dev-server task log before dissecting the DOM (§3.3); in-memory handoffs don't survive navigation — same mount for sms send+submit, dev `import()` of the live module + SPA nav for singletons, never a hard reload (§7.2); an SSR error face is a server-chain symptom and boundary retry never re-runs the loader (§4).
- Traps: time-critical runs (reaction windows) are seam-driven — injected clock/instance controls, never real-time input (§1.4).
- Close a walk by writing newly-diagnosed traps back into the conventions file verdict-style, the moment they are understood.
