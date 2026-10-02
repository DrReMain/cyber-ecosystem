# Custom PostgreSQL Image (pgvector + PostGIS)

Extended PostgreSQL image based on `postgres:18.6-trixie` with pgvector and PostGIS pre-installed.

## Build

```bash
docker build -t pg-extended:18.6 .
```

## Usage

The dev stack references this image by name (`deploy/docker/compose.dev.yaml`, service `postgresql`); compose builds it automatically when absent.

The init SQL (`deploy/docker/postgres/01-extensions.sql`) uses fault-tolerant `DO` blocks that create the `vector` and `postgis` extensions on first initialization; with a vanilla `postgres` image they are silently skipped.

For the k8s deployment shape, load the image into the cluster's registry or a private registry the cluster can pull from.
