package redis

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

func TestKV(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	resetKeys(t, c, "k1", "a", "b")
	ctx := context.Background()
	kv := c.KV

	// miss → ErrCacheMiss
	if _, err := kv.Get(ctx, "missing"); !errors.Is(err, cache.ErrCacheMiss) {
		t.Fatalf("Get missing: want ErrCacheMiss, got %v", err)
	}

	// GetTTL on absent key → ErrCacheMiss (validates the -2s mapping)
	if _, err := kv.GetTTL(ctx, "missing"); !errors.Is(err, cache.ErrCacheMiss) {
		t.Fatalf("GetTTL missing: want ErrCacheMiss, got %v", err)
	}

	// Set + Get
	if err := kv.Set(ctx, "k1", []byte("v1"), 10*time.Second); err != nil {
		t.Fatal(err)
	}
	got, err := kv.Get(ctx, "k1")
	if err != nil || !bytes.Equal(got, []byte("v1")) {
		t.Fatalf("Get k1: err=%v val=%q", err, got)
	}

	// GetTTL on key with expiry → positive, ≤ set TTL
	ttl, err := kv.GetTTL(ctx, "k1")
	if err != nil || ttl <= 0 || ttl > 10*time.Second {
		t.Fatalf("GetTTL k1: err=%v ttl=%v", err, ttl)
	}

	// Exist
	if ok, _ := kv.Exist(ctx, "k1"); !ok {
		t.Fatal("Exist k1: want true")
	}

	// MSet (no expiry) + GetTTL → 0 (no expiry, validates the -1s mapping)
	if err := kv.MSet(ctx, map[string][]byte{"a": []byte("1"), "b": []byte("2")}, 0); err != nil {
		t.Fatal(err)
	}
	if ttl, err := kv.GetTTL(ctx, "a"); err != nil || ttl != 0 {
		t.Fatalf("GetTTL a (no expiry): err=%v ttl=%v", err, ttl)
	}

	// MGet incl. a missing key → nil entry
	ms, err := kv.MGet(ctx, "a", "b", "missing")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ms[0], []byte("1")) || !bytes.Equal(ms[1], []byte("2")) || ms[2] != nil {
		t.Fatalf("MGet: %v", ms)
	}

	// Del
	if err := kv.Del(ctx, "k1"); err != nil {
		t.Fatal(err)
	}
	if _, err := kv.Get(ctx, "k1"); !errors.Is(err, cache.ErrCacheMiss) {
		t.Fatalf("Get after Del: want ErrCacheMiss, got %v", err)
	}
}
