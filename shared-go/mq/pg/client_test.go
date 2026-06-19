package pg

import (
	"context"
	"errors"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/mq"
)

var testSeq atomic.Uint64

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

// uniqTopic returns a process-unique topic and registers cleanup that wipes its
// rows (messages CASCADE their deliveries; subscribers/dlq for the topic too).
func uniqTopic(t *testing.T, base string) string {
	t.Helper()
	topic := "t-" + base + "-" + strconv.FormatUint(testSeq.Add(1), 10)
	t.Cleanup(func() {
		h, c, _ := NewClient(testConfig())
		if h == nil {
			return
		}
		defer c()
		ctx := context.Background()
		_, _ = h.pool.Exec(ctx, `DELETE FROM messages WHERE topic=$1`, topic)
		_, _ = h.pool.Exec(ctx, `DELETE FROM subscribers WHERE topic=$1`, topic)
		_, _ = h.pool.Exec(ctx, `DELETE FROM dlq WHERE topic=$1`, topic)
	})
	return topic
}

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
	_, _, err := NewClient(&Config{DSN: "postgres://postgres:postgres@127.0.0.1:1/mq?sslmode=disable&connect_timeout=2"})
	if !errors.Is(err, mq.ErrUnavailable) {
		t.Fatalf("dead dsn: got %v, want ErrUnavailable", err)
	}
}
