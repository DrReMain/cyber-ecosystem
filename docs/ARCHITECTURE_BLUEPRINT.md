# Cyber Ecosystem 架构演进设计草案

> 文档目标：为仓库所有者与后续 AI agent 提供一份共享的顶层架构蓝图、演进边界、执行顺序与约束基线。
>
> 当前状态：仓库已完成目录结构演进第一阶段，建立了 `contracts/`、`examples/`、`shared-go/`、`gen/` 等顶层目录，Go/Kratos 基座已稳定，正在推进平台化治理能力建设。

## 1. 文档定位

本草案不是单个服务的开发规范，而是整个 monorepo 的平台级设计说明。

适用对象：

- 仓库所有者
- 负责后端、前端、客户端、基础设施的开发者
- 后续参与执行任务的 AI agent

本草案解决四个问题：

1. 这个仓库最终要长成什么样
2. 哪些能力应该放进仓库，哪些不该混在一起
3. 应该先做什么，再做什么
4. AI agent 后续接手任务时应该遵守哪些架构边界

---

## 2. 顶层愿景

将当前仓库演进为一个可支撑以下目标的产品级 monorepo 平台：

- 快速孵化不同业务形态的系统项目
- 支持多语言后端：Go、Rust、Python
- 支持多协议传输：gRPC、HTTP/JSON、Connect-RPC、WebSocket、SSE、流式传输
- 支持多终端客户端：Web SSR、Web SPA、iOS、Android、Flutter、Desktop、Game Client
- 支持多种高复杂业务场景：AI、IoT、数字孪生、无人机/机器人控制、流媒体、游戏
- 在保持开发效率的同时，逐步具备生产级治理能力

核心结论：

- 仓库不应只是一组服务模板的集合
- 仓库应演进为“平台工程 + 契约工程 + 业务工程”的统一载体

---

## 3. 当前阶段判断

### 3.1 已具备的基础

- 已采用 `Buf + Proto` 作为契约中心
- 已采用 `Nx` 作为 monorepo 顶层编排工具
- 已沉淀 Kratos 服务骨架与共享 Go 工具库
- 已具备基础数据库、追踪、代码生成流程
- 已形成 `template1`、`template2` 两个可复用参考实现
- **已完成目录结构演进**：建立了 `contracts/`、`examples/`、`shared-go/`、`gen/` 等顶层目录
- **契约层已独立**：公共 Proto 定义已迁移至 `contracts/proto/`
- **示例服务已归档**：`template1`、`template2` 已迁移至 `examples/` 目录

### 3.2 当前主要短板

- 服务治理能力不完整
- 多语言支持尚未产品化
- 多传输支持尚未形成统一抽象
- 客户端与 SDK 体系尚未落地
- 文件、媒体、实时通信、IoT、游戏等专项能力尚未分层设计

---

## 4. 架构原则

后续所有新增能力，默认遵守以下原则。

### 4.1 契约优先

- 所有跨进程、跨语言、跨端通信优先以 Proto 契约定义
- OpenAPI 只是衍生产物，不是唯一契约
- Connect-RPC、gRPC、HTTP、SDK 生成都应围绕统一契约层展开

### 4.2 平台与业务分离

- 平台能力要沉淀为通用基础设施或共享组件
- 业务服务不能重复实现鉴权、链路追踪、限流、分页、传输适配等通用能力

### 4.3 控制面与数据面分离

- 配置、发布、治理、审计、观测属于控制面
- 文件传输、流媒体、实时连接、设备接入、游戏同步属于数据面
- 控制面和数据面不能混在一个服务模型中

### 4.4 同步与异步分离

- 同步 RPC 用于查询、低复杂编排、低延迟交互
- 事件总线、任务调度、流式处理用于解耦、削峰、耗时处理
- 不允许把所有跨服务协作都设计为同步 RPC

### 4.5 多语言职责明确

- Go：主业务服务、常规平台服务、BFF、管理型服务
- Rust/Tonic：高性能、低延迟、流式、实时、网关型服务
- Python：AI 推理、模型编排、智能处理、实验性能力

### 4.6 客户端直连优先考虑契约一致性

- 能直接通过契约生成 SDK 的场景，不应长期依赖“OpenAPI -> 手写转换 -> 客户端”
- Connect-RPC 应作为 Web 与移动端的重要一等能力，而不是附属能力

---

## 5. 目标能力版图

后续 monorepo 需要覆盖的能力版图如下。

### 5.1 基础平台能力

- 统一构建编排
- 代码生成与脚手架
- 配置管理
- 服务发现
- 流量治理
- 日志、指标、链路追踪
- 鉴权、授权、租户、审计
- 消息总线与任务调度
- 发布、回滚、灰度、兼容性检查

### 5.2 后端服务能力

- Kratos 微服务
- Rust/Tonic 高性能 gRPC 服务
- Python AI gRPC 服务
- BFF / Gateway
- Realtime Gateway
- Device Gateway
- File / Media Platform
- Game Realtime Service

### 5.3 传输能力

- gRPC
- HTTP/JSON
- Connect-RPC
- WebSocket
- Server-Sent Events
- gRPC Streaming
- 二进制上传下载
- 流媒体切片与播放

### 5.4 客户端能力

- Next.js SSR
- Vite SPA
- iOS
- Android
- Flutter
- Tauri
- Qt
- Slint
- GPUI
- Godot
- UE5

### 5.5 业务场景能力

- AI 应用
- 物联网
- 数字孪生
- 无人机与机器人控制
- 文件与媒体系统
- 实时协作
- 游戏业务

---

## 6. 仓库顶层结构

### 6.1 当前结构（已完成第一阶段演进）

```text
cyber-ecosystem/
├── docs/                       # 文档与设计草案
├── contracts/                  # 契约定义
│   └── proto/                  # 公共 Proto 定义
│       ├── common/             # 公共类型
│       └── errors/             # 错误定义
├── gen/                        # 生成产物
│   ├── go/                     # 生成的 Go 代码
│   └── oas/                    # OpenAPI 规范
├── examples/                   # 示例服务
│   ├── template1/              # 示例服务1（Blog CRUD）
│   └── template2/              # 示例服务2（Reading）
├── shared-go/                  # Go 共享库
│   ├── kratos/                 # Kratos 相关工具
│   └── orm/                    # ORM 相关
├── clients/                    # 客户端项目
│   └── admin/                  # 管理后台（Next.js）
├── tools/                      # 开发工具与脚本
│   └── docker/                 # Docker Compose 配置
├── go.mod                      # Go 模块定义
├── buf.yaml                    # Buf 配置
└── nx.json                     # Nx 配置
```

### 6.2 目标结构（长期演进方向）

```text
cyber-ecosystem/
├── docs/                       # 文档与设计草案
├── contracts/                  # 所有契约定义
│   ├── proto/                  # RPC / API Proto
│   ├── events/                 # 事件契约
│   ├── models/                 # 公共领域模型
│   └── errors/                 # 错误模型
├── gen/                        # 生成产物
│   ├── go/
│   ├── ts/
│   ├── swift/
│   ├── kotlin/
│   ├── dart/
│   ├── rust/
│   └── oas/
├── services/                   # 业务服务
├── gateways/                   # 网关与接入层
│   ├── api-gateway/
│   ├── realtime-gateway/
│   ├── connect-gateway/
│   └── device-gateway/
├── platform/                   # 平台服务
│   ├── identity/
│   ├── config/
│   ├── observability/
│   ├── messaging/
│   ├── scheduler/
│   └── file-media/
├── sdk/                        # 多端 SDK
│   ├── ts/
│   ├── swift/
│   ├── kotlin/
│   ├── dart/
│   ├── csharp/
│   └── rust/
├── clients/                    # 客户端项目
│   ├── web-next/
│   ├── web-vite/
│   ├── mobile-flutter/
│   ├── desktop-tauri/
│   ├── desktop-qt/
│   ├── game-godot/
│   └── game-ue5/
├── shared-go/                  # Go 共享库（已落地）
├── shared-ts/                  # TypeScript 共享库
├── shared-rust/                # Rust 共享库
├── shared-python/              # Python 共享库
├── infra/                      # IaC / 部署 / 运行环境
│   ├── docker/
│   ├── k8s/
│   ├── terraform/
│   └── helm/
├── tools/                      # 开发工具与脚本
└── nx.json
```

### 6.3 设计决策

- **契约层**：`contracts/` 只包含公共 Proto 定义，服务特定的 API 定义保留在服务目录内
- **示例服务**：`examples/` 存放参考实现，不作为生产服务模板
- **共享库命名**：使用 `shared-{lang}` 格式（如 `shared-go`），而非 `shared/{lang}`
- **服务目录**：暂不按语言分目录，服务数量增加后再考虑

---

## 7. 契约层设计

契约层是未来所有扩展的基础，必须优先稳定。

### 7.1 契约分类

建议明确拆分四类契约：

- 公共模型契约：如分页、时间范围、错误模型、身份上下文
- 内部 RPC 契约：服务间调用
- 外部 API 契约：面向客户端或公网
- 事件契约：消息总线、异步任务、流处理

### 7.2 契约治理要求

- 所有 proto 必须具备版本策略
- 启用 breaking change 检查
- 统一命名、包名、目录规则
- 统一代码生成路径
- 所有语言 SDK 以契约生成优先

### 7.3 Connect-RPC 策略

Connect-RPC 建议作为正式 transport 引入，目标如下：

- Kratos 服务对外可同时提供 gRPC、HTTP、Connect-RPC
- Web 与移动端优先直连 Connect-RPC
- OpenAPI 保留用于生态兼容和文档展示，不再承担唯一客户端接入职责

---

## 8. 服务分层与语言分工

### 8.1 Go / Kratos

适合承载：

- 业务微服务
- 管理后台 BFF
- 统一鉴权与平台服务
- 常规 CRUD 与后台系统
- 对外 HTTP / gRPC API

### 8.2 Rust / Tonic

适合承载：

- 高性能 gRPC 服务
- 实时网关
- 长连接核心通道
- 低延迟设备接入
- 游戏房间与状态同步
- 流式数据处理

约束：

- Rust 不应用来抢占所有通用业务服务
- 只有对吞吐、延迟、内存占用有明确要求时才优先选 Rust

### 8.3 Python

适合承载：

- AI 推理服务
- 模型编排
- 多模型路由
- 文本、图像、音频等智能处理
- 实验性智能能力

约束：

- Python 不承担高频实时主链路
- Python 服务必须以 gRPC/Proto 方式接入平台

---

## 9. 传输层架构

当前最大的架构缺口之一，是传输层能力尚未统一规划。

### 9.1 建议的传输分工

- gRPC：服务间调用、高性能客户端、内部标准 RPC
- HTTP/JSON：公共开放接口、浏览器兼容接口、生态兼容
- Connect-RPC：Web/移动端直连契约接口
- WebSocket：双工长连接、实时协作、设备控制、游戏
- SSE：服务端单向推送、轻量实时通知
- Streaming：大文件、媒体流、增量数据、实时数据面

### 9.2 架构约束

- 不能在每个业务服务里随意实现 WS/SSE/Streaming
- Realtime 相关能力应集中在 `realtime-gateway`
- 大文件与媒体流应集中在 `file-media` 平台
- Device 接入应集中在 `device-gateway`

---

## 10. 文件、媒体与流式能力规划

文件与流媒体能力不能当作普通业务接口的附属功能。

建议拆出独立平台能力：

- 对象存储接入
- 分片上传
- 断点续传
- 权限签名 URL
- 下载加速
- 文件元数据管理
- 媒体转码
- 缩略图/预览图
- HLS/DASH 切片
- 流式播放
- 内容审核接口

设计原则：

- 文件元数据走业务服务
- 文件实际传输走对象存储或媒体平台
- 大流量传输不穿透普通业务服务

---

## 11. Realtime、IoT、数字孪生、机器人与无人机

这些场景不应沿用“普通 CRUD 微服务”思维。

### 11.1 Realtime Gateway

职责：

- WebSocket / SSE 统一接入
- 连接管理
- 订阅与广播
- 在线状态管理
- 心跳与断线恢复
- 背压与消息投递策略

### 11.2 Device Gateway

职责：

- 设备鉴权
- 设备协议适配
- 命令下发
- 遥测上报
- 时序数据采集
- 边缘节点接入

建议支持的协议方向：

- gRPC
- WebSocket
- MQTT
- 自定义二进制协议适配层

### 11.3 数字孪生与控制场景

应拆成三个域：

- 设备接入域
- 状态建模域
- 控制调度域

约束：

- 控制命令、遥测数据、业务查询不要混在一个服务里
- 需要独立定义时序数据存储、事件流与控制确认机制

---

## 12. 游戏场景规划

游戏业务后端与常规企业后台不是一类系统。

### 12.1 可复用常规服务的部分

- 账号体系
- 支付
- 背包与资产
- 配置中心
- 活动系统
- 管理后台

### 12.2 需要单独设计的部分

- 匹配服务
- 房间服务
- 状态同步
- 实时事件广播
- 帧同步或状态同步逻辑
- 语音/实时交互相关能力

建议：

- Godot、UE5 客户端接入优先考虑 gRPC/Connect/WS 统一协议层
- 房间服和实时同步层更适合独立性能链路，优先考虑 Rust

---

## 13. 客户端体系规划

客户端未来会非常多样，必须提前定义接入模型。

### 13.1 Web

- Next.js：面向 SSR、SEO、BFF、内容型和混合交互场景
- Vite：面向后台工具、工作台、高交互 SPA

### 13.2 Mobile

- 原生 iOS / Android：高性能、高平台能力需求
- Flutter：跨平台业务客户端

### 13.3 Desktop

- Tauri：轻量桌面应用
- Qt：传统桌面与工业控制场景
- Slint / GPUI：高性能桌面 UI 探索方向

### 13.4 Game

- Godot：轻量与中型项目
- UE5：高表现力与重交互项目

### 13.5 客户端统一原则

- 客户端优先消费统一契约生成的 SDK
- 客户端不直接依赖业务服务内部细节
- 客户端与服务端通信优先通过 gateway 或稳定外部契约

---

## 14. 平台治理能力规划

这是从 demo 仓库走向生产平台的关键部分。

### 14.1 必须补齐的治理能力

- 服务发现
- 配置中心
- 统一日志规范
- 统一 tracing / metrics / profiling
- 超时、重试、熔断、限流
- 灰度发布
- 统一错误码体系
- 鉴权与权限模型
- 租户隔离
- 审计日志
- 幂等控制
- 任务调度
- 事件总线

### 14.2 平台优先级原则

- 没有统一鉴权前，不要大规模扩服务
- 没有统一观测前，不要大规模引入异步链路
- 没有统一契约治理前，不要大量接客户端

---

## 15. 数据与存储策略

不同业务场景需要不同存储模型，不能假设一个 PostgreSQL 足够覆盖全部。

### 15.1 建议的存储分层

- PostgreSQL：事务型业务数据
- Redis：缓存、会话、热点数据、轻量队列
- 对象存储：大文件、媒体、附件
- 时序数据库：设备遥测、工业数据、监控数据
- 搜索引擎：全文检索、日志检索、内容检索
- 消息系统：事件流、解耦、异步处理

### 15.2 数据原则

- 业务库不直接承担大文件存储
- 高频时序数据不直接落常规关系库
- 实时状态与历史归档分开处理

---

## 16. 安全与权限模型

至少拆分三套身份模型：

- 用户身份
- 服务身份
- 设备身份

必须支持：

- JWT / Session / Token 模式的统一抽象
- 服务间凭证
- 设备注册与绑定
- 权限策略与资源授权
- 审计与追责

---

## 17. 可观测性与 SLO

未来扩到 AI、媒体、IoT、游戏后，没有统一可观测性会很快失控。

建议标准化以下内容：

- Trace ID 全链路透传
- 日志字段规范
- 统一 metrics 命名规范
- error budget 与 SLO 定义
- p95 / p99 延迟分层指标
- 队列积压与实时通道健康指标
- 连接数、广播量、流量峰值指标

---

## 18. 构建、生成与发布体系

`Nx` 后续应从“命令入口”升级为“平台编排核心”。

### 18.1 目标能力

- affected build/test/lint
- 依赖图分析
- 远程缓存
- 多语言 target 规范化
- 契约生成统一流水线
- SDK 发布流水线
- 文档生成流水线

### 18.2 代码生成方向

- 新服务脚手架
- 新 proto 脚手架
- 新 gateway 脚手架
- 新 SDK 生成任务
- Connect-RPC 相关生成任务
- Rust/Python 服务模板生成任务

---

## 19. 演进阶段规划

以下阶段是建议的执行顺序，不建议跳步。

### 阶段 A：稳定当前 Go/Kratos 基座 ✅ 已完成

目标：

- 统一目录命名
- 固化服务模板
- 固化契约规范
- 固化共享库边界
- 补齐基础 CI 与代码生成任务

产出：

- ✅ Go 服务模板 v1（`examples/template1`、`examples/template2`）
- ✅ 契约规范 v1（`contracts/proto/`）
- ✅ 项目脚手架 v1（Nx 项目配置）
- ✅ 平台命令规范 v1（`docs/AGENTS.md`、`docs/DEVELOPMENT.md`）

### 阶段 B：把 monorepo 升级为平台仓 ✅ 已完成

目标：

- 引入项目分类结构
- 建立平台层与业务层边界
- 建立统一 SDK 与生成物目录
- 建立统一文档体系

产出：

- ✅ 顶层目录重构（`contracts/`、`examples/`、`shared-go/`、`gen/`）
- ✅ 文档体系（`docs/ARCHITECTURE_BLUEPRINT.md`、`docs/AGENTS.md`、`docs/DEVELOPMENT.md`）
- 平台层项目清单（待补充）

### 阶段 C：补齐治理能力 🔄 进行中

目标：

- 配置中心
- 观测体系
- 服务治理
- 鉴权权限
- 任务调度
- 消息总线

产出：

- 平台基础服务 v1
- 通用中间件栈 v1
- 服务治理基线 v1

### 阶段 D：引入 Connect-RPC 与多端 SDK

目标：

- 为 Kratos 增加 Connect-RPC transport
- 建立 TypeScript 优先的客户端契约直连链路
- 形成 SDK 生成与发布流程

产出：

- Connect transport PoC
- Web SDK v1
- 移动端 SDK 方案 v1

### 阶段 E：引入 Rust/Tonic 与 Python gRPC

目标：

- 建立 Rust 服务模板
- 建立 Python AI 服务模板
- 建立多语言服务准入规范

产出：

- Rust/Tonic 模板 v1
- Python gRPC 模板 v1
- 多语言服务边界文档 v1

### 阶段 F：建设 Realtime / File / Device 专项平台

目标：

- Realtime Gateway
- File / Media Platform
- Device Gateway

产出：

- 长连接接入层 v1
- 文件媒体平台 v1
- 设备接入平台 v1

### 阶段 G：专项场景拓展

目标：

- 数字孪生
- 机器人/无人机控制
- 游戏服务能力

产出：

- 场景专项参考架构
- 对应服务模板与协议规范

---

## 20. 未来 12 个月的建议优先级

优先级从高到低如下：

1. ✅ 稳定契约层与服务模板
2. ✅ 建立目录重构与脚手架能力
3. 🔄 补齐服务治理与可观测性
4. 引入 Connect-RPC
5. 建立 Web SDK 和客户端接入规范
6. 引入 Rust/Tonic 模板
7. 引入 Python AI 模板
8. 建设 Realtime Gateway
9. 建设 File / Media Platform
10. 建设 Device Gateway
11. 扩展数字孪生与机器人控制能力
12. 扩展游戏专项能力

说明：

- 不建议一开始就并行做所有方向
- 先统一基座，再接专项能力

---

## 21. AI Agent 执行约束

后续任何 AI agent 参与本仓库建设时，默认遵守以下规则。

### 21.1 允许优先推进的事项

- 契约治理
- 模板固化
- 共享库抽象
- 目录重构
- 构建与生成流程
- Connect-RPC 接入
- Rust/Python 模板
- Realtime/File/Device 平台骨架

### 21.2 不允许擅自推进的事项

- 未经确认的大规模目录迁移
- 直接把所有服务升级成多语言混合实现
- 在没有边界文档前引入新的基础设施
- 在没有统一抽象前让每个服务自行实现 WebSocket / SSE / 上传下载
- 在没有契约治理前大量生成客户端 SDK

### 21.3 新增能力时的决策顺序

1. 先判断能力属于平台还是业务
2. 再判断能力属于控制面还是数据面
3. 再判断是否需要新契约类型
4. 再判断应使用 Go、Rust 还是 Python
5. 最后才选择服务模板与目录位置

---

## 22. 近期建议立项清单

建议立即拆成以下可执行项目：

- ✅ 项目 1：monorepo 顶层目录演进方案（已完成）
- ✅ 项目 2：Go/Kratos 服务脚手架生成器（已完成基础模板）
- ✅ 项目 3：Proto/Buf 契约治理规范（已完成）
- 🔄 项目 4：服务治理基础中间件包（进行中）
- 项目 5：Connect-RPC transport 设计与 PoC
- 项目 6：Rust/Tonic 服务模板
- 项目 7：Python AI gRPC 服务模板
- 项目 8：Realtime Gateway 设计
- 项目 9：File / Media Platform 设计
- 项目 10：Device Gateway 设计

---

## 23. 本草案的使用方式

你和后续 AI agent 可以这样使用这份草案：

- 做大方向判断时，把本文件作为顶层原则
- 做具体实现时，再结合 `docs/AGENTS.md` 与 `docs/DEVELOPMENT.md`
- 遇到新业务场景时，先判断是否已被本草案覆盖
- 若未覆盖，先更新本草案，再推进大规模实现

---

## 24. 当前建议结论

当前仓库已完成第一阶段演进，Go/Kratos 基座已稳定，目录结构已优化。

下一步方向是：

- ✅ 把 Go/Kratos 基座做稳（已完成）
- ✅ 把 monorepo 变成平台仓（已完成第一阶段）
- 🔄 补治理、Connect、多语言与专项能力（进行中）
- 最后扩展到 IoT、媒体、实时、游戏等高复杂场景

如果后续要让 AI agent 辅助执行，应默认以本草案为总纲，以阶段任务为切分单位推进，而不是让 agent 在没有边界的情况下自由扩张仓库。
