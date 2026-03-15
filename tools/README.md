# Cyber Ecosystem Tools

本目录包含项目开发所需的工具配置，包括中间件的 Docker 配置。

## 目录结构

```
tools/
├── docker/                    # Docker Compose 配置（本地开发）
│   ├── docker-compose.yaml    # 中间件编排配置
│   ├── prometheus/            # Prometheus 配置
│   │   └── prometheus.yml
│   └── grafana/               # Grafana 配置
│       └── provisioning/
│           └── datasources/
│               └── datasources.yml
└── project.json               # Nx 项目配置
```

## 中间件列表

| 服务 | 端口 | 说明 |
|------|------|------|
| PostgreSQL | 5432 | 主数据库 |
| Redis | 6379 | 缓存 |
| Jaeger | 16686 (UI), 14268 (HTTP), 4317 (OTLP gRPC) | 链路追踪 |
| Prometheus | 9090 | 监控指标收集 |
| Grafana | 3000 | 监控可视化 |

## Docker 使用方法

### 前置条件

确保已安装 Docker 和 Docker Compose：

```bash
# 检查 Docker 版本
docker --version

# 检查 Docker Compose 版本
docker compose version
```

### Nx 命令

本项目使用 Nx 管理命令，所有命令都在 `tools` 项目中定义：

```bash
# 启动所有中间件
nx docker:up tools

# 停止所有中间件
nx docker:down tools

# 查看日志
nx docker:logs tools

# 查看运行状态
nx docker:ps tools

# 重启所有中间件
nx docker:restart tools

# 清理（删除容器和数据卷）
nx docker:clean tools

# 单独启动某个服务
nx docker:jaeger tools     # 仅启动 Jaeger
nx docker:postgres tools   # 仅启动 PostgreSQL
nx docker:redis tools      # 仅启动 Redis
nx docker:prometheus tools # 仅启动 Prometheus
nx docker:grafana tools    # 仅启动 Grafana

# 启动监控栈（Prometheus + Grafana）
nx docker:monitoring tools
```

### 直接使用 Docker Compose

也可以直接使用 Docker Compose 命令：

```bash
cd tools/docker

# 启动所有服务
docker compose up -d

# 停止所有服务
docker compose down

# 查看日志
docker compose logs -f

# 查看状态
docker compose ps

# 启动单个服务
docker compose up -d jaeger
docker compose up -d postgres
docker compose up -d redis
docker compose up -d prometheus
docker compose up -d grafana
```

## Jaeger 链路追踪

### 访问 UI

- 本地 Docker: http://localhost:16686

### 配置应用

在应用配置中设置 Jaeger 端点：

```yaml
trace:
  endpoint: "http://localhost:14268/api/traces"
```

### OTLP 配置

Jaeger 支持 OTLP 协议，可以使用以下端口：

- OTLP gRPC: 4317
- OTLP HTTP: 4318

## Prometheus 监控

### 访问 UI

- 本地 Docker: http://localhost:9090

### 配置应用监控

1. 在 Go 应用中添加 Prometheus metrics 端点（使用 `github.com/prometheus/client_golang`）
2. 修改 [`prometheus/prometheus.yml`](docker/prometheus/prometheus.yml) 添加应用抓取配置

### 数据保留

默认保留 15 天，可在 docker-compose.yaml 中修改 `--storage.tsdb.retention.time` 参数。

## Grafana 可视化

### 访问 UI

- 本地 Docker: http://localhost:3000
- 默认用户名: `admin`
- 默认密码: `admin`

### 数据源

Grafana 已预配置以下数据源：

- **Prometheus**: 用于查询监控指标
- **Jaeger**: 用于关联追踪数据

### 导入仪表盘

推荐导入以下社区仪表盘：

- [Go Processes](https://grafana.com/grafana/dashboards/6671): Go 应用进程监控
- [Prometheus 2.0 Overview](https://grafana.com/grafana/dashboards/3662): Prometheus 概览

## 故障排除

### Docker 容器无法启动

1. 检查端口是否被占用：
   ```bash
   lsof -i :5432   # PostgreSQL
   lsof -i :6379   # Redis
   lsof -i :16686  # Jaeger
   lsof -i :9090   # Prometheus
   lsof -i :3000   # Grafana
   ```

2. 检查 Docker 日志：
   ```bash
   docker compose logs <service-name>
   ```

### 数据库连接失败

1. 确认 PostgreSQL 容器已启动：
   ```bash
   docker compose ps postgres
   ```

2. 检查健康状态：
   ```bash
   docker compose exec postgres pg_isready -U postgres
   ```

### Grafana 无法连接 Prometheus

1. 确认两个服务都在运行：
   ```bash
   docker compose ps prometheus grafana
   ```

2. 检查网络连接：
   ```bash
   docker compose exec grafana wget -q -O- http://prometheus:9090/-/healthy
   ```

## 生产环境注意事项

1. **密码安全**: 生产环境请使用强密码
2. **存储**: 生产环境建议使用持久化存储
3. **高可用**: 生产环境建议配置 PostgreSQL 和 Redis 的高可用方案
4. **Jaeger 存储**: 生产环境建议使用 Elasticsearch 或 Cassandra 作为 Jaeger 的后端存储
5. **Prometheus 存储**: 生产环境建议配置更长的数据保留时间和远程存储
6. **Grafana 安全**: 生产环境建议配置 HTTPS 和 OAuth 认证
