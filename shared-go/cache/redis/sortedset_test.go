package redis

import (
	"context"
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/cache"
)

func TestSortedSet(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	resetKeys(t, c, "z1")
	ctx := context.Background()
	z := c.SortedSet
	const key = "z1"

	// Add {a:1, b:3, c:2}
	if err := z.Add(ctx, key,
		cache.Member{Score: 1, Member: "a"},
		cache.Member{Score: 3, Member: "b"},
		cache.Member{Score: 2, Member: "c"},
	); err != nil {
		t.Fatal(err)
	}
	if n, _ := z.Card(ctx, key); n != 3 {
		t.Fatalf("Card: %d", n)
	}

	// RevRange top 2 → [b(3), c(2)]
	top, err := z.RevRange(ctx, key, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(top) != 2 || top[0].Member != "b" || top[0].Score != 3 || top[1].Member != "c" {
		t.Fatalf("RevRange: %+v", top)
	}

	// Score + Rank (ascending: a=1 lowest → rank 0)
	if s, _ := z.Score(ctx, key, "b"); s != 3 {
		t.Fatalf("Score b: %v", s)
	}
	if r, _ := z.Rank(ctx, key, "a"); r != 0 {
		t.Fatalf("Rank a: %d", r)
	}

	// Score on missing member → ErrKeyNotFound
	if _, err := z.Score(ctx, key, "nope"); !errors.Is(err, cache.ErrKeyNotFound) {
		t.Fatalf("Score missing: want ErrKeyNotFound, got %v", err)
	}

	// IncrBy a +10 → 11
	if v, err := z.IncrBy(ctx, key, "a", 10); err != nil || v != 11 {
		t.Fatalf("IncrBy a: v=%v err=%v", v, err)
	}

	// RangeByScore [0,5] → b(3),c(2); a(11) excluded
	inRange, err := z.RangeByScore(ctx, key, 0, 5, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(inRange) != 2 {
		t.Fatalf("RangeByScore [0,5]: %+v", inRange)
	}

	// Remove b → Card 2
	if err := z.Remove(ctx, key, "b"); err != nil {
		t.Fatal(err)
	}
	if n, _ := z.Card(ctx, key); n != 2 {
		t.Fatalf("Card after remove: %d", n)
	}
}
