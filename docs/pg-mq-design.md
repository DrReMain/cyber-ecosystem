# PG MQ 后端设计（`shared-go/mq/pg`）

> 状态：**已实现**（2026-06-19），**生产硬化**（2026-06-20）。`shared-go/mq/pg`（config/error/schema/client/publisher/consumer/mq + 测试，全绿含 `-count=5 -race`）；平台接入 `NewMQ` 按 nats/pg 二选一；PG 后端服务可启动（独立 `mq` 库，4 张表启动幂等创建）。实现计划见 `docs/pg-mq-plan.md`，硬化审查见 `docs/pg-mq-review-findings.md`。
>
> **2026-06-20 生产硬化**（4 专家交叉审查 + 对抗式验证）：fetch 时自增 `deliveries`（可见性超时重投计入，对齐 NATS `NumDelivered`）+ stall 闸门 + reaper 兜底（卡死/永不 ack 的消息也能 DLQ）；补 `rows.Err()`；settle/轮询错误不再静默；轮询失败退避 + down→up 日志；backfill 分批；Subscribe ctx 改派生自 handle 生命源（PG+NATS）；关停闸门；轮询索引覆盖排序；nak 退避上限。P1-4（reaper 删在途）经核实改为「P1-1 已缓解 + NATS MaxAge parity」文档化（不加会泄漏的守卫）；P2-7（无界在途）经核实为误报（串行 pollLoop 下 batch_size 即上限）。

## 1. 背景与目标

`shared-go/mq` 是后端无关的消息队列能力层，已有 NATS JetStream 实现（`shared-go/mq/nats`）。本文设计**第二个后端：PostgreSQL 实现**，落到同一份契约（`Publisher`/`Consumer`/`Subscription`/`Message`）上，让"不想跑 NATS 的部署"也能用 MQ。

定位（与 IoT 路线图决策一致）：PG-MQ 是**次要后端 / 无 NATS 部署的后备**，只实现最小契约；IoT 场景仍走 NATS（leafnode/request-reply 是 NATS 专属，PG 做不了）。两个后端**可观察行为一致**（业务无感知切换）。

契约已具备可移植性：上一轮生产可用修复把 `Message.ID` 文档化为"后端本地不透明值"，retry/DLQ 机制（`decideAck`/`NumDelivered`）完全在 NATS impl 内部——契约只规定"handler error → 重试到 MaxRetries 后 DLQ"。PG 用自己的 `deliveries` 列、自己的 bigserial ID 实现同一行为即可。

## 2. 关键决策

1. **投递模型：纯轮询**。后台 goroutine 每 `poll_interval`（可配，默认 ~500ms）用 `FOR UPDATE SKIP LOCKED` 批量取消息。简单可靠、无连接复杂度；延迟≈轮询间隔。匹配"后备后端 + 通常低频"定位。（不走 `LISTEN/NOTIFY`——独占连接/重连/payload≤8KB 的复杂度对后备后端不值。）
2. **队列模型：per-(group,message) 投递表（方案 A）**，非游标/offset。投递表让**竞争消费干净**（按投递行 SKIP LOCKED，PG 工作队列标准范式），广播/回放自然成立。游标模型在同 group 内竞争上别扭（共享游标，没法按行 SKIP LOCKED）。
3. **广播语义：历史回放（对齐 NATS）**。新 group 订阅时收保留窗口内的所有历史消息；代价是首次 Subscribe 时 backfill 投递行。保证两后端语义一致（可移植）。
4. **独立 `mq` 库 + 独立 pgx 连接池**（不复用业务池），与 `edge_mobile`/glitchtip 库隔离，MQ 负载不影响业务查询。
5. **自管表，不走 atlas**：`NewClient` 启动时幂等 `CREATE TABLE IF NOT EXISTS`（GlitchTip 式）。atlas/ent 只管业务库。

## 3. 包结构（镜像 nats 后端）

```
shared-go/mq/pg/
  config.go     # Config + *OrDefault（dsn/poll_interval/visibility_timeout/max_retries/retention）
  client.go     # NewClient(*Config) (*handle, func(), error)：建池 + 建表/索引 + cleanup
  schema.go     # CREATE TABLE IF NOT EXISTS 语句 + 索引
  publisher.go  # Publish：插 message + 扇出 deliveries
  consumer.go   # Subscribe：注册 subscriber +（新 group）backfill + 轮询循环 + ack/nak/dlq
  error.go      # mapError：pgx 错误 → mq sentinel
  mq.go         # New(*handle) *mq.MQ
```

平台接入：`app/services/edge_mobile/internal/platform/platform_pg_mq.go`（`NewPGMQ` + `toPGConfig`），`NewMQ` 按配置块选 nats 或 pg。

## 4. 表结构（`mq` 库，启动时幂等建表）

```sql
-- 只追加的消息存储
CREATE TABLE IF NOT EXISTS messages (
  id         BIGSERIAL PRIMARY KEY,
  topic      TEXT NOT NULL,
  payload    BYTEA NOT NULL,
  headers    JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_messages_topic_id ON messages (topic, id);

-- 每个 (group, message) 一行：group 要投递这条消息
CREATE TABLE IF NOT EXISTS deliveries (
  id         BIGSERIAL PRIMARY KEY,
  group_name TEXT NOT NULL,
  topic      TEXT NOT NULL,                       -- 反范式，便于按 (group,topic) 轮询
  message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  deliveries INT NOT NULL DEFAULT 0,             -- 投递次数（= NATS NumDelivered）
  due_at     TIMESTAMPTZ NOT NULL DEFAULT now(), -- 可见性闸门：due_at <= now() 才可取
  UNIQUE (group_name, message_id)                -- 去重「扇出 vs 回填」竞态
);
CREATE INDEX IF NOT EXISTS idx_deliveries_poll
  ON deliveries (group_name, topic, due_at);

-- 已订阅的 (group, topic)：发布扇出 + 新 group 判定
CREATE TABLE IF NOT EXISTS subscribers (
  group_name TEXT NOT NULL,
  topic      TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_name, topic)
);

-- 死信
CREATE TABLE IF NOT EXISTS dlq (
  id         BIGSERIAL PRIMARY KEY,
  topic      TEXT NOT NULL,
  group_name TEXT NOT NULL,
  payload    BYTEA NOT NULL,
  headers    JSONB NOT NULL DEFAULT '{}',
  deliveries INT NOT NULL,
  error      TEXT NOT NULL,
  dead_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

清理：按 `created_at` 年龄清理过期的 `messages`（`ON DELETE CASCADE` 自动删其 `deliveries`）；DLQ 自管保留（类比 NATS 的 `mq-dlq` 独立长保留）。

## 5. 流程

**Publish(topic, msg)**（单事务）：
```sql
INSERT INTO messages (topic, payload, headers) VALUES ($1,$2,$3) RETURNING id;
INSERT INTO deliveries (group_name, topic, message_id)
  SELECT group_name, $1, $id FROM subscribers WHERE topic=$1
  ON CONFLICT (group_name, message_id) DO NOTHING;
```
返回 `messages.id`（字符串）作为 `Message.ID`。

**Subscribe(topic, group)**：
```sql
INSERT INTO subscribers (group_name, topic) VALUES (G,T) ON CONFLICT DO NOTHING;
-- 若新建（RowsAffected=1）→ 历史回放 backfill：
INSERT INTO deliveries (group_name, topic, message_id)
  SELECT G, T, id FROM messages WHERE topic=T
  ON CONFLICT (group_name, message_id) DO NOTHING;
```
新 group = 回放保留窗口内历史；已存在 = 续传（其 deliveries 持久在）。然后启动该 (group,topic) 的轮询 goroutine。

**消费循环**（每 `poll_interval`，按 (group,topic)）：
1. 取事务：
   ```sql
   SELECT d.id, d.deliveries, m.id AS msg_id, m.payload, m.headers, m.created_at
     FROM deliveries d JOIN messages m ON d.message_id=m.id
     WHERE d.group_name=$G AND d.topic=$T AND d.due_at <= now()
     ORDER BY d.message_id LIMIT $batch
     FOR UPDATE SKIP LOCKED;
   UPDATE deliveries SET due_at = now()+$visibility_timeout WHERE id IN (…取到的…);
   COMMIT;
   ```
   取到的行在 `visibility_timeout` 内对其他消费者不可见（防同 group 并发重复，慢 handler 风险与 NATS AckWait 一致，文档注明）。
2. 事务外逐条调 handler。
3. Ack → `DELETE FROM deliveries WHERE id=?`。
   Nak → `UPDATE deliveries SET due_at=now()+$backoff, deliveries=deliveries+1 WHERE id=?`；
   若 `deliveries+1 >= MaxRetries` → `INSERT INTO dlq …` + `DELETE FROM deliveries WHERE id=?`。

与 NATS `decideAck` 对应：visibility_timeout=AckWait、ack=删除、nak=延时重投、达上限进 DLQ（DLQ 写失败则保留、成功则删——不静默丢失）。

**投递计数语义（生产硬化，2026-06-20）**：`deliveries` 在 **fetch 时**自增（`UPDATE deliveries SET due_at=…, deliveries=deliveries+1`），因此**可见性超时重投也计入**——与 NATS `NumDelivered` 对齐（此前版本只在 nak 路径自增，慢 handler 的超时重投不计，会偏离 NATS）。相应地：
- `settle` 的 nak 路径不再自增（fetch 已计），DLQ 判定用「本次 attempt = fetch 前的 deliveries + 1 ≥ maxRetries」。
- **stall 闸门**：fetch 时若某投递已 `deliveries ≥ maxRetries` 仍未 ack（handler 卡过可见性 / 永不返回），直接在 fetch 阶段 DLQ，不进 handler——对齐 NATS server-side MaxDeliver。
- **reaper 兜底**：后台周期把 `due_at 已过期 且 deliveries ≥ maxRetries` 的投递原子 DLQ+删除，捕获单消费者 goroutine 卡死、或无活跃订阅者等「poll 循环触达不到」的情况（NATS server-side 的完全对等）。

## 6. 生命周期与配置

- `NewClient(*Config) (*handle, func(), error)`：连独立 pgx 池 → 幂等建表/索引 → 返回；cleanup 关池。`Config{ DSN, PollInterval, VisibilityTimeout, MaxRetries, Retention }` + `*OrDefault`。
- `Subscribe` 启动注册的轮询 goroutine（注册表 + 关停 drain，与 NATS 后端同构）。**轮询 ctx 派生自 handle 生命源 ctx，不是调用方 ctx**（调用方 ctx 只管同步 setup），故常驻订阅不会因发起请求结束而静默死亡；关停开始后 `closed` 标志拒绝新 Subscribe，避免漏 drain。
- `conf.Data.MQ.PG { dsn, poll_interval, visibility_timeout, max_retries, retention }`；`NewMQ` 按配置了哪个块选 nats 或 pg（一个后端生效，符合"一句柄一后端"）。

## 7. 错误模型

复用 `shared-go/mq` sentinel + `HandleMQError`→34xx。`mapError` 映射：`context.DeadlineExceeded`/pgx 超时 → `ErrTimeout`；连接类（`pgx.ErrConnDone`/pool 已关）→ `ErrUnavailable`；其余 → `ErrUnavailable`。handler 业务错误不进 mapError（重试/DLQ，同 NATS）。

## 8. 测试策略

镜像 NATS 后端的测试结构：round-trip、retry→DLQ、competing（同 group）、broadcast（不同 group）、durable 续传（重启进程后续传）、DLQ 头保真、fault（DSN 不可达 → ErrUnavailable）、config 校验。白盒（package pg）复用 `newTestMQ`/`uniqTopic` 等价物（PG 里每个测试用唯一 topic，测试后清理）。需要可达的 PG（k3s 已有 PG；用一个独立 `mq_test` 库或 schema 隔离，跑完清理）。

## 9. 不在本次范围（与既有决策一致）

- LISTEN/NOTIFY 低延迟推送（需低延迟用 NATS）。
- per-message TTL / 优先级 / seek-replay（YAGNI）。
- 延时投递（归 scheduler，不进 MQ）。
- 发布去重（Nats-Msg-Id 等价物）——两后端都暂不做，at-least-once + handler 幂等。
- 跨后端同时启用（一句柄一后端）。

## 10. NATS ↔ PG 语义对照

| 行为 | NATS | PG |
|---|---|---|
| 消息存储 | JetStream stream（mq-<topic>） | `messages` 表 |
| 竞争消费 | 同 durable（group） | 同 group_name，按 deliveries 行 SKIP LOCKED |
| 广播 | 不同 durable | 不同 group_name，各有 deliveries 行 |
| 新 group 历史回放 | durable DeliverAll | subscribers 新建时 backfill deliveries |
| 续传 | durable 持久位置 | deliveries 行持久 |
| 重投等待 | AckWait（精确） | visibility_timeout = AckWait；重投节奏 = vis + poll_interval（略晚于 NATS 的精确 AckWait） |
| 投递计数 | NumDelivered | deliveries（**fetch 时自增**，含可见性超时重投，与 NumDelivered 对齐） |
| 重试退避 | NakWithDelay 线性 | nak 时 due_at=now()+线性 backoff |
| 毒消息 | DLQ（mq-dlq）+ Term | dlq 表 + 删 delivery |
| 卡死/永不 ack 兜底 | server-side MaxDeliver（NumDelivered 达上限即停投） | fetch 阶段 stall 闸门（deliveries≥maxRetries 直接 DLQ）+ reaper 周期兜底（原子 DLQ+删过期超限投递） |
| 消息 ID | 流序号 | bigserial |
| 在途上限 | MaxAckPending（默认 256） | batch_size（默认 16）= 每订阅在途上限（pollLoop 单 goroutine 串行，取一批处理完才取下一批；无独立 MaxAckPending） |

关联：`docs/mq-capability-design.md` §6（PG 后端设计预留）、`docs/capability-audit.md`（能力包一致性）。
