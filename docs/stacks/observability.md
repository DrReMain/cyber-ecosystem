# Observability Infrastructure

Operations guide for the local development observability stack.

---

## 1. Starting the Observability Stack

Observability services start via Docker Compose profile and do not auto-start with core services:

```bash
cd infra/docker
docker compose --profile observability up -d
```

### Service Ports

| Service | Port | Purpose |
|---------|------|---------|
| SigNoz | `localhost:8080` | Traces / Metrics / Logs query UI |
| GlitchTip | `localhost:8000` | Error reporting (Sentry-compatible) |
| OTel Collector | `localhost:4317` (gRPC) / `localhost:4318` (HTTP) | OTLP data ingestion |

### Default Credentials

Both services share the same email:

- **SigNoz**: `admin@cyber-ecosystem.com` / `admin@Signoz123!`
- **GlitchTip**: `admin@cyber-ecosystem.com` / `Cyber-Ecosystem123`

---

## 2. GlitchTip Error Reporting

Each service (including frontend clients) needs its own GlitchTip project and DSN.

### 2.1 Create a Project

1. Log in at `http://localhost:8000`
2. Navigate to Organization → Projects → Create Project
3. Select the platform:
   - Go services → **Go**
   - React (TanStack) clients → **React**
   - React Native clients → **React Native**
4. Suggested naming: `base`, `admin_bff`, `mobile_bff`, `admin_client`

### 2.2 Get the DSN

After creating the project, go to Settings → Client Keys. The DSN format is:

```
http://<key>@localhost:8000/<project_id>
```

### 2.3 Configure the DSN

**Go services** — add to `configs/config.yaml`:

```yaml
error_report:
  enabled: true
  dsn: "http://<key>@localhost:8000/<project_id>"
  environment: "development"
  sample_rate: 1.0
```

**React (TanStack) clients** — add to `.env`:

```
VITE_GLITCHTIP_DSN=http://<key>@localhost:8000/<project_id>
```

**React Native clients** — add to `.env`:

```
EXPO_PUBLIC_GLITCHTIP_DSN=http://<key>@localhost:8000/<project_id>
```

### 2.4 Verify

1. Start the service
2. Trigger an error (e.g., call a dependent service while it is stopped)
3. Check GlitchTip → Issues page — confirm the error event appears with an `error.reason` tag

---

## 3. OTel Collector Metrics Scraping

New services need their Prometheus metrics manually added as scrape targets in the OTel Collector config.

### 3.1 Add a Scrape Job

Edit `infra/docker/signoz/otel-collector-config.yaml` and add under `receivers.prometheus.config.scrape_configs`:

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

- `<service>`: service name (e.g., `base`, `admin_bff`)
- `<ops_port>`: ops server port (the port specified by `ops.addr` in `config.yaml`)

### 3.2 Port Assignments

| Service | Ops Port | Scrape Target |
|---------|----------|---------------|
| base | 14000 | `host.docker.internal:14000` |
| admin_bff | 14001 | `host.docker.internal:14001` |
| mobile_bff | 14002 | `host.docker.internal:14002` |

New services increment by +1.

### 3.3 Verify

1. Restart the OTel Collector: `docker compose restart signoz-otel-collector`
2. Wait 30 seconds (one scrape interval)
3. SigNoz → Services page — confirm the new service appears
4. SigNoz → Metrics Explorer — verify metric data is flowing

---

## 4. Tracing and Logs

Traces and logs are collected automatically via the OTLP protocol — no manual scrape config needed. Services only need to configure the OTLP endpoint in `config.yaml`:

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

### Verify

- **Traces**: SigNoz → Traces Explorer, filter by `service.name`
- **Logs**: SigNoz → Logs Explorer, confirm structured logs contain a `trace_id` field
- **Trace-Log correlation**: In a Trace detail view, click a span → Related Signals → Logs, confirm the jump links to the corresponding log entries

---

## 5. New Service Observability Checklist

When adding a new Go service, complete observability integration in this order:

1. Create a GlitchTip project and get the DSN → write to `config.yaml` under `error_report.dsn`
2. Add a Prometheus scrape job to the OTel Collector → edit `otel-collector-config.yaml`
3. Restart the OTel Collector
4. Start the service and verify it appears on the SigNoz Services page
5. Send a request and verify that Traces, Logs, and Metrics are all collected correctly
