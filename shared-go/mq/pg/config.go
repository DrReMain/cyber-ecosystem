package pg

import "time"

// Config 映射 conf.Data.MQ.PG 到 pgx 客户端。零值字段在 NewClient 内回退到默认。
type Config struct {
	DSN               string        // postgres://user:pass@host:5432/mq?sslmode=disable
	PollInterval      time.Duration // 消费轮询间隔；0 → 500ms
	VisibilityTimeout time.Duration // 取出后不可见时长（=NATS AckWait）；0 → 30s，须大于 handler p99
	MaxRetries        int           // 投递上限，超过进 DLQ；0 → 5
	Retention         time.Duration // messages 保留时长（=NATS MaxAge），到期清理；0 → 7d
	// BatchSize bounds how many deliveries one poll iteration locks+processes. Since
	// the poll loop is a single goroutine that processes a batch serially before
	// fetching the next, this is also the per-subscription in-flight cap — the PG
	// analogue of NATS MaxAckPending (there is no separate unbounded accumulation).
	BatchSize int // 0 → 16
}

const (
	defaultPollInterval = 500 * time.Millisecond
	defaultVisibility   = 30 * time.Second
	defaultMaxRetries   = 5
	defaultRetention    = 7 * 24 * time.Hour
	defaultBatchSize    = 16
)

func pollIntervalOrDefault(v time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return defaultPollInterval
}
func visibilityOrDefault(v time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return defaultVisibility
}
func maxRetriesOrDefault(v int) int {
	if v > 0 {
		return v
	}
	return defaultMaxRetries
}
func retentionOrDefault(v time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return defaultRetention
}
func batchSizeOrDefault(v int) int {
	if v > 0 {
		return v
	}
	return defaultBatchSize
}
