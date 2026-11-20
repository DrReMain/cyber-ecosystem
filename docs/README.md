# Documentation — Map & Law

`docs/` holds the durable knowledge base, written for humans and coding agents at once. This file is the law of what lives where.

Entry points: `AGENTS.md` (root rules + area routing index; `CLAUDE.md` imports it) · `.claude/skills/` (workflow dispatchers) · this file (doc law).

## Lanes — one home per kind of knowledge

| Lane | Home | Holds | Language |
|---|---|---|---|
| **Durable** | `docs/conventions/<area>/CONVENTIONS.md` | rules — what you MUST follow, how the system works | EN |
| **Working** | `docs/roadmap/*.md` | active fronts, in-flight design, acceptance records | ZH |
| **Procedural** | `.claude/skills/<name>/SKILL.md` | dispatchers — a trigger and compressed pointers or procedure; rule text is owned by the durable lane, only cited | EN |

**Single-home law.** Content is never copied — one fact, one home, everything else a pointer. Pointed-at headings are stable anchors: a rename fixes every pointer in the same change.

## The laws

1. **Admission.** A rule enters the durable lane only when it has been applied for real ≥2 times or is enforced by a test/tool — never pre-written for imagined futures, never narrative or decision logs: a durable entry is a pit with its prescription, or a load-bearing rule. A new area directory is born with the first real code of that tech stack; areas are tech stacks only.
2. **Destination.** Every roadmap doc declares near its top where it ends up — backfilled / cleared / archived. `docs/roadmap/` holds only `vision.md` plus the active fronts (target ≤4). A durable rule found in a working doc *moves* out (never copies) — backfill is the only bridge.
3. **Instance state lives in its owning source.** Rule text stays fork-neutral: state the invariant, point at where the current value lives (settings file, compose files, directory tree). Never transcribe a current enumeration into a rule — "every locale registered in the settings file" cannot rot; "all five locales" rots with the first fork.
4. **Roadmap isolation.** The durable lane — `AGENTS.md`, `CLAUDE.md`, skills, conventions — never references `docs/roadmap/`: no paths, no section or ruling citations. A durable rule never depends on a working doc to be understood or complete.

## Sentinel — `./nx run docs:check`

Machine-enforced, exit 1 on any violation: a literal `docs/` or `.claude/` path referenced from AGENTS.md, CLAUDE.md, or any doc or skill body that does not exist (paths into `docs/roadmap/` excepted — the lane is never committed, so its files exist only on machines holding local ledgers); a `docs/<file>.md <section N>` pointer that does not resolve to a real heading; the `AGENTS.md` area index and the conventions directories disagreeing; a conventions doc missing its `**Status:**` line; the durable lane referencing roadmap content. Roadmap content itself is never machine-checked — the destination declaration (law 2) is prose law, held at review.
