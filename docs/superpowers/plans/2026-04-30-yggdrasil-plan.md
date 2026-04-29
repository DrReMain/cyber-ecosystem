# Yggdrasil 实现计划

基于 `docs/superpowers/specs/2026-04-30-yggdrasil-scaffold-design.md` 的粗粒度计划。
执行过程中交互式细化，不预设详细步骤。

## 阶段 A：基础设施准备

**目标：** 重组 shared-go，更新 template 适配新模式

### A1. shared-go 重组
- 删除 `kratos/middleware/auth/`（matcher 后续在 access 能力中重建）
- 新增 `kratos/access/` 包骨架（只有接口和类型定义，无实现）
- 新增 `kratos/audit/` 包骨架
- 新增 `kratos/storage/` 包骨架
- 更新 template 中对 `kratos/middleware/auth/` 的引用

### A2. template 适配 platform 模式
- `internal/data/store.go` → `internal/platform/platform.go`（提升 Store 为 Platform）
- `internal/data/store_ent.go` → `internal/platform/platform_ent.go`
- `internal/data/store_cache.go` → `internal/platform/platform_cache.go`
- 更新所有 repos 从 `*Store` 改为 `*platform.Platform`
- 更新 wire.go 和 ProviderSet
- 验证 template 仍然能正常 build 和 dev

**验收：** `./nx run template_base:build` 通过，`./nx run template_base:dev` 正常运行

## 阶段 B：脚手架骨架（闭环 0）

**目标：** 创建 yggdrasil app，跑通全链路

- 从 template 复制基础结构，创建 `apps/yggdrasil/`
- 调整为 yggdrasil 的命名和配置
- 包含 platform/ + 一个 demo 业务 entity
- Nx targets 可用

**验收：** `./nx run yggdrasil_admin:dev` 能启动，能 CRUD demo entity

## 阶段 C：Session + RBAC（闭环 1）

**目标：** 最小可用的身份认证 + 角色权限

- 实现 `shared-go/kratos/access/` 的核心逻辑（Subject, JWT, Casbin 适配）
- 实现 Session + Authorizer 中间件
- 实现 public_access matcher（selector 用）
- app 侧：user/role/permission ent schemas + 登录/角色管理 API
- 管理前端：登录页 + 角色/权限管理页面

**验收：** 未登录被拦截，登录后按角色访问资源

## 阶段 D：Audit 基础（闭环 2）

**目标：** 统一审计采集和存储

- 实现 `shared-go/kratos/audit/`（Event, Recorder, Sink）
- 请求级审计中间件
- app 侧：audit_event 表 + DB Sink + 查询 API

**验收：** 登录失败、权限拒绝自动记录

## 阶段 E：Storage 基础（闭环 3）

**目标：** 文件上传下载 + 生命周期

- 实现 `shared-go/kratos/storage/`（FileStore, lifecycle, backend）
- app 侧：file_meta 表 + 上传/下载 API

**验收：** 上传 → 关联 → 确认 完整流程

## 阶段 F：Ent Hooks（闭环 4）

**目标：** 数据级横切能力

- access/enthook.go：Ownership + SoftDelete
- audit/enthook.go：数据变更审计

**验收：** 业务代码无感知，数据自动带审计和归属

## 阶段 G：ABAC + DataScope（闭环 5）

**目标：** 环境条件判断 + 数据范围过滤

- ABAC Plugin 体系 + 内置插件
- DataScope interceptor

**验收：** 配置部门过滤规则 → 自动生效

---

每个阶段开始时细化具体步骤，执行过程中交互校准。
