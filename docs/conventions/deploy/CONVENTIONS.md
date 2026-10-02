# Deploy Conventions

**Status:** ACTIVE

**Scope:** containerization and deployment of everything in this repo — app images, `deploy/`, the compose stacks, the edge. Prescriptive — `MUST` / `SHOULD` / `MAY`.

This repo is a **generic skeleton**: every rule below is a per-app pattern, not tied to any specific app. Current instances: one kratos service (`app/services/system`), one tanstack client (`app/clients/admin`) — when they are renamed, removed, or joined by more apps, the rules apply unchanged. Instance state (service names, URLs, ports) lives in the compose files and `deploy/docker/traefik/dynamic.yaml` themselves — read those for what exists, read this for how it fits together.

---

## 1. Shapes & dials

| Shape | Stack | Access |
|---|---|---|
| dev | `deploy/docker/compose.dev.yaml` (middleware) + apps bare-metal | host-direct ports |
| pre (= local prod rehearsal) | dev stack + `deploy/docker/compose.app.yaml` + `deploy/docker/compose.edge.yaml` | edge-only, `*.cyber.test` |
| prod compose / prod k3s | parameterized compose / k8s translation (`deploy/k8s/`, empty placeholder) | edge / ingress |

- pre has two middleware dials: **minimal** (`deploy:pre:start` = the app layer's hard deps only — currently db + storage + app + edge; mq runs in pg mode, so NATS/realtime/media/observability are all optional) and **full** (`deploy:pre:full:start` = all six middleware profiles + app + edge).
- pre runs **on top of** the dev stack (one compose project, shared volumes) — deliberate; split only when data isolation actually hurts. The edge, by contrast, is its **own compose project** joining the middleware network as `external` — it can only start after the middleware stack exists (the `pre:*` targets encode that order).
- Named volumes persist across `down`; only `down -v` (`deploy:reset`) wipes them.
- `MUST` keep the shapes aligned: same discovery names, same edge semantics, same images; compose artifacts carry **zero k8s residue**.
- App containers publish **no host ports** in pre — the edge is the only way in.

## 2. Usage (all through Nx)

| Target | Effect |
|---|---|
| `<app>:image` (on each app project) | build that app's image → `cyber-ecosystem/<name>:local`; follows host `GOPROXY` / npm registry |
| `deploy:start` / `deploy:stop` | middleware core (db+storage+mq) |
| `deploy:<domain>:start/stop` | middleware per domain (`db storage mq realtime media observability`) |
| `deploy:app:start/stop/status` | app layer + its hard middleware deps; start builds the app images first (Nx `dependsOn` → `<app>:image`, version-stamped) — compose `up` never builds, so the stamp survives |
| `deploy:pre:start` / `deploy:pre:full:start` | pre, minimal / full dial (app layer + edge) |
| `deploy:pre:stop` | edge down + app containers **removed** (middleware untouched) |
| `deploy:reset` | both compose projects down `-v` — clean slate, images and unrelated projects kept |

App-layer stop is **remove, not stop**: `deploy:stop` (dev-file-only `down`) removes the shared network, and stopped app containers left referencing it make the next `up` fail with "network not found".

Every `app:start` **rolls all app containers**: docker builds stamp a fresh timestamp into the image config, so even a content-identical rebuild yields a new image ID and compose recreates every app service (a seconds-long blip). Per-service rollout granularity, if ever needed, is a per-service `up`.

## 3. Ownership: build self-contained, orchestration central

- Each app owns its image definition: `Dockerfile` + `image` Nx target at the app root. Adding an app touches its own project plus §7's deploy-side wiring — nothing else.
- `deploy/` owns orchestration: compose files, edge, migration wiring, start/stop targets.
- `deploy/images/<name>/` = infra-derived custom images only (e.g. `pg-extended`), never app images.
- Both image kinds use the **workspace root as build context** (single Go module + pnpm workspace). The root `.dockerignore` is the **single trim surface** — `MUST NOT` scatter per-directory ignores.
- pnpm-workspace installs inside image builds `MUST` run with the **full workspace source present** (`COPY` before `install`; a BuildKit store cache mount recovers the lost layer cache). A manifests-only install layer changes pnpm's node_modules shape, and vite's server environment then hashes `?url` css assets differently than the client emits — phantom css names that 404 at runtime (first-paint flash). Verified both directions; the admin Dockerfile comment marks the spot.

## 4. Discovery names

- Discovery is **by naming discipline, not a registry**: every compose service carries a `<service>.<domain>` network alias (`postgresql.db`, `system.api`, `admin.web`). These are exactly the two-label short names k8s Service DNS resolves — a container-mode config (§6) serves every container shape unchanged.
- `MUST NOT` introduce a service registry / config center at the current scale. Revisit triggers: (a) ≥3 mutually-calling services with dynamic topology, (b) a real need to change config without a release.

## 5. Ports

- Middleware publishes host ports (dev direct access); the app layer publishes nothing.
- The edge never rewrites `Host` (S3 presign SigV4 depends on it) and terminates nothing (plain :80; TLS is §9).

## 6. App configuration

- **kratos services** keep two bootstrap configs of identical shape at the app root: `configs/config.yaml` (bare-metal dev, `localhost` endpoints) and `container/config.yaml` (network aliases, baked into the image). **Mirror every key change across the two.** `configs/` is the dev source directory — kratos reads *every* file in it, so it MUST contain nothing but the dev config.
- Deployment-sensitive values are kratos `${VAR:default}` placeholders, overridden at runtime by `APP_`-prefixed env vars (`APP_DB_HOST=…`). This is the **only** env-override channel: kratos's env source does flat keys, not nested overrides — `APP_DATA_DATABASE_HOST` is silently discarded; the yaml placeholder is what makes an env var reachable. Booleans included (`WithResolveActualTypes` is wired in `main.go`); keep placeholders opt-in — only keys a deployment actually flips.
- **tanstack clients** read runtime env only (API URL, PORT) and derive the site origin per request (`getSiteUrl`: request URL server-side, `window.location.origin` client-side) — nothing environment-specific is baked; one image serves every deployment. A `server.mjs` host mounts the fetch-handler build output (the vite plugin emits `export default { fetch }`, not a self-starting server): zero-dep node:http + static client serving + WHATWG bridging.

## 7. Adding an application

1. App side (self-contained): `Dockerfile` + `image` target at the app root — use the existing kratos service / tanstack client as the template; kratos adds `container/config.yaml` (§6).
2. `deploy/docker/compose.app.yaml`: service block under the `app` profile — `<service>.<domain>` alias, healthcheck (kratos: `/healthz` on its http port; client: a static asset), `depends_on` (incl. its migration job, if any), resource limits, `image: cyber-ecosystem/<name>:local`, **no host ports**.
3. Edge (`deploy/docker/traefik/dynamic.yaml`): `Host(<name>.cyber.test)` → the service; for clients calling same-origin `/connect`, add the `Host && PathPrefix(/connect)` rule → the target service's connect port + `stripprefix` middleware (mirrors the dev vite proxy; length-based priority needs no explicit ranks).
4. Migrations (kratos + ent): one-shot atlas job (§10).
5. Local pre access: one hosts entry per edge router — the `cyber.test` line in `/etc/hosts` mirrors `dynamic.yaml`'s router set, extend both together.

## 8. The presign dual-path pattern

A service that hands **presigned URLs to browsers** (S3 today) must sign against the *external* name, while the same URL is also used server-side inside the network:

1. the external edge name doubles as an **in-network alias** on the target container (`s3.cyber.test` on seaweedfs);
2. the target listens on container port **80**, so the port-less URL connects on both paths (browser → hosts/DNS → edge; in-network → alias → container);
3. the host-port publish adapts (`8333:80` keeps dev's `localhost:8333` working).

`MUST` apply this pattern before pointing any presign-capable endpoint at an edge name.

## 9. Stances

- **Edge engine: traefik** (measured 17MB with the file provider; k3s first-party ingress). Switch only on: k8s support dropped from the product plan; a ≤2GB customer VM where the footprint is decisive; edge complexity outgrowing the file provider. Pre-TLS, when needed, prefers mkcert certs mounted into traefik.
- **CI: intentionally none in the skeleton phase.** The eventual shape is GitHub Actions on a backup mirror (origin is GitLab) and/or an **offline install bundle** (images + parameterized compose + scripts, USB/file-transfer delivery to customer LAN servers); the Nx image targets and placeholder parameterization are its foundation.
- **Environment matrix** (no-domain / intranet IP / plain http): minimal mode (§1) is the resource-limited-server baseline; full offline delivery lands with the install bundle.
- **k8s manifests:** keep `deploy/k8s/` empty until the compose shapes settle.
- **Build args** (`GOPROXY`, `NPM_REGISTRY`, `VERSION`, client site URL): defaults stay official/neutral; hosts override at invocation.

## 10. Migrations

- Schema migrations run as a **one-shot compose job**: `arigaio/atlas:<version aligned with the local @ariga/atlas CLI>` (the official image; the npm package serves only the bare-metal dev workflow), bound read-only to the service's `internal/ent/migrate/migrations` directory — an offline bundle ships that directory alongside the compose files.
- The job precedes its service via `depends_on: service_completed_successfully`, so `--wait` stacks wait for schema readiness too. On a retained volume the job is a no-op ("No migration files to execute").
- A fresh postgres volume is bootstrapped by `deploy/docker/postgres/*.sql` (extensions + database shells); database names are shared facts with the consuming services' configs. Schemas are owned by their consumers: atlas migrations for kratos services, first-connect for the mq capability, setup jobs for the observability tools. Extension ownership follows the same consumer line: init SQL covers the default database and the diff scratch `atlas_dev` (fault-tolerant, skipped on a vanilla image); a service database's extensions land in its own first atlas migration.
