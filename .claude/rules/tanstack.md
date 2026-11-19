# Tanstack law card

- Before writing: read the read map at the top of `docs/conventions/tanstack/CONVENTIONS.md` — it names the sections that govern the artifact.
- Read-first by surface: §1 (layering & transport JSON shapes), §3 (route discipline), §5 (component file discipline), §6 (i18n ownership).
- Trap: transport responses are JSON-shaped, not message instances — consume accordingly (§1).
- Trap: e2e interactions after a cold `goto` wait on hydration first — the input value tracker, or a loader-RPC wait armed before the navigation; a suite with real-stack legs closes serial (`--workers=1`) (§11).
- Spec first: logic-face tables (`<file>.test.ts` beside flat `.ts` modules) land before the module body fills in — code aligns to the table, never the reverse (§11).
- Surfaces reachable without login need live browser verification before done (§11 reachable-surface rule); close with the three-face declaration (§11 Close).
