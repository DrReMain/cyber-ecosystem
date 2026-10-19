# Custom PostgreSQL Image (pgvector + PostGIS)

Extended PostgreSQL image based on `postgres:18.6-trixie` with pgvector and PostGIS pre-installed.

## Build

```bash
./nx run deploy:image:pg-extended
```

The target carries an existence guard: it builds only when `pg-extended:18.6` is absent locally, so a fresh clone gets the image on its first infrastructure start without a manual step. Every start target that activates the `db` profile (`deploy:db:start`, `deploy:start`, `deploy:observability:start`, `deploy:app:start`, `deploy:pre:start`, `deploy:pre:full:start`) runs it via Nx `dependsOn` before `docker compose up`, so compose never has to resolve this local-only tag against a registry.

The guard skips present images: after editing this Dockerfile under the same tag, remove the stale image first (`docker image rm pg-extended:18.6`) and rerun the target. A base-image bump changes the tag itself, which the guard misses naturally.

## Usage

The dev stack references this image by name (`deploy/docker/compose.dev.yaml`, service `postgresql`); the service also keeps a `build` section as provenance and last-resort fallback.

The init SQL (`deploy/docker/postgres/01-extensions.sql`) uses fault-tolerant `DO` blocks that create the `vector` and `postgis` extensions on first initialization; with a vanilla `postgres` image they are silently skipped.

For the k8s deployment shape, load the image into the cluster's registry or a private registry the cluster can pull from.
