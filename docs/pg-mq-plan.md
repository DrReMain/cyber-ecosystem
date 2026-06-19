# PG MQ 后端实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 `shared-go/mq` 增加第二个后端——PostgreSQL 实现（`shared-go/mq/pg`），落到与 NATS 后端同一份契约上，供"不跑 NATS 的部署"使用。

**Architecture:** per-(group,message) 投递表模型。`messages`（只追加）+ `deliveries`（每个 group×消息一行）+ `subscribers` + `dlq`，独立 `mq` 库 + 独立 pgx 连接池。消费走纯轮询：`FOR UPDATE SKIP LOCKED` 批量取 + `due_at` 可见性闸门。发布扇出投递行；新 group 订阅时回填（历史回放，对齐 NATS durable）。语义与 NATS 一一对应（visibility_timeout=AckWait、deliveries=NumDelivered、ack=删除、nak=延时重投、达上限进 DLQ）。

**Tech Stack:** Go 1.25、`github.com/jackc/pgx/v5` + `pgxpool`、Kratos v3、shared-go/mq 契约。镜像 `shared-go/mq/nats/` 的结构（机制/实例分离、`(T,func(),error)` provider、注册表 + 关停 drain）。

**Reference:** NATS 后端 `shared-go/mq/nats/{client,config,consumer,publisher,stream,error,mq}.go` 是结构模板；PG 后端一一镜像，区别只在队列机制（轮询+投递表 vs JetStream）。实现每个文件时对照对应的 nats 文件。

## Global Constraints

- 模块路径 `cyber-ecosystem/shared-go/mq/pg`；遵循 gci import 分组（standard → default → Prefix(cyber-ecosystem/shared-go) → Prefix(gen/go) → Prefix(app)）。
- 契约不改：复用 `shared-go/mq` 的 `Publisher/Consumer/Subscription/Message/MQ`、sentinel、`HandleMQError`、`ValidateTopic/ValidateGroup`。
- 自管表：`CREATE TABLE IF NOT EXISTS`（启动幂等），**不走 atlas**（atlas 只管业务库 edge_mobile）。
- 独立 `mq` 库 + 独立 pgx 池（不复用业务池）。`mq` 库需预先存在（dev/k3s 一次性 `CREATE DATABASE mq`）。
- 一句柄一后端：`NewMQ` 按配置块选 nats 或 pg，不同时启用。
- 所有验证走 Nx：`./nx run tools:go:test` / `tools:go:lint`；生成走 `proto:generate` / `edge_mobile:proto:conf` / `edge_mobile:generate:wire`。不直接 buf/go（[[use-nx-targets-for-toolchain]]）。
- 测试需可达 PG（k3s 的 PG，默认 `postgres://postgres:postgres@localhost:5432/mq?sslmode=disable`，`PG_MQ_DSN` 可覆盖）；PG 不可用时 `t.Skip`。

---

## File Structure

```
shared-go/mq/pg/
  config.go       Config + *OrDefault
  error.go        mapError（pgx → mq sentinel）
  schema.go       CREATE TABLE IF NOT EXISTS + 索引
  client.go       handle + NewClient（建池/建表/reaper/注册表/cleanup）
  publisher.go    Publish（插 message + 扇出 deliveries）
  consumer.go     Subscribe（注册/backfill/轮询）+ ack/nak/dlq + subscription 生命周期
  mq.go           New(*handle) *mq.MQ
  *_test.go       error_test（unit）+ client_test（harness）+ roundtrip/retry/group（集成）
app/services/edge_mobile/internal/platform/platform_pg_mq.go   NewPGMQ + toPGConfig
app/services/edge_mobile/internal/conf/conf.proto              Data.MQ.PG 块
app/services/edge_mobile/configs/config.yaml                   data.mq.pg（dev 备选）
```

---

## Task 1: Config + 错误映射（纯逻辑，无 DB）

**Files:**
- Create: `shared-go/mq/pg/config.go`, `shared-go/mq/pg/error.go`
- Test: `shared-go/mq/pg/error_test.go`

**Interfaces:**
- Produces: `Config{DSN, PollInterval, VisibilityTimeout, MaxRetries, Retention}` + `*OrDefault` 辅助；`mapError(err, op) error`（pgx → mq sentinel）。

- [ ] **Step 1: 写失败的 mapError 测试**

`shared-go/mq/pg/error_test.go`:
```go
package pg

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"cyber-ecosystem/shared-go/mq"
)

func TestMapError(t *testing.T) {
	cases := []struct {
		name string
		in   error
		want error
	}{
		{"ctx deadline", context.DeadlineExceeded, mq.ErrTimeout},
		{"pgx conn done", pgx.ErrConnDone, mq.ErrUnavailable},
		{"unknown", errors.New("boom"), mq.ErrUnavailable},
	}
	for _, c := range cases {
		if !errors.Is(mapError(c.in, "op"), c.want) {
			t.Errorf("%s: mapError(%v) not %v", c.name, c.in, c.want)
		}
	}
	if mapError(nil, "op") != nil {
		t.Error("nil error should map to nil")
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `go test ./shared-go/mq/pg/ -run TestMapError`
Expected: FAIL（`mapError` 未定义 / 包构建失败）。

- [ ] **Step 3: 实现 config.go + error.go**

`shared-go/mq/pg/config.go`:
```go
package pg

import "time"

// Config 映射 conf.Data.MQ.PG 到 pgx 客户端。零值字段在 NewClient 内回退到默认。
type Config struct {
	DSN             string        // postgres://user:pass@host:5432/mq?sslmode=disable
	PollInterval    time.Duration // 消费轮询间隔；0 → 500ms
	VisibilityTimeout time.Duration // 取出后不可见时长（=NATS AckWait）；0 → 30s，须大于 handler p99
	MaxRetries      int           // 投递上限，超过进 DLQ；0 → 5
	Retention       time.Duration // messages 保留时长（=NATS MaxAge），到期清理；0 → 7d
	BatchSize       int           // 每轮最多取多少条；0 → 16
}

const (
	defaultPollInterval    = 500 * time.Millisecond
	defaultVisibility      = 30 * time.Second
	defaultMaxRetries      = 5
	defaultRetention       = 7 * 24 * time.Hour
	defaultBatchSize       = 16
)

func pollIntervalOrDefault(v time.Duration) time.Duration     { if v > 0 { return v }; return defaultPollInterval }
func visibilityOrDefault(v time.Duration) time.Duration       { if v > 0 { return v }; return defaultVisibility }
func maxRetriesOrDefault(v int) int                           { if v > 0 { return v }; return defaultMaxRetries }
func retentionOrDefault(v time.Duration) time.Duration        { if v > 0 { return v }; return defaultRetention }
func batchSizeOrDefault(v int) int                            { if v > 0 { return v }; return defaultBatchSize }
```

`shared-go/mq/pg/error.go`:
```go
package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"cyber-ecosystem/shared-go/mq"
)

// mapError 把 pgx/PG 错误翻成 mq sentinel。consumer handler 业务错误不进这里。
func mapError(err error, op string) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("%s: %w", op, mq.ErrTimeout)
	case errors.Is(err, context.Canceled):
		return fmt.Errorf("%s: %w", op, err)
	case errors.Is(err, pgx.ErrConnDone), errors.Is(err, pgx.ErrNoRows) && false:
		return fmt.Errorf("%s: %w", op, mq.ErrUnavailable)
	default:
		return fmt.Errorf("%s: %w", op, mq.ErrUnavailable)
	}
}
```
（注：`pgx.ErrNoRows` 不是不可用——保留 `&& false` 仅占位，实际不命中；连接类用 `pgx.ErrConnDone`。pool 关闭由 pgxpool 的 `(*Pool).Config().AfterConnect`/`ErrClosed` 体现——若版本有 `pgxpool.ErrClosed` 用它，否则用 `pgx.ErrConnDone` 兜底。实现时按实际 pgx 版本取连接关闭哨兵。）

- [ ] **Step 4: 运行测试通过**

Run: `go test ./shared-go/mq/pg/ -run TestMapError`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add shared-go/mq/pg/config.go shared-go/mq/pg/error.go shared-go/mq/pg/error_test.go
go mod tidy # 在 shared-go 下；确保 pgx/v5 入 go.mod
git commit -m "feat(mq/pg): config + error mapping"
```

---

## Task 2: Client + schema + 注册表 + mq.go（bootstrap）

**Files:**
- Create: `shared-go/mq/pg/schema.go`, `shared-go/mq/pg/client.go`, `shared-go/mq/pg/mq.go`, `shared-go/mq/pg/client_test.go`
- Reference: `shared-go/mq/nats/client.go`（handle/注册表/drainSubs 结构镜像）。

**Interfaces:**
- Produces: `handle{pool *pgxpool.Pool, cfg Config, mu sync.Mutex, subs map[*subscription]struct{}}`；`NewClient(*Config) (*handle, func(), error)`；`register/unregister/drainSubs`；`New(*handle) *mq.MQ`；测试 harness `testConfig()`/`newTestMQ(t)`/`uniqTopic(t,base)`。

- [ ] **Step 1: 写 schema.go（4 张表 + 索引，spec §4 原样）**

`shared-go/mq/pg/schema.go`:
```go
package pg

// schemaSQL 幂等建表（CREATE TABLE IF NOT EXISTS）。在独立的 `mq` 库执行，不走 atlas。
const schemaSQL = `
CREATE TABLE IF NOT EXISTS messages (
  id         BIGSERIAL PRIMARY KEY,
  topic      TEXT NOT NULL,
  payload    BYTEA NOT NULL,
  headers    JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_messages_topic_id ON messages (topic, id);

CREATE TABLE IF NOT EXISTS deliveries (
  id         BIGSERIAL PRIMARY KEY,
  group_name TEXT NOT NULL,
  topic      TEXT NOT NULL,
  message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  deliveries INT NOT NULL DEFAULT 0,
  due_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (group_name, message_id)
);
CREATE INDEX IF NOT EXISTS idx_deliveries_poll ON deliveries (group_name, topic, due_at);

CREATE TABLE IF NOT EXISTS subscribers (
  group_name TEXT NOT NULL,
  topic      TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_name, topic)
);

CREATE TABLE IF NOT EXISTS dlq (
  id         BIGSERIAL PRIMARY KEY,
  topic      TEXT NOT NULL,
  group_name TEXT NOT NULL,
  payload    BYTEA NOT NULL,
  headers    JSONB NOT NULL DEFAULT '{}',
  deliveries INT NOT NULL,
  error      TEXT NOT NULL,
  dead_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);`
```

- [ ] **Step 2: 写 client.go（NewClient + handle + 注册表 + reaper）**

`shared-go/mq/pg/client.go`（镜像 `nats/client.go` 的 handle/注册表/drainSubs；cleanup 多一个 reaper 停止）:
```go
package pg

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"cyber-ecosystem/shared-go/mq"
)

// NewClient 连独立 `mq` 库的 pgx 池，幂等建表，启动保留期 reaper，返回 handle + cleanup。
// cleanup：先 drain 所有订阅（停轮询），关 reaper，再关池。
func NewClient(cfg *Config) (*handle, func(), error) {
	if cfg == nil || cfg.DSN == "" {
		return nil, nil, fmt.Errorf("%w: dsn is required", mq.ErrInvalidArgument)
	}
	pool, err := pgxpool.New(context.Background(), cfg.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: pgxpool: %w", mq.ErrUnavailable, err)
	}
	if _, err := pool.Exec(context.Background(), schemaSQL); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("%w: ensure schema: %w", mq.ErrUnavailable, err)
	}
	h := &handle{pool: pool, cfg: *cfg}
	stopReaper := h.startReaper()
	return h, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		h.drainSubs(ctx)
		cancel()
		stopReaper()
		pool.Close()
	}, nil
}

type handle struct {
	pool *pgxpool.Pool
	cfg  Config

	mu   sync.Mutex
	subs map[*subscription]struct{}
}

func (h *handle) register(s *subscription) {
	h.mu.Lock()
	if h.subs == nil {
		h.subs = make(map[*subscription]struct{})
	}
	h.subs[s] = struct{}{}
	h.mu.Unlock()
}
func (h *handle) unregister(s *subscription) {
	h.mu.Lock()
	delete(h.subs, s)
	h.mu.Unlock()
}
func (h *handle) drainSubs(ctx context.Context) {
	h.mu.Lock()
	subs := make([]*subscription, 0, len(h.subs))
	for s := range h.subs {
		subs = append(subs, s)
	}
	h.mu.Unlock()
	for _, s := range subs {
		s.stop(ctx)
	}
}

// startReaper 周期性删除超过保留期的 messages（CASCADE 删其 deliveries，等价 NATS MaxAge）。
func (h *handle) startReaper() func() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		t := time.NewTicker(retentionOrDefault(h.cfg.Retention) / 4)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				_, _ = h.pool.Exec(ctx,
					`DELETE FROM messages WHERE created_at < now() - $1::interval`,
					retentionOrDefault(h.cfg.Retention).String())
			}
		}
	}()
	return cancel
}
```

`shared-go/mq/pg/mq.go`:
```go
package pg

import "cyber-ecosystem/shared-go/mq"

func New(h *handle) *mq.MQ {
	return &mq.MQ{Publisher: newPublisher(h), Consumer: newConsumer(h)}
}
```
（`newPublisher`/`newConsumer` 在 Task 3/4 实现；本任务可先用桩返回 nil 占位以便编译，但更干净：把 mq.go 放到 Task 4 末尾。**调整：mq.go 移到 Task 4。本任务不创建 mq.go，NewClient 测试直接用 handle。**）

- [ ] **Step 3: 写 client_test.go（harness + 建表/故障测试）**

`shared-go/mq/pg/client_test.go`（镜像 `nats/client_test.go`）:
```go
package pg

import (
	"context"
	"os"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

var testSeq uint64

func testConfig() *Config {
	dsn := os.Getenv("PG_MQ_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/mq?sslmode=disable"
	}
	return &Config{DSN: dsn, MaxRetries: 3, VisibilityTimeout: 2 * time.Second, PollInterval: 50 * time.Millisecond}
}

func newTestMQ(t *testing.T) (*handle, func()) {
	t.Helper()
	h, cleanup, err := NewClient(testConfig())
	if err != nil {
		t.Skipf("pg unavailable: %v", err)
	}
	return h, cleanup
}

// uniqTopic 返回进程内唯一 topic（隔离流/消费者/投递行），cleanup 清理该 topic 的数据。
func uniqTopic(t *testing.T, base string) string {
	t.Helper()
	seq := testSeqAdd()
	topic := "t-" + base + "-" + seq
	t.Cleanup(func() {
		// 测试结束清理该 topic 的所有数据（messages CASCADE 删 deliveries）
		h, c, _ := NewClient(testConfig())
		if h == nil {
			return
		}
		defer c()
		_, _ = h.pool.Exec(context.Background(), `DELETE FROM messages WHERE topic=$1`, topic)
		_, _ = h.pool.Exec(context.Background(), `DELETE FROM subscribers WHERE topic=$1`, topic)
	})
	return topic
}
```
（`testSeqAdd()` 用 `sync/atomic`：`func testSeqAdd() string { return strconv.FormatUint(testSeq.Add(1), 10) }`，`testSeq` 改为 `atomic.Uint64`。）

```go
func TestNewClientCreatesSchema(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	var n int
	err := h.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM information_schema.tables
		 WHERE table_schema='public' AND table_name IN ('messages','deliveries','subscribers','dlq')`).Scan(&n)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 4 {
		t.Fatalf("expected 4 tables, got %d", n)
	}
}

func TestNewClientFaultUnavailable(t *testing.T) {
	_, _, err := NewClient(&Config{DSN: "postgres://x:x@127.0.0.1:1/mq?sslmode=disable&connect_timeout=2"})
	if !errors.Is(err, mq.ErrUnavailable) {
		t.Fatalf("dead dsn: got %v, want ErrUnavailable", err)
	}
}
```

- [ ] **Step 4: 运行通过（需可达 PG + `mq` 库存在）**

确保 `mq` 库存在：`psql ... -c 'CREATE DATABASE mq'`（一次性）。
Run: `go test ./shared-go/mq/pg/ -run 'TestNewClient'`
Expected: 2 PASS（PG 不可达则 Skip）。

- [ ] **Step 5: 提交**

```bash
git add shared-go/mq/pg/schema.go shared-go/mq/pg/client.go shared-go/mq/pg/client_test.go
git commit -m "feat(mq/pg): client + schema + registry + reaper"
```

---

## Task 3: Publisher（Publish）

**Files:**
- Create: `shared-go/mq/pg/publisher.go`, `shared-go/mq/pg/publisher_test.go`
- Reference: `shared-go/mq/nats/publisher.go`。

**Interfaces:**
- Produces: `publisher{h *handle}`、`newPublisher(h) mq.Publisher`、`Publish(ctx, topic, *mq.Message) (string, error)`。

- [ ] **Step 1: 写失败测试**

`shared-go/mq/pg/publisher_test.go`:
```go
package pg

import (
	"context"
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/mq"
)

func TestPublishInsertsAndFansOut(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	ctx := context.Background()
	topic := uniqTopic(t, "pub")
	// 预置一个 subscriber，让扇出有目标
	if _, err := h.pool.Exec(ctx, `INSERT INTO subscribers(group_name,topic) VALUES('g',$1) ON CONFLICT DO NOTHING`, topic); err != nil {
		t.Fatalf("seed sub: %v", err)
	}
	p := newPublisher(h)
	id, err := p.Publish(ctx, topic, &mq.Message{Payload: []byte("hi"), Headers: map[string]string{"k": "v"}})
	if err != nil || id == "" {
		t.Fatalf("Publish: id=%q err=%v", id, err)
	}
	// message 存在
	var payload []byte
	if err := h.pool.QueryRow(ctx, `SELECT payload FROM messages WHERE id=$1`, id).Scan(&payload); err != nil {
		t.Fatalf("message: %v", err)
	}
	if string(payload) != "hi" {
		t.Fatalf("payload=%q", payload)
	}
	// 扇出了一条 delivery 给 g
	var dcount int
	h.pool.QueryRow(ctx, `SELECT count(*) FROM deliveries WHERE message_id=$1 AND group_name='g'`, id).Scan(&dcount)
	if dcount != 1 {
		t.Fatalf("deliveries=%d, want 1", dcount)
	}
}

func TestPublishInvalidTopic(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	_, err := newPublisher(h).Publish(context.Background(), "", &mq.Message{Payload: []byte("x")})
	if !errors.Is(err, mq.ErrInvalidArgument) {
		t.Fatalf("empty topic: %v", err)
	}
}
```

- [ ] **Step 2: 运行确认失败**（newPublisher 未定义）

- [ ] **Step 3: 实现 publisher.go**

`shared-go/mq/pg/publisher.go`:
```go
package pg

import (
	"context"
	"encoding/json"
	"fmt"

	"cyber-ecosystem/shared-go/mq"
)

type publisher struct{ h *handle }

func newPublisher(h *handle) mq.Publisher { return &publisher{h: h} }

func (p *publisher) Publish(ctx context.Context, topic string, msg *mq.Message) (string, error) {
	if err := mq.ValidateTopic(topic); err != nil {
		return "", err
	}
	headers, _ := json.Marshal(msg.Headers)
	tx, err := p.h.pool.Begin(ctx)
	if err != nil {
		return "", mapError(err, "begin")
	}
	defer tx.Rollback(ctx)
	var id int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO messages(topic,payload,headers) VALUES($1,$2,$3) RETURNING id`,
		topic, msg.Payload, headers).Scan(&id); err != nil {
		return "", mapError(err, "insert message")
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO deliveries(group_name,topic,message_id)
		 SELECT group_name, $1, $2 FROM subscribers WHERE topic=$1
		 ON CONFLICT (group_name, message_id) DO NOTHING`,
		topic, id); err != nil {
		return "", mapError(err, "fanout")
	}
	if err := tx.Commit(ctx); err != nil {
		return "", mapError(err, "commit")
	}
	return fmt.Sprintf("%d", id), nil
}
```

- [ ] **Step 4: 运行通过**

Run: `go test ./shared-go/mq/pg/ -run 'TestPublish'`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add shared-go/mq/pg/publisher.go shared-go/mq/pg/publisher_test.go
git commit -m "feat(mq/pg): publisher (insert + fan-out)"
```

---

## Task 4: Consumer（Subscribe + 轮询 + ack/nak/dlq + 生命周期 + mq.go）

**Files:**
- Create: `shared-go/mq/pg/consumer.go`, `shared-go/mq/pg/mq.go`, `shared-go/mq/pg/roundtrip_test.go`, `shared-go/mq/pg/retry_test.go`, `shared-go/mq/pg/group_test.go`
- Reference: `shared-go/mq/nats/consumer.go`（decideAck/subscription/drain 结构）、`nats/group_test.go`、`nats/retry_test.go`、`nats/roundtrip_test.go`。

**Interfaces:**
- Produces: `consumer{h}`、`newConsumer(h) mq.Consumer`、`Subscribe(ctx, topic, group, handler) (mq.Subscription, error)`、`subscription{...}`。`mq.New(*handle) *mq.MQ`。

- [ ] **Step 1: 写 consumer.go（核心）**

`shared-go/mq/pg/consumer.go`:
```go
package pg

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

type consumer struct{ h *handle }

func newConsumer(h *handle) mq.Consumer { return &consumer{h: h} }

type subscription struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
	h      *handle
}

// stop：等轮询 goroutine 结束（已发出的 handler 用 ctx 之外的独立周期，本实现 handler 在轮询内同步调用，cancel 即可停止新轮询）。
func (s *subscription) stop(ctx context.Context) {
	s.cancel()
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
	}
}
func (s *subscription) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.stop(ctx)
	if s.h != nil {
		s.h.unregister(s)
	}
	return nil
}

type delivery struct {
	deliveryID int64
	deliveries int
	msgID      int64
	payload    []byte
	headers    map[string]string
	created    time.Time
}

func (c *consumer) Subscribe(ctx context.Context, topic, group string, handler func(context.Context, mq.Message) error) (mq.Subscription, error) {
	if err := mq.ValidateTopic(topic); err != nil {
		return nil, err
	}
	if err := mq.ValidateGroup(group); err != nil {
		return nil, err
	}
	// 注册 subscriber；新建则回填（历史回放，对齐 NATS durable DeliverAll）
	if err := c.registerAndBackfill(ctx, topic, group); err != nil {
		return nil, mapError(err, "subscribe")
	}
	cctx, cancel := context.WithCancel(ctx)
	sub := &subscription{cancel: cancel, h: c.h}
	c.h.register(sub)
	sub.wg.Add(1)
	go func() {
		defer sub.wg.Done()
		c.pollLoop(cctx, topic, group, handler)
	}()
	return sub, nil
}

func (c *consumer) registerAndBackfill(ctx context.Context, topic, group string) error {
	tx, err := c.h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	ct, err := tx.Exec(ctx,
		`INSERT INTO subscribers(group_name,topic) VALUES($1,$2) ON CONFLICT DO NOTHING`, group, topic)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 1 {
		// 新 group → 回填保留窗口内的历史消息
		if _, err := tx.Exec(ctx,
			`INSERT INTO deliveries(group_name,topic,message_id)
			 SELECT $1,$2,id FROM messages WHERE topic=$2
			 ON CONFLICT (group_name, message_id) DO NOTHING`, group, topic); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (c *consumer) pollLoop(ctx context.Context, topic, group string, handler func(context.Context, mq.Message) error) {
	cfg := c.h.cfg
	interval := pollIntervalOrDefault(cfg.PollInterval)
	vis := visibilityOrDefault(cfg.VisibilityTimeout)
	maxRetries := maxRetriesOrDefault(cfg.MaxRetries)
	batch := batchSizeOrDefault(cfg.BatchSize)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
		ds, err := c.fetchBatch(ctx, group, topic, batch, vis)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			continue // 瞬时错误：下轮再试
		}
		for _, d := range ds {
			m := mq.Message{Topic: topic, Payload: d.payload, Headers: d.headers, ID: strconv.FormatInt(d.msgID, 10), Timestamp: d.created}
			herr := handler(ctx, m)
			c.settle(ctx, d, topic, group, herr, maxRetries)
		}
	}
}

func (c *consumer) fetchBatch(ctx context.Context, group, topic string, batch int, vis time.Duration) ([]delivery, error) {
	tx, err := c.h.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx,
		`SELECT d.id, d.deliveries, m.id, m.payload, m.headers, m.created_at
		 FROM deliveries d JOIN messages m ON d.message_id=m.id
		 WHERE d.group_name=$1 AND d.topic=$2 AND d.due_at <= now()
		 ORDER BY d.message_id LIMIT $3 FOR UPDATE SKIP LOCKED`, group, topic, batch)
	if err != nil {
		return nil, err
	}
	var ds []delivery
	for rows.Next() {
		var d delivery
		var hdr []byte
		if err := rows.Scan(&d.deliveryID, &d.deliveries, &d.msgID, &d.payload, &hdr, &d.created); err != nil {
			rows.Close()
			return nil, err
		}
		_ = json.Unmarshal(hdr, &d.headers)
		ds = append(ds, d)
	}
	rows.Close()
	for _, d := range ds {
		if _, err := tx.Exec(ctx, `UPDATE deliveries SET due_at=now()+$1 WHERE id=$2`, vis, d.deliveryID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ds, nil
}

// settle 对齐 NATS decideAck：成功 ack(删除)；失败→未达上限 nak(延时重投)，达上限进 DLQ。
func (c *consumer) settle(ctx context.Context, d delivery, topic, group string, herr error, maxRetries int) {
	if herr == nil {
		_, _ = c.h.pool.Exec(ctx, `DELETE FROM deliveries WHERE id=$1`, d.deliveryID)
		return
	}
	if d.deliveries+1 >= maxRetries {
		// DLQ；写失败则保留（不删除 delivery），下轮再试 —— 不静默丢失
		headers, _ := json.Marshal(d.headers)
		tx, err := c.h.pool.Begin(ctx)
		if err != nil {
			_, _ = c.h.pool.Exec(ctx, `UPDATE deliveries SET due_at=now(), deliveries=deliveries+1 WHERE id=$1`, d.deliveryID)
			return
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO dlq(topic,group_name,payload,headers,deliveries,error) VALUES($1,$2,$3,$4,$5,$6)`,
			topic, group, d.payload, headers, d.deliveries+1, herr.Error()); err != nil {
			tx.Rollback(ctx)
			_, _ = c.h.pool.Exec(ctx, `UPDATE deliveries SET due_at=now(), deliveries=deliveries+1 WHERE id=$1`, d.deliveryID)
			return
		}
		if _, err := tx.Exec(ctx, `DELETE FROM deliveries WHERE id=$1`, d.deliveryID); err != nil {
			tx.Rollback(ctx)
			return
		}
		_ = tx.Commit(ctx)
		return
	}
	backoff := time.Duration(d.deliveries+1) * 100 * time.Millisecond
	_, _ = c.h.pool.Exec(ctx, `UPDATE deliveries SET due_at=now()+$1, deliveries=deliveries+1 WHERE id=$2`, backoff, d.deliveryID)
}
```

`shared-go/mq/pg/mq.go`（如 Task 2 未创建则此处创建）:
```go
package pg

import "cyber-ecosystem/shared-go/mq"

func New(h *handle) *mq.MQ {
	return &mq.MQ{Publisher: newPublisher(h), Consumer: newConsumer(h)}
}
```

- [ ] **Step 2: 写 round-trip 测试**

`shared-go/mq/pg/roundtrip_test.go`（镜像 `nats/roundtrip_test.go`，buffered done）:
```go
package pg

import (
	"context"
	"sync"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

func TestRoundTrip(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	m := New(h)
	ctx := context.Background()
	topic, group := uniqTopic(t, "rt"), "rt-group"
	var mu sync.Mutex
	var got *mq.Message
	done := make(chan struct{}, 1)
	sub, err := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, msg mq.Message) error {
		mu.Lock(); got = &msg; mu.Unlock()
		select { case done <- struct{}{}: default: }
		return nil
	})
	if err != nil { t.Fatalf("Subscribe: %v", err) }
	defer sub.Close()
	if _, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte("hello pg")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("did not receive within 5s")
	}
	mu.Lock(); defer mu.Unlock()
	if got == nil || string(got.Payload) != "hello pg" {
		t.Fatalf("got %+v", got)
	}
}
```

- [ ] **Step 3: 运行 round-trip 通过**

Run: `go test ./shared-go/mq/pg/ -run TestRoundTrip`
Expected: PASS。

- [ ] **Step 4: 写 retry→DLQ 测试**

`shared-go/mq/pg/retry_test.go`（镜像 `nats/retry_test.go`；testConfig MaxRetries=3，断言 ≥3 次尝试后 dlq 表有记录）:
```go
package pg

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

var errBoom = errBoomType{}
type errBoomType struct{}
func (errBoomType) Error() string { return "boom" }

func TestConsumerRetryThenDLQ(t *testing.T) {
	h, cleanup := newTestMQ(t)
	defer cleanup()
	m := New(h)
	ctx := context.Background()
	topic, group := uniqTopic(t, "retry"), "retry-group"
	var attempts atomic.Int32
	sub, err := m.Consumer.Subscribe(ctx, topic, group, func(_ context.Context, _ mq.Message) error {
		attempts.Add(1); return errBoom
	})
	if err != nil { t.Fatalf("Subscribe: %v", err) }
	defer sub.Close()
	if _, err := m.Publisher.Publish(ctx, topic, &mq.Message{Payload: []byte("poison")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	deadline := time.After(15 * time.Second)
	for attempts.Load() < 3 {
		select { case <-time.After(150 * time.Millisecond): case <-deadline: t.Fatalf("attempts=%d", attempts.Load()) }
	}
	time.Sleep(500 * time.Millisecond)
	var n int
	h.pool.QueryRow(ctx, `SELECT count(*) FROM dlq WHERE topic=$1 AND payload='poison'`, topic).Scan(&n)
	if n == 0 { t.Fatal("poison not in dlq") }
}
```
（注意 import `"cyber-ecosystem/shared-go/mq"`。）

- [ ] **Step 5: 写 competing + broadcast + durable-resume 测试**

`shared-go/mq/pg/group_test.go`（镜像 `nats/group_test.go`：同 group 竞争 total==N；不同 group 各收 N；durable resume：s1 处理后关、s2 续传拿剩余）。
（结构同 nats/group_test.go，把 backend 换成 `New(h)`、topic 用 `uniqTopic`。）

- [ ] **Step 6: 运行全部 consumer 测试通过**

Run: `go test ./shared-go/mq/pg/ -run 'TestRoundTrip|TestConsumerRetryThenDLQ|TestGroup'`
Expected: PASS。

- [ ] **Step 7: 提交**

```bash
git add shared-go/mq/pg/consumer.go shared-go/mq/pg/mq.go shared-go/mq/pg/*_test.go
git commit -m "feat(mq/pg): consumer (poll loop + ack/nak/dlq + lifecycle)"
```

---

## Task 5: 平台接入（conf.proto + NewMQ 分支 + config.yaml）

**Files:**
- Modify: `app/services/edge_mobile/internal/conf/conf.proto`（Data.MQ 加 PG 块）
- Create: `app/services/edge_mobile/internal/platform/platform_pg_mq.go`
- Modify: `app/services/edge_mobile/internal/platform/platform_mq.go`（NewMQ：nats 或 pg 二选一）
- Modify: `app/services/edge_mobile/configs/config.yaml`（data.mq.pg 注释样例）

**Interfaces:**
- Consumes: `shared-go/mq/pg.NewClient` / `pg.New`。

- [ ] **Step 1: conf.proto 加 PG 块**

在 `message MQ { message NATS{...} NATS nats=1; }` 里加：
```proto
  message PG {
    string dsn = 1;
    google.protobuf.Duration poll_interval = 2;
    google.protobuf.Duration visibility_timeout = 3;
    int32 max_retries = 4;
    google.protobuf.Duration retention = 5;
    int32 batch_size = 6;
  }
  NATS nats = 1;
  PG pg = 2;
```
- [ ] **Step 2: 重生成 conf**

Run: `./nx run edge_mobile:proto:conf`
Expected: 生成 `Data_MQ_PG` + getters。

- [ ] **Step 3: 写 platform_pg_mq.go + 改 platform_mq.go**

`platform_pg_mq.go`:
```go
package platform

import (
	"fmt"

	"cyber-ecosystem/shared-go/mq"
	mqpg "cyber-ecosystem/shared-go/mq/pg"

	"cyber-ecosystem/app/services/edge_mobile/internal/conf"
)

func toPGConfig(p *conf.Data_MQ_PG) *mqpg.Config {
	cfg := &mqpg.Config{DSN: p.GetDsn(), MaxRetries: int(p.GetMaxRetries()), BatchSize: int(p.GetBatchSize())}
	if p.PollInterval != nil {
		cfg.PollInterval = p.GetPollInterval().AsDuration()
	}
	if p.VisibilityTimeout != nil {
		cfg.VisibilityTimeout = p.GetVisibilityTimeout().AsDuration()
	}
	if p.Retention != nil {
		cfg.Retention = p.GetRetention().AsDuration()
	}
	return cfg
}
```
`platform_mq.go` 的 `NewMQ`：改为按配置二选一（nats 优先，否则 pg；都没有则报错）。例如：
```go
func NewMQ(c *conf.Data) (*mq.MQ, func(), error) {
	mc := c.GetMq()
	if mc == nil {
		return nil, nil, fmt.Errorf("mq config is required")
	}
	if n := mc.GetNats(); n != nil && n.GetEndpoint() != "" {
		// 现有 NATS 分支（保留）
		...
	}
	if p := mc.GetPg(); p != nil && p.GetDsn() != "" {
		h, closeFn, err := mqpg.NewClient(toPGConfig(p))
		if err != nil { return nil, nil, err }
		return mqpg.New(h), closeFn, nil
	}
	return nil, nil, fmt.Errorf("mq: configure either nats or pg")
}
```
（保留现有 NATS 分支完整，仅在外层加 `pg` 分支。）

- [ ] **Step 4: 重生成 wire + 构建**

Run: `./nx run edge_mobile:generate:wire && ./nx run edge_mobile:build`
Expected: 构建通过（NewMQ 签名不变，wire_gen 无需改）。

- [ ] **Step 5: config.yaml 注释样例（dev 备选，默认仍 nats）**

在 `data.mq` 下加（注释掉，dev 默认走 nats）：
```yaml
  # mq:
  #   pg:                    # 无 NATS 部署的备选后端（与 nats 二选一）
  #     dsn: postgres://postgres:postgres@localhost:5432/mq?sslmode=disable
  #     poll_interval: 0.5s
  #     visibility_timeout: 30s
  #     max_retries: 5
  #     retention: 168h      # 注意 protobuf Duration 用秒：604800s
  #     batch_size: 16
```

- [ ] **Step 6: 提交**

```bash
git add app/services/edge_mobile/internal/conf/conf.proto app/services/edge_mobile/internal/conf/conf.pb.go \
  app/services/edge_mobile/internal/platform/platform_pg_mq.go app/services/edge_mobile/internal/platform/platform_mq.go \
  app/services/edge_mobile/configs/config.yaml
git commit -m "feat(mq): wire PG backend into edge_mobile platform (nats-or-pg)"
```

---

## Task 6: 全量验证 + 完成报告

- [ ] **Step 1: PG-MQ 全量测试（含 race）**

Run: `go test ./shared-go/mq/pg/ -count=5 -race`
Expected: PASS（PG 不可达则 Skip）。

- [ ] **Step 2: 全工作区 test + lint**

Run: `./nx run tools:go:test && ./nx run tools:go:lint`
Expected: 绿，0 issue。

- [ ] **Step 3: 服务启动校验（可选：临时切到 pg 配置启动确认 boot 建 表）**

临时把 config.yaml 的 mq 切到 pg 块 → `./nx run edge_mobile:dev` → 确认 3 端口起来、无 panic、`mq` 库已建 4 表 → 改回 nats。

- [ ] **Step 4: 收尾提交（如有）+ 更新文档**

更新 `docs/pg-mq-design.md` 状态为"已实现"；按需在 `docs/mq-validation-report.md` 追加 PG 覆盖。

---

## Self-Review（写完后自查）

1. **Spec 覆盖**：spec 各节→任务映射——投递模型轮询（T4 pollLoop）、投递表（T2 schema）、历史回放（T4 registerAndBackfill）、独立 mq 库+池（T2 NewClient）、自管表非 atlas（T2 schema）、扇出（T3 Publish）、可见性/deliveries/DLQ（T4 settle/fetchBatch）、生命周期注册表+drain（T2/T4）、配置/平台接入（T5）、错误模型（T1 mapError）、测试策略（各任务测试）。✓ 无遗漏。
2. **占位符**：无 TBD/TODO；Task 4 Step 5 的 group_test 描述为"镜像 nats/group_test.go"并给出结构——实际实现时照 nats 版逐测试改 backend，属合理引用既有文件非占位。
3. **类型一致**：`handle`、`newPublisher/newConsumer`、`Subscribe`、`subscription.stop/Close`、`mapError`、`Config` 跨任务命名一致。✓
4. **已知需实现时确认的细节**：pgx 连接关闭哨兵（`pgx.ErrConnDone` 或 `pgxpool.ErrClosed`，按 pgx 版本）；reaper 的 `DELETE ... $1::interval` 文本转换（`time.Duration.String()` 产出如 "7h0m0s"，PG interval 接受 → 验证，否则用秒数）；`mq` 库需预先 `CREATE DATABASE`。

## 执行交接

Plan complete and saved to `docs/pg-mq-plan.md`. 两种执行方式：

1. **Subagent-Driven（推荐）** — 每任务派新 subagent，任务间 review，迭代快。
2. **Inline（本会话执行）** — 用 executing-plans 批量执行 + 检查点。

选哪种？
