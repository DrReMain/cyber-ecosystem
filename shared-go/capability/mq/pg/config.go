package pg

import "time"

// Config maps conf.Data.MQ.PG to the pgx client. Zero fields fall back to defaults
// inside NewClient.
type Config struct {
	DSN               string        // postgres://user:pass@host:5432/mq?sslmode=disable
	PollInterval      time.Duration // consume poll interval; 0 → 500ms
	VisibilityTimeout time.Duration // invisible-after-fetch duration (= NATS AckWait); 0 → 30s, must exceed handler p99
	MaxRetries        int           // delivery cap, exceeded → DLQ; 0 → 5
	Retention         time.Duration // messages retention (= NATS MaxAge), age-evicted on expiry; 0 → 7d
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
