# Kratos runbooks

Procedures for rare, tool-heavy moments — the law lives in `CONVENTIONS.md` (§0 routes here). Every runbook closes with its gate rows (§7 item IDs); a runbook never restates law.

## 1. New domain module (aggregate root)

1. `module/<d>/biz.go`: `DO` → `Port` → `UC` → `Method` (→ `Private`); UC embeds `shared.UC` (`uc.Tm` / `uc.Log`).
2. ent `schema/<do>.go`: fields + mixins; `./nx run <service>:generate:ent`, `:migrate:diff`, `:migrate:apply` (persisted only).
3. `module/<d>/data.go`: `<do>RP` + `New<DO>RP` + `map<DO>` (embed `shared.RP`, use `rp.Platform`). **Omit** this file for a pure ACL facade (§3 of this file).
4. `module/<d>/service.go`: `<Name>Service` + handlers + Registrar. `module/<d>/module.go`: `ProviderSet{New<Name>UC, New<DO>RP, New<Name>Service}`.
5. Wire it in: add the module's `ProviderSet` to `wireApp`; add its `*<Name>Service` param to `server.NewRegistrarList`.
6. `./nx run <service>:generate:wire` then `:build`. If it imports another module, keep it a **read** (port query) — cross-module writes go via domain event. Verify dependency direction (UC→port; UC→UC only toward foundation modules — user/authz/file). **Wire prunes unconsumed providers**: a startup-shaped component (seed job, engine watch loop) must hang off a live dependency chain — a `newApp` param or self-registration on `shared.HookRegistry` in a UC constructor — or it silently never instantiates.

Gate: full §7 (rows K-NAM K-STR K-CMT K-LAY K-UTL K-ERR).

## 2. Remote-service dependency (treat as a repo)

1. biz: declare `<Remote>RP` with neutral ACL payloads; inject into the UC like any repo.
2. client `<remote>.go`: `<remote>Client` + `New<Remote>Client` + `map<ProviderDO>` over one connection. Register `New<Remote>Client` in `client.ProviderSet`. One file per remote.
3. conf `conf.proto` `Remote`: add endpoint field; `configs/config.yaml`: set it. `./nx run <service>:proto:conf`.
4. `./nx run <service>:generate:wire` then `:build`. Verify biz never imports the provider's proto.

Gate: rows K-NAM K-LAY K-ERR (plus K-CMT K-UTL on the added files).

## 3. Pure ACL facade (no own persistence)

The module has **no `data.go`**; its UC depends on a `<Remote>RP` satisfied by the `client` layer. Its `module.go` `ProviderSet` lists only `New<Name>UC` + `New<Name>Service`.

## 4. Deriving a new service

From `system`, the mandatory base and reference skeleton:

1. **Copy-set**: `project.json` / `Dockerfile` / both `config.yamls` / `conf.proto`(+buf.gen.conf) / `cmd/app` / `bootstrap` (lifecycle only, no seed) / `shared` (kernel trimmed + `dependency_guard_test` + `conventions_guard_test` rewritten) / `platform` (trimmed to the module set — §5 platform law) / `server` (five files) / `client` (once the service calls out) / `ent` schema + `local_mixins`. Deliberately NOT copied: `seed.go`, system business modules, the authz engine internals (casbin/attrs/policies), audit.
2. **Dials**: the port triple advances **+1 per service** (system 11001/12001/13001 → agent 11002/…; the service configs are the truth). A service assembling redis isolates its db index from non-service consumers of the same redis (the current pairing lives in the compose files).
3. **ent first generation is a two-pass stub dance**: `local_mixins` import the generated tree (chicken-and-egg), so stub the mixins (nil `Hooks`/`Interceptors` over entgo-level types) → generate → restore the real mixins → regenerate. entc also refuses an empty schema package ("no Go files"), so the derive slice must land its first aggregate's schema — deferring all schemas to the data slice is not an option. After removing an entity, nuke the whole generated tree first — the loader reads the stale tree.
4. **wire cold start**: the `go:generate` directive lives in `wire_gen.go` itself — a fresh service runs `go run github.com/google/wire/cmd/wire` manually; wire rejects interface bindings with no consumer in the graph (comment the bind out until its consumer exists).
5. **First migration**: `NAME=init <svc>:migrate:diff` → hand-prepend `CREATE EXTENSION` lines the atlas scratch already has (a service database owns its extensions) → `:migrate:hash` → `:migrate:apply`. Recreate the service database rather than fighting a drifted init. Squashing the directory into one init requires resetting the atlas dev database first (`DROP DATABASE atlas_dev; CREATE DATABASE atlas_dev`) — the diff runs against its residual schema, so a stale dev database yields a delta fragment instead of the full schema.

Deploy wiring follows: `docs/conventions/deploy/CONVENTIONS.md` §7. Gate: full §7 over the derived tree.

## 5. Operating facts

### Editor diagnostics lag codegen

After ent/proto/wire regeneration the editor's LSP frequently reports `undefined` on freshly generated symbols (stale index). The compile truth is `go build` / the Nx targets, never the diagnostics panel: red LSP + green build means a stale panel, not a broken tree.

### Live API probe

REST probes hit the HTTP gateway port (11xxx; the `google.api.http` routes); the connect port (13xxx) serves only `/connect/<pkg.Service>/<Method>`. A login's `session_token` rides the response body — the HttpOnly cookie is a web-client concern. The gateway binds nested query params dotted (`page.page_no`) and renders wrapper fields flattened (`nickname`, not `nickname.value`).
