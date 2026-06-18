package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

func TestCounter(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	resetKeys(t, c, "ct1")
	ctx := context.Background()
	ct := c.Counter
	const key = "ct1"

	// Get missing → ErrKeyNotFound
	if _, err := ct.Get(ctx, key); !errors.Is(err, cache.ErrKeyNotFound) {
		t.Fatalf("Get missing: want ErrKeyNotFound, got %v", err)
	}

	// Incr accumulates
	if v, err := ct.Incr(ctx, key, 5); err != nil || v != 5 {
		t.Fatalf("Incr 5: v=%d err=%v", v, err)
	}
	if v, err := ct.Incr(ctx, key, 3); err != nil || v != 8 {
		t.Fatalf("Incr 3: v=%d err=%v", v, err)
	}
	if v, _ := ct.Get(ctx, key); v != 8 {
		t.Fatalf("Get: %d", v)
	}

	// Decr
	if v, err := ct.Decr(ctx, key, 2); err != nil || v != 6 {
		t.Fatalf("Decr 2: v=%d err=%v", v, err)
	}

	// Set overwrites
	if err := ct.Set(ctx, key, 100); err != nil {
		t.Fatal(err)
	}
	if v, _ := ct.Get(ctx, key); v != 100 {
		t.Fatalf("Get after Set: %d", v)
	}

	// Expire sets TTL (>0); Expire 0 removes it
	if err := ct.Expire(ctx, key, 10*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := ct.Expire(ctx, key, 0); err != nil {
		t.Fatal(err)
	}
	if v, _ := ct.Get(ctx, key); v != 100 { // still present after Persist
		t.Fatalf("Get after Persist: %d", v)
	}
}
