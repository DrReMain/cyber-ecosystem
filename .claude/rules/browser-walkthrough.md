# Browser walkthrough law card

- Before any simulated walkthrough / live-proof pass / browser verification: read `docs/conventions/browser-walkthrough/CONVENTIONS.md` in full — it is tool-bound to chrome-devtools-mcp.
- Trap: the fill tool does not trigger React onChange — drive controlled inputs with the native value setter + `dispatchEvent("input")` (§1.1); programmatic `.click()` is the fallback after the CDP click fails, not the first move (§1.2). Icon-only controls match by `aria-label`, never by text (§1.2). Checkboxes: click the input body, located via its enclosing row's text — index picks and label clicks are unreliable (§1.1).
- Trap: transient feedback (toasts, countdowns) MUST be observed inside one evaluate_script (trigger → sleep → query), reading the toast container selector whole — never filter by an invented copy wordlist (§2.1). A "silent UI" verdict requires every feedback surface enumerated empty first (§2.2).
- Trap: for a broken page, read the dev-server task log before dissecting the DOM (§3.3).
- Trap: client in-memory handoffs don't survive navigation — keep sms send+submit in one mount (no nav between calls); inject module singletons via dev `import()` of the live instance + SPA-side nav, never a hard reload (§6.2).
- Close a walk by writing newly-diagnosed traps back into the conventions file verdict-style, the moment they are understood.
