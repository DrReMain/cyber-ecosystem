package redis

import (
	"context"
	"testing"
)

func TestSet(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	resetKeys(t, c, "key1", "key2")
	ctx := context.Background()
	s := c.Set

	// key1={a,b}, key2={b,c}
	if n, err := s.SAdd(ctx, "key1", "a", "b"); err != nil || n != 2 {
		t.Fatalf("SAdd key1: n=%d err=%v", n, err)
	}
	if n, err := s.SAdd(ctx, "key2", "b", "c"); err != nil || n != 2 {
		t.Fatalf("SAdd key2: n=%d err=%v", n, err)
	}

	// SInter(key1,key2) = {b}
	inter, err := s.SInter(ctx, "key1", "key2")
	if err != nil {
		t.Fatal(err)
	}
	if len(inter) != 1 || inter[0] != "b" {
		t.Fatalf("SInter: %v", inter)
	}

	// SUnion(key1,key2) = {a,b,c}
	union, err := s.SUnion(ctx, "key1", "key2")
	if err != nil {
		t.Fatal(err)
	}
	if len(union) != 3 {
		t.Fatalf("SUnion len: %d (%v)", len(union), union)
	}
	got := map[string]bool{}
	for _, m := range union {
		got[m] = true
	}
	for _, want := range []string{"a", "b", "c"} {
		if !got[want] {
			t.Fatalf("SUnion missing %q: %v", want, union)
		}
	}

	// SIsMember + SCard
	if ok, _ := s.SIsMember(ctx, "key1", "a"); !ok {
		t.Fatal("SIsMember key1 a: want true")
	}
	if n, _ := s.SCard(ctx, "key1"); n != 2 {
		t.Fatalf("SCard key1: %d", n)
	}

	// SRem + SMembers
	if n, err := s.SRem(ctx, "key1", "a"); err != nil || n != 1 {
		t.Fatalf("SRem: n=%d err=%v", n, err)
	}
	members, _ := s.SMembers(ctx, "key1")
	if len(members) != 1 || members[0] != "b" {
		t.Fatalf("SMembers key1 after rem: %v", members)
	}
}
