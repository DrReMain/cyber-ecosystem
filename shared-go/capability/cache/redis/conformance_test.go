package redis

import (
	"context"
	"errors"
	"os"
	"sort"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/cache"
)

// Live conformance probe against a real redis-compatible backend; skipped
// unless REDIS_CONFORMANCE_ADDR is set. One run produces a drift profile: each
// subtest probes one standard-redis surface this capability targets, and a
// failure names the drift plus what breaks downstream. Drift is absorbed in
// config switches or platform instantiation — never inside this package.
//
//	REDIS_CONFORMANCE_ADDR=localhost:6379 \
//	go test ./shared-go/capability/cache/redis/ -run TestLiveConformance -count=1 -v
func TestLiveConformance(t *testing.T) {
	addr := os.Getenv("REDIS_CONFORMANCE_ADDR")
	if addr == "" {
		t.Skip("REDIS_CONFORMANCE_ADDR not set; live conformance is opt-in")
	}
	// Values pinned to config.yaml's redis block (zeros would silently resolve
	// to go-redis defaults and diverge this instantiation from the service's).
	cfg := &Config{
		Network:         "tcp",
		Addr:            addr,
		Password:        os.Getenv("REDIS_CONFORMANCE_PASSWORD"),
		DB:              0,
		PoolSize:        50,
		MinIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ReadTimeout:     200 * time.Millisecond,
		WriteTimeout:    200 * time.Millisecond,
		DialTimeout:     5 * time.Second,
		PoolTimeout:     1200 * time.Millisecond,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client, cleanup, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer cleanup()
	c := New(client)
	// unique prefix for this run so repeated runs never collide
	prefix := "conf" + time.Now().Format("150405.000000000")
	mk := func(k string) string { return prefix + ":" + k }
	defer func() {
		for _, k := range []string{"ttl", "nottl", "gd", "m1", "m2", "z", "h", "l", "lock", "rl"} {
			_ = c.KV.Del(ctx, mk(k))
		}
	}()

	t.Run("ttl-sentinels", func(t *testing.T) {
		// GetTTL's contract rests on the -2ns/-1ns sentinel encoding; drift here
		// breaks every TTL-observing caller (session refresh, cache warming).
		if err := c.KV.Set(ctx, mk("ttl"), []byte("v"), 30*time.Second); err != nil {
			t.Fatalf("Set: %v", err)
		}
		ttl, err := c.KV.GetTTL(ctx, mk("ttl"))
		if err != nil || ttl <= 0 || ttl > 30*time.Second {
			t.Errorf("ttl'd key TTL = %v, %v; want (0, 30s]", ttl, err)
		}
		if err := c.KV.Set(ctx, mk("nottl"), []byte("v"), 0); err != nil {
			t.Fatalf("Set: %v", err)
		}
		if ttl, err := c.KV.GetTTL(ctx, mk("nottl")); ttl != 0 || err != nil {
			t.Errorf("no-expiry key TTL = %v, %v; want 0, nil — sentinel drift", ttl, err)
		}
		if _, err := c.KV.GetTTL(ctx, mk("absent")); !errors.Is(err, cache.ErrCacheMiss) {
			t.Errorf("absent key TTL err = %v, want ErrCacheMiss", err)
		}
	})

	t.Run("expire-on-missing-noerror", func(t *testing.T) {
		// EXPIRE on a missing key reports false, not an error; an error here
		// turns every benign TTL slide into a failed request.
		if err := c.KV.Expire(ctx, mk("absent"), time.Minute); err != nil {
			t.Errorf("Expire(missing): %v — want nil (benign no-op)", err)
		}
	})

	t.Run("getdel-atomic", func(t *testing.T) {
		// GetDel must return the value and delete in one step; drift (value
		// without delete, or delete without value) breaks one-shot reads.
		if err := c.KV.Set(ctx, mk("gd"), []byte("once"), time.Minute); err != nil {
			t.Fatalf("Set: %v", err)
		}
		v, err := c.KV.GetDel(ctx, mk("gd"))
		if err != nil || string(v) != "once" {
			t.Fatalf("GetDel = %q, %v; want once", v, err)
		}
		if _, err := c.KV.Get(ctx, mk("gd")); !errors.Is(err, cache.ErrCacheMiss) {
			t.Errorf("post-GetDel Get err = %v, want ErrCacheMiss — delete half drifted", err)
		}
	})

	t.Run("mget-nil-slots", func(t *testing.T) {
		// Missing entries decode to nil slots, not errors; an error (or dropped
		// slot) breaks positional multi-key reads.
		if err := c.KV.Set(ctx, mk("m1"), []byte("a"), time.Minute); err != nil {
			t.Fatalf("Set: %v", err)
		}
		vals, err := c.KV.MGet(ctx, mk("m1"), mk("m2"))
		if err != nil {
			t.Fatalf("MGet: %v", err)
		}
		if len(vals) != 2 || string(vals[0]) != "a" || vals[1] != nil {
			t.Errorf("MGet = %q; want [a nil] — missing-slot drift", vals)
		}
	})

	t.Run("scan-session-namespace", func(t *testing.T) {
		// The whole Session surface rides on SCAN pattern matching over the
		// "session:{id}:{key}" namespace; drift breaks Keys/Refresh/Destroy
		// (isolation included — Destroy must not touch other sessions).
		sid := "conf-" + prefix
		other := "conf-other-" + prefix
		for i, f := range []string{"f1", "f2", "f3"} {
			if err := c.Session.Set(ctx, sid, f, []byte{byte('0' + i)}, 30*time.Second); err != nil {
				t.Fatalf("Session.Set: %v", err)
			}
		}
		if err := c.Session.Set(ctx, other, "f1", []byte("x"), 30*time.Second); err != nil {
			t.Fatalf("Session.Set(other): %v", err)
		}
		keys, err := c.Session.Keys(ctx, sid)
		if err != nil {
			t.Fatalf("Keys: %v", err)
		}
		sort.Strings(keys)
		want := []string{"f1", "f2", "f3"}
		if len(keys) != 3 {
			t.Fatalf("Keys = %v, want %v — scan leaked the other session's keyspace", keys, want)
		}
		for i := range want {
			if keys[i] != want[i] {
				t.Errorf("Keys[%d] = %q, want %q", i, keys[i], want[i])
			}
		}
		if err := c.Session.Refresh(ctx, sid, 90*time.Second); err != nil {
			t.Fatalf("Refresh: %v", err)
		}
		ttl, err := c.KV.GetTTL(ctx, sessionKey(sid, "f1"))
		if err != nil || ttl <= 30*time.Second {
			t.Errorf("post-Refresh TTL = %v, %v; want extended past 30s — pipeline drift", ttl, err)
		}
		if err := c.Session.Destroy(ctx, sid); err != nil {
			t.Fatalf("Destroy: %v", err)
		}
		if ok, err := c.Session.Exists(ctx, sid); ok || err != nil {
			t.Errorf("Exists after Destroy = %v, %v; want false", ok, err)
		}
		if ok, err := c.Session.Exists(ctx, other); !ok || err != nil {
			t.Errorf("other session destroyed too = %v, %v — isolation drift", ok, err)
		}
		_ = c.Session.Destroy(ctx, other)
	})

	t.Run("zset-float-roundtrip", func(t *testing.T) {
		// Scores survive a float64→string→float64 round-trip exactly, and
		// RangeByScore bounds are inclusive; drift breaks leaderboards and
		// delayed queues (score is the schedule).
		x, y, z := -1234567.875, 0.1, 3.141592653589793
		if err := c.SortedSet.Add(ctx, mk("z"),
			cache.Member{Score: x, Member: "x"},
			cache.Member{Score: y, Member: "y"},
			cache.Member{Score: z, Member: "z"}); err != nil {
			t.Fatalf("Add: %v", err)
		}
		for m, want := range map[string]float64{"x": x, "y": y, "z": z} {
			got, err := c.SortedSet.Score(ctx, mk("z"), m)
			if err != nil || got != want {
				t.Errorf("Score(%s) = %v, %v; want %v — float precision drift", m, got, err, want)
			}
		}
		got, err := c.SortedSet.RangeByScore(ctx, mk("z"), y, z, 0, -1)
		if err != nil || len(got) != 2 || got[0].Member != "y" || got[1].Member != "z" {
			t.Errorf("RangeByScore(inclusive) = %+v, %v; want [y z] — bound drift", got, err)
		}
		if r, err := c.SortedSet.Rank(ctx, mk("z"), "z"); err != nil || r != 2 {
			t.Errorf("Rank(z) = %d, %v; want 2 (score-ascending)", r, err)
		}
	})

	t.Run("hash-incrby", func(t *testing.T) {
		// HIncrBy on a fresh field returns the delta and accumulates; drift
		// breaks multi-dim counters.
		if v, err := c.Hash.HIncrBy(ctx, mk("h"), "n", 5); err != nil || v != 5 {
			t.Errorf("HIncrBy(fresh) = %d, %v; want 5", v, err)
		}
		if v, err := c.Hash.HIncrBy(ctx, mk("h"), "n", -2); err != nil || v != 3 {
			t.Errorf("HIncrBy(accumulate) = %d, %v; want 3", v, err)
		}
		if n, err := c.Hash.HLen(ctx, mk("h")); err != nil || n != 1 {
			t.Errorf("HLen = %d, %v; want 1", n, err)
		}
		if err := c.Hash.HDel(ctx, mk("h"), "n"); err != nil {
			t.Errorf("HDel: %v", err)
		}
		if n, err := c.Hash.HLen(ctx, mk("h")); err != nil || n != 0 {
			t.Errorf("post-HDel HLen = %d, %v; want 0", n, err)
		}
	})

	t.Run("list-trim-range", func(t *testing.T) {
		// LRange negative indices and LTrim retention shape the recent-N
		// window; drift breaks bounded activity feeds.
		if _, err := c.List.RPush(ctx, mk("l"), []byte("a"), []byte("b"), []byte("c"), []byte("d")); err != nil {
			t.Fatalf("RPush: %v", err)
		}
		vals, err := c.List.LRange(ctx, mk("l"), 0, -1)
		if err != nil || len(vals) != 4 || string(vals[0]) != "a" {
			t.Fatalf("LRange(0,-1) = %q, %v — ordering drift", vals, err)
		}
		if err := c.List.LTrim(ctx, mk("l"), 1, 2); err != nil {
			t.Fatalf("LTrim: %v", err)
		}
		vals, err = c.List.LRange(ctx, mk("l"), 0, -1)
		if err != nil || len(vals) != 2 || string(vals[0]) != "b" || string(vals[1]) != "c" {
			t.Errorf("post-LTrim = %q, %v; want [b c] — trim drift", vals, err)
		}
		if v, err := c.List.LPop(ctx, mk("l")); err != nil || string(v) != "b" {
			t.Errorf("LPop = %q, %v; want b", v, err)
		}
	})

	t.Run("lock-lifecycle", func(t *testing.T) {
		// SET NX ownership + lua release/refresh; drift breaks the distributed
		// mutex (double-ownership or lost locks).
		rel, err := c.Lock.TryLock(ctx, mk("lock"), 30*time.Second)
		if err != nil {
			t.Fatalf("TryLock: %v", err)
		}
		if _, err := c.Lock.TryLock(ctx, mk("lock"), 30*time.Second); !errors.Is(err, cache.ErrLockNotAcquired) {
			t.Errorf("second TryLock = %v, want ErrLockNotAcquired — mutual exclusion drift", err)
		}
		if err := rel.Unlock(ctx); err != nil {
			t.Errorf("Unlock: %v", err)
		}
		rel2, err := c.Lock.TryLock(ctx, mk("lock"), 30*time.Second)
		if err != nil {
			t.Fatalf("re-acquire after Unlock: %v — release drift", err)
		}
		if err := rel2.Unlock(ctx); err != nil {
			t.Errorf("Unlock(2): %v", err)
		}
		if err := rel2.Extend(ctx, 30*time.Second); !errors.Is(err, cache.ErrLockNotAcquired) {
			t.Errorf("Extend after release = %v, want ErrLockNotAcquired", err)
		}
	})

	t.Run("ratelimiter-gcra", func(t *testing.T) {
		// redis_rate's GCRA lua; drift in the script result shape breaks the
		// Allow/RetryAfter contract (burst = limit, refill = limit/window).
		limit, window := int64(2), time.Second
		for i := int64(1); i <= limit; i++ {
			res, err := c.RateLimiter.Allow(ctx, mk("rl"), limit, window)
			if err != nil || !res.Allowed || res.Remaining != limit-i {
				t.Fatalf("Allow #%d = %+v, %v; want allowed, remaining %d", i, res, err, limit-i)
			}
		}
		res, err := c.RateLimiter.Allow(ctx, mk("rl"), limit, window)
		if err != nil {
			t.Fatalf("Allow over: %v", err)
		}
		if res.Allowed || res.RetryAfter <= 0 || res.RetryAfter > window {
			t.Errorf("denied Allow = %+v; want denied with RetryAfter in (0, %v]", res, window)
		}
	})

	t.Run("pubsub-pattern-delivery", func(t *testing.T) {
		// Exact and pattern subscriptions deliver with correct Channel/Pattern
		// tagging; drift breaks real-time fan-out (centrifugo-side events).
		exact := mk("pub/exact")
		patTopic := mk("pub/pat/a")
		sub, err := c.PubSub.Subscribe(ctx, exact)
		if err != nil {
			t.Fatalf("Subscribe: %v", err)
		}
		psub, err := c.PubSub.PSubscribe(ctx, mk("pub/pat/")+"*")
		if err != nil {
			t.Fatalf("PSubscribe: %v", err)
		}
		defer func() { _ = psub.Close() }()
		time.Sleep(150 * time.Millisecond) // let both subscriptions register server-side
		if err := c.PubSub.Publish(ctx, exact, []byte("e")); err != nil {
			t.Fatalf("Publish(exact): %v", err)
		}
		if err := c.PubSub.Publish(ctx, patTopic, []byte("p")); err != nil {
			t.Fatalf("Publish(pattern): %v", err)
		}
		recv := func(s cache.Subscription) cache.Message {
			t.Helper()
			select {
			case m := <-s.Channel():
				return m
			case <-time.After(2 * time.Second):
				t.Fatal("no delivery within 2s — subscription/pattern drift")
				return cache.Message{}
			}
		}
		if m := recv(sub); m.Channel != exact || string(m.Payload) != "e" || m.Pattern != "" {
			t.Errorf("exact delivery = %+v — tagging drift", m)
		}
		if m := recv(psub); m.Channel != patTopic || string(m.Payload) != "p" || m.Pattern == "" {
			t.Errorf("pattern delivery = %+v — Pattern tagging drift", m)
		}
		if err := sub.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
		if err := sub.Close(); err != nil {
			t.Errorf("Close idempotency: %v", err)
		}
		if _, open := <-sub.Channel(); open {
			t.Error("delivery channel not closed after Close — goroutine/channel leak")
		}
	})
}
