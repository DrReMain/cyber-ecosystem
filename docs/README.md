# Documentation — Map & Governance Law

This directory is the durable knowledge base of the skeleton, written for two audiences at once: humans and coding agents. This file holds the map and the law — what lives where, and how content moves between homes.

Entry points:

- **`AGENTS.md`** (repo root) — root rules + area routing index; auto-loaded by coding agents. `CLAUDE.md` is a Claude Code import shell over it.
- **`.claude/rules/`** — path-scoped delivery shells; injected automatically when Claude reads or edits a file matching the frontmatter `paths` globs; pointer-only (§2.9).
- **`.claude/skills/`** — procedural workflow dispatchers; Claude Code triggers them by description, everyone else reads them as plain markdown.
- **`docs/environment.md`** — the recommended agent environment: optional tooling, each entry naming what its absence degrades; a root-level durable doc (no stack area owns it).
- **This file** — the governance law; read on demand.

---

## 1) Three lanes — one home per kind of knowledge

| Lane | Home | Holds | Language | Lifecycle |
|---|---|---|---|---|
| **Durable** | `docs/conventions/<area>/CONVENTIONS.md` | rules and mechanisms — what you MUST follow, how the system works | EN | long-lived; changes follow the paradigm-change discipline (AGENTS.md §6) |
| **Working** | `docs/roadmap/*.md` (single doc) or `docs/roadmap/<front>/` (directory: `README.md` front index + sub-docs split by concern) | in-flight design, milestone acceptance records, backlog pools — one front = one doc or one doc set, plus `vision.md` (north-star) | the owning team's language — any, per front (§4) | temporary by contract: every item carries a destination (§2.4) |
| **Procedural** | `.claude/skills/<name>/SKILL.md` | workflow dispatchers — trigger, pointers, cross-doc assembly, one-line trap annotations (warning + home pointer, no steps); never rule text or step text | EN | evolves with code; git-versioned; directory names are stable identifiers |

**Single-source law.** Content is never copied. Where two audiences need the same knowledge, both read the same file. A skill contributes only a trigger, pointers, and one-line trap annotations (a warning naming the trap and its home section — never the rule text or the steps); procedures (checklists) live in docs, and skills point at them. Duplication is the one failure mode this law exists to prevent — one fact, one home, everything else a pointer.

**Root tier of the Durable lane.** Durable knowledge with no stack area lives in the root rules (`AGENTS.md`) when it is rules, or as a root-level `docs/<topic>.md` registered in the entry list when it is a recommendation layer — currently `docs/environment.md`. This is the Durable lane's root tier, not a fourth lane.

---

## 2) The laws

### 2.1 Admission

- A **rule** enters the durable lane when it has been applied for real ≥2 times, or is enforced by a test/tool. Do not pre-write rules for imagined futures.
- A **workflow** earns a skill when it has been walked ≥2 times.
- A **new area directory** is born with the first real code of that **tech stack** — never pre-created empty. Areas are tech stacks only (kratos, tanstack, proto, deploy); a topic or domain does not get its own area — its rules live in the owning stack doc or the root rules.

### 2.2 Promotion — the same bar applied upward

- A rule with one consumer lives in the most specific place that serves it. On the **second real consumer**, it *moves* (never copies) to the shared layer:
  - stack-agnostic client rules graduate from `docs/conventions/tanstack/` to a shared client area when a second client stack (RN, Taro, …) lands;
  - capability rules have no area of their own: dependency architecture (self-containment) lives in the root rules (`AGENTS.md` §4), service-facing usage (error choke points, injection) in the `kratos` stack doc — there are no non-stack areas to graduate to;
  - durable rules found trapped in a working doc are backfilled immediately (§2.3), regardless of milestone state.
- Doc promotion mirrors code promotion (the shared-ts/shared-go second-consumer triggers) and ships in the same change.

### 2.3 Backfill — front closure

When a milestone (or a ruling) closes:

1. Extract durable rules into the owning CONVENTIONS section.
2. Mark the source item `backfilled <date> → <target>`.
3. Prune acceptance logs to one-line closure notes (commit hash + verdict); long histories are archived or deleted.

### 2.4 Destination

Every working-lane item declares where it ends up: **backfilled / cleared / archived**. The declaration is a `> destination:` line near the top of the front's index — the doc itself for a single-doc front, the `README.md` for a directory front: an English protocol keyword (§4) whose payload is the team's own language. A closed front's doc (or directory) is pruned or deleted; `docs/roadmap/` holds only `vision.md` plus the active fronts (target: ≤4 fronts — a directory front counts as one front, its `README.md` is the index).

### 2.5 Pointer discipline

Content moves; it is never duplicated. The one sanctioned instantiation is a skill's gates line — the owning chain's concrete commands, instantiated from the conventions that stay the source; rule shells carry no commands at all. Sections that skills point to owe **stable anchors**: headings are not renamed casually; a rename fixes every pointer in the same change (`docs:check` turns red otherwise).

### 2.6 Drift checking

`./nx run docs:check` verifies:

1. every file path referenced by `AGENTS.md` / `CLAUDE.md` / the root READMEs / skills exists;
2. every heading targeted by a skill pointer exists;
3. the `AGENTS.md` area index ↔ `docs/conventions/*/CONVENTIONS.md` agree in both directions;
4. durable docs carry a `Status:` line; a working front's index (top-level doc or `<front>/README.md`) carries a `destination:` line;
5. the durable lane is roadmap-clean (§2.7): no `docs/roadmap/` reference in `AGENTS.md`, `CLAUDE.md`, skills, or conventions;
6. every gate row ID (`K-XXX-N` / `T-XXX-N`) cited by a skill or rule shell exists in the owning conventions gate (kratos §7, tanstack §12).

The sentinel walks all of `docs/` (conventions and roadmap) plus `AGENTS.md`, `CLAUDE.md`, the root READMEs, and every skill and rule-shell body.

This extends the definition of done (AGENTS.md §6 item 6): a paradigm/mechanism change is done when the durable lane is synced and affected skill pointers verify green.

### 2.7 Roadmap isolation & conventions self-containment

The working lane is temporary by contract and **fully isolated**: `AGENTS.md`, `CLAUDE.md`, skills, and conventions `MUST NOT` reference roadmap content — no `docs/roadmap/` paths, no section or ruling citations. A durable rule never depends on a roadmap doc to be understood or complete; backfill (§2.3) is the only bridge, and it *moves* text out of the working lane. Active fronts are introduced by the user in conversation, not by root docs. Conventions record rules only — self-contained, no narrative or decision logs (`docs:check` check 5 enforces the path ban).

### 2.8 Instance state lives in its owning source

The skeleton forks: the service set, client set, locale set, ports, and middleware dials are per-fork values a deployment can change without a rule changing. Conventions state the invariant and point at where the current value lives (settings file, compose files, directory tree); rule text stays fork-neutral. Never transcribe a current enumeration into a rule — "copy in all five locales" rots the moment a fork registers one or ten; "copy in every locale registered in `project.inlang/settings.json`" cannot rot. A current value may appear only as orientation that names its source ("currently …; the compose files are the truth"), never as the rule. `e.g.`-marked examples of a rule's shape are fine; enumeration as the rule's content is drift by construction.

### 2.9 Delivery shells (`.claude/rules/`)

A rule shell is the path axis of delivery: frontmatter `paths` (repo-tree globs) decides when its card injects — automatically, as Claude reads or edits a matching file. The body holds pointers and one-line trap annotations only (same content law as skills, §1); the conventions doc it points to stays the single source. English. Dead pointers are `docs:check` failures; directories with no card route through the area index (`AGENTS.md` §7) as before.

### 2.10 Enforcement admission

A verification mechanism (hook, CI gate, checker script) is admitted only on evidence: a demonstrated silent-failure face — the failure it catches passes every other gate silently — or ≥2 real misses of the standing law it would enforce. Mechanisms are cost: rarely-changing areas get no standing mechanism. Evidence declarations (`AGENTS.md` §6.1) are reviewed, never regex-linted — the vocabulary is dense with domain false-friends.

---

## 3) Reading order (cold start)

1. `AGENTS.md` — root rules and the routing table.
2. The CONVENTIONS of every area you will touch — **before** writing code there.
3. The roadmap doc of the active front you work on — for in-flight context; its landed sections are history, its design sections are contract-until-backfilled.
4. `.claude/skills/` — workflows fire automatically for Claude Code; humans read them as recipes.

---

## 4) Language law

English is the platform's first language: durable and procedural docs are written in English, and so are the protocol keywords mechanisms verify — `Status:` lines, gate row IDs (`K-XXX-N`), the working lane's `destination:` marker. Protocol keywords are to docs what identifiers are to code: fixed vocabulary, regardless of the prose language around them. The working lane is written in the owning team's language — any language, chosen per front; the platform constrains nothing there. Roadmap isolation (§2.7) is what makes this safe: working-lane language never leaks into the durable surface, and the public snapshot strips the lane entirely (github-publish exclusion). Root README translations (`README.zh-CN.md`) are courtesy mirrors of the English source, kept section-for-section in lockstep.
