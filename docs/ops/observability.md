# 可观测性基础设施

本地开发环境的可观测性栈操作指南。

---

## 1. 启动可观测性栈

可观测性服务通过 Docker Compose profile 启动，不会随核心服务自动启动：

```bash
cd infra/docker
docker compose --profile observability up -d
```

### 服务端口

| 服务 | 端口 | 用途 |
|------|------|------|
| SigNoz | `localhost:8080` | Traces / Metrics / Logs 查询 |
| GlitchTip | `localhost:8000` | 错误上报 (Sentry 兼容) |
| OTel Collector | `localhost:4317` (gRPC) / `localhost:4318` (HTTP) | OTLP 数据接收 |

### 默认账号

两个服务共用：

- **SigNoz**: `admin@cyber-ecosystem.com` / `admin@Signoz123!`
- **GlitchTip**: `admin@cyber-ecosystem.com` / `Cyber-Ecosystem123`

---

## 2. GlitchTip 错误上报

每个服务（包括前端客户端）需要独立的 GlitchTip 项目和 DSN。

### 2.1 创建项目

1. 登录 `http://localhost:8000`
2. 进入组织 → Projects → Create Project
3. 选择平台：
   - Go 服务 → **Go**
   - React (TanStack) 客户端 → **React**
   - React Native 客户端 → **React Native**
4. 命名建议：`base`、`admin_bff`、`mobile_bff`、`admin_client`

### 2.2 获取 DSN

项目创建后，Settings → Client Keys 中获取 DSN，格式：

```
http://<key>@localhost:8000/<project_id>
```

### 2.3 配置 DSN

**Go 服务** — 写入 `configs/config.yaml`：

```yaml
error_report:
  enabled: true
  dsn: "http://<key>@localhost:8000/<project_id>"
  environment: "development"
  sample_rate: 1.0
```

**React (TanStack) 客户端** — 写入 `.env`：

```
VITE_GLITCHTIP_DSN=http://<key>@localhost:8000/<project_id>
```

**React Native 客户端** — 写入 `.env`：

```
EXPO_PUBLIC_GLITCHTIP_DSN=http://<key>@localhost:8000/<project_id>
```

### 2.4 验证

1. 启动服务
2. 触发一个错误（如调用依赖服务时停止被依赖服务）
3. 检查 GlitchTip → Issues 页面，确认错误事件出现且包含 `error.reason` tag

---

## 3. OTel Collector 指标采集

新服务的 Prometheus 指标需要在 OTel Collector 配置中手动添加 scrape target。

### 3.1 添加 scrape job

编辑 `infra/docker/signoz/otel-collector-config.yaml`，在 `receivers.prometheus.config.scrape_configs` 下添加：

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

- `<service>`: 服务名（如 `base`、`admin_bff`）
- `<ops_port>`: ops server 端口（`config.yaml` 中 `ops.addr` 指定的端口）

### 3.2 端口分配

| 服务 | Ops 端口 | Scrape target |
|------|---------|---------------|
| base | 14000 | `host.docker.internal:14000` |
| admin_bff | 14001 | `host.docker.internal:14001` |
| mobile_bff | 14002 | `host.docker.internal:14002` |

新增服务按 +1 递增。

### 3.3 验证

1. 重启 OTel Collector：`docker compose restart signoz-otel-collector`
2. 等待 30 秒（scrape interval）
3. SigNoz → Services 页面检查新服务是否出现
4. SigNoz → Metrics Explorer 验证指标数据

---

## 4. Tracing 和 Logs

Tracing 和 Logs 通过 OTLP 协议自动采集，无需手动配置 scrape target。服务只需在 `config.yaml` 中配置 OTLP endpoint：

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

### 验证

- **Traces**: SigNoz → Traces Explorer，按 service.name 过滤
- **Logs**: SigNoz → Logs Explorer，检查结构化日志是否包含 `trace_id` 字段
- **Trace-Log 关联**: 在 Trace 详情页点击 span → Related Signals → Logs，确认能跳转到对应日志

---

## 5. 新增服务可观测性清单

新增一个 Go 服务时，按以下顺序完成可观测性接入：

1. GlitchTip 创建项目，获取 DSN → 写入 `config.yaml` 的 `error_report.dsn`
2. OTel Collector 添加 prometheus scrape job → 写入 `otel-collector-config.yaml`
3. 重启 OTel Collector
4. 启动服务，验证 SigNoz Services 页面出现新服务
5. 触发请求，验证 Traces / Logs / Metrics 三种数据均正常采集
