package redis

import (
	"bytes"
	"context"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/cache"
)

// TestCounterConcurrentAtomicity: N goroutines each Incr(1) → final value == N.
// Validates redis INCR atomicity under real concurrency (run with -race).
func TestCounterConcurrentAtomicity(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	const N = 100
	resetKeys(t, c, "stress:counter")

	var wg sync.WaitGroup
	for range N {
		wg.Go(func() {
			_, _ = c.Counter.Incr(ctx, "stress:counter", 1)
		})
	}
	wg.Wait()
	got, err := c.Counter.Get(ctx, "stress:counter")
	if err != nil || got != N {
		t.Fatalf("concurrent Incr: got %d err=%v, want %d (atomicity broken)", got, err, N)
	}
}

// TestLockConcurrentMutualExclusion: N goroutines TryLock the same key → exactly 1 acquires.
func TestLockConcurrentMutualExclusion(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	const N = 20
	resetKeys(t, c, "stress:lock")

	var acquired atomic.Int32
	var winner cache.Release
	var wmu sync.Mutex
	var wg sync.WaitGroup
	for range N {
		wg.Go(func() {
			rel, err := c.Lock.TryLock(ctx, "stress:lock", 10*time.Second)
			if err == nil {
				acquired.Add(1)
				wmu.Lock()
				winner = rel
				wmu.Unlock()
			} else if !errors.Is(err, cache.ErrLockNotAcquired) {
				t.Errorf("TryLock unexpected err: %v", err)
			}
		})
	}
	wg.Wait()
	if got := acquired.Load(); got != 1 {
		t.Fatalf("mutual exclusion broken: %d acquirers, want exactly 1", got)
	}
	if winner != nil {
		_ = winner.Unlock(ctx)
	}
}

// TestKVConcurrentDistinctKeys: N goroutines Set/Get distinct keys concurrently (race detector).
func TestKVConcurrentDistinctKeys(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	const N = 50
	keys := make([]string, N)
	for i := range N {
		keys[i] = "stress:kv:" + itoa(i)
	}
	resetKeys(t, c, keys...)

	var wg sync.WaitGroup
	for i := range N {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			val := []byte("v" + itoa(i))
			if err := c.KV.Set(ctx, keys[i], val, time.Minute); err != nil {
				t.Errorf("Set %d: %v", i, err)
			}
			got, err := c.KV.Get(ctx, keys[i])
			if err != nil || !bytes.Equal(got, val) {
				t.Errorf("Get %d: got %q err=%v", i, got, err)
			}
		}(i)
	}
	wg.Wait()
}

// TestKVBigValue: a ~1MB value round-trips intact.
func TestKVBigValue(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	resetKeys(t, c, "stress:big")
	big := bytes.Repeat([]byte("X"), 1<<20) // 1 MiB
	if err := c.KV.Set(ctx, "stress:big", big, time.Minute); err != nil {
		t.Fatal(err)
	}
	got, err := c.KV.Get(ctx, "stress:big")
	if err != nil || !bytes.Equal(got, big) {
		t.Fatalf("big value round-trip: len_got=%d len_want=%d err=%v", len(got), len(big), err)
	}
}

// TestKVBinary: values with null bytes / 0xFF / control chars round-trip intact.
func TestKVBinary(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	resetKeys(t, c, "stress:bin")
	bin := []byte{0x00, 0x01, 0xFF, 0xFE, '\n', '\r', 0x7F}
	if err := c.KV.Set(ctx, "stress:bin", bin, time.Minute); err != nil {
		t.Fatal(err)
	}
	got, err := c.KV.Get(ctx, "stress:bin")
	if err != nil || !bytes.Equal(got, bin) {
		t.Fatalf("binary round-trip: got %v want %v err=%v", got, bin, err)
	}
}

// TestKVSpecialCharKey: keys with unicode / spaces / punctuation (non-forbidden) work.
func TestKVSpecialCharKey(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	keys := []string{"用户:资料", "key with space", "key/with/slash", "key.with.dot", "键值_1"}
	resetKeys(t, c, keys...)
	for _, k := range keys {
		if err := c.KV.Set(ctx, k, []byte("ok"), time.Minute); err != nil {
			t.Fatalf("Set %q: %v", k, err)
		}
		got, err := c.KV.Get(ctx, k)
		if err != nil || string(got) != "ok" {
			t.Fatalf("Get %q: got %q err=%v", k, got, err)
		}
	}
}

// TestSortedSetLarge: 500 members round-trip with correct cardinality + range length.
func TestSortedSetLarge(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	resetKeys(t, c, "stress:zset")
	const N = 500
	members := make([]cache.Member, N)
	for i := range N {
		members[i] = cache.Member{Score: float64(i), Member: itoa(i)}
	}
	if err := c.SortedSet.Add(ctx, "stress:zset", members...); err != nil {
		t.Fatal(err)
	}
	if n, _ := c.SortedSet.Card(ctx, "stress:zset"); n != N {
		t.Fatalf("Card: %d, want %d", n, N)
	}
	all, err := c.SortedSet.Range(ctx, "stress:zset", 0, -1)
	if err != nil || len(all) != N {
		t.Fatalf("Range all: len=%d err=%v, want %d", len(all), err, N)
	}
}

// TestRangeByScoreNaNInf: NaN/Inf bounds are rejected at the boundary (not pushed to redis).
func TestRangeByScoreNaNInf(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	resetKeys(t, c, "stress:zset2")
	cases := []struct{ min, max float64 }{
		{math.NaN(), 5}, {0, math.NaN()}, {math.Inf(1), 5}, {0, math.Inf(-1)},
	}
	for _, tc := range cases {
		if _, err := c.SortedSet.RangeByScore(ctx, "stress:zset2", tc.min, tc.max, 0, 10); !errors.Is(err, cache.ErrInvalidArgument) {
			t.Fatalf("RangeByScore(%v,%v): want ErrInvalidArgument, got %v", tc.min, tc.max, err)
		}
	}
}

// TestCacheRedisDown: a dead redis addr surfaces as ErrUnavailable (boot ping path).
func TestCacheRedisDown(t *testing.T) {
	_, _, err := NewClient(&Config{Addr: "127.0.0.1:39998", ReadTimeout: time.Second, WriteTimeout: time.Second})
	if err == nil {
		t.Fatal("dead addr: want error, got nil")
	}
	if !errors.Is(err, cache.ErrUnavailable) {
		t.Fatalf("dead addr: want ErrUnavailable wrap, got %v", err)
	}
}

// itoa avoids strconv import in this file.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}
