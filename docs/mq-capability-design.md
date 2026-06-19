# MQ 平台能力设计（`shared-go/mq`）

> 状态：设计稿（待审）。位置遵循本仓 spec 惯例（与 `storage-capability-design.md` 同目录）。

> **实现状态（生产可用修订，2026-06-19）**：经 3 专家 × 2 轮审查（见
> `mq-review-findings.md`）后的修订落地。本文以下章节为初稿意图，实现已对齐/收窄：
> (1) `Nak` 用 `NakWithDelay` 线性退避（`NakBackoffStep`）；(2) durable 名 = `<group>-<topic>`；
> (3) 毒消息到 retry 上限后先写 DLQ，**成功才 `Term`、失败则 `Nak`（绝不静默丢失）**，
> `meta==nil` 时直接 `Term`（防死循环）；(4) 消费者显式 `MaxAckPending`（默认 256，限内存+重投）；
> (5) 共享 `mq-dlq` 流用**独立的更长保留**（`DLQMaxAge`/`DLQMaxBytes`，默认 30d），防毒消息被业务流挤掉；
> (6) `Creds` = NATS 凭据文件路径（`UserCredentials`），非 bearer token；(7) `Message.ID` 为后端本地不透明值，
> 幂等去重走 `Headers`；(8) `ValidateTopic/Group` 拒绝 `.`/`>`/`*`/空白/控制字符（stream 名安全）。
> 错误映射补 `nats.ErrTimeout` + 504 `APIError` → `ErrTimeout`。实现细节以
> `shared-go/mq/` 代码为准。

## 1. 目标与范围

构建 `shared-go/mq` —— 后端无关的消息队列**能力层** + NATS JetStream 实现，克隆 `shared-go/cache` / `shared-go/storage` 的模式（机制/实例分离、Platform 同构接入）。它是平台骨架的异步事件层：通知、支付事件、领域事件广播、消费者侧可靠处理。

**核心语义**：可靠工作队列（at-least-once 投递 + 消费者 ack + 失败重试 + DLQ）+ 组语义（同组竞争 = 工作队列；不同组各收全量 = 广播）。

**覆盖**：publish / subscribe / ack / 重试 N 次 + 死信 / durable 消费者崩溃恢复。

**不在本次**：scheduler（定时/到期提醒，独立能力，后补）、outbox（producer 可靠性模式，支付真做时）、dtm（分布式事务协调器，独立平台件）、**PG 后端的实现**（设计预留、本次只建抽象 + NATS 后端）。延迟投递归 scheduler，**不进 MQ 抽象**（MQ 只管"可靠送达"，"何时送"是调度器的事）。

**基建**：NATS JetStream 已部署于 k3s（namespace `mq`，host:4222 永久可达，monitoring 8222），见 `deploy/k8s/mq/`。

---

## 2. 架构与包布局（克隆 cache / storage）

```
shared-go/mq/                  # 后端无关契约层
  mq.go          # MQ struct { Publisher; Consumer }
  message.go     # Message
  publisher.go   # Publisher 接口
  consumer.go    # Consumer 接口 + Subscription
  errors.go      # sentinels（后端无关）
  error.go       # MQDefaultError + ValidateMQDefaultError + HandleMQError（机制）
shared-go/mq/nats/             # NATS JetStream 实现（本次）
  client.go      # NewClient(*Config) (*nats.Conn + JetStream, func(), error)
  config.go      # Config + 默认值/校验
  mq.go          # New(...) *mq.MQ（装配 Publisher + Consumer）
  publisher.go   # js.Publish
  consumer.go    # durable consumer + fetch 循环 + ack/nak + 重试 + DLQ
  error.go       # mapError: nats/jetstream err → sentinel
shared-go/mq/pg/               # PG(Skip Locked)实现 —— 本次【不实现】，设计预留
  （独立 mq 库 + 自管表 messages/dlq/per-group ack，boot auto-DDL，不走 atlas）
```

**机制/实例分离**（同 cache/storage）：`shared-go/mq` 持 `switch`+`WithCause` 映射机制；具体 proto `*errors.Error` 实例由服务 platform 层注入。mq 包不依赖任何服务错误 proto。

**一句柄一后端**（镜像旧仓库 cache 的 memory+redis 模式）：一个项目通常实例化一个 `*mq.MQ` 用一个后端（NATS 或 PG）；特殊情况下开发者用两个句柄接两个后端，但**骨架不写这种多后端示例**。

---

## 3. 接口

```go
// message.go
type Message struct {
	Topic     string
	Payload   []byte
	Headers   map[string]string
	ID        string    // 后端填充（JetStream sequence/ack token）
	Timestamp time.Time // 后端填充
}

// publisher.go
type Publisher interface {
	Publish(ctx context.Context, topic string, msg *Message) (id string, err error)
}

// consumer.go
type Subscription interface {
	Close() error
}
type Consumer interface {
	// Subscribe 启动一个后台 goroutine 持续消费（push 模型：库跑 fetch 循环、调 handler）。
	// group = 消费者身份：同 group 竞争消费（工作队列，每条只被组内一个消费者处理）；
	// 不同 group 各自收全量（广播）。
	// handler 返 nil → ack；返 error → 重试（达 max_retries 后进 DLQ）。
	Subscribe(ctx context.Context, topic, group string, handler func(ctx context.Context, msg Message) error) (Subscription, error)
}

// mq.go
type MQ struct {
	Publisher Publisher
	Consumer  Consumer
}
```

**push 模型说明**：底层（NATS / PG）网络层永远是 pull；"push" 是 API 包法 —— 库在后台 goroutine 跑 fetch 循环，每拉到一条调 handler（类比 `http.HandleFunc`）。订阅者只写 handler，不管循环。

**与 4 类非 unary 传输原语无关**：MQ 是平台件（Platform 持有），不直接挂在 transport 上；业务通过 Platform 用 MQ。MQ 自身的消息消费是后台 goroutine（随 App 生命周期启停），不是 RPC。

---

## 4. 错误模型（机制/实例分离，克隆 cache/storage）

### 4.1 sentinels（`errors.go`）

```go
var (
	ErrInvalidArgument = errors.New("mq: invalid argument") // 非法 topic/payload/group
	ErrUnavailable     = errors.New("mq: unavailable")      // NATS 连接/发布失败
	ErrTimeout         = errors.New("mq: timeout")          // 发布/操作超时
)
```

> 注意：**消费者 handler 返回的业务错误不属于 MQ 错误**——那是业务错误，MQ 对其重试/DLQ，不经 `HandleMQError`。`HandleMQError` 只映射 MQ **基础设施**错误（publish/connect）。

### 4.2 proto 错误码（`infra.proto`，新增 34xx MQ 段）

```proto
// 34xx: MQ
INFRA_ERROR_MQ_INVALID_ARGUMENT = 3400 [(.errors.code) = 400];
INFRA_ERROR_MQ_UNAVAILABLE      = 3401 [(.errors.code) = 503];
INFRA_ERROR_MQ_TIMEOUT          = 3402 [(.errors.code) = 504];
```

走错误模型的 Nx 生成 target 重生成 → `errorspb.ErrorInfraErrorMqXxx` 可用。

### 4.3 机制（`error.go`，克隆 cache/storage）

```go
type MQDefaultError struct {
	InvalidArgument *kratoserrors.Error
	Unavailable     *kratoserrors.Error
	Timeout         *kratoserrors.Error // optional：nil → 未知错误原样透传
}
func ValidateMQDefaultError(errs *MQDefaultError) error
func HandleMQError(err error, errs *MQDefaultError) error // switch + WithCause
```

---

## 5. NATS 后端实现（`shared-go/mq/nats`）

**Config**：`Endpoint`(nats://host:4222)、`Creds`(可选鉴权，空=无鉴权 dev)、`MaxAge`(stream 消息 TTL，默认 7d)、`MaxBytes`(stream 容量上限，默认如 1GB)、`MaxRetries`(消费者重试上限，默认 5)。

**Publisher**：
- `js.Publish(topic, payload)`，主题映射：**一 topic 一 stream**（stream 名 `mq-<topic>`，subject=`<topic>`）。
- stream 首次发布时**惰性创建**（`StreamCreate` 或 `StreamUpdate` 幂等）：retention=`Limits`、`MaxAge`=config、`MaxBytes`=config，防 stream 无限增长占盘。
- 返回 JetStream sequence 作为 Message.ID。

**Consumer（Subscribe）**：
- 建 **durable consumer**（崩溃恢复位置不丢）：filter-subject=`<topic>`；durable 名 = `<group>-<topic>`；`DeliverGroup`=`<group>`（同 group 竞争 = 工作队列；不同 durable 名 = 广播）。
- 后台 goroutine 循环：取消息 → 调 handler：
  - handler 返 `nil` → `msg.Ack()`。
  - handler 返 error → 看 `msg.Metadata.NumDelivered`（JetStream 内置投递计数）：
    - `< MaxRetries` → `msg.Nak(delay)`（带退避延迟重投）。
    - `>= MaxRetries` → **发 DLQ**（`js.Publish("<topic>-dlq", payload, headers=retry计数/original-meta)`）+ `msg.Ack()` 原（隔离毒消息，不阻塞）。
- `Subscription.Close()` → drain + 取消订阅（消费者 durable 状态保留，重启可续）。

**mapError**：nats/jetstream 错误 → sentinel（`nats.ErrConnectionClosed`/连接错误 → ErrUnavailable；`context.DeadlineExceeded` → ErrTimeout；非法参数 → ErrInvalidArgument）。

---

## 6. PG 后端（设计预留，本次不实现）

`shared-go/mq/pg`，将来补：
- **独立 `mq` 数据库**（同一 PG 实例，与 `edge_mobile`/glitchtip 库隔离）+ 自己的连接池。
- **自管表**（boot 时幂等 `CREATE TABLE IF NOT EXISTS`）：`messages(topic, payload, headers, due_at, created_at)`、`dlq(...)`、广播用的 per-group ack 表。**不走 atlas**（atlas/ent 只管业务库 edge_mobile；这是 GlitchTip 式的"用 PG 但自管表"模式）。
- 工作队列天然贴合抽象（`FOR UPDATE SKIP LOCKED WHERE topic=?`）；广播需 per-group ack 跟踪表（impl 细节）。
- 接口层与 NATS 完全等价（抽象无感）。

---

## 7. 配置（`conf.Data.MQ`）

`app/services/edge_mobile/internal/conf/conf.proto`：
```proto
message MQ {
  message NATS {
    string endpoint = 1;                          // nats://localhost:4222
    string creds = 2;                             // 可选鉴权（creds/token）；空=dev 无鉴权
    google.protobuf.Duration max_age = 3;         // stream 消息 TTL；0 → 默认 7d
    int64 max_bytes = 4;                          // stream 容量上限；0 → 默认
    int32 max_retries = 5;                        // 消费者重试上限；0 → 默认 5
  }
  NATS nats = 1;
}
// message Data 内：... MQ mq = 4;
```
`configs/config.yaml` 加 `data.mq.nats:` 块（endpoint `nats://localhost:4222`，dev 无鉴权）。Kratos 结构化映射，无需 main.go 手工 mapping。

---

## 8. Platform 接入（克隆 cache/storage）

- `platform_mq.go`：`NewMQ(c *conf.Data) (*mq.MQ, func(), error)` + `toMQConfig`；连 NATS（`nats.NewClient`）+ cleanup（`nc.Drain()`/`Close()`）。返回 `(T, func(), error)` 让 wire 链 cleanup。
- `platform_mq_handler.go`：`defaultMQError`（填 34xx）+ `NewMQErrorHandler()`（启动期校验 sentinel slot）。
- `platform.go`：`MQ mq.MQ` + `handleMQError MQErrorHandler` 字段、ctor 入参、`GetMQ()` + `HandleMQError()` 访问器；`ProviderSet` 加 `NewMQ`/`NewMQErrorHandler`。
- `NewPlatform` 不持资源（cleanup 靠 provider 的 `func()` 经 wire 链）。

---

## 9. 验证策略（克隆 cache/storage 那套）

MQ 无现成消费者（业务尚未用），穿栈验证用**临时 RPC**（probe）：
1. 加临时 RPC（如 `Debug.MQProbe`），proto + `nx run proto:generate:go` 重生成。
2. 栈内跑：publish → subscribe round-trip（收发一致）、handler 失败 → 重试 → 达上限进 DLQ、durable 消费者重启续投。
3. 真实 grpcurl/curl 打**三协议**，验 `HandleMQError` → 三协议渲染一致（INVALID_ARGUMENT/UNAVAILABLE）。
4. 详细测试报告（结果 + 过程）。
5. **永久硬化测试**（`shared-go/mq/nats`，对真实 NATS）：publish/consume/ack、handler 失败重试到 DLQ、同组竞争（多 subscriber 一条只被一个处理）、不同组广播（各组收全量）、durable 崩溃恢复、NATS 断连 fault、特殊 payload（二进制/大消息/特殊字符 topic）。
6. **撤** 临时 RPC（revert proto + regen + 临时代码），保留硬化测试。

---

## 10. 可观测性

**留到阶段三**（与 cache/storage 一致）：MQ 现在只做逻辑 + 错误映射，**不接 otel span、不接慢操作日志**。阶段三 db/cache/storage/mq 四后端 span + 慢查询统一接。MQ span 机制（每 publish/consume 一个 span，属性 topic/group）阶段三补。

---

## 11. 非目标

- scheduler（cron/到期提醒，独立能力，后补）。
- outbox（producer 可靠性，支付真做时）。
- dtm（分布式事务协调器，独立平台件）。
- PG 后端实现（设计预留，本次只 NATS）。
- 延迟投递（归 scheduler）。
- 多后端同实例（一句柄一后端；多后端是调用方职责，骨架不示例）。
- exactly-once（at-least-once + 幂等消费者是行业标准；broker 级 exactly-once 不做）。

## 12. 待定 / 延后

- stream retention 细节（max_age/max_bytes 实测调优，默认先 7d/1GB）。
- 重试退避策略（nak delay 线性 vs 指数，默认先线性小步）。
- PG 后端（按需实现：独立库 + 自管表 + Skip Locked）。
- MQ 错误码扩展（若实际多出错误类型再加 34xx）。
