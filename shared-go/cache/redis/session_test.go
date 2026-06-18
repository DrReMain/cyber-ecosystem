package redis

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

func TestSession(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	s := c.Session
	const sid = "s1"
	// Prefix-pollution guard: SCAN-based destroy clears any prior session:s1:*
	// at start and on cleanup (resetKeys can't do patterns).
	_ = s.Destroy(ctx, sid)
	t.Cleanup(func() { _ = s.Destroy(ctx, sid) })

	// Set two keys
	if err := s.Set(ctx, sid, "a", []byte("1"), time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(ctx, sid, "b", []byte("2"), time.Minute); err != nil {
		t.Fatal(err)
	}

	// Keys contains a, b
	keys, err := s.Keys(ctx, sid)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(keys, "a") || !slices.Contains(keys, "b") {
		t.Fatalf("Keys: %v", keys)
	}

	// Get
	if v, err := s.Get(ctx, sid, "a"); err != nil || !bytes.Equal(v, []byte("1")) {
		t.Fatalf("Get a: err=%v val=%q", err, v)
	}

	// Exists
	if ok, _ := s.Exists(ctx, sid); !ok {
		t.Fatal("Exists: want true")
	}

	// Del one
	if err := s.Del(ctx, sid, "a"); err != nil {
		t.Fatal(err)
	}

	// Refresh (session still exists via b)
	if err := s.Refresh(ctx, sid, 2*time.Minute); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	// Destroy → Keys empty
	if err := s.Destroy(ctx, sid); err != nil {
		t.Fatal(err)
	}
	if keys, _ := s.Keys(ctx, sid); len(keys) != 0 {
		t.Fatalf("Keys after destroy: %v", keys)
	}

	// Refresh on destroyed session → ErrSessionNotFound
	if err := s.Refresh(ctx, sid, time.Minute); !errors.Is(err, cache.ErrSessionNotFound) {
		t.Fatalf("Refresh after destroy: want ErrSessionNotFound, got %v", err)
	}
}

// TestSessionInvalidID pins the interface-boundary isolation invariant: ':' in
// id/key (cross-namespace collision) and glob metacharacters in id (SCAN
// injection) are rejected for EVERY backend.
func TestSessionInvalidID(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	s := c.Session

	if err := s.Set(ctx, "a:b", "k", []byte("x"), time.Minute); !errors.Is(err, cache.ErrInvalidArgument) {
		t.Fatalf("Set ':' in id: want ErrInvalidArgument, got %v", err)
	}
	if err := s.Set(ctx, "sid", "k:v", []byte("x"), time.Minute); !errors.Is(err, cache.ErrInvalidArgument) {
		t.Fatalf("Set ':' in key: want ErrInvalidArgument, got %v", err)
	}
	if err := s.Destroy(ctx, "*"); !errors.Is(err, cache.ErrInvalidArgument) {
		t.Fatalf("Destroy glob id: want ErrInvalidArgument, got %v", err)
	}
}
