---
name: observability
description: Use when adding observability to a new service or client, setting up error reporting, metrics scraping, or traces — typically after scaffolding and before E2E testing
---

# Observability Setup

Quick-reference for wiring a new service or client into the local observability stack.

---

## New Service Checklist

1. Start the stack: `cd infra/docker && docker compose --profile observability up -d`
2. Create a GlitchTip project → get DSN from Settings → Client Keys
3. Write the DSN to the appropriate config file (see below)
4. Add a Prometheus scrape job to `infra/docker/signoz/otel-collector-config.yaml`
5. Restart OTel Collector: `docker compose restart signoz-otel-collector`
6. Start the service → verify on SigNoz Services page
7. Send a request → verify Traces, Logs, and Metrics all collected

---

## DSN Configuration

### Go Service (`configs/config.yaml`)

```yaml
error_report:
  enabled: true
  dsn: "http://<key>@localhost:8000/<project_id>"
  environment: "development"
  sample_rate: 1.0
```

### TanStack Client (`.env`)

```
VITE_GLITCHTIP_DSN=http://<key>@localhost:8000/<project_id>
```

### Expo Client (`.env`)

```
EXPO_PUBLIC_GLITCHTIP_DSN=http://<key>@localhost:8000/<project_id>
```

---

## Prometheus Scrape Job

Add to `infra/docker/signoz/otel-collector-config.yaml` under `receivers.prometheus.config.scrape_configs`:

```yaml
- job_name: genesis-service-<service>
  metrics_path: /metrics
  scrape_interval: 15s
  static_configs:
  - targets:
      - host.docker.internal:<ops_port>
    labels:
      app: genesis
      service: <service>
      env: development
      component: ops
```

Replace `<service>` with the service name and `<ops_port>` with the ops server port.

---

## Tracing & Logs Configuration

Add to the service's `configs/config.yaml`:

```yaml
trace:
  endpoint: "localhost:4318"
  insecure: true

log:
  otlp_log:
    enabled: true
    endpoint: "localhost:4318"
    insecure: true
```

Traces and logs are collected automatically via OTLP — no manual scrape config needed beyond the trace/log endpoint configuration.

---

## Port Assignments

| Service | Ops Port | Scrape Target |
|---------|----------|---------------|
| base | 14000 | `host.docker.internal:14000` |
| admin_bff | 14001 | `host.docker.internal:14001` |
| mobile_bff | 14002 | `host.docker.internal:14002` |

New services increment by +1. Next available: 14003.

---

## Verification

1. Restart OTel Collector: `docker compose restart signoz-otel-collector`
2. Wait 30 seconds (one scrape interval)
3. SigNoz (`localhost:8080`) → Services page — confirm the new service appears
4. Send a request → check Traces Explorer and Logs Explorer
5. Trigger an error (e.g., call a dependent service while it is stopped) → GlitchTip (`localhost:8000`) → Issues page

---

## Stack Access

| Service | URL | Credentials |
|---------|-----|-------------|
| SigNoz | `localhost:8080` | local-dev defaults — see `infra/docker/docker-compose.yaml` |
| GlitchTip | `localhost:8000` | local-dev defaults — see `infra/docker/docker-compose.yaml` |
| OTel Collector | `localhost:4317` (gRPC) / `localhost:4318` (HTTP) | — |

---

For detailed config explanations, see `docs/stacks/observability.md`.
