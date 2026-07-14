# Custom PostgreSQL Image (pgvector + PostGIS)

Extended PostgreSQL image based on `postgres:18.3-trixie` with pgvector and PostGIS pre-installed.

This directory is independent of the repository's Nx toolchain. Build and import are manual operations.

## Build

```bash
docker build -t pg-extended:18.3 .
```

## Import into k3d

```bash
k3d image import pg-extended:18.3 -c cyber-ecosystem
```

## Usage

`deploy/k8s/db/postgresql.yaml` contains two image lines — one active, one commented:

```yaml
          image: postgres:18.3-trixie       # vanilla (default)
          # image: pg-extended:18.3         # pgvector + PostGIS
```

To switch to the extended image: comment the vanilla line, uncomment the extended line, then restart:

```bash
./nx run deploy:db:stop && ./nx run deploy:db:start
```

The `postgresql-extensions` ConfigMap is always present. Its init SQL uses fault-tolerant `DO` blocks that create `vector` and `postgis` extensions on first initialization when the extended image is used, and silently skip them when the vanilla image is used.
