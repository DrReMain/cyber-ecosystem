---
paths: ["app/clients/**", "shared-ts/**"]
---

# Tanstack law card

- Before writing: load the dispatch row for this artifact — docs/conventions/tanstack/CONVENTIONS.md §0 "Execution contract".
- Before reporting done: run docs/conventions/tanstack/CONVENTIONS.md §12 "Pre-completion gate" over the combined diff and emit its report table.
- Read-first by surface: §1 (layering & transport JSON shapes), §3 (route tiers & the read-set contract), §5 (component & naming grammar), §6 (i18n ownership).
- Surfaces reachable without login need e2e verification before done (§11 reachable-surface rule).
