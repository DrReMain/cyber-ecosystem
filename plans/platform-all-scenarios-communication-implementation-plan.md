# Cyber Ecosystem 全场景通信能力实施方案

## 1. 目标与范围

本方案用于将当前仓库从“CRUD + 多协议共存”升级为“可覆盖全场景业务通信能力”的平台基线，覆盖以下场景：

- 普通 CRUD / 查询 / 命令调用
- 文件上传下载（含大文件、分片、断点续传、Range 下载）
- 实时推送（SSE）
- 双向实时（WebSocket）
- 流媒体与会议（WebRTC 信令 + HLS/LL-HLS 分发治理）
- IoT 设备接入（MQTT）
- 服务间异步事件（EventBus）

本方案遵循仓库硬约束：

- 继续保持 `server -> service -> biz -> data` 依赖方向
- Proto 作为控制面契约 SSOT
- 生成流程和构建流程必须通过 Nx Target 执行
- 不手改受保护生成文件

---

## 2. 核心架构：三平面 + 一致能力层

### 2.1 控制面 Control Plane（Proto SSOT）

定位：

- 统一承接业务意图与策略判定
- 暴露 gRPC / HTTP / Connect 控制接口
- 负责认证鉴权、ACL、配额、会话创建、能力令牌签发、审计

原则：

- 所有需要权限约束的行为，必须先进入控制面
- 控制面决定“谁可以做什么”，不负责高吞吐字节搬运

### 2.2 数据面 Data Plane（字节与媒体）

定位：

- 承载高吞吐字节与低延迟媒体传输

范围：

- 文件分片上传、合并、下载、Range
- WebSocket 全双工消息通道
- SSE 服务端推送
- WebRTC 媒体通道（音视频）
- HLS/LL-HLS 播放分发

原则：

- 数据面只执行被授权能力，不做复杂业务编排
- 通过短时能力令牌与资源绑定实现权限控制

### 2.3 事件面 Event Plane（异步与广播）

定位：

- 作为服务间解耦总线与外部实时推送桥

范围：

- 内部事件发布订阅（建议先落地 NATS 或 Kafka 之一）
- 对接 WS/SSE/MQTT 的实时分发
- 统一失败重试、死信、幂等语义

原则：

- 事件格式标准化，跨服务禁止自定义散乱事件结构

### 2.4 一致能力层 Shared Capabilities（shared-go）

定位：

- 抽象跨协议复用的基础能力，避免在每个服务重复实现

能力基线：

- 认证鉴权
- 能力令牌（capability token）
- 限流与配额
- 幂等键
- 统一错误模型
- 可观测性（Trace / Metric / Log）
- 审计日志

---

## 3. 通信模型与协议选型矩阵

| 场景 | 首选模型 | 协议建议 | 备注 |
|---|---|---|---|
| CRUD / 查询 | Unary | gRPC + HTTP/Connect | 当前已具备 |
| 长任务进度 | 单向推送 | SSE / gRPC server stream | 浏览器优先 SSE |
| 实时互动 | 双向通道 | WebSocket | 浏览器生态最稳 |
| 大文件上传下载 | 控制+数据分离 | Proto 控制接口 + 分片上传/Range 下载 | 支持断点续传 |
| 在线会议 | 信令+媒体分离 | WS/gRPC 信令 + WebRTC 媒体 | 媒体不走普通 RPC |
| 视频点播/直播播放 | 分发协议 | HLS/LL-HLS | 支持 CDN |
| IoT 设备 | 设备协议 | MQTT（设备）+ Proto 控制面 | 管理面仍走控制面 |
| 服务间解耦 | 事件驱动 | NATS 或 Kafka | 统一事件 Envelope |

---

## 4. 仓库目录与模块落位设计

建议新增（或逐步补齐）以下平台目录：

1. `shared-go/auth/capability/`
- 能力令牌签发与校验
- 支持资源绑定、操作绑定、TTL、`jti` 防重放、租户隔离

2. `shared-go/storage/`
- `ObjectStore` 抽象接口
- 实现 `s3minio/` 与 `localfs/`
- 支持 multipart、etag/checksum、range

3. `shared-go/transport/stream/`
- `ws/` 连接管理、心跳、背压
- `sse/` 重连语义与 Last-Event-ID
- `upload/` 分片协议适配
- `range/` 下载与断点续传支持

4. `shared-go/eventbus/`
- 统一 `Publish/Subscribe/Ack/Nack/Retry/DLQ`
- 首期实现 `nats/` 或 `kafka/` 一种

5. `apps/app_1/api/capabilities/*.proto`
- 控制面契约，不承载大字节
- 建议新增：
  - `upload_control.proto`
  - `download_control.proto`
  - `realtime_control.proto`
  - `media_session_control.proto`
  - `iot_session_control.proto`

---

## 5. 分层映射规范（必须遵守）

### 5.1 `internal/server`

- 仅做协议入口、路由、编解码、中间件装配
- 新增 WS/SSE/上传下载 endpoint 也放在此层
- 禁止业务规则与仓储逻辑进入 server

### 5.2 `internal/service`

- 控制协议 DTO 与 usecase 编排
- 保持 proto 契约与 biz 入参出参映射
- 禁止直接依赖 data

### 5.3 `internal/biz`

- 定义业务策略与抽象接口
- 示例：
  - `UploadUC`：上传会话、分片完成、合并确认
  - `RealtimeUC`：订阅、广播、在线状态
  - `MediaUC`：会话协商、流状态编排
- 事务边界在 biz 抽象层定义

### 5.4 `internal/data`

- 实现 biz 定义的仓储与外部适配
- 包括：
  - 对象存储适配（MinIO/S3/LocalFS）
  - 事件总线适配（NATS/Kafka）
  - 会话仓储（Postgres/Redis）

---

## 6. 权限与安全模型（重点）

### 6.1 控制面判定 + 数据面执行

统一流程：

1. 客户端请求控制面（如 `InitiateUpload` / `IssueDownloadGrant`）
2. 控制面完成鉴权、ACL、配额、租户校验
3. 控制面签发短时能力令牌（或预签名 URL）
4. 客户端携带令牌访问数据面
5. 完成后回调控制面确认（如 `CompleteUpload`）

### 6.2 能力令牌最小约束

- 绑定 `subject`（用户/设备/服务身份）
- 绑定 `resource`（对象键、流 ID、频道 ID）
- 绑定 `action`（upload/download/subscribe/publish/play）
- 绑定 `ttl`（默认 5-15 分钟）
- 绑定 `jti`（防重放）
- 可选绑定 `ip/device_fingerprint`

### 6.3 文件场景安全要求

- 分片必须有 etag/checksum
- 合并前必须二次校验分片完整性
- 下载必须支持授权过期与撤销
- 审计记录包含 who/what/when/trace_id

---

## 7. 控制面 Proto 契约建议（草案）

以下为契约方向，字段可按业务细化：

1. 上传控制
- `InitiateUpload(object_name, content_type, size, hash_hint)`
- `IssueUploadPartGrant(upload_id, part_number, part_size)`
- `CompleteUpload(upload_id, parts[])`
- `AbortUpload(upload_id)`
- `GetUploadStatus(upload_id)`

2. 下载控制
- `IssueDownloadGrant(file_id, disposition, expires_in)`
- `RevokeDownloadGrant(grant_id)`
- `GetDownloadAudit(file_id, page, page_size)`

3. 实时控制
- `CreateChannel(channel_type, ttl)`
- `IssueSubscribeGrant(channel_id)`
- `IssuePublishGrant(channel_id)`
- `CloseChannel(channel_id)`

4. 媒体会话控制
- `CreateMediaSession(scene, profile)`
- `JoinMediaSession(session_id, role)`
- `NegotiateSignal(session_id, sdp/ice)`
- `EndMediaSession(session_id)`

5. IoT 会话控制
- `RegisterDevice(device_id, metadata)`
- `IssueDeviceConnectGrant(device_id)`
- `RotateDeviceCredential(device_id)`
- `DisableDevice(device_id)`

---

## 8. Nx 执行与流水线约束

所有改造必须先走 Nx target 检查与执行，不使用 ad-hoc 作为标准路径。

### 8.1 每次 proto 变更的标准命令序列

1. `./nx run tools:buf:format`
2. `./nx run tools:buf:lint`
3. `./nx run contracts:proto`（若改 `contracts/**`）
4. `./nx run app_1_api:proto:api`（若改 `apps/app_1/api/**`）
5. `./nx run app_1_service_1:proto:conf`（若改 service_1 conf proto）
6. `./nx run app_1_service_2:proto:conf`（若改 service_2 conf proto）

### 8.2 每次服务实现变更的标准命令序列

1. `./nx run app_1_service_1:generate` 或 `./nx run app_1_service_2:generate`
2. `./nx run app_1_service_1:build` 或 `./nx run app_1_service_2:build`
3. 运行触达包测试（至少触达新增/改动包）

---

## 9. 分三期实施路线图（建议 8-12 周）

### Phase 1：文件与实时基础能力（优先）

目标：

- 统一文件上传下载控制面
- 打通 WS/SSE 基础设施
- 建立 capability token 基础能力

交付项：

- `shared-go/auth/capability`
- `shared-go/storage`（`localfs` + `s3minio`）
- `shared-go/transport/stream/{ws,sse,upload,range}`
- `upload_control.proto`、`download_control.proto`、`realtime_control.proto`

验收：

- 支持分片上传、断点续传、Range 下载
- 支持频道订阅推送与鉴权
- 鉴权失败、令牌过期、重放攻击测试通过

### Phase 2：事件面统一

目标：

- 平台化事件总线并统一事件 Envelope

交付项：

- `shared-go/eventbus` 接口与首个后端实现（NATS 或 Kafka）
- 服务内发布订阅改造
- WS/SSE 从事件面消费并下发

验收：

- 重试、死信、幂等行为可配置
- 关键事件具备 trace 贯通

### Phase 3：媒体与 IoT 扩展

目标：

- 支撑会议/直播/IoT 复杂场景

交付项：

- `media_session_control.proto`、`iot_session_control.proto`
- WebRTC 信令控制链路
- HLS/LL-HLS 治理接口
- MQTT 接入网关与设备会话管理

验收：

- 会议信令稳定、媒体链路可观测
- 设备接入生命周期可管理（注册、授权、吊销、轮转）

---

## 10. 质量门禁与 DoD

每次能力交付必须满足：

1. 相关 source-of-truth（proto/schema/wire）已按 Nx 规则生成
2. 受影响服务 build target 全通过
3. 测试覆盖触达路径：
- 正常流
- 权限失败
- 令牌过期
- 幂等冲突
- 网络中断恢复（续传/重连）
4. 观测指标接入：
- 请求量、错误率、P95/P99 时延
- 上传吞吐、下载吞吐
- WS/SSE 活跃连接数
- EventBus 消费延迟与堆积
5. 无非预期生成文件改动
6. 分层依赖无违规（禁止 `service -> data` 直连）

---

## 11. 风险与规避策略

1. 风险：把所有流量都塞回业务服务导致扩展瓶颈  
规避：采用控制面与数据面分离，大字节默认直达存储或专用网关

2. 风险：协议增多导致治理失控  
规避：统一 capability token、审计模型、错误码模型、可观测模型

3. 风险：事件语义碎片化  
规避：强制统一 Event Envelope 与 SDK 包装

4. 风险：本地磁盘在多副本环境不可用  
规避：`localfs` 仅开发/单机，生产使用对象存储或共享存储方案

---

## 12. 首次落地建议（下一个迭代可直接开工）

建议先从 `app_1_service_1` 启动最小闭环：

1. 新增 3 个控制面 proto：
- 上传控制
- 下载控制
- 实时频道控制

2. 新增 2 个共享包：
- `shared-go/auth/capability`
- `shared-go/storage`（先 `localfs`，保留 `s3minio` 接口）

3. 新增 2 个 server 通道：
- HTTP 上传/下载（含 Range）
- WS 或 SSE（二选一先落地，建议 SSE 先行）

4. 跑通完整 Nx 流程并补齐触达测试

当上述最小闭环稳定后，再进入 EventBus 和媒体/IoT 扩展。

