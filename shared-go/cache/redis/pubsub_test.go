package redis

import (
	"context"
	"testing"
	"time"
)

func TestPubSub(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	ps := c.PubSub

	// Subscribe, then publish, then receive
	sub, err := ps.Subscribe(ctx, "ch1")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	// Subscribe is async in redis; give the subscription a moment to register.
	time.Sleep(80 * time.Millisecond)

	if err := ps.Publish(ctx, "ch1", []byte("hello")); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case msg := <-sub.Channel():
		if msg.Channel != "ch1" || string(msg.Payload) != "hello" {
			t.Fatalf("received: %+v", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("did not receive published message within 2s")
	}

	// Close → channel closes, forwarding goroutine exits (no leak)
	if err := sub.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, ok := <-sub.Channel(); ok {
		t.Fatal("channel should be closed after Close")
	}
}

func TestPubSubPattern(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	ps := c.PubSub

	sub, err := ps.PSubscribe(ctx, "evt:*")
	if err != nil {
		t.Fatalf("PSubscribe: %v", err)
	}
	time.Sleep(80 * time.Millisecond)

	if err := ps.Publish(ctx, "evt:login", []byte("ok")); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	select {
	case msg := <-sub.Channel():
		if msg.Pattern != "evt:*" || msg.Channel != "evt:login" || string(msg.Payload) != "ok" {
			t.Fatalf("received: %+v", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("did not receive pattern-published message within 2s")
	}
	_ = sub.Close()
}

// TestPubSubCloseWhileFull pins the F3 leak fix: with the out buffer full (slow
// consumer), Close must still unblock the forwarder and close the channel. A
// stranded goroutine (the pre-fix bug) would leave the channel open forever.
func TestPubSubCloseWhileFull(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()
	ps := c.PubSub

	sub, err := ps.Subscribe(ctx, "chfull1")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	time.Sleep(80 * time.Millisecond) // let the subscription register

	// Publish well past the 100-cap out buffer WITHOUT consuming -> the
	// forwarder fills s.out then parks on the blocked send.
	for range 200 {
		if err := ps.Publish(ctx, "chfull1", []byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(120 * time.Millisecond) // let the forwarder fill + park

	if err := sub.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// Drain any buffered messages; the channel MUST close (goroutine exited).
	deadline := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-sub.Channel():
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("out channel never closed after Close — forwarder goroutine leaked")
		}
	}
}

// TestPubSubCtxCancelWhileFull pins the ctx-cancel trigger of F3: cancelling
// the subscription's ctx while the out buffer is full must also close the
// channel (forwarder exits via ctx.Done()).
func TestPubSubCtxCancelWhileFull(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	cctx, cancel := context.WithCancel(context.Background())
	ps := c.PubSub

	sub, err := ps.PSubscribe(cctx, "evtfull:*")
	if err != nil {
		t.Fatalf("PSubscribe: %v", err)
	}
	time.Sleep(80 * time.Millisecond)

	for range 200 {
		if err := ps.Publish(context.Background(), "evtfull:x", []byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(120 * time.Millisecond) // forwarder fills + parks

	cancel() // ctx cancellation must unblock the parked forwarder
	deadline := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-sub.Channel():
			if !ok {
				_ = sub.Close()
				return
			}
		case <-deadline:
			t.Fatal("out channel never closed after ctx cancel — forwarder goroutine leaked")
		}
	}
}
