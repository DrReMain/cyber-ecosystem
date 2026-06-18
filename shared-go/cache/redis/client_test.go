package redis

import (
	"context"
	"os"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

// testConfig returns a Config pointing at the test redis (REDIS_ADDR env, or
// localhost:6379). Tests t.Skip when NewClient cannot connect, so the suite is
// safe to run without a live redis.
func testConfig() *Config {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return &Config{Network: "tcp", Addr: addr, ReadTimeout: time.Second, WriteTimeout: time.Second}
}

// newTestCache builds a wired *cache.Cache against the test redis, skipping if
// redis is unavailable. Returned cleanup closes the client.
func newTestCache(t *testing.T) (*cache.Cache, func()) {
	t.Helper()
	client, cleanup, err := NewClient(testConfig())
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	return New(client), cleanup
}

// resetKeys clears the given keys immediately (removes pollution from prior
// runs) and again on test completion, keeping the shared redis hermetic across
// repeated suite runs.
func resetKeys(t *testing.T, c *cache.Cache, keys ...string) {
	t.Helper()
	ctx := context.Background()
	del := func() {
		for _, k := range keys {
			_ = c.KV.Del(ctx, k)
		}
	}
	del()
	t.Cleanup(del)
}

func TestNewClientPing(t *testing.T) {
	client, cleanup, err := NewClient(testConfig())
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	defer cleanup()
	if client == nil {
		t.Fatal("nil client")
	}
}
