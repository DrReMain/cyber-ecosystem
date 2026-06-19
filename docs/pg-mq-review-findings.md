# PG MQ 后端审查报告（`shared-go/mq/pg`）

> 阵容：PG 并发/事务隔离专家 + MQ 语义对等专家 + 分布式架构师（生命周期）+ Go/pgx 专家，4 路并行独立审查 → 交叉对抗式验证。
> **事实来源 = 代码**。`docs/pg-mq-design.md` 的 §5/§10 对等表此前已漂移出错，本审查以代码为准，凡文档与代码冲突处单独列为发现。
> 立场基线（用户既定）：可观测性分阶段统一接入（非缺陷）；跨能力/调用方配合场景加注释而非当缺陷修；只有真实缺陷才修。

---

## 0. 关键裁定：P0 vs P1 的分歧

四路专家在「慢/stuck handler 的 DLQ 行为」上正面冲突——PG 专家判 P0（无限循环），对等专家判 P1（仅永不返回的 handler 才循环）。我追代码（`consumer.go` `settle` / `fetchBatch`）裁定如下：

- `deliveries` **只在 nak 路径**自增（`consumer.go:180`、`:184`）；`fetchBatch` 的可见性续期 `UPDATE … SET due_at=now()+$vis`（`:160`）**不自增**。
- **handler 会返回错误（即便很慢）** → 命中 nak → `deliveries++` → 达 `maxRetries` 进 DLQ。**不构成无限循环。**
- **handler 永不返回**（死锁 / 无超时网络挂起）→ `pollLoop` 卡在 `consumer.go:124` 的 `handler(ctx, m)`，永远到不了 `settle`，`deliveries` 永不增 → **永不进 DLQ**，并把消费 goroutine 楔死。

**裁定：P1（非 P0）。** 理由：(1) 无数据丢失；(2) "无限"仅发生在契约违规（handler 永不返回）时；(3) 常见慢 handler 场景是「DLQ 时机 + 重复执行面」与 NATS 的**语义偏离**，非正确性断裂。但它仍是**头条**：既是与 NATS 的真实语义偏离，又缺 NATS 的 server-side 兜底，且文档对等声明与代码不符。

---

## 1. 确认的真实问题（按优先级）

### P1-1（头条）可见性超时重投不计入投递计数 + 无租约续期 → 与 NATS 语义偏离 + 阻塞 handler 永不 DLQ

**位置**：`consumer.go:160`（续期不增 `deliveries`）、`consumer.go:173-186`（`settle`，仅 nak 路径自增）。

**NATS 对比**：JetStream 的 `NumDelivered` 对**每一次**重投（含 `AckWait` 超时）都 +1，因此慢/卡 handler 经若干次 `AckWait` 周期后达 `MaxDeliver` → DLQ（`nats/consumer.go` `decideAck` 以 `meta.NumDelivered` 为闸）。PG 的 `deliveries` 只在显式 nak 时 +1，可见性超时重投不计。

**两类后果**：
- **慢但会返回的 handler**：PG 会在 `due_at` 到期后重投（单消费者 + `batch>1` 时为串行重复；多消费者同 group 时为并发重复），**但不计入 maxRetries** → DLQ 出现得比 NATS 晚；重复执行面也不同。触发条件：`batch × handler_p99 > visibility_timeout`（默认 16×…>30s，即 handler p99 > ~2s 即触发批次内重复）。
- **永不返回的 handler**：见 §0 裁定——楔死 goroutine，永不 DLQ。

**修复方向**（需设计决策，见 §3）：在 `fetchBatch` 重新锁定已到期行时 `deliveries = deliveries + 1`，并以自增后的值判定 DLQ；可选再加 handler 运行期间的可见性心跳续期（真正对齐 NATS 的"在途可续"语义）。专家在此点收敛。

**关联文档错误**：`docs/pg-mq-design.md` §5 行 123「`deliveries`=NumDelivered」、§10 行 157「投递计数 | NumDelivered | deliveries 列」**与代码不符**，须改。

---

### P1-2 `fetchBatch` 未检查 `rows.Err()` → 流式错误被静默当成"空批次"

**位置**：`consumer.go:148-158`。`for rows.Next() {}` 结束后直接 `rows.Close()`，未读 `rows.Err()`。

pgx v5 中 `rows.Next()` 返回 false 既可能是行耗尽，也可能是流式获取中途出错（连接中断、ctx 取消、后续行权限错误）。错误挂在 `Rows` 上须用 `rows.Err()` 取回。此处永不读 → 连接/池健康信号丢失，且 ctx 中途取消时仍会对半残事务提交可见性续期 UPDATE。教科书级 pgx v5 修正：

```go
rows.Close()
if err := rows.Err(); err != nil {
    return nil, err
}
```

（注：`rows.Close()` 在扫描错误分支 `:152` 与正常路径 `:158` 各调一次——pgx v5 幂等，**非 bug**，已核。）

---

### P1-3 `settle` 的 nak UPDATE 静默吞错 → 可冻结 `deliveries` 计数，叠加 P1-1

**位置**：`consumer.go:174-185`。

逐路径裁定：
- **ack（`DELETE`）吞错**：可接受（at-least-once，行未删则下次重投再 ack）。✓ 故意。
- **nak（`UPDATE …+1`）吞错**：**掩盖真实问题**。UPDATE 失败时 `deliveries` 不自增、`due_at` 仍停在 `fetchBatch` 设的 `now()+vis` → 该消息每轮要等满 `visibility_timeout`（默认 30s）才重投，且计数不涨 → 可绕过 `maxRetries` 上限。与 P1-1 叠加放大。
- **DLQ 失败回退 UPDATE 吞错**：`dlqAndRemove` 失败 + 回退 UPDATE 也失败 → 消息既不在 DLQ、又卡在 deliveries，**无声搁浅**（非丢失，行仍在）。

**修复**：nak/DLQ 回退路径至少 warn 级日志（ack 路径可保持静默）。at-least-once 正确性不破，但 `maxRetries` 上限保证与可运维性需要这个。

---

### P1-4 reaper 按 `created_at` 删 message，CASCADE 连带删在途/重试中的 delivery → 慢 group 静默丢消息

**位置**：`client.go:101` `DELETE FROM messages WHERE created_at < $1` + `schema.go:20` `ON DELETE CASCADE`。

**场景**：`retention=7d`（默认）。一条消息尚在重试（`deliveries` 表里有行、未达 DLQ），到第 7 天被 reaper 连 delivery 行 CASCADE 删除 → 消息既未 ack 也未 DLQ，凭空消失。消费方离线 >7d 的积压同理蒸发，而契约宣称支持 durable resume。NATS 的 `MaxAge` 语义上也会过期，但 NATS 把重试计数放在 consumer（`NumDelivered` 在 consumer 上），消息过期不重置重试；PG 的重试状态唯一存在 delivery 行里，过期即连同重试状态一起丢——这是 PG 独有的新失败面。

**修复方向**：reaper 加守卫 `AND NOT EXISTS (SELECT 1 FROM deliveries WHERE deliveries.message_id = messages.id)`（保留仍有投递/重试的消息），或对临期且仍待投递的消息先迁入 DLQ 再删。默认 7d 下概率低，但仍须堵。

---

### P1-5 轮询循环对 fetch 错误静默 `continue` → PG 故障期零可观测

**位置**：`consumer.go:112-118`。

PG 宕机/池打满时 `fetchBatch` 每轮失败，`continue` 直接到下一 tick。**已核：非紧密自旋**（`time.After(interval)` 以 500ms 门控），但仍**零日志/零指标** → 故障与"空闲 topic"无法区分，消费方可能静默失败数小时。

**立场张力**：可观测性按用户既定分阶段接入——但此处是"故障静默"，非单纯"缺指标"，属于可靠性盲区。建议**最小**处理：down→up 状态转换时打一条 warn（避免每 500ms 刷屏），完整 metrics 留待可观测性阶段。

---

### P1-6（跨后端·契约）Subscribe 把轮询生命周期绑在调用方 ctx 上

**位置**：`consumer.go:65` `cctx, cancel := context.WithCancel(ctx)`；同形 bug 也在 NATS 后端 `nats/consumer.go`。

**场景**：若用启动期的 RPC/请求 ctx 调 `Subscribe`，该请求一返回 ctx 即取消 → `cctx` 取消 → `pollLoop` 静默退出，订阅永久死亡（无错无日志，handle 注册表里仍"活着"）。

**裁定**：这是**契约/用法**问题，非 PG 专属代码 bug——(a) NATS 后端同病；(b) 当前无生产调用方（仅测试用 `context.Background()`），纯潜伏；(c) 契约 doc 措辞（"ctx 在订阅关闭时取消"）暗示 ctx 是 close 的**结果**而非**生命源**，实现却把调用方 ctx 当生命源。**建议**：契约 doc 明确"Subscribe 的 ctx 必须是进程级长生命 ctx，不得用请求 ctx"；可选地在**两个**后端把 pollLoop 从调用方 ctx 解耦（派生自 handle/进程级 lifetime ctx），调用方 ctx 只用于同步 setup。优先级随首个真实常驻消费者的接入而升高。

---

### P2-7 无 `MaxAckPending` 对等物 → 慢 handler + 快生产者可累积无界在途

**位置**：`pollLoop` 每 `poll_interval` 取 `batch` 条、行锁在 fetch-commit 即释放。NATS 有 `MaxAckPending`（默认 256）硬限每消费者在途；PG 无对等闸。**对等表 §10 行 161「在途上限」夸大了对等**。低频后备定位下可接受；文档注明 + 可选加信号量。

### P2-8 新订阅 backfill 为无界单条 INSERT

**位置**：`consumer.go:90-95`。高频 topic + 7 天窗口下，首次 Subscribe 单事务插入百万级行，长锁、WAL 膨胀、阻塞并发 publish。建议分批（`id > cursor LIMIT N` 循环提交）或限制回放窗口。

### P2-9 轮询查询强制排序（索引未覆盖 ORDER BY）

**位置**：`consumer.go:139-143` `ORDER BY d.message_id` vs 索引 `(group_name, topic, due_at)`（`schema.go:25`）。ORDER BY 列不在索引内 → 收集后显式 Sort。改为 `ORDER BY due_at, message_id` + 索引 `(group_name, topic, due_at, message_id) INCLUDE (deliveries)` 可走纯索引扫描。

### P2-10 `stop()` 在 15s ctx 超时后返回，不保证 poll goroutine 已退出 → `pool.Close()` 可能与在途 goroutine 竞态

**位置**：`consumer.go:26-34`、`client.go:39-45`。仅当 handler 不尊重 ctx 且关停 >15s 时触发；正常 handler 毫秒级退出。`pool.Close()` 会阻塞至连接归还，故非 panic，但可能拖长关停。`stop()` 注释"waits for the poll goroutine to exit (bounded by ctx)"在超时分支下不成立。

### P2-11 drain 快照后注册的订阅逃逸 drain → 跑进正在关闭的池

**位置**：`client.go:74-84` vs `:58-65`。`Subscribe` 与关停并发时，drainSubs 快照之后才注册的订阅不会被停。潜伏（当前 Subscribe 仅测试同步调用）。建议关停开始后拒绝新 Subscribe。

### P2-12 错误映射把 `*pgconn.PgError` 一律归 `ErrUnavailable`

**位置**：`error.go`。唯一契约哨兵是 `ErrInvalidArgument`/`ErrUnavailable`/`ErrTimeout`，PG 语义错误（如 `23505` 唯一冲突）归 Unavailable 对调用方略有误导，但原始错误经 `%w` 保留可检。轻微。

### P2-13 Nak 退避无上限

**位置**：`consumer.go:179` `(deliveries+1)*100ms`。`maxRetries=5` 下峰值 ~600ms，永不触及；仅当 `maxRetries` 配极高时才显现。理论问题。

---

## 2. 已核掉的误报（明确记录，避免重复）

- **`UNIQUE(group_name, message_id)` 去重竞态**：约束在 `schema.go:23` 实际存在，fanout（`publisher.go:37`）与 backfill（`consumer.go:93`）的 `ON CONFLICT DO NOTHING` 正确去重 publish-fanout vs subscribe-backfill 竞态。✓ 非 bug。
- **`FOR UPDATE SKIP LOCKED` 在并发同 group 消费者下的幻读/丢更新**：行锁持有贯穿整个 fetch 事务，可见性续期在同一事务内，两并发 fetch 不会抓同一行。✓ fetch 层正确——问题在 COMMIT 之后（见 P1-1）。
- **`fetchBatch` 中途失败引发热循环**：deferred Rollback 回滚整事务（SELECT FOR UPDATE + 全部 UPDATE），行回到 `due_at<=now()`，但 `poll_interval` 门控重取，非热循环。✓
- **resume 丢/重复消息**：ack 即 `DELETE`（`consumer.go:175`），已有 subscribers 行 → RowsAffected=0 → 不 backfill；`TestDurableResume` 证实。✓
- **`drainSubs` 快照-锁外迭代的并发**：锁内拷贝、锁外迭代，`unregister` 并发 delete 不影响快照；`cancel()` 幂等、`wg.Wait()` 归零安全。✓ 无 data race。
- **`NewClient`/reaper 用 `context.Background()`**：构造期/长生命后台操作，正确。✓
- **数值/时间边界**：bigserial→`%d`/`FormatInt`、`time.Now().Add(-retention)` 符号、`max()` builtin、退避无溢出。✓ 全部正确；`maxRetries=N` 恰好 N 次进 DLQ，已追全程。
- **`sync.WaitGroup.Go`**：Go 1.25（go.mod 已声明），有效。✓
- **Commit 后 deferred Rollback**：pgx v5 返回 `ErrTxClosed` 类错误，`_ =` 丢弃正确，提交结果已定。✓
- **`pgconn.ErrConnClosed`**：v5 正确符号（非 v4 的 `pgx.ErrConnDone`）。✓
- **`pool.Close()` 后使用 panic**：pgxpool 不 panic，`Acquire`/`Exec` 返回错误。✓
- **双重 stop**：`cancel()` 幂等、每次 stop 各建 `done` chan，无重复关闭。✓
- **NATS 无 meta 的毒消息无限循环**：`decideAck` 在 `meta==nil` 返回 `termMsg` 显式中断。✓
- **PG DLQ 有"只入 DLQ 不删 delivery"窗口**：`dlqAndRemove` 单事务（`consumer.go:191-207`），PG 此处**强于** NATS（NATS 的 DLQ-publish + Term 是分离可竞态操作）。✓
- **`json.Unmarshal(headers)` 吞错**：publisher 是 headers 列唯一写者（私有表，恒写 `map[string]string` 的 marshal），失败则 handler 得空 header（降级非不安全）。✓ 故意。

---

## 3. 修复决策点（需用户拍板）

1. **P1-1 修法选择**（头条，影响可观察行为）：
   - (a) **对齐 NATS**：`fetchBatch` 重锁到期行时 `deliveries++`，DLQ 判定改用自增后值。→ 阻塞 handler 也能在若干可见性周期后进 DLQ，真正语义对齐。
   - (b) 仅文档化"PG 的 maxRetries 只计显式失败，不计可见性超时重投"，承认与 NATS 的轻微偏离。
   - 推荐 **(a)**：这是真实语义偏离 + 阻塞 handler 兜底缺失，(a) 一次性解决 P1-1 及其引发的批次重复面（P1-1 的"慢但返回"子项同源于此）。
2. **P1-4 reaper**：加 `NOT EXISTS` 守卫（推荐）还是接受 NATS-MaxAge 等价语义仅文档化？
3. **P1-5 静默 continue**：按可观测性阶段**推迟**，仅加一条 down→up warn？还是完整留待统一接入？
4. **P1-6 Subscribe ctx**：仅文档化契约，还是顺带在两个后端解耦 pollLoop 生命源？
5. **P2 批**：多数是文档/可选硬化，是否本轮一并处理，还是随"注释优化"批次一起？

---

## 4. 建议的修复批次

- **现在修（真实代码缺陷）**：P1-1（按决策 1a）、P1-2（`rows.Err()`，一行）、P1-3（nak/DLQ 回退日志）、P1-4（reaper 守卫，按决策 2）。
- **文档修（与代码不符）**：`pg-mq-design.md` §5 行 123、§10 行 157/156/161。
- **按立场推迟**：P1-5（可观测性阶段）、P2 批（硬化/文档）。
- **契约层**：P1-6 视首个常驻消费者接入时机决定。

关联：`docs/mq-review-findings.md`（NATS 后端审查）、`docs/pg-mq-design.md`（设计，§5/§10 需据本报告修正）、`docs/mq-validation-report.md`（测试覆盖——P1-1/2/3/4 均未被现有测试触达，因其用快同步 handler）。
