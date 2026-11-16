# Tanstack law card

- Before writing: read the read map at the top of `docs/conventions/tanstack/CONVENTIONS.md` — it names the sections that govern the artifact.
- Read-first by surface: §1 (layering & transport JSON shapes), §3 (route discipline), §5 (component file discipline), §6 (i18n ownership).
- Trap: transport responses are JSON-shaped, not message instances — consume accordingly (§1).
- Surfaces reachable without login need live browser verification before done (§11 reachable-surface rule); close with the evidence declaration (§11 Close).
