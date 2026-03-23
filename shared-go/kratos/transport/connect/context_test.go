package connect

import (
	"context"
	"errors"
	"testing"
	"time"
)

type ctxKey string

func TestMergeValuePriority(t *testing.T) {
	k := ctxKey("k")
	c1 := context.WithValue(context.Background(), k, "from1")
	c2 := context.WithValue(context.Background(), k, "from2")
	m, cancel := Merge(c1, c2)
	defer cancel()
	if got := m.Value(k); got != "from1" {
		t.Fatalf("value=%v", got)
	}
}

func TestMergeDeadline(t *testing.T) {
	c1, cancel1 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel1()
	c2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()
	m, cancel := Merge(c1, c2)
	defer cancel()

	d, ok := m.Deadline()
	if !ok {
		t.Fatal("deadline should exist")
	}
	if d.After(time.Now().Add(1500 * time.Millisecond)) {
		t.Fatal("deadline should pick earlier parent")
	}
}

func TestMergeCancelFunc(t *testing.T) {
	m, cancel := Merge(context.Background(), context.Background())
	cancel()
	select {
	case <-m.Done():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("merged context not canceled")
	}
	if !errors.Is(m.Err(), context.Canceled) {
		t.Fatalf("err=%v", m.Err())
	}
}

func TestMergeParentCancel(t *testing.T) {
	p1, cancel1 := context.WithCancel(context.Background())
	m, cancel := Merge(p1, context.Background())
	defer cancel()
	cancel1()

	select {
	case <-m.Done():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("merged context should follow parent cancel")
	}
	if !errors.Is(m.Err(), context.Canceled) {
		t.Fatalf("err=%v", m.Err())
	}
}
