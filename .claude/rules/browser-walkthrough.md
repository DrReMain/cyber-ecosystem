# Browser walkthrough law card

- Before walking: read the read map at the top of `docs/conventions/browser-walkthrough/CONVENTIONS.md` — it names the sections that govern the occasion.
- Trap: interactions wait on hydration first — the input value tracker, or the loader RPC having fired; a pre-hydration click is a lost event (§1).
- Trap: the interaction tools ride CDP input (isTrusted) — the human-like baseline; programmatic el.click() / the native-setter pair is a named descent, a refusal records the natural timeout (§1).
- Trap: one-shot UI gets a self-contained observer installed before the event, never after-the-fact polling — and a concurrent text-wait still races the toast (§2).
- Trap: evidence screenshots come from the front-most tab after paint; blob assets are copied out, never linked (§5).
- "It persisted" closes against server truth — the store or the wire, not the happy paint (§6).
- A pit caught in the wild is pinned in the doc the day it is caught, judgement form, business-neutral.
