package redis

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/cache"
)

// Argument guards must reject invalid input before any network call, so these
// run against impls constructed with a nil client: a guard miss would
// nil-deref, a guard hit returns the sentinel.
func TestArgumentGuards(t *testing.T) {
	ctx := context.Background()
	oversize := strings.Repeat("k", 1025)
	isInvalid := func(name string, err error) {
		t.Helper()
		if !errors.Is(err, cache.ErrInvalidArgument) {
			t.Errorf("%s: %v, want ErrInvalidArgument", name, err)
		}
	}

	k := newKV(nil)
	isInvalid("kv.Get(empty)", func() error { _, err := k.Get(ctx, ""); return err }())
	isInvalid("kv.Set(oversize)", k.Set(ctx, oversize, nil, 0))
	isInvalid("kv.MGet(one bad)", func() error { _, err := k.MGet(ctx, "ok", ""); return err }())

	h := newHash(nil)
	isInvalid("hash.HGet(empty)", func() error { _, err := h.HGet(ctx, "", "f"); return err }())
	if err := h.HMSet(ctx, "k", nil); err != nil {
		t.Errorf("HMSet(empty fields) should be a no-op: %v", err)
	}
	if err := h.HDel(ctx, "k"); err != nil {
		t.Errorf("HDel(no fields) should be a no-op: %v", err)
	}

	l := newList(nil)
	isInvalid("list.LPop(empty)", func() error { _, err := l.LPop(ctx, ""); return err }())
	if n, err := l.LPush(ctx, "k"); n != 0 || err != nil {
		t.Errorf("LPush(no vals) = %d, %v; want 0, nil", n, err)
	}

	s := newSet(nil)
	isInvalid("set.SAdd(empty)", func() error { _, err := s.SAdd(ctx, ""); return err }())
	if m, err := s.SInter(ctx); m != nil || err != nil {
		t.Errorf("SInter(no keys) = %v, %v; want nil, nil", m, err)
	}
	if m, err := s.SUnion(ctx); m != nil || err != nil {
		t.Errorf("SUnion(no keys) = %v, %v; want nil, nil", m, err)
	}

	z := newSortedSet(nil)
	isInvalid("zset.Add(empty)", z.Add(ctx, ""))
	if err := z.Add(ctx, "k"); err != nil {
		t.Errorf("Add(no members) should be a no-op: %v", err)
	}
	// NaN/Inf produce strings redis rejects at runtime — fail fast instead.
	_, err := z.RangeByScore(ctx, "k", math.NaN(), 1, 0, -1)
	isInvalid("zset.RangeByScore(NaN)", err)
	_, err = z.RangeByScore(ctx, "k", 0, math.Inf(1), 0, -1)
	isInvalid("zset.RangeByScore(+Inf)", err)

	c := newCounter(nil)
	isInvalid("counter.Incr(empty)", func() error { _, err := c.Incr(ctx, "", 1); return err }())

	sess := newSession(nil)
	isInvalid("session.Get(injection id)", func() error { _, err := sess.Get(ctx, "a*b", "k"); return err }())
	isInvalid("session.Get(key with ':')", func() error { _, err := sess.Get(ctx, "sid", "a:b"); return err }())

	p := newPubSub(nil)
	isInvalid("pubsub.Publish(empty channel)", p.Publish(ctx, "", nil))
	_, err = p.Subscribe(ctx)
	isInvalid("pubsub.Subscribe(no channels)", err)

	rl := newRateLimiter(nil)
	_, err = rl.Allow(ctx, "k", 1, 0)
	isInvalid("ratelimiter.Allow(window<=0)", err)
	// A degenerate limit denies everything rather than feeding the GCRA lua.
	res, err := rl.Allow(ctx, "k", 0, time.Second)
	if err != nil || res == nil || res.Allowed {
		t.Errorf("Allow(limit=0) = %+v, %v; want denied, nil error", res, err)
	}

	lk := newLock(nil)
	_, err = lk.TryLock(ctx, "", time.Second)
	isInvalid("lock.TryLock(empty key)", err)
}
