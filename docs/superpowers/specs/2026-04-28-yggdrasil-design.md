# Yggdrasill - 通用管理后台骨架基座设计

## 概述

Yggdrasill 是一个通用管理后台骨架基座，不绑定任何真实业务。目标是让后续开发真实业务时能快速迭代，而不需要花过多成本在基础功能搭建上。

本设计替代现有的 `apps/singleton/`，解决其三大痛点：代码耦合（所有能力捆绑在一起）、部署粒度粗（无法按需裁剪）、扩展困难（业务定制如 DataScope 不够灵活）。

## 架构核心决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 能力部署模式 | 纯 Remote（能力服务始终是独立服务） | 架构简单，一种实现路径，服务边界清晰 |
| App 定义 | 一个 App = 一个完整业务体系 | App 内服务可互通，跨 App 调用禁止 |
| IAM 策略 | 各端独立 IAM（admin/saas/consumer） | 用户体系解耦，多租户隔离可实现 |
| 数据库 | 每个服务独立 DB，不跨库访问 | 独立演进，互不影响 |
| 通信协议 | Go 服务间 ConnectRPC/gRPC；第三方 HTTP | 与现有技术栈一致 |
| 分布式事务 | 不需要，使用最终一致性 | 管理后台场景不需要强一致性 |

## Monorepo 结构

```
cyber-ecosystem/
  contracts/                          # 跨 App 共享 proto 基础类型
    common/
      page.proto
      errors.proto
    buf.yaml

  shared-go/                          # 跨 App 共享 Go 模块
    capabilities/                     # 能力 thin client 模块
      auth/                           # JWT 校验 + session 验证（调 iam）
        auth.go                       # 接口定义
        client.go                     # gRPC 客户端实现
        middleware.go                  # 认证中间件
      rbac/                           # 鉴权决策客户端（调 iam）
        rbac.go
        client.go
        middleware.go
      audit/                          # 审计事件发送（调 audit）
        audit.go
        client.go
        interceptor.go
      storage/                        # 文件操作客户端（调 storage）
        storage.go
        client.go
      datascope/                      # 从 iam 拉取规则 + 本地 Ent interceptor
        datascope.go
        client.go
        mixin.go                      # Ent mixin（owner_id, dept_id 字段）
        interceptor.go                # Ent interceptor（从 context 取规则过滤查询）
      condition/                      # 从 iam 拉取条件 + 本地中间件执行
        condition.go
        client.go
        middleware.go
      security/                       # 中间件编排，组合以上能力到统一请求链
        security.go
        options.go                    # 配置项（启用/禁用哪些能力、服务地址等）
    platform/                         # 平台工具（不调远程服务，本地运行）
      scheduler/                      # 定时任务封装（robfig/cron）
      messaging/                      # MQ 抽象层（Redis Streams / NATS）
      websocket/                      # WebSocket 辅助
      sse/                            # SSE 辅助
    kratos/                           # Kratos 框架封装（已有）
    orm/ent/                          # Ent 公共 mixin（已有）
    utils/                            # 工具库（已有）

  apps/
    yggdrasil/                       # 通用管理后台骨架基座
      api/v1/                         # App 级 proto 定义（平铺 + 服务前缀）
        iam_account_auth.proto
        iam_account_detail.proto
        iam_user.proto
        iam_role.proto
        iam_role_binding.proto
        iam_permission_binding.proto
        iam_department.proto
        iam_department_binding.proto
        iam_user_attribute.proto
        iam_condition.proto
        iam_data_scope.proto
        iam_resource.proto
        iam_internal_auth.proto
        iam_internal_authorization.proto
        iam_internal_scope.proto
        iam_internal_condition.proto
        audit_log.proto
        audit_sink.proto
        storage_file.proto
        worklog.proto
        error_reason.proto
      gen/
        go/v1/                        # 生成的 Go stubs
        oas/                          # OpenAPI spec
      clients/
        admin/                        # Admin 前端（React + TanStack）
      services/
        iam/                          # Admin IAM 服务
        audit/                        # 审计服务（独立）
        storage/                      # 文件存储服务（独立）
        worklog/                      # 工作日志（业务模板）
    singleton/                        # 旧项目（保留参考）

  infra/                              # 共享基础设施
  tools/                              # 开发工具
```

## 服务架构

### 服务清单与职责

| 服务 | 数据域 | 独立 DB | 依赖 |
|------|--------|---------|------|
| **iam** | 用户、角色、部门、策略、条件规则、DataScope 规则 | `yggdrasil_iam` | Redis, audit |
| **audit** | 审计日志 | `yggdrasil_audit` | Redis（可选缓冲）, iam |
| **storage** | 文件元数据 | `yggdrasil_storage` | MinIO / S3, iam |
| **worklog** | 工作日志业务数据 | `yggdrasil_worklog` | iam, audit, storage |

### 服务间调用关系

```
                  ┌─────────┐
                  │   iam   │◄─── 所有服务的认证/鉴权/规则查询
                  └────┬────┘
                       │ 自身也产生审计事件
                       ▼
  ┌──────────┐   ┌─────────┐   ┌──────────┐
  │ worklog  │──►│  audit  │◄──│ storage  │
  │          │──►│         │──►│          │
  │          │──►└─────────┘   └────▲─────┘
  │          │──►──────────────────┘
  └──────────┘
```

- **iam** 被所有服务依赖（认证、鉴权、规则查询）
- **audit** 被所有服务依赖（接收事件），自身也依赖 iam 做请求认证
- **storage** 按需使用（有文件需求的业务才调用），自身也依赖 iam 做请求认证
- 业务服务之间互不依赖
- IAM 自身也向 audit 投递审计事件（登录/登出/权限变更等）

### 通信协议

| 场景 | 协议 |
|------|------|
| 一方 Go 客户端（Admin 前端） | ConnectRPC（一元接口）；HTTP（非一元接口） |
| Go 服务间 | gRPC / ConnectRPC |
| 三方客户端 | HTTP |
| Rust/Python 服务 | HTTP（调 iam/audit/storage 的 API） |

### Admin 前端路由

Admin 前端直接与多个后端服务通信，不经过单一 BFF 代理：

| 前端操作 | 目标服务 | 说明 |
|----------|----------|------|
| 登录、用户管理、角色管理、部门管理 | iam | IAM 管理接口 |
| 审计日志查看 | audit | 审计查询接口 |
| 工作日志 CRUD | worklog | 业务接口 |
| 文件上传/下载 | storage | 文件操作接口 |

前端通过配置知道各服务的地址。各服务独立暴露 ConnectRPC/HTTP 端口。

### 服务端口分配

| 服务 | HTTP | gRPC | ConnectRPC | Ops |
|------|------|------|------------|-----|
| iam | 11000 | 12000 | 13000 | 14000 |
| audit | 11001 | 12001 | 13001 | 14001 |
| storage | 11002 | 12002 | 13002 | 14002 |
| worklog | 11003 | 12003 | 13003 | 14003 |

## 服务内部架构规则

以下规则适用于所有 yggdrasil 内部服务的 Go 实现，替代 singleton_backend 中发现的问题模式。

### Domain Event 系统

服务内部使用 in-process 事件总线进行 UC 之间的协调，替代直接方法调用。

**原理：** 当一个 UC 完成操作后，相关副作用（策略重载、缓存失效、级联清理）通过事件订阅自动触发，而非在代码中手动调用。这消除了遗漏调用的 bug，也解除了 UC 之间的具体类型依赖。

**事件列表（IAM 服务）：**

| Topic | Publisher | Subscribers | 副作用 |
|-------|-----------|-------------|--------|
| `role.deleted` | RoleUC | PolicyUC, DataScopeUC, ConditionUC | 删除关联策略、scope 规则、条件规则 |
| `role_binding.created` | RoleBindingUC | PolicyUC, DataScopeUC | 重载策略、失效用户 scope 缓存 |
| `role_binding.deleted` | RoleBindingUC | PolicyUC, DataScopeUC | 重载策略、失效用户 scope 缓存 |
| `permission_binding.created` | PermissionBindingUC | PolicyUC | 重载策略 |
| `permission_binding.deleted` | PermissionBindingUC | PolicyUC | 重载策略 |
| `department_binding.changed` | DepartmentBindingUC | DataScopeUC | 失效用户 scope 缓存 |
| `user_attribute.changed` | UserAttributeUC | DataScopeUC | 失效用户 scope 缓存 |
| `user.deleted` | UserUC | DataScopeUC | 失效用户 scope 缓存 |

**实现位置：** `internal/event/bus.go` — `Bus` 接口 + `SyncBus` 同步实现。

**替代的 singleton 模式：**

| Singleton | Yggdrasil | 改进 |
|-----------|-----------|------|
| `RoleCascadeRegistry` 直接调用 3 个 UC | 事件 `role.deleted` + 独立订阅者 | 无循环依赖 |
| `ScopeCacheInvalidator` 接口泄漏到 service 层 | 事件 `department_binding.changed` → DataScopeUC 订阅 | 分层清晰 |
| service 层手动调 `InvalidateUser` | UC 内部自动发布事件 | 无遗漏 |

### 严格分层注入规则

```
┌─────────────────────────────────────────────────────────────┐
│ server/                                                      │
│   middleware ONLY depends on pkg/* interfaces                 │
│   shared buildMiddlewares() — NOT triplicated                 │
├─────────────────────────────────────────────────────────────┤
│ service/                                                     │
│   ONLY depends on biz/* UC interfaces                         │
│   NEVER imports data/* or biz internal adapters               │
├─────────────────────────────────────────────────────────────┤
│ biz/                                                         │
│   UC depends on: RP interfaces (defined in biz) + event.Bus  │
│   UC NEVER depends on another UC concrete type                │
│   Adapters split into adapter_*.go (one per concern)          │
│   Event handlers registered in ProviderSet                    │
├─────────────────────────────────────────────────────────────┤
│ data/                                                        │
│   Returns biz.*RP interfaces from constructors                │
│   ONLY depends on Store/Ent                                   │
├─────────────────────────────────────────────────────────────┤
│ event/                                                       │
│   Bus interface + SyncBus implementation                      │
├─────────────────────────────────────────────────────────────┤
│ pkg/                                                         │
│   Pure interfaces and types — no UC/RP imports                │
└─────────────────────────────────────────────────────────────┘
```

**规则：**

1. **middleware 只依赖 pkg/ 接口** — 不注入 UC 或 RP，通过适配器转换
2. **service 只依赖 biz UC** — 不依赖 RP、不依赖 biz 内部适配器、不依赖 event.Bus
3. **UC 不依赖其他 UC 具体类型** — 通过 event.Bus 发布事件协调副作用
4. **适配器只做 UC → pkg 接口转换** — 不直接导入 RP（修复 singleton 中 ConditionChecker 直接引用 UserAttributeRP 的问题）
5. **中间件链只构建一次** — `buildMiddlewares()` 函数被 http/grpc/connect 共享调用

### Singleton 已知问题与改进对照

| 问题 | Singleton 代码 | Yggdrasil 改进 |
|------|---------------|---------------|
| UC 循环依赖 | DataScopeUC→PolicyUC→ConditionUC | event.Bus 解耦，无直接引用 |
| 级联遗漏 bug | 权限变更后策略不更新 | 事件订阅保证每个副作用都被触发 |
| biz 接口泄漏 | ScopeCacheInvalidator 被 service 引用 | service 只依赖 UC，副作用通过事件 |
| adapter 混杂 | adapter.go 一个文件三种适配器 | adapter_session.go, adapter_authorizer.go 等拆分 |
| 跨层引用 | ConditionChecker 导入 UserAttributeRP | 适配器只包装 UC，不直接导入 RP |
| 中间件三重复制 | http/grpc/connect 各写一遍 | buildMiddlewares() 共享 |
| main.go 过大 | 256 行混合多种初始化 | 按职责拆分辅助文件 |
| rp_policy.go 过大 | 481 行混合 adapter/repo/seeder/pubsub | 拆分为独立文件 |

## IAM 服务

### 定位

只服务管理后台（Admin Realm），处理内部员工的身份认证和访问控制。SaaS 和 C 端未来有独立的 IAM 服务。

### 内部结构

```
apps/yggdrasil/services/iam/
  cmd/app/
    main.go
    wire.go
    wire_gen.go
  configs/
    config.yaml
  internal/
    conf/                   # 配置 schema (proto)
    server/                 # 传输层
      server.go
      http.go               # Kratos HTTP server
      grpc.go               # Kratos gRPC server
      connect.go            # ConnectRPC server
      ops.go                # 监控 (Prometheus, pprof)
      i18n.go               # i18n bundle
      locales/              # zh-CN, en-US
      middleware/            # IAM 自身中间件
        auth.go             # JWT 认证
        i18n.go
    service/                # Service 层（proto → biz 映射）
    biz/                    # 业务逻辑层（use cases）
    data/                   # 数据访问层（Ent repos）
      ent/
        schema/             # IAM 实体 schemas
          user.go
          role.go
          department.go
          policy.go
          condition.go
          data_scope.go
          user_attribute.go
    pkg/                    # 内部工具
```

### 能力清单

| 能力 | 描述 | Proto (管理 API) | Proto (内部 API) |
|------|------|-------------------|-------------------|
| 认证 | JWT + Session 管理 | `iam_account_auth.proto` | `iam_internal_auth.proto` |
| 用户管理 | 用户 CRUD | `iam_user.proto` | - |
| RBAC | 角色、权限、Casbin 策略 | `iam_role.proto`, `iam_role_binding.proto`, `iam_permission_binding.proto` | `iam_internal_authorization.proto` |
| 部门 | 层级部门管理 | `iam_department.proto`, `iam_department_binding.proto` | - |
| 用户属性 | KV 属性存储 | `iam_user_attribute.proto` | - |
| ABAC 条件 | 条件规则管理 | `iam_condition.proto` | `iam_internal_condition.proto` |
| DataScope | 行级数据权限 | `iam_data_scope.proto` | `iam_internal_scope.proto` |
| 资源发现 | 服务/方法自省 | `iam_resource.proto` | - |

注意：资源发现只能自省 IAM 自身注册的服务/方法。业务服务如需资源发现，各服务自行实现（参照 IAM 的模式），或未来可引入统一的服务注册中心。

### 内部 API 设计

内部 API 面向业务服务的 thin client，关注高性能决策，与管理 API 的区别：

| 维度 | 管理 API | 内部 API |
|------|----------|----------|
| 消费者 | Admin 前端 | 业务服务 thin client |
| 关注点 | CRUD、分页查询 | 高性能决策、规则获取 |
| 部署 | 可公网暴露 | 仅内网 |
| 示例 | 列出所有角色（分页） | 检查用户 X 是否有 worklog.create 权限 |

## Audit 服务

### 定位

独立的审计日志服务，接收来自所有服务的事件并提供查询 API。

### 审计维度

| 维度 | 覆盖内容 | 触发点 |
|------|----------|--------|
| **操作审计** | 谁、何时、做了什么（create/update/delete/login/logout） | Service 层 |
| **访问审计** | 谁、何时、看了什么（read/list/download/export） | Service 层 |
| **安全审计** | 登录失败、权限拒绝、条件拒绝、Session 失效 | 中间件层 |

### 审计事件模型

```go
type AuditEvent struct {
    Timestamp     time.Time
    ActorID       string
    ActorName     string
    DepartmentID  string
    Action        string       // create/read/update/delete/login/logout/login_failed/access_denied/...
    ResourceType  string       // user/role/worklog/...
    ResourceID    string
    ServiceName   string       // iam/worklog/...
    Result        string       // success/failure/denied
    IP            string
    UserAgent     string
    Detail        string       // JSON 补充上下文
}
```

### 事件采集方式

- **iam / audit 服务自身**：直接写入 DB
- **业务服务**：通过 `audit_sink.proto` 异步投递
- **thin client**：内部使用 buffered channel 批量发送，避免每个请求一次 gRPC 调用

### 失败/拒绝场景

RBAC 中间件拒绝请求时，audit interceptor 自动记录 `access_denied` 事件。Condition 中间件拒绝时，自动记录 `condition_denied` 事件。这些不需要业务代码手动处理。

## Storage 服务

### 定位

独立的文件存储服务，管理文件元数据和物理存储。

### 存储 Backend

- 开发/测试：本地文件系统
- 生产：MinIO / S3
- 通过配置切换，业务服务无需感知

## 能力 Thin Client 模块

### 设计原则

每个 thin client 模块由三部分组成：

1. **接口定义**：Go interface，定义能力契约
2. **gRPC 客户端实现**：调用对应服务的内部 API
3. **本地应用逻辑**：中间件、拦截器、缓存等

### 横切型能力的特殊处理

DataScope 和 Condition 是横切型能力——规则管理在 IAM 服务，但规则执行必须在业务服务进程内（因为要挂 Ent interceptor / 中间件）。

| 能力 | 规则管理 | 规则执行 |
|------|----------|----------|
| DataScope | IAM 服务 | 业务服务 Ent interceptor（thin client） |
| Condition | IAM 服务 | 业务服务中间件（thin client） |

### 业务服务接入方式

通过 Wire ProviderSet 组合：

```go
// services/worklog/cmd/app/wire.go
wire.Build(
    security.ProviderSet,    // 中间件编排
    auth.ProviderSet,        // 认证
    rbac.ProviderSet,        // 鉴权
    audit.ProviderSet,       // 审计
    datascope.ProviderSet,   // 数据权限
    condition.ProviderSet,   // ABAC 条件
    // storage.ProviderSet,  // 需要时取消注释

    biz.ProviderSet,
    data.ProviderSet,
    service.ProviderSet,
    server.ProviderSet,
)
```

配置：

```yaml
iam:
  address: "localhost:12000"     # gRPC 端口（服务间通信）
audit:
  address: "localhost:12001"
storage:
  address: "localhost:12002"

capabilities:
  auth:
    enabled: true
  rbac:
    enabled: true
  audit:
    enabled: true
  datascope:
    enabled: true
  condition:
    enabled: true
```

### 中间件链（security 模块编排）

```
i18n → recovery → ratelimit → metrics → tracing → metadata → logging
  → [selector: auth → rbac → condition → scope_injector] → validate
```

`selector` 根据 proto annotation（`auth.public_access`）跳过公开接口的认证。

## Worklog 服务（业务模板）

### 定位

Worklog 是一个参考实现，展示业务服务如何接入所有能力模块。后续新业务服务参照此结构创建。

### 目录结构

```
apps/yggdrasil/services/worklog/
  cmd/app/
    main.go
    wire.go
    wire_gen.go
  configs/
    config.yaml
  internal/
    conf/
    server/
      server.go
      http.go
      grpc.go
      connect.go
      ops.go
    service/
      worklog.go
    biz/
      worklog.go              # WorklogUC + WorklogRP 接口
    data/
      store.go                # Store (Transaction, Ent Client)
      worklog.go              # WorklogRP 实现
      ent/
        schema/
          worklog.go          # 含 DataScope mixin
```

### Worklog Ent Schema

```go
type Worklog struct {
    ent.Schema
}

func (Worklog) Mixin() []ent.Mixin {
    return []ent.Mixin{
        sharedmixins.IDString{},
        sharedmixins.CreatedUpdated{},
        sharedmixins.SoftDelete{},
        datascope.Mixin{},             // owner_id, dept_id
    }
}

func (Worklog) Fields() []ent.Field {
    return []ent.Field{
        field.String("title"),
        field.Text("content"),
        field.Strings("attachment_ids").Optional(),
    }
}
```

### 能力接入清单

| 操作 | Auth | RBAC | Condition | DataScope | Audit | Storage |
|------|------|------|-----------|-----------|-------|---------|
| CreateWorklog | 验证身份 | `worklog.create` | 时间/IP条件 | 填充 owner_id, dept_id | 创建事件 | 关联附件 |
| ListWorklogs | 验证身份 | `worklog.list` | 时间/IP条件 | Ent interceptor 过滤 | 查询事件 | - |
| GetWorklog | 验证身份 | `worklog.get` | - | 拦截非授权记录 | 查看事件 | 返回附件信息 |
| UpdateWorklog | 验证身份 | `worklog.update` | 时间/IP条件 | 只能改本部门 | 修改事件 | 更新附件关联 |
| DeleteWorklog | 验证身份 | `worklog.delete` | - | 只能删本部门 | 删除事件 | - |

### 通信流（CreateWorklog 示例）

```
Admin 前端
  │  ConnectRPC: CreateWorklog
  ▼
Worklog Service
  ├─ 1. auth middleware     → iam (验证 JWT + session)
  ├─ 2. rbac middleware     → iam (检查 worklog.create 权限)
  ├─ 3. condition middleware → iam (检查时间/IP 条件)
  ├─ 4. service.CreateWorklog()
  │     ├─ biz: 写入 Ent (datascope interceptor 自动填充 owner_id/dept_id)
  │     └─ biz: 调 storage (关联 attachment_ids)
  └─ 5. audit interceptor   → audit (异步发送审计事件)
```

### 审计事件采集

在 Service 层手动记录成功操作，在中间件层自动捕获失败/拒绝：

```go
func (s *WorklogService) CreateWorklog(ctx context.Context, req *pb.CreateWorklogRequest) (*pb.CreateWorklogResponse, error) {
    resp, err := s.uc.Create(ctx, ...)
    if err != nil {
        return nil, err
    }
    s.audit.Record(ctx, &audit.AuditEvent{
        Action:       "create",
        ResourceType: "worklog",
        ResourceID:   resp.Id,
    })
    return resp, nil
}
```

## 数据库策略

| 服务 | 数据库名 | 说明 |
|------|----------|------|
| iam | `yggdrasil_iam` | 用户、角色、部门、策略等 |
| audit | `yggdrasil_audit` | 审计日志 |
| storage | `yggdrasil_storage` | 文件元数据 |
| worklog | `yggdrasil_worklog` | 工作日志业务数据 |

- 每个服务拥有独立的 PostgreSQL 数据库
- 服务不跨库访问，需要其他服务的数据时通过 API 调用
- 每个服务独立管理自己的 Ent schema 和迁移

## 跨服务一致性

不引入分布式事务。各场景使用最终一致性：

| 场景 | 一致性方案 |
|------|-----------|
| 审计事件 | thin client buffered channel + 重试，短暂失败不丢事件 |
| 用户删除 | IAM 软删除，发事件通知业务服务异步清理 |
| 权限变更 | IAM 更新策略，Redis Pub/Sub 通知业务服务刷新缓存；缓存 TTL 兜底 |
| 部门变更 | DataScope 每次查询实时从 IAM 拉取规则（或缓存 + TTL） |
| 文件 + 业务实体 | 先上传拿 ID，再创建实体关联；孤儿文件定期清理 |

`shared-go/` 提供重试、幂等、Redis Pub/Sub 等工具模式，但不作为强制框架。

Storage 服务内部（DB 元数据 + MinIO 物理文件）的一致性方案在实现 storage 服务时再具体设计。

## 未来扩展

### 同 App 内的扩展（yggdrasil 内）

```
apps/yggdrasil/services/
  iam/                # Admin IAM (MVP)
  audit/
  storage/
  worklog/
  # ── 未来按需添加 ──
  saas_iam/           # SaaS IAM（租户感知、自注册、多租户隔离）
  consumer_iam/       # C 端 IAM（手机 OTP、OAuth 社交登录）
  approval/           # 审批流引擎（流程定义、实例、会签）
  portal/             # SaaS 业务服务
  member/             # C 端业务服务
```

Admin 查看租户数据通过 App 内服务间调用：Admin 业务服务 → SaaS IAM 内部 API。

### 跨 App 扩展

为不同客户创建新 App 时：

```
apps/
  yggdrasil/         # 产品级骨架（可能自用）
  crm/                # A 客户 CRM
  erp/                # B 客户 ERP
```

每个 App 复用 `shared-go/` 模块和 `contracts/` 类型，但 App 间禁止互相调用。

### 平台工具扩展

`shared-go/platform/` 按需添加，不影响已有服务：

| 工具 | 说明 | 触发条件 |
|------|------|----------|
| scheduler | robfig/cron 封装 | 需要定时任务时 |
| messaging | Redis Streams / NATS 抽象 | Redis Pub/Sub 不够用时 |
| websocket | 连接管理、心跳、重连 | 需要实时推送时 |
| sse | SSE 辅助 | 需要服务端推送时 |

## 技术栈

| 类别 | 选型 | 说明 |
|------|------|------|
| 语言 | Go 1.25 | 主力开发语言 |
| 框架 | go-kratos/kratos v2 | 微服务框架 |
| 传输 | ConnectRPC + gRPC + HTTP | 多协议支持 |
| ORM | entgo.io/ent | 类型安全 ORM |
| 数据库 | PostgreSQL | 各服务独立实例 |
| 缓存 | Redis | 缓存 + Pub/Sub |
| RBAC | casbin/casbin v2 | 策略引擎 |
| Auth | golang-jwt/jwt v5 | JWT |
| DI | google/wire | 编译时依赖注入 |
| 日志 | uber.org/zap | 结构化日志 |
| 追踪 | opentelemetry/otel | 分布式追踪 |
| 指标 | prometheus | 监控指标 |
| i18n | nicksnyder/go-i18n v2 | 国际化 |
| Proto | buf.build | Proto 管理和代码生成 |
| 构建 | Nx | Monorepo 构建编排 |
| 对象存储 | MinIO / S3 | 文件存储 |
| 前端 | React + TanStack (Start, Router, Query) | Admin 前端 |
