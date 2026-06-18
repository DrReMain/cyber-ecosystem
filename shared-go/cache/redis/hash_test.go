package redis

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/cache"
)

func TestHash(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	resetKeys(t, c, "h1")
	ctx := context.Background()
	h := c.Hash
	const key = "h1"

	// HMSet + HGetAll
	if err := h.HMSet(ctx, key, map[string][]byte{"a": []byte("1"), "b": []byte("2")}); err != nil {
		t.Fatal(err)
	}
	all, err := h.HGetAll(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 || !bytes.Equal(all["a"], []byte("1")) || !bytes.Equal(all["b"], []byte("2")) {
		t.Fatalf("HGetAll: %v", all)
	}

	// HSet single + HGet
	if err := h.HSet(ctx, key, "c", []byte("9")); err != nil {
		t.Fatal(err)
	}
	got, err := h.HGet(ctx, key, "c")
	if err != nil || !bytes.Equal(got, []byte("9")) {
		t.Fatalf("HGet c: err=%v val=%q", err, got)
	}

	// HGet missing field → ErrCacheMiss
	if _, err := h.HGet(ctx, key, "nope"); !errors.Is(err, cache.ErrCacheMiss) {
		t.Fatalf("HGet missing field: want ErrCacheMiss, got %v", err)
	}

	// HIncrBy on numeric field ("1" + 5 = 6)
	if v, err := h.HIncrBy(ctx, key, "a", 5); err != nil || v != 6 {
		t.Fatalf("HIncrBy a: v=%d err=%v", v, err)
	}

	// HExists
	if ok, _ := h.HExists(ctx, key, "a"); !ok {
		t.Fatal("HExists a: want true")
	}

	// HDel + HLen (a,b removed → only c remains)
	if err := h.HDel(ctx, key, "a", "b"); err != nil {
		t.Fatal(err)
	}
	if n, _ := h.HLen(ctx, key); n != 1 {
		t.Fatalf("HLen after del: %d", n)
	}
}
