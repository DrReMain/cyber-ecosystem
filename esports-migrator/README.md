# esports-migrator

One-shot importer for the esports cutover: reads the legacy MySQL store,
writes the new `system` (identity core) and `esports` (business) Postgres
databases. The rehearsal loads a production backup into local MySQL and
migrates it into local Postgres (no desensitized copy — user ruling
2026-11-02; post-rehearsal cleanup details TBD); the cutover window runs
it once more against production. After cutover this directory is
deleted in a single commit — it owns no product schema, and the
`migration_map` table it leaves behind in the esports database is the
persisted legacy→new id evidence.

Standalone Go module, zero imports from the platform tree, deliberately
outside Nx (user ruling 2026-11-02). Both target databases are written as
raw SQL whose column lists are pinned to the services' Atlas migrations;
the rehearsal's row-count and per-field sampling pass is the drift check
for that seam.

## Run

DSNs live in env files under `~/.xgwl-secrets/esports-mig/` (deploy-plane
facts — never inline in commands, never in the repo): `source.env` holds
`APP_MIG_SOURCE_DSN` (old MySQL); the write phases additionally read
`APP_MIG_SYSTEM_DSN` / `APP_MIG_ESPORTS_DSN` from the same file. The `run`
nx target is the execution plane — direct `go` shell calls are forbidden in
this repo — and it sources the env file itself:

```sh
./nx run esports-migrator:run -- --phase all --dry-run   # census only, no target credentials needed
./nx run esports-migrator:run -- --phase all             # guard + identity + entitlement + behavior (later cuts add phases)
```

> TODO(mig1 temporary instrumentation, teardown listed): `project.json` is a
> rehearsal-run-up wrapper only — the module stays deliberately outside the
> build graph (user ruling 2026-11-02). Tear it down with the directory's
> post-cutover single-commit removal; it must not outlive the tool.

Phases: `guard` (halt on any school-domain rows), `identity`, `entitlement`
(legacy unlock flags, activation records, and unredeemed codes → grant
ledger + bare code rows), `behavior` (game preferences → profile.games,
test attempts → the fact stream with detail-else-summary payloads; the
retired `skill_prediction` rows drop, login_record stays cold — no
new-schema home — and user_referral waits on the pending referral ruling),
and `reconcile` — the read-only closing assertion pass (count identities
scoped through `migration_map`, full-column invariants such as
grade=band(score) and re-derived attempt ordinals and age bands, per-field
sampling with canonical JSON comparison; `--sample N` sizes it). `snapshot`
waits on the legacy-report carrier ruling and `recompute` on the
calibration-curve window (both filed as P26/P26.1). Every phase is
idempotent — re-running after a failure resumes where the durable markers
left off.

## Checks

```sh
./nx run esports-migrator:check
```
