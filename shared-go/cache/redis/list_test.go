package redis

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/cache"
)

func TestList(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	resetKeys(t, c, "l1")
	ctx := context.Background()
	l := c.List
	const key = "l1"

	// RPush returns new length
	if n, err := l.RPush(ctx, key, []byte("a"), []byte("b"), []byte("c")); err != nil || n != 3 {
		t.Fatalf("RPush: n=%d err=%v", n, err)
	}
	if n, _ := l.LLen(ctx, key); n != 3 {
		t.Fatalf("LLen: %d", n)
	}

	// LTrim keeps window [0,1] → [a,b]
	if err := l.LTrim(ctx, key, 0, 1); err != nil {
		t.Fatal(err)
	}
	rng, err := l.LRange(ctx, key, 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(rng) != 2 || !bytes.Equal(rng[0], []byte("a")) || !bytes.Equal(rng[1], []byte("b")) {
		t.Fatalf("LRange after trim: %v", rng)
	}

	// LPop → a, RPop → b
	if v, err := l.LPop(ctx, key); err != nil || !bytes.Equal(v, []byte("a")) {
		t.Fatalf("LPop: v=%q err=%v", v, err)
	}
	if v, err := l.RPop(ctx, key); err != nil || !bytes.Equal(v, []byte("b")) {
		t.Fatalf("RPop: v=%q err=%v", v, err)
	}

	// empty now → LPop → ErrCacheMiss
	if _, err := l.LPop(ctx, key); !errors.Is(err, cache.ErrCacheMiss) {
		t.Fatalf("LPop empty: want ErrCacheMiss, got %v", err)
	}
}
