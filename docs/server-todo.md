# 后端平台后续工作（TODO）

后续路线：mobile 业务、可观测收尾、edge_admin（多服务）。平台基础已就绪（db + 三 transport + 中间件链 + cache/storage/mq + 可观测全链路 → SigNoz，均接入验证），本文只列**未完成**项。

---

## 架构与边界（两服务）

> 两个服务：**edge_mobile（边缘/消费面）+ edge_admin（管理面）**。**主权数据**：各管自己的 DB，无跨库直读。

- **edge_mobile**：终端用户系统。owns `mobile_user` + 消费侧业务。**必须能独立运行**（每次请求鉴权，不依赖 admin）—— 稳定内核。
- **edge_admin**：内部管理系统。owns `admin_user` + RBAC/ABAC/datascope。通过 **RPC 管理 mobile 的 `mobile_user`**（不直读 mobile 库）。`mobile_user` ≠ `admin_user`。
- **依赖方向**：**admin → mobile（单向，不环）**。mobile 先行。

**mobile 设计时带上 admin 意识**：
1. 数据模型带上 admin 要的字段（status/软删、scope、审计）—— 加字段现在便宜，迁移贵。
2. admin 要调的内部管理 RPC 挂 proto —— 只实现现在用得到的，不投机建整套。

**分布式原则**：
- 分布式事务设计上回避（mobile 库内原子 + admin 记审计，最终一致；真要跨库原子才 saga/outbox）。
- 服务发现 = k3s DNS（`<svc>.<ns>.svc.cluster.local`），不自建 discovery/etcd。

---

## 路线图

| 阶段 | 事项 | 状态 |
|---|---|---|
| **三** | [可观测收尾：慢查询 log + 后端异常上报](#阶段三可观测收尾--异常上报) | **下一步** |
| **四** | [edge_admin + 多服务能力（F）](#阶段四edge_admin--多服务能力f) | 骨架成型后 |
| **五** | [生产运维（H）+ 错误守卫（A）](#阶段五生产运维h--错误守卫a) | 主线 / 延后 |
| | [mobile auth / 会话子系统](#mobile-业务-backlogauth--会话子系统后续独立设计) | 后续 |

> 顺序：**三 → 四 → 五**。mobile auth 是独立业务特性，按业务节奏排，不绑平台阶段。

---

## 阶段三：可观测收尾 + 异常上报

trace/metric/log 全链路 + db/cache/storage/mq 四后端 span 已完成（SigNoz 验证）。剩余两件小事 + 异常上报分流架构：

**异常上报架构（后端 vs 客户端）**：
- **后端异常 → SigNoz**（OTel log：panic/未受控错误 → 结构化 error log，带 trace_id 关联；recovery 中间件落地）。后端异常天然跟 trace 相关，留 SigNoz 保持统一（trace/metric/log/error 一处）。权衡：错误 dedup/grouping 不如 Sentry 级，早期够用。
- **客户端异常 → GlitchTip**（自建 Sentry 协议，sentry-go SDK；crash 报告主场）。
- **客户端全链路 OTel**：各客户端引入 OTel SDK，链路从客户端起。**后端已就绪**（kratos tracing 中间件 W3C Extract，客户端带 traceparent 即续链，后端零改动）。属各客户端项目投入，非本仓库事。

**待办（后端残余）**：
- **慢查询 log**：db/cache/storage/mq 操作超阈值 → slog warn → OTLP log → SigNoz。阈值进 conf。
- **后端异常上报 → SigNoz**：recovery 中间件把 panic/未受控错误发结构化 error log（OTel log sink 已在）。

---

## 阶段四：edge_admin + 多服务能力（F）

后端骨架成型后起 edge_admin。**复用 shared-go**（cache/observability/transport/错误模型/ent 工具），净新增：

- **admin 本体**：kratos 服务骨架 + `admin_user` 员工账号体系 + **RBAC/ABAC/datascope**。
- **F 远程服务客户端**：admin→mobile 客户端 + 客户端中间件链（`recovery → circuitbreaker → metrics → tracing → metadata → logging → status 转换`）。传输协议**二选一**（别像旧仓库 server=connect / client=grpc 两套并存）。
- **mobile 侧**：暴露**面向 admin 的内部管理 RPC**（管 mobile_user），与给 App 的公开 API 分开，走内部鉴权（mTLS / internal token），不挂在公开面。
- **解锁验证**：多服务链路追踪、熔断、（分布式事务按需）。

**datascope 跨边界（设计点）**：策略在 admin、数据在 mobile —— 两条路：
- **push**：admin 把 scope 下推给 mobile 查询 RPC（如"只返 dept=X 的用户"）—— 不拉冗余，但 mobile 要懂一部分 admin scope 模型。
- **pull**：mobile 全量返、admin 端过滤 —— 简单无耦合，但过度拉数据、分页难。

按 scope 复杂度选（结构化维度 push、轻量 pull），做到那里再定。

---

## 阶段五：生产运维（H）+ 错误守卫（A）

### H. 健康检查 + 安全关机 + pprof

**现状**：connect 有 `/healthz`（仅 liveness 恒 200）；http 无探针；`StopTimeout` 默认 0（无 drain）；无 pprof。k8s Deployment 清单未写。

**① 健康检查 + ② 安全关机**（一起做，共用「准备好了 / 要收摊了」开关）：
- http `/healthz`（liveness 恒 200）+ `/readyz`（`ready` 标志驱动 200/503）；`AfterStart`→MarkReady、`BeforeStop`→MarkNotReady。
- 停机时 **liveness 保持 200**（防 SIGKILL 打断 drain）、**readiness 翻 503**（摘流量）。
- `kratos.StopTimeout(drain 窗口)` 默认 15s；扩 `conf.Server.graceful_stop_timeout`；`newApp` 加 `*conf.Server` 参数。
- 时序：SIGTERM → BeforeStop（ready=false）→ cancel → 并发 drain → cleanup。
- k8s 探针指 `:11002`；`terminationGracePeriodSeconds` > `graceful_stop_timeout`。

**③ pprof** —— `/debug/pprof`，配置/env 门控默认关，port-forward 抓取；注意 prod 暴露风险。

> 设计依据（已核对源码）：kratos App 默认 `stopTimeout=0`；`App.Stop()` 先 beforeStop 再 cancel；各 server `Stop()` 并发执行，不能靠注册顺序。

### A. 错误出口守卫（延后，安全兜底）

**设计原则**：透传给客户端的永远是**受控的模糊信息**（reason 级）；精确内容走 cause → 日志排查。校验类字段提示是客户端职责（buf.validate 生成器守第一道）；透到后端的，后端只给模糊信息，细节进日志。

> 现状 validator `WithCause(verr)` 已是正确实现，**无需改动**。错误机制已落地：proto 枚举 + 工厂；中间件 `init()` 映射；三协议错误内容实测一致。

**唯一要堵的泄漏点** —— kratos `FromError` 兜底：业务随手返回的裸 error（非受控 `*Error`）会被 `New(UnknownCode, UnknownReason, err.Error())`，把原始 `err.Error()`（如 `pq: duplicate key...`）直接透传给客户端。

落两件事（安全兜底，非紧急）：
- **构造纪律**：`Message` 默认空（模糊）；非空 = 故意直出。防开发随手把内部细节塞进 `Message`。手段：lint / 显式 `ErrorXxxDirect` 工厂。
- **出口兜底**：错误透传前（中间件层），非受控 `*Error` → 转通用安全文案，原文进 cause 日志。
- （低优先）三协议 ReplyHeader 错误路径位置对齐（connect unary 在 `*connect.Error.Meta()`）。

---

## mobile 业务 backlog：auth / 会话子系统（后续，独立设计）

mobile 目前只有 mobile_user CRUD（无登录）。后续做真实 auth，**核心是双 token（OAuth2 风格）**：

- **双 token**：access（短 TTL **JWT**，无状态本地验签，不可吊销但短 TTL 封顶）+ refresh（长效，**服务端 DB 记录**、可吊销）。Google / Auth0 / Okta / Cognito 范式，也是"管理端查看/管理签发凭证"的基础。
- **密码 login** → 签 access + refresh。
- **refresh 端点**：换新 access，**轮换**（签新 refresh、吊销旧）+ **复用检测**（被吊销的旧 refresh 再被用 → 视为被盗，吊销整条链）。
- **auth 中间件**：验 access（JWT 本地验签）放行受保护 RPC。
- **logout**：吊销 refresh（DB）+ access 进黑名单（redis，TTL = access 剩余有效期）。
- **refresh-token 实体**（mobile ent + Atlas，**存哈希不存原文**）：id / user_id / token_hash / device·client / ip / issued·last_used·expires / revoked / replaced_by。
- **多端策略**（可配）：多客户端并发 vs 单点登录（踢旧端）。
- **第三方登录**（OAuth：Google/Apple/微信）—— 可插拔 auth method，签同样的 access+refresh（加法）。
- **admin 管理凭证**：列表/吊销单个/吊销全部 —— admin 经 RPC 调 mobile 操作 refresh-token 表。

**关键设计点（做时定）**：access 签名密钥（HMAC vs RSA/Ed25519，多服务场景非对称更合适）；refresh 轮换+复用检测（核心安全特性）；refresh 存哈希；多端策略实现。

**cache 在 auth 的角色**：login 限频（RateLimiter）、access 黑名单（KV/Set）、refresh 热查缓存（KV）。注：双 token 下 Session 接口不被用。

**分步**：① login 核心 → ② refresh-token 管理 RPC → ③ 第三方登录 → ④ admin 管理凭证 UI/RPC。

---

## 实现纪律

- 共享能力进 `shared-go/`，服务特有进 `internal/`；不引入跨服务依赖。
- conf 改动走 proto 源 → Nx `proto:conf` 重生成。
- 每部分做完：`go build ./...` 过 + 相关 Nx target 跑过 + 按关注点单独提交。
- mobile 实体带上 admin 意识；管理 RPC 契约挂 proto 但 admin 本体后建。
