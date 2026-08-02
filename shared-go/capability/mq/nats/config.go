package nats

import "time"

// Config maps conf.Data.MQ.NATS to the nats.go client. Zero-value Duration/
// size/count fields fall back to the defaults below inside NewClient/New.
type Config struct {
	Endpoint       string        // nats://localhost:4222
	Creds          string        // optional NATS credentials file path; empty → no auth
	MaxAge         time.Duration // stream message TTL; 0 → 7d
	MaxBytes       int64         // stream size cap; 0 → 1 GiB
	MaxRetries     int           // delivery attempts before DLQ; 0 → 5 (see consumer.go)
	AckWait        time.Duration // redelivery wait on no-ack; 0 → 30s; must exceed handler p99
	MaxAckPending  int           // max in-flight unacked per consumer; 0 → 256; bounds memory + redelivery
	DLQMaxAge      time.Duration // DLQ stream TTL; 0 → 30d (longer than live streams)
	DLQMaxBytes    int64         // DLQ stream cap; 0 → 1 GiB
	NakBackoffStep time.Duration // linear Nak redelivery delay increment; 0 → 500ms
}

const (
	defaultMaxAge         = 7 * 24 * time.Hour
	defaultMaxBytes       = 1 << 30 // 1 GiB
	defaultMaxRetries     = 5
	defaultAckWait        = 30 * time.Second
	defaultMaxAckPending  = 256
	defaultDLQMaxAge      = 30 * 24 * time.Hour
	defaultDLQMaxBytes    = 1 << 30 // 1 GiB
	defaultNakBackoffStep = 500 * time.Millisecond
)

func maxAgeOrDefault(v time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return defaultMaxAge
}
func maxBytesOrDefault(v int64) int64 {
	if v > 0 {
		return v
	}
	return defaultMaxBytes
}
func maxRetriesOrDefault(v int) int {
	if v > 0 {
		return v
	}
	return defaultMaxRetries
}
func ackWaitOrDefault(v time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return defaultAckWait
}
func maxAckPendingOrDefault(v int) int {
	if v > 0 {
		return v
	}
	return defaultMaxAckPending
}
func dlqMaxAgeOrDefault(v time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return defaultDLQMaxAge
}
func dlqMaxBytesOrDefault(v int64) int64 {
	if v > 0 {
		return v
	}
	return defaultDLQMaxBytes
}
func nakBackoffStepOrDefault(v time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return defaultNakBackoffStep
}
