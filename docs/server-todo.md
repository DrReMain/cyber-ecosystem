# 后端平台后续工作（TODO）

排期 backlog。新仓库从零重构，平台能力参考旧仓库 `/Users/shijiao/Desktop/lab/cyber-ecosystem` 对照。本文记录后续路线 —— 含 **mobile 业务丰富**、**可观测收尾**、**第二个服务 edge_admin** 的搭建（解锁多服务能力）。

**已接入**：db（ent + postgres + Atlas 版本化迁移）、三 transport server（http :11002 / grpc :12002 / connect :13002）、中间件链（recovery→tracing→metrics→logging→ratelimit→metadata→validator）、**可观测性全链路**（trace/metric/log OTLP→SigNoz；db/cache/storage/mq 四后端的 span+metric 全接入并经 SigNoz 控制面验证；`shared-go/kratos/observability` 提供仪表化 helper、后端包不耦合；多 sink 日志含 file 轮转）、**cache**（`shared-go/cache`，redis，10 接口）、**storage**（`shared-go/storage`，S3/MinIO + Presign/Stat/Copy）、**mq**（`shared-go/mq`，NATS JetStream + PostgreSQL 双后端）。三能力均接入 edge_mobile Platform 并经对抗式真实后端验证（无 P0/P1，证据见 `docs/integration-test-report.md`）。

---

## 架构与边界（两服务）

> 平台演进为两个服务：**edge_mobile（边缘/消费面）+ edge_admin（管理面）**。**主权数据**：各管自己的 DB，无跨库直读。

- **edge_mobile**：终端用户系统。owns `mobile_user`（用户账号）+ 消费侧业务。**必须能独立运行**（每次请求鉴权，不依赖 admin）—— 它是稳定内核。
- **edge_admin**：内部管理系统。owns `admin_user`（员工账号）+ RBAC/ABAC/datascope。通过 **RPC 管理 mobile 的 `mobile_user`**（不直接读 mobile 的库）。`mobile_user` ≠ `admin_user`（两类人、两个库，零重叠）。
- **依赖方向**：**admin → mobile（单向，不环）**。被调方（mobile 的管理 API）必须先存在，所以顺序天然 mobile 先行。

**mobile 设计时带上 admin 意识（便宜两层）**：
1. **数据模型**：设计 mobile 实体（mobile_user 等）时，带上 admin 管理明确需要的字段（status/软删、scope 维度、审计字段）—— 加字段现在便宜，迁移贵。datascope 模型未定就先留 hook。
2. **管理 RPC 契约**：mobile 成型时把 admin 要调的内部管理 RPC 挂 proto —— 只实现现在用得到的，不投机建整套 admin API。

**分布式相关设计原则**：
- **分布式事务 —— 设计上回避**：admin 调 mobile 改用户 = mobile 库内原子写 + admin 另记审计（最终一致）。真要跨库原子（少）才上 saga/outbox；平台准备该能力，但别滥用（分布式 tx 是最贵的东西）。
- **服务发现 —— k3s DNS 即发现**：`<svc>.<ns>.svc.cluster.local`，不自建 discovery/etcd（除非脱离 k8s 部署）。
- **一致性姿态**：跨服务审计/日志走最终一致。

---

## 路线图（按阶段）

| 阶段 | 事项 | 状态 |
|---|---|---|
| ~~B/C/D~~ | 日志 / 可观测底座 / cache | ✅ 已完成 |
| ~~一~~ | cache 服务栈集成验证 | ✅ 已完成（一次性 campaign 探针，见报告）|
| ~~二~~ | storage 平台能力（S3/MinIO + Presign/Stat/Copy）| ✅ 已完成 |
| | ~~mq（G，可选）~~ | ✅ 已完成（已建 NATS+PG 双后端，非"待业务"）|
| **三** | [可观测收尾 + 异常上报](#阶段三可观测收尾--异常上报) | **进行中（主体完成）**：trace/metric/log + 四后端 span ✅（SigNoz 验证）；残余：慢查询 log + 后端异常上报（→SigNoz）|
| **四** | [edge_admin 搭建 + 多服务能力（F）](#阶段四edge_admin--多服务能力f) | 骨架成型后 |
| **五** | [生产运维（H）+ 错误守卫（A）](#阶段五生产运维h--错误守卫a) | 主线 / 延后 |
| | [mobile auth / 会话子系统（业务，后续独立设计）](#mobile-业务-backlogauth--会话子系统后续独立设计) | 后续 |

> 推荐顺序：**三残余（慢查询 log + 后端异常上报）→ 四 → 五**。三的主体（trace/metric/log 全链路 + 四后端 span）已完成并经 SigNoz 验证。**mobile auth/会话**是独立业务特性，按业务节奏排，不绑平台阶段顺序。
>
> 注：cache/storage/mq 的"做什么"细节已随实现落地而归档（接口/实现/错误映射/部署均已完成），不再在本文展开。

---

## mobile 业务 backlog：auth / 会话子系统（后续，独立设计）

mobile 目前只有 mobile_user CRUD（无登录）。后续做真实 auth，**核心是双 token（OAuth2 风格）**。能力清单（做时分步，独立 brainstorm/spec）：

- **双 token**：access（短 TTL **JWT**，无状态本地验签，不可吊销但短 TTL 封顶）+ refresh（长效，**服务端 DB 记录**、可吊销）。这是 Google / Auth0 / Okta / Cognito 的范式，也是"管理端能查看/管理签发凭证"的基础。
- **密码 login** → 签 access + refresh。
- **refresh 端点**：凭 refresh 换新 access，**轮换**（签新 refresh、吊销旧）+ **复用检测**（被吊销的旧 refresh 再被用 → 视为被盗，吊销整条链；Auth0/Okta 做法）。
- **auth 中间件**：验 access（JWT 本地验签）放行受保护 RPC。
- **logout / 吊销**：吊销 refresh（DB）+ access 进黑名单（redis，TTL = access 剩余有效期）。
- **refresh-token 实体**（mobile ent + Atlas，**存哈希不存原文**，跟 password_hash 同理）：id / user_id / token_hash / device·client / ip / issued·last_used·expires / revoked / replaced_by（轮换链）。
- **多端登录策略**（可配）：同一账号**允许多客户端并发**（每端独立 refresh）vs **单点登录**（新登录踢掉旧端）—— 策略可选。
- **第三方登录**（OAuth：Google / Apple / 微信）—— 可插拔 auth method，签同样的 access+refresh（加法，不重构）。
- **admin 管理凭证**：列表 / 吊销单个 / 吊销全部（"在所有设备登出"）—— admin 经 RPC 调 mobile 操作 refresh-token 表（主权数据，admin→mobile）。

**关键设计点（做时定）**：
- **access 签名密钥**：HMAC（对称，简单）vs RSA/Ed25519（非对称，其他服务只验不签，多服务场景更合适）。
- **refresh 轮换 + 复用检测**：双 token 的核心安全特性，建议核心就含。
- **refresh 存哈希**：DB 泄露 ≠ token 可用。
- **多端策略实现**：per-user session 集合 / token family / 单点登录踢人机制。

**cache 在 auth 里的角色**：login 暴力限频（RateLimiter）、access 黑名单（KV/Set，logout）、refresh 热查缓存（KV）—— cache 在 auth 里被真实用上，key/TTL/失效 pattern 在此暴露。注：refresh 是 DB 实体，access 是 JWT 本地验，**双 token 下 Session 接口不被用**。

**排序（分步）**：① login 核心（access+refresh+中间件+吊销）→ ② refresh-token 管理 RPC（列表/吊销，给 admin + 自查"我在哪登录"）→ ③ 第三方登录 → ④ admin 管理凭证 UI/RPC。

---

## 阶段三：可观测收尾 + 异常上报

**已完成（✅ SigNoz 验证）**：trace/metric/log 全链路 + db/cache/storage/mq 四后端的 span+metric 全接入（`shared-go/kratos/observability/instrument.go` 仪表化 helper + platform 接线；后端包不耦合 observability 库）。三信号 OTLP→SigNoz，控制面 `/services` 出现 `edge_mobile`（spanmetric 派生 P99/错误率）+ ingestion active。

**异常上报架构（后端 vs 客户端分流）**：
- **后端异常 → SigNoz**（经 OTel log：panic/未受控错误 → 结构化 error log，带 trace_id 关联；recovery 中间件落地）。后端异常天然跟 trace 强相关，留在 SigNoz 保持后端可观测统一（trace/metric/log/error 一处、全关联）。权衡：错误 dedup/grouping/版本回归不如 Sentry 级，早期够用，真要重 error workflow 再补。
- **客户端异常 → GlitchTip**（自建 Sentry 协议，sentry-go SDK；breadcrumbs / 崩溃 grouping / release）—— 移动/Web crash 报告是它的主场，SigNoz 不做。
- **客户端全链路 OTel**：各客户端（iOS/Android/Web）引入 OTel SDK，链路从客户端起。**后端已就绪**（kratos tracing 中间件对进来的请求做 W3C tracecontext `Extract`，客户端带 `traceparent` 即续链，后端零改动）。属各**客户端项目**的投入（OTel SDK + on-device 采样 + 客户端可达的 OTLP 接收网关），非本后端仓库事。

**待办（后端残余，小）**：
- **慢查询 log**：db/cache/storage/mq 操作超阈值 → slog warn → OTLP log → SigNoz。阈值进 conf。
- **后端异常上报 → SigNoz**：recovery 中间件把 panic/未受控错误发结构化 error log（OTel log sink 已在）。

client tracing 随阶段四 F 客户端中间件；客户端全链路 OTel 是各客户端项目事项。

---

## 阶段四：edge_admin + 多服务能力（F）

后端骨架成型后（mobile 业务+能力齐、可观测收尾完），起 edge_admin。**复用 shared-go**（cache/observability/transport/错误模型/ent 工具），净新增：

- **admin 本体**：kratos 服务骨架 + `admin_user` 员工账号体系 + **RBAC/ABAC/datascope**。
- **F 远程服务客户端**：admin→mobile 客户端 + 客户端中间件链（`recovery → circuitbreaker → metrics → tracing → metadata → logging → status 转换`）。传输协议**二选一**（别像旧仓库 server=connect / client=grpc 两套并存）。
- **mobile 侧**：暴露**面向 admin 的内部管理 RPC**（管 mobile_user），与给 App 的公开 API 分开，走内部鉴权（mTLS / internal token），**不挂在公开面**。
- **解锁验证**：多服务链路追踪、熔断、（分布式事务按需）。

**datascope 跨边界（设计点）**：策略在 admin（哪个员工能看哪些用户）、数据在 mobile —— 两条路：
- **push**：admin 把 scope 下推给 mobile 查询 RPC（如"只返 dept=X 的用户"）—— 不拉冗余，但 mobile 要懂一部分 admin scope 模型。
- **pull**：mobile 全量返、admin 端过滤 —— 简单无耦合，但过度拉数据、分页难。

按 scope 复杂度选（结构化维度 push、轻量 pull）。无银弹，做到那里再定。

> 服务发现：k3s Service DNS 直连，不自建 discovery。扩 `conf.Data.BaseService`（addr/timeout）。
> 借鉴旧仓库：客户端中间件链顺序。舍弃：自研 discovery（k3s 直连）、grpc/connect 并存的半迁移态。

---

## 阶段五：生产运维（H）+ 错误守卫（A）

### H. 健康检查 + 安全关机 + pprof

**现状**：connect 有 `/healthz`（仅 liveness 恒 200）；http 无探针；`StopTimeout` 默认 0（无 drain）；无 pprof。k8s Deployment 清单未写。

**① 健康检查 + ② 安全关机**（一起做，共用「准备好了 / 要收摊了」开关）：

> 比喻：值班护士定时问「病人清醒吗」（健康检查）；打烊时先挂「暂停接客」、让店里客人吃完再锁门，而不是直接赶人（安全关机）。

- http `/healthz`（liveness 恒 200）+ `/readyz`（`ready` 标志驱动 200/503）；`AfterStart`→MarkReady、`BeforeStop`→MarkNotReady。
- 停机时 **liveness 保持 200**（防 SIGKILL 打断 drain）、**readiness 翻 503**（摘流量）。
- `kratos.StopTimeout(drain 窗口)` 默认 15s；扩 `conf.Server.graceful_stop_timeout`；`newApp` 加 `*conf.Server` 参数。
- 时序：SIGTERM → BeforeStop（ready=false）→ cancel → 并发 drain → cleanup。
- k8s 探针指 `:11002`；`terminationGracePeriodSeconds` > `graceful_stop_timeout`。

**③ pprof** —— `/debug/pprof`，配置 / env 门控默认关，port-forward 抓取；注意 prod 暴露风险（profile 开销 + 状态泄漏）。

> **设计依据（已核对源码）**：kratos App 默认 `stopTimeout=0`（`app.go:106`）；`App.Stop()` 先 beforeStop 再 cancel；各 server `Stop()` 并发执行，不能靠注册顺序。

### A. 错误出口守卫（延后，安全兜底）

**设计原则**：透传给客户端的永远是**受控的模糊信息**（reason 级）；精确内容走 cause → 日志排查。校验类字段提示是客户端职责——客户端用 buf.validate 生成器守第一道，自带友好精准提示；透到后端才报错的，后端只给模糊信息，细节进日志。

> 现状 validator `WithCause(verr)` 已是这个设计的正确实现（verr 进 cause / 日志，客户端只见模糊的 `VALIDATION_FAILED`），**无需改动**。错误机制已落地：proto 枚举 + 工厂；中间件错误 `init()` 映射枚举；三协议错误内容实测一致。

**唯一要堵的泄漏点** —— kratos `FromError` 兜底（`errors.go:137`）：业务随手返回的裸 error（非受控 `*Error`）会被 `New(UnknownCode, UnknownReason, err.Error())`，把原始 `err.Error()`（如 `pq: duplicate key...`）直接透传给客户端。

落两件事（安全兜底，非紧急）：

- **构造纪律**：`Message` 默认空（模糊）；非空 = 故意直出。防开发随手把内部细节塞进 `Message`（会被客户端原样展示给用户）。手段：lint / 显式 `ErrorXxxDirect` 工厂。
- **出口兜底**：错误透传前（中间件层），非受控 `*Error`（裸 error / Unknown reason）→ 转通用安全文案，原文进 cause 日志。不管业务返回啥，出口都干净。
- （低优先）三协议 ReplyHeader 错误路径位置对齐（connect unary 在 `*connect.Error.Meta()`）。

---

## 总览：旧仓库对照

| 能力 | 旧仓库技术栈 | 借鉴 | 舍弃 / 改进 |
|---|---|---|---|
| cache | go-redis/v9 + memory，5 接口 + otel hook | 接口拆分、otel hook、慢查询日志 | 砍 memory 后端；补分布式锁（bsm/redislock）；限流 redis_rate |
| mq | 无 | cache 的抽象模式 | 实落 NATS JetStream + PG 双后端（旧仓库无）|
| storage | aws-sdk-go-v2/s3（MinIO） | 接口骨架、smithy 错误映射、otel span | 补 Presign/Stat/Copy；文件表归 admin |
| 远程调用 | 自研 connect transport + kgrpc client | 客户端中间件链 | discovery（k3s 直连）；grpc/connect 二选一 |
| logging | zap + lumberjack + otel log | 多输出、慢查询、字段约定 | 切 slog；v3 `log` facade + `contrib/otel/log`（内置 trace 关联） |
| 可观测 | otel trace/metric/log + otelsql + sentry | 三 provider、newResource、W3C propagator | 官方 `contrib/otel/{tracing,metrics,log}` middleware（三协议统一）；metric 统一 OTLP；**后端 trace/metric/log + 异常 → SigNoz；客户端 crash + 全链路 OTel → GlitchTip（自建 Sentry 协议）；不引商业 Sentry SaaS** |
| 架构 | 单服务 genesis | — | 两服务（mobile/admin）主权数据；admin→mobile 单向依赖 |

---

## 实现纪律

- 共享能力进 `shared-go/`，服务特有进 `internal/`；不引入跨服务依赖。
- conf 改动走 proto 源 → Nx `proto:conf` 重生成。
- 每个部分做完：`go build ./...` 过 + 相关 Nx target 跑过 + 按关注点单独提交。
- mobile 实体设计带上 admin 意识（status/软删、scope 维度、审计字段）；管理 RPC 契约挂 proto 但 admin 本体后建。
