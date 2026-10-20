---
paths: ["proto/**/*.proto", "gen/**"]
---

# Proto law card

- `gen/` is derived output — never hand-edit it; fix the source and regenerate (`AGENTS.md` §4).
- Read first by concern — docs/conventions/proto/CONVENTIONS.md §3 (optionality & the null/zero contract), §4 (error enums, reserved law, locale copy), §5 (access annotation is mandatory on every business RPC).
- Changing a contract: follow the §8 checklist exactly — it carries the gates.
