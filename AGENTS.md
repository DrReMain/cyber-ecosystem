# Cyber Ecosystem

Monorepo for the Cyber Ecosystem platform — a **generic development skeleton**: one shared base, business apps fork and grow on top. Protobuf contracts (`proto/`) are the source of truth; `gen/` is derived output, never hand-edited.

| Path | Holds |
|---|---|
| `proto/` | protobuf contracts — the source of truth |
| `gen/` | generated code (Go, connect-ts, OpenAPI) — derived |
| `app/services/<name>/` | Kratos Go services; `system` = the mandatory base (IAM) |
| `app/clients/<name>/` | TanStack Start web clients (`admin`) |
| `shared-go/` / `shared-ts/` | cross-service capabilities — never business aggregates |
| `deploy/` | compose stacks, edge, migration wiring |
| `tools/` | repo bootstrap and tooling targets |

## Hard laws

1. **Workflows run through Nx** — `./nx run <project>:<target>`; every valid target is declared in the project's `project.json`. Package-local binaries go through `pnpm exec <bin>`. A workflow with no target gets one added, never an ad-hoc command.
2. **pnpm is the only package manager.** A `package-lock.json` or `yarn.lock` anywhere means a command bypassed pnpm — delete the artifact, fix the command.
3. **Never hand-edit `gen/`.** Fix the source or generator and regenerate (proto conventions §7).
4. **No cross-service code dependencies.** Cross-service reads go through the owning service's RPCs, behind a `<Remote>RP` ACL adapter (kratos conventions §5); capability packages stay self-contained (kratos conventions §1).
5. **No secrets or environment topology in code** — endpoints and keys are deploy-plane facts (deploy conventions §6).

## Working contract

- **Read before write.** Each convention opens with a read map: the artifact you are about to touch names the sections that are its law. Read those sections — the text itself, not your memory of it.
- **Structural actions verify first.** Branch, commit, mirror, delete: check the standing law (branch discipline, face separation) before acting, not after.
- **Gaps are questions.** Where no law speaks, or the call is an author's rather than a mechanic's, stop and ask. A choice you cannot trace to a law line is a question, not a decision.
- **Close with the counted declaration.** One line stating what ran and what was observed: chain green, probes counted, residues zero.

## Conventions

| Area | Doc | Read when |
|---|---|---|
| proto (contracts) | `docs/conventions/proto/CONVENTIONS.md` | touching any `.proto`, or code consuming `gen/` |
| kratos (Go services) | `docs/conventions/kratos/CONVENTIONS.md` | writing Go under `app/services/*` or `shared-go/` |
| tanstack (web clients) | `docs/conventions/tanstack/CONVENTIONS.md` | writing under `app/clients/*` or `shared-ts/` |
| deploy (containers) | `docs/conventions/deploy/CONVENTIONS.md` | touching `deploy/`, Dockerfiles, compose, edge, migrations |
| skin (antd presets) | `docs/conventions/skin/CONVENTIONS.md` | touching `shared-ts/antd/src/skin-presets/**` or authoring a skin |
