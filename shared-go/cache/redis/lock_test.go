package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

func TestLock(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	lk := c.Lock
	resetKeys(t, c, "lock1")

	// TryLock acquires
	r1, err := lk.TryLock(ctx, "lock1", 10*time.Second)
	if err != nil {
		t.Fatalf("TryLock 1: %v", err)
	}
	// Second TryLock on held key → ErrLockNotAcquired
	if _, err := lk.TryLock(ctx, "lock1", 10*time.Second); !errors.Is(err, cache.ErrLockNotAcquired) {
		t.Fatalf("TryLock 2: want ErrLockNotAcquired, got %v", err)
	}
	// Extend the held lock
	if err := r1.Extend(ctx, 20*time.Second); err != nil {
		t.Fatalf("Extend: %v", err)
	}
	// Release then reacquire
	if err := r1.Unlock(ctx); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	r3, err := lk.TryLock(ctx, "lock1", 10*time.Second)
	if err != nil {
		t.Fatalf("TryLock after release: %v", err)
	}

	// Blocking Lock waits for a concurrent release, then acquires
	go func() {
		time.Sleep(80 * time.Millisecond)
		_ = r3.Unlock(ctx)
	}()
	ctx2, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	r4, err := lk.Lock(ctx2, "lock1", 10*time.Second)
	if err != nil {
		t.Fatalf("Lock blocking: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Fatalf("Lock returned too fast (%v); didn't block on held lock", elapsed)
	}
	_ = r4.Unlock(ctx)
}

// TestLockExtendExpired pins the cache.Release contract: Extend on an
// already-expired lock returns ErrLockNotAcquired (redislock returns
// ErrNotObtained here, not ErrLockNotHeld — both must map to the contract).
func TestLockExtendExpired(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	lk := c.Lock
	resetKeys(t, c, "lockexp")

	r, err := lk.TryLock(ctx, "lockexp", 200*time.Millisecond)
	if err != nil {
		t.Fatalf("TryLock: %v", err)
	}
	time.Sleep(350 * time.Millisecond) // let the lock expire
	if err := r.Extend(ctx, time.Second); !errors.Is(err, cache.ErrLockNotAcquired) {
		t.Fatalf("Extend after expiry: want ErrLockNotAcquired, got %v", err)
	}
}
