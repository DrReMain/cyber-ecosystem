# Yggdrasil Master Restructuring Plan

## 设计原则

本计划替代原有的 5 份水平分层计划（plan1-iam, plan2-audit, plan3-storage, plan4-thinclients, plan5-worklog）。

原计划的问题：按层组织（所有 proto → 所有 schema → 所有 repo → ...），导致执行时等同于重写整个 singleton，且直到最后才能验证架构是否正确。

本计划采用**垂直闭环验证**：每个 Slice 完成一个可独立编译、运行、测试的功能闭环。

## 依赖关系

```
Slice 0: App Skeleton
  └─► Slice 1: Audit Service (最简服务，验证全栈)
       └─► Slice 2: IAM Auth (JWT 登录/登出)
            ├─► Slice 3: IAM RBAC + Event Bus (角色权限 + 事件总线)
            │    └─► Slice 4: IAM Advanced (条件/数据权限/部门/用户属性)
            │         └─► Slice 5: Thin Clients (shared-go/capabilities)
            │              ├─► Slice 6: Storage Service
            │              └─► Slice 7: Worklog Service
            └───────────────────────────────────────────────────────┘
```

## Slice 总览

| Slice | 目标 | 新增文件 | 检查点 |
|-------|------|----------|--------|
| 0 | App 骨架 + Nx/Buf 配置 | ~5 | `buf lint` 通过 |
| 1 | Audit 服务（完整闭环） | ~25 | gRPC 提交事件 + HTTP 查询 |
| 2 | IAM 认证（登录/登出/内部API） | ~30 | HTTP 登录 + 内部 ValidateToken |
| 3 | IAM RBAC + Event Bus | ~40 | 删除角色 → 事件级联清理 |
| 4 | IAM 高级能力 | ~35 | 条件/数据权限集成测试 |
| 5 | Thin Client 模块 | ~30 | `shared-go/capabilities/` 编译 + 测试 |
| 6 | Storage 服务 | ~25 | 文件上传/下载 |
| 7 | Worklog 服务（业务模板） | ~25 | 完整 CRUD + 全能力集成 |

## 每个 Slice 的闭环标准

一个 Slice **完成** 的定义：

1. `go build` 编译通过
2. 服务可启动（或模块可导入）
3. 至少一个端到端验证命令执行成功
4. 所有变更已提交

## 约定

### 模块路径

```
cyber-ecosystem/apps/yggdrasil/...      # 应用代码
cyber-ecosystem/shared-go/...            # 共享库
cyber-ecosystem/contracts/go/...         # 共享 Proto 类型
```

### 端口分配

| 服务 | HTTP | gRPC | ConnectRPC | Ops |
|------|------|------|------------|-----|
| iam | 11000 | 12000 | 13000 | 14000 |
| audit | 11001 | 12001 | 13001 | 14001 |
| storage | 11002 | 12002 | 13002 | 14002 |
| worklog | 11003 | 12003 | 13003 | 14003 |

### 目录结构约定

```
apps/yggdrasil/
  api/v1/                    # App 级 proto 定义
  gen/go/v1/                 # 生成的 Go stubs
  gen/oas/                   # OpenAPI spec
  clients/admin/             # Admin 前端
  services/<service>/
    cmd/app/                 # main.go, wire.go
    configs/config.yaml
    internal/
      conf/                  # 配置 schema (proto)
      server/                # 传输层
        middleware/
        locales/
      service/               # Service 层
      biz/                   # 业务逻辑层
      data/                  # 数据访问层
        ent/
          schema/
      pkg/                   # 内部工具（仅 IAM 需要）
      event/                 # 事件总线（仅 IAM 需要）
```

### 通用模式参考

以下模式在所有服务中一致使用，各 Slice 计划中不再重复完整代码，只说明要适配的路径和名称：

- **Store 模式**: `data/store.go` — Store 结构体持有 `*cache.Cache` + `*ent.Client`，提供 `InTx`, `GetClient`, `GetCache`
- **Ent 客户端模式**: `data/store_ent.go` — `NewEntClient` 使用 `client.NewEntClient(DBConfig{...})`，`HandleError` 使用 `entutil.HandleEntError`
- **缓存模式**: `data/store_cache.go` — `NewCache` 分发到 Redis 或 Memory 实现
- **Server 模式**: `server/server.go` — `init()` 重映射框架错误到 proto 错误码，`ProviderSet` 包含所有 server 构造函数
- **Ops 模式**: `server/ops.go` — 条件注册 Prometheus metrics + pprof
- **i18n 模式**: `server/i18n.go` — `//go:embed locales/*.yaml` + `i18n.NewBundleFS`
- **Main 模式**: `cmd/app/main.go` — config 加载 → logger → OTel → wire → run
- **Wire 模式**: `cmd/app/wire.go` — `wire.Build(server.ProviderSet, service.ProviderSet, biz.ProviderSet, data.ProviderSet, newApp)`
- **Service 基础模式**: `service/service.go` — `Registrar` 接口 + `NewRegistrarList` + `ProviderSet`

### 通用文件模板

每个服务都需要的通用文件，按以下模式创建（只改路径和名称）：

**data/data.go:**
```go
package data

import (
    "github.com/google/wire"
    "github.com/go-kratos/kratos/v2/log"
    "<module>/internal/biz"
)

type RP struct {
    log   *log.Helper
    store *Store
}

var ProviderSet = wire.NewSet(
    NewStore,
    NewCache,
    NewEntClient,
    wire.Bind(new(biz.Transaction), new(*Store)),
    // + 各 RP 构造函数
)
```

**service/service.go:**
```go
package service

import (
    "github.com/google/wire"
    "github.com/go-kratos/kratos/v2/transport/grpc"
    krahttp "github.com/go-kratos/kratos/v2/transport/http"
    "cyber-ecosystem/shared-go/kratos/transport/connect"
)

type Registrar interface {
    RegisterGRPC(*grpc.Server)
    RegisterHTTP(*krahttp.Server)
    RegisterConnect(*connect.Server)
}

var ProviderSet = wire.NewSet(
    NewRegistrarList,
    // + 各 Service 构造函数
)

func NewRegistrarList(/* 各 Service 指针 */) []Registrar {
    return []Registrar{ /* ... */ }
}
```

## Slice 计划文件

| 文件 | 内容 |
|------|------|
| `2026-04-29-yggdrasil-slice0-skeleton.md` | App 骨架 |
| `2026-04-29-yggdrasil-slice1-audit.md` | Audit 服务 |
| `2026-04-29-yggdrasil-slice2-iam-auth.md` | IAM 认证 |
| `2026-04-29-yggdrasil-slice3-iam-rbac.md` | IAM RBAC + Event Bus |
| `2026-04-29-yggdrasil-slice4-iam-advanced.md` | IAM 高级能力 |
| `2026-04-29-yggdrasil-slice5-thinclients.md` | Thin Client 模块 |
| `2026-04-29-yggdrasil-slice6-storage.md` | Storage 服务 |
| `2026-04-29-yggdrasil-slice7-worklog.md` | Worklog 服务 |
