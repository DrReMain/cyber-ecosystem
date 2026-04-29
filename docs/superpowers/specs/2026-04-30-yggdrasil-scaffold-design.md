# Yggdrasil 通用后台管理脚手架设计规格

## 1. 背景

### 1.1 singleton 的教训

之前的 singleton app 尝试一次性实现完整 IAM（Session + RBAC + ABAC + DataScope）+ Audit + Storage，导致三个核心问题：

1. **能力代码和业务代码交织** — 没有模块边界，IAM 逻辑散落在 biz/data/service 各层
2. **横切关注点散落各处** — 认证检查、权限判断、审计采集分散在多个文件中
3. **一次性设计过多，没有闭环** — 还没跑通基础 RBAC 就开始做 ABAC 和 DataScope

singleton 已删除。本设计从零开始，遵循小闭环慢演进。

### 1.2 设计目标

Yggdrasil 是一个**代码模板脚手架**，为任何需要内部管理后台的项目提供开箱即用的 Access / Audit / Storage 能力。新项目从 Yggdrasil 初始化，在此基础上开发业务功能。

## 2. 定位与场景

### 2.1 定位

- 永远是**内部管理大后台**，不涉及多租户
- 代码模板模式：新项目复制即用，在项目内定制
- 比纯初始化脚手架（template）丰富，提供可插拔的平台能力

### 2.2 典型场景

| 场景 | 描述 |
|------|------|
| 内部 CRM | 纯团队内部使用，后端 + 1 个 web 客户端，需要完整 IAM + Audit + Storage + CRM 业务 |
| 平台型应用 | Yggdrasil 作为内部管理大后台，前台有面向 C 端的客户端和服务，前台不需要复杂 IAM |
| SaaS 平台 | Yggdrasil 作为内部管理大后台，前台有多租户 SaaS（有自己独立的 IAM/Audit），可能有面向租户客户的 C 端 |

## 3. 架构决策

### 3.1 交付形态

**代码模板（复制即用）**。Yggdrasil 在 monorepo 中维护为一份参考实现（`apps/yggdrasil/`），新项目从它初始化。

能力核心代码收敛到 `shared-go/` 作为共享包，app 只保留数据层和注册代码。

### 3.2 服务模型

灵活：简单项目一个 app 一个 service，复杂项目可以一个 app 多个 service。`apps/` 目录承载各种无关的内部系统和客户定制系统，共享仓库级能力。

### 3.3 数据隔离

每个 service 拥有自己的 schema（数据主权跟着 service 走）。简单项目一个 service 就一个 schema。

### 3.4 管理客户端

一个统一的管理前端，用户不应该使用两个客户端来分别进行配置和业务操作。

### 3.5 模块组织

能力模块垂直自包含：每个能力包（access/audit/storage）包含自己的类型定义、Kratos 中间件、Ent Hook。

### 3.6 依赖管理

**核心规则：UC 永远不依赖另一个 UC。**

依赖路径：`service → UC → [repos, DomainServices]`

三种解法共存：
- **共享业务规则** → DomainService（无 repo 依赖的纯计算）
- **流程编排** → Orchestrator 模式（UC 内部直接编排多个 repo 和 DomainService）
- **副作用通知** → Domain Event（只用于不需要同步返回值的场景）

## 4. shared-go 重组

### 4.1 组织原则

- 每个能力包自包含：类型 + Kratos 中间件 + Ent Hook 全在一起
- 成熟开源库利用 + 最多封装适配层
- Ent interceptors/hooks 实现数据级横切能力

### 4.2 完整结构

```
shared-go/
  cache/                              # 已有，保持不变
  utils/                              # 已有，保持不变
  orm/                                # 已有，通用 ent 工具
    ent/
      mixins/                         # ent mixins（created_updated, id_string）
      entutil/                        # 分页、错误处理、事务
      client/                         # ent client 初始化
      logging/                        # 慢查询日志

  kratos/
    transport/connect/                 # 已有，保持不变
    logging/zap/                       # 已有，保持不变
    middleware/
      validate/                        # 已有，保持不变
      i18n/                            # 已有，保持不变

    access/                            # 自包含：身份 + 授权
    ├── access.go                      #   核心接口（Checker, RuleStore, SubjectProvider）
    ├── subject.go                     #   Subject 类型 + WithSubject / SubjectFromContext
    ├── decision.go                    #   Decision{allow, deny, filters, obligations}
    ├── middleware.go                   #   Kratos 中间件（session 校验 + 统一授权决策）
    ├── casbin.go                      #   Casbin → RuleStore 适配层
    ├── matcher.go                     #   proto public_access 白名单匹配器
    ├── enthook.go                     #   Ent hooks（datascope 过滤、ownership 注入、soft delete）
    └── abac/
        ├── plugin.go                  #   ABACPlugin 接口
        └── builtin/
            ├── timerange.go           #   内置：时间范围条件
            └── iprange.go             #   内置：IP 范围条件

    audit/                             # 自包含：审计
    ├── audit.go                       #   核心接口（Recorder, Sink）+ AuditEvent 类型
    ├── middleware.go                   #   Kratos 请求级审计采集中间件
    └── enthook.go                     #   Ent 数据变更审计 hook（before/after diff）

    storage/                           # 自包含：文件
    ├── storage.go                     #   核心接口（FileStore）+ FileMeta 类型
    ├── lifecycle.go                   #   文件生命周期（pending → committed → orphaned → deleted）
    └── backend/
        ├── local.go                   #   本地文件系统 backend
        └── minio.go                   #   MinIO backend（MinIO SDK 适配层）
```

### 4.3 变化点

| 操作 | 说明 |
|------|------|
| 删除 `kratos/middleware/auth/` | matcher 移入 `access/matcher.go` |
| 不新增 `orm/ent/hook/` | hooks 各归其能力包（`access/enthook.go`、`audit/enthook.go`） |
| 新增 `kratos/access/` | 自包含授权能力 |
| 新增 `kratos/audit/` | 自包含审计能力 |
| 新增 `kratos/storage/` | 自包含文件能力 |

### 4.4 开源库利用

| 能力 | 开源库 | 集成方式 |
|------|--------|----------|
| RBAC + ABAC | Casbin | 适配为 RuleStore 接口，ABAC 插件通过 Casbin 自定义函数注册 |
| JWT | golang-jwt/jwt/v5（已在 go.mod） | 直接用于 SubjectProvider 的 token 签发和本地验签 |
| 对象存储 | MinIO SDK | 适配为 FileStore 接口的 backend 实现 |

## 5. app 侧结构

```
apps/yggdrasil/
  api/                                # proto 定义（access, audit, storage, 业务示例）
  clients/admin/                      # 管理前端
  gen/                                # 生成代码
  services/admin/
    cmd/app/                          # 入口（main.go, wire.go）
    configs/config.yaml

    internal/
      platform/                       # 基础设施容器（替代 data.Store）
      │   ├── platform.go             #   Platform{db, cache} + ProviderSet
      │   ├── platform_ent.go
      │   └── platform_cache.go

      access/                         # app 级授权配置
      │   ├── data/                   #   ent schemas（user, role, permission...）
      │   ├── plugin/                 #   app 特有 ABAC plugin（如部门插件）
      │   └── register.go             #   组装 Casbin + JWT + plugins → shared-go 接口

      audit/
      │   ├── data/                   #   ent schemas（audit_event 表）
      │   └── register.go             #   组装 Sink + Recorder

      storage/
      │   ├── data/                   #   ent schemas（file_meta 表）
      │   └── register.go             #   选择 backend + repo → FileStore

      biz/                            # 业务 UC（完全不感知 access/audit/storage）
      data/                           # 业务数据层（接收 *platform.Platform）
      service/                        # 业务 protobuf impls
      server/                         # transport（通过 selector 组装中间件链）
      conf/                           # 配置
```

## 6. Access 能力设计

### 6.1 核心类型

```go
// subject.go
type Subject struct {
    UserID    string
    SessionID string
    Realm     string
    Roles     []string
}

func WithSubject(ctx context.Context, s *Subject) context.Context
func SubjectFromContext(ctx context.Context) (*Subject, bool)

// decision.go
type Decision struct {
    Allow       bool
    Reason      string
    Filters     []ScopeFilter   // 结构化过滤条件（DataScope 输出，不拼接 SQL）
    Obligations []string
}

// checker.go
type Checker interface {
    Check(ctx context.Context, sub *Subject, resource, action string) (*Decision, error)
}

type RuleStore interface {
    LoadPolicies(ctx context.Context) ([]Policy, error)  // Policy 为 Casbin 策略行模型
    LoadRoles(ctx context.Context, userID string) ([]string, error)
}

type SubjectProvider interface {
    VerifyToken(token string) (*Subject, error)
}
```

### 6.2 中间件链

通过 Kratos selector 中间件串联：

```
请求 → selector.Match(PublicAccessMatcher)
  ├── public → 跳过
  └── 非public → SessionMiddleware → AuthorizerMiddleware → 业务逻辑
```

1. **Session Middleware**：从 HTTP header / gRPC metadata 取 token → SubjectProvider.VerifyToken → WithSubject 注入 context
2. **Authorizer Middleware**：SubjectFromContext → 从 proto method descriptor 提取 resource + action → Checker.Check → Decision.Allow → Decision.Filters 注入 context

### 6.3 Ent Hooks

在 ent client 初始化时注册，所有 schema 自动获得：

| Hook | 功能 | 触发时机 |
|------|------|----------|
| DataScope | 从 context 取 Decision.Filters，自动给查询加 WHERE 条件 | Query (interceptor) |
| Ownership | 从 context 取 Subject.UserID，自动填充 created_by / updated_by | Create / Update (hook) |
| SoftDelete | Delete 改为 UPDATE deleted_at，Query 自动加 WHERE deleted_at IS NULL | Delete / Query (interceptor) |

### 6.4 ABAC Plugin 体系

```go
type ABACPlugin interface {
    Name() string
    Evaluate(ctx context.Context, condition map[string]any) (bool, error)
}
```

- 内置插件：TimeRange（工作时间检查）、IPRange（内网 IP 检查）
- app 自定义插件：实现接口，注册到 Casbin。成熟后可上浮到 `builtin/`
- 不需要 ABAC 的 app：不注册任何插件即可

### 6.5 app 侧注册

```go
// access/register.go
func NewChecker(repo *data.AccessRepo) access.Checker {
    return access.NewCasbinChecker(
        repo,
        abac.WithPlugins(
            builtin.TimeRange(),
            builtin.IPRange(),
            plugin.NewDepartmentPlugin(repo),
        ),
    )
}

var ProviderSet = wire.NewSet(
    NewChecker,
    data.NewAccessRepo,
    data.NewEntSchemas,
    plugin.NewDepartmentPlugin,
)
```

## 7. Audit 能力设计

### 7.1 核心类型

```go
type Event struct {
    Timestamp time.Time
    SubjectID string
    Resource  string
    Action    string
    Result    string            // success / deny / error
    Detail    map[string]any
    RequestID string
}

type Recorder interface {
    Record(ctx context.Context, event *Event) error
}

type Sink interface {
    Write(ctx context.Context, events []*Event) error
}
```

### 7.2 两层采集

| 层 | 机制 | 采集内容 |
|----|------|----------|
| Kratos 中间件 | 请求级 | 登录失败、鉴权拒绝、请求完成事件 |
| Ent Hook | 数据级 | Create/Update/Delete 的 before/after diff |

### 7.3 可靠性

- 异步发送，不阻塞业务请求
- 不能静默丢失，至少区分"已写入""待重试""最终失败"
- Sink 可插拔（DB / 日志 / 外部系统）

## 8. Storage 能力设计

### 8.1 核心类型

```go
type FileMeta struct {
    ID         string
    Name       string
    MIMEType   string
    Size       int64
    Status     FileStatus    // pending / committed / orphaned / deleted
    CreatedBy  string
    CreatedAt  time.Time
}

type FileStore interface {
    Upload(ctx context.Context, name string, reader io.Reader) (*FileMeta, error)
    Download(ctx context.Context, id string) (io.ReadCloser, *FileMeta, error)
    Commit(ctx context.Context, id string) error
    Delete(ctx context.Context, id string) error
}
```

### 8.2 文件生命周期

```
pending → committed（业务确认关联）
pending → orphaned（超时未确认，定期清理）
committed → deleted（业务删除）
```

### 8.3 Backend 可插拔

- `local.go`：本地文件系统，开发环境用
- `minio.go`：MinIO，生产环境用（MinIO SDK 适配层）

### 8.4 特点

Storage 没有中间件和 Ent Hook——它是 API 调用而非横切关注点。

## 9. 依赖规则

### 9.1 允许的依赖方向

- server → shared-go/kratos/access（通过 selector 注册中间件）
- service → biz
- biz → repos, DomainServices, shared-go/kratos/access（SubjectFromContext 等工具函数）
- data → platform, ent
- access/data → platform, ent
- audit/data → platform, ent
- storage/data → platform, ent
- wire.go → 所有 ProviderSet

### 9.2 禁止的依赖方向

- server → app 内的 access/audit/storage 包（用 shared-go 的中间件）
- biz/UC → biz/其他UC
- access → audit（能力间不互引）
- access → biz（能力不引业务）
- biz → app 内的 access/audit/storage 包（用 shared-go 的接口和 context 工具）
- data → biz（反向依赖）

### 9.3 server 层中间件组装

```go
// server/http.go
hs := http.NewServer(
    http.Middleware(
        selector.Server(
            access.SessionMiddleware(provider),
            access.AuthorizerMiddleware(checker),
        ).Match(access.PublicAccessMatcher()),
        selector.Server(
            audit.CollectMiddleware(recorder),
        ).Match(audit.MutationMatcher()),  // 可选：只在 mutation 操作时采集
        validate.ProtoValidate(),
        i18n.Server(bundle),
    ),
)
```

## 10. 演进路径

每个闭环独立可验收，不依赖后续闭环。

### 闭环 0：脚手架骨架

- apps/yggdrasil/ 目录结构（api, clients, gen, services/admin）
- internal/platform/ 基础设施容器
- shared-go/kratos/access/ 包骨架（只有接口和类型，无实现）
- 一个 demo 业务 entity，验证 biz/data/service 全链路
- Nx targets（proto, generate, build, dev）

**验收：** `./nx run yggdrasil_admin:dev` 能启动，能 CRUD demo entity

### 闭环 1：Session + RBAC

- shared-go/kratos/access/ 实现：Subject, Checker, SubjectProvider 接口
- JWT 签发 + 本地验签
- Casbin 集成：RBAC 策略加载和评估
- Session 中间件 + Authorizer 中间件 + public_access matcher
- app 侧：ent schemas（user, role, permission, role_binding）+ register.go
- 登录/登出 API + 角色/权限管理 CRUD API

**验收：** 未登录请求被拦截，登录后按角色决定能否访问资源

### 闭环 2：Audit 基础

- shared-go/kratos/audit/ 实现：Event, Recorder, Sink 接口
- 请求级审计中间件
- app 侧：ent schema（audit_event 表）+ DB Sink
- 审计查询 API

**验收：** 登录失败、权限拒绝、业务操作自动记录到审计表

### 闭环 3：Storage 基础

- shared-go/kratos/storage/ 实现：FileStore 接口 + lifecycle
- Backend：local + MinIO
- app 侧：ent schema（file_meta 表）+ register.go
- 上传/下载/确认/删除 API

**验收：** 上传文件 → 获得文件 ID → 业务关联 → 确认提交

### 闭环 4：Ent Hooks

- access/enthook.go：Ownership（created_by/updated_by 自动注入）
- access/enthook.go：SoftDelete
- audit/enthook.go：数据变更审计（before/after diff）

**验收：** 业务代码不写任何审计/ownership 逻辑，数据自动带审计和归属

### 闭环 5：ABAC + DataScope

- access/abac/：Plugin 接口 + 内置插件（TimeRange, IPRange）
- Casbin 自定义函数注册 ABAC 插件
- access/enthook.go：DataScope interceptor（自动查询过滤）
- app 侧：DataScope 规则配置 + 自定义 plugin

**验收：** 配置"销售部只能看自己部门的客户" → 自动过滤生效

**注意：** DataScope 风险最高，最后做，等前面的模式稳定。

## 11. Wire 组装

```go
// wire.go
func wireApp(...) (*kratos.App, func(), error) {
    panic(wire.Build(
        platform.ProviderSet,       // → *Platform (db, cache)
        access.ProviderSet,         // → Checker, SubjectProvider, Middleware
        audit.ProviderSet,          // → Recorder, Sink
        storage.ProviderSet,        // → FileStore
        data.ProviderSet,           // → 业务 repos
        biz.ProviderSet,            // → 业务 use cases
        service.ProviderSet,        // → 业务 service impls
        server.ProviderSet,         // → transport
        newApp,
    ))
}
```

不想要某种能力？删掉对应 ProviderSet 行即可。

注意：`data.ProviderSet` 是业务 repos 的集合（不含 Store，Store 已提升到 platform）。`platform.ProviderSet` 负责基础设施（db、cache 初始化和清理）。
