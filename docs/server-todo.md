# edge_mobile 后续工作（TODO）

记录已设计 / 部分实施但尚未完成的工作，作为排期 backlog。新仓库从零重构，平台能力参考旧仓库 `/Users/shijiao/Desktop/lab/cyber-ecosystem` 的封装方式（逐部分对照）。

**已接入**：db（ent + postgres + Atlas 版本迁移）、三 transport server（http :11002 / grpc :12002 / connect :13002）、中间件链（recovery→tracing→metrics→logging→ratelimit→metadata→validator）、**可观测性（trace/metrics/log OTLP→SigNoz，抽到 `shared-go/kratos/observability`，多 sink 日志含 file 轮转）**。

## 全貌

分三层、八个部分。横切关注点（可观测性）已打底完成，接下来平台能力，最后生产运维：

| 部分 | 事项 | 状态 |
|---|---|---|
| **B** | 日志体系（trace_id 关联 + 多 sink fanout logger） | ✅ 完成 |
| **C** | 可观测性（trace/metrics/log OTLP→SigNoz） | ✅ 完成；⏳ db-span 待做 |
| **D** | [cache](#d-cache)（redis 抽象层 + otel hook） | 主线 |
| **E** | [storage](#e-storage)（S3/MinIO + 预签名） | 主线 |
| **F** | [远程服务客户端](#f-远程服务客户端)（connect/grpc client + 客户端中间件链） | 主线 |
| **H** | [生产运维](#h-健康检查--安全关机--pprof)（健康检查 + 安全关机 + pprof） | 主线 |
| **A** | [错误出口守卫](#a-错误出口守卫)（出口 sanitize，防内部信息泄漏） | 延后（安全兜底） |
| **G** | [mq](#g-mq可选--延后)（从零选型） | 延后（待业务） |

> 推荐顺序：~~B → C~~（✅ 已完成）→ **D / E /（F）→ 可观测收尾（db-span + 后端 span + 慢查询）→ H**。A、G 延后。

---

## 第一层：横切关注点（已完成）

**B 日志体系 + C 可观测性（trace/metrics/log）已落地**，抽到 `shared-go/kratos/observability` 包：

- `Init` 统一建 trace/metrics/log provider（otlphttp exporter、共享 resource 带 host+env、可选采样、聚合 shutdown）+ 多 sink fanout logger（console / otlp / file+lumberjack 轮转）。
- `MetricsServer` 中间件（contrib `metrics.Server` + 默认仪表）。
- 三协议中间件链 `tracing→metrics→logging→recovery→ratelimit→metadata→validator`（三个记录型在 recovery 外，panic/429 不丢信息）。
- 生产硬化：resource `host.name`+`deployment.environment`、采样可配（默认不采）、错误记 Error（链顺序）、log 默认 info。
- conf 嵌套 `Observability{Trace{enabled,sampling_ratio}, Metrics{enabled}, Log{level,console,otlp,file{...}}}`。

---

## 第二层：平台能力

> `conf.Data` 现只有 `Database` + `Redis`（Redis 段已定义但无代码）。新增能力都要扩 `conf.Data` 或新增 `conf.Cache` / `conf.Storage` 等。

### D. cache

**现状**：`conf.Data.Redis` 已定义，无任何缓存代码。

**做什么**：`shared-go/cache` 抽象层 + redis 实现。

- **接口拆分**（借鉴旧仓库，按需取用）：`KV`（Get/Set/Del/Exist/MGet/MSet）、`Counter`、`SortedSet`、`RateLimiter`；聚合成一个 `Cache` struct + `io.Closer`。
- **后端**：**只做 redis**（go-redis/v9），**砍 memory 后端**（旧仓库有，生产不用、工作量大；留接口位，单测用 fake）。
- **otel hook**：`redis.AddHook` 挂 tracing（每命令一个 span，带 operation/key_count/duration）。
- **补分布式锁**（旧仓库无）：SETNX 或 redlock 封装。
- 激活已定义的 `conf.Data.Redis`；key 前缀由调用方管（抽象层不绑）。

> **借鉴**：5 接口拆分 + 聚合 struct + otel hook + 慢查询日志。**舍弃**：memory 后端。

### E. storage

**现状**：无。k3s 已部署 MinIO。

**做什么**：`shared-go/storage` 接口 + S3 实现（aws-sdk-go-v2，`UsePathStyle` 指向 MinIO）。

- **接口**（借鉴旧仓库并补全）：`Upload / Download / Delete` + **补 `Presign`（上传/下载预签名，前端直传必需）+ `Stat` + `Copy`**。
- 错误映射：smithy `APIError` → 应用错误。otel span（每操作一个 span，带 bucket/key/size）。
- 扩 `conf.Data.Storage`（endpoint/access_key/secret_key/bucket/region/max_file_size）。

> **借鉴**：接口骨架 + smithy 错误映射 + otel span。**补**：Presign/Stat/Copy。

### F. 远程服务客户端

**现状**：新仓库已有自研 connect **server**（transport/connect + health + reflection）。缺 **client** 端（BFF 调下游服务）。

**做什么**：connect/grpc client + 客户端中间件链。

- **客户端中间件链**（借鉴旧仓库，教科书级）：`recovery → circuitbreaker → metrics → tracing → metadata → logging → status 转换`。
- **传输**：和下游服务协议对齐，**二选一，别像旧仓库 server=connect / client=grpc 两套并存**。
- **服务发现**：**先不做**。k3s 里下游用 Service name 直连（`<svc>.<ns>.svc.cluster.local`）；旧仓库自研 discovery 但 BFF 实际也直连。discovery/etcd 延后。
- 扩 `conf.Data.BaseService`（addr/timeout）。

> **借鉴**：客户端中间件链顺序。**舍弃**：自研 discovery（k3s 直连）、grpc/connect 并存的半迁移态。

### G. mq（可选 / 延后）

**现状**：旧仓库也没有 MQ，当前无业务驱动。

**做什么（待业务出现再选型）**：套 cache 的抽象模式。

- `mq.Publisher` / `mq.Consumer` 接口；后端按场景选 **redis-stream（轻，复用已有 redis）** 或 **kafka（重，segmentio/kafka-go）**。
- 重试 / 死信在抽象层或装饰器；otel hook（每消息一个 span）。

---

## 可观测性收尾：db-span + 后端 span + 慢查询（排在平台能力之后）

trace 目前只有 RPC span，**无后端子 span**（DB/cache/storage 调用都不可见）。等 D/E 落地后统一接，构成完整后端追踪：

- **db-span**：`github.com/XSAM/otelsql` wrap SQL 驱动 → 每个查询一个 span。
- **cache-span**：随 D（redis `AddHook`，已在 D 的设计里）。
- **storage-span**：随 E（每操作一个 span，已在 E 的设计里）。
- **慢查询日志**：补 B 的慢查询缺口（duration + slow 标记 + 阈值，db/cache/storage 各接）。
- **client tracing**：`tracing.Client()` 进 F 的客户端中间件链。

复用 shared-go observability 已设的全局 tracer。**为什么排在 D/E 之后**：避免半截 trace（只显示 DB 不显示 cache）误导；后端 span 是一组内聚工作，一起做更干净。

---

## 第三层：生产运维

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

---

## 延后 / 可选

### A. 错误出口守卫

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
| cache | go-redis/v9 + memory，5 接口 + otel hook | 接口拆分、otel hook、慢查询日志 | 砍 memory 后端；补分布式锁 |
| mq | 无 | cache 的抽象模式 | — |
| storage | aws-sdk-go-v2/s3（MinIO） | 接口骨架、smithy 错误映射、otel span | 补 Presign/Stat/Copy |
| 远程调用 | 自研 connect transport + kgrpc client | 客户端中间件链 | discovery（k3s 直连）；grpc/connect 二选一 |
| logging | zap + lumberjack + otel log | 多输出、慢查询、字段约定 | 切 slog；v3 `log` facade + `contrib/otel/log`（内置 trace 关联） |
| 可观测 | otel trace/metric/log + otelsql + sentry | 三 provider、newResource、W3C propagator | 官方 `contrib/otel/{tracing,metrics,log}` middleware（三协议统一）；metric 统一 OTLP；弃 sentry |

---

## 实现纪律

- 共享能力进 `shared-go/`，服务特有进 `internal/`；不引入跨服务依赖。
- conf 改动走 proto 源 → Nx `proto:conf` 重生成。
- 每个部分做完：`go build ./...` 过 + 相关 Nx target 跑过 + 按关注点单独提交。
