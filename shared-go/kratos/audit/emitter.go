package audit

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

// PublishFunc delivers one event to the audit carrier. It runs on the
// drainer goroutine, never on the request path, and must give up on its own
// — a slow or dead carrier costs queue headroom first, then drops.
type PublishFunc func(ctx context.Context, ev *Event) error

var (
	defaultCapacity = 1024
	publishTimeout  = 5 * time.Second
	stopTimeout     = 3 * time.Second
)

type EmitterOption func(*Emitter)

// Capacity overrides the queue size (default 1024).
func Capacity(n int) EmitterOption {
	return func(e *Emitter) { e.ch = make(chan *Event, n) }
}

// Emitter is the asynchronous publish component the application's
// persistence side drives: an in-process queue drained by a background
// publisher. Audit availability yields to service availability — a full
// queue drops events (counted, sampled WARN) instead of blocking or
// rejecting the request.
type Emitter struct {
	publish PublishFunc
	log     *slog.Logger
	ch      chan *Event
	quit    chan struct{}
	done    chan struct{}
	dropped atomic.Uint64
}

// NewEmitter starts the drainer and returns the emitter plus a cleanup that
// stops it after a best-effort drain of the queued events.
func NewEmitter(publish PublishFunc, logger *slog.Logger, opts ...EmitterOption) (*Emitter, func()) {
	e := &Emitter{
		publish: publish,
		log:     logger,
		ch:      make(chan *Event, defaultCapacity),
		quit:    make(chan struct{}),
		done:    make(chan struct{}),
	}
	for _, opt := range opts {
		opt(e)
	}
	go e.drain()
	return e, e.stop
}

// Emit enqueues without ever blocking the caller. The context is accepted
// for the interface and deliberately not retained: the queue exists exactly
// so that delivery outlives the request.
func (e *Emitter) Emit(_ context.Context, ev *Event) {
	select {
	case e.ch <- ev:
	default:
		// Sampling the log keeps a dead carrier from amplifying itself into
		// a log storm; the count stays exact.
		n := e.dropped.Add(1)
		if n == 1 || n%128 == 0 {
			e.log.Warn("audit: queue full, dropping events", "dropped", n)
		}
	}
}

// Dropped reports how many events the queue has discarded. Intended for
// diagnostics; the counter never resets.
func (e *Emitter) Dropped() uint64 { return e.dropped.Load() }

func (e *Emitter) drain() {
	defer close(e.done)
	for {
		select {
		case <-e.quit:
			e.flush()
			return
		case ev := <-e.ch:
			e.deliver(ev)
		}
	}
}

// deliver publishes under an independent timeout: the drainer owns its own
// deadline, unrelated to any request lifetime.
func (e *Emitter) deliver(ev *Event) {
	ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
	defer cancel()
	if err := e.publish(ctx, ev); err != nil {
		e.log.Warn("audit: publish failed", "operation", ev.Operation, "error", err)
	}
}

func (e *Emitter) flush() {
	for {
		select {
		case ev := <-e.ch:
			e.deliver(ev)
		default:
			return
		}
	}
}

func (e *Emitter) stop() {
	close(e.quit)
	select {
	case <-e.done:
	case <-time.After(stopTimeout):
		// A publish stuck past the budget abandons the remaining queue —
		// shutdown must not be held hostage by the audit carrier.
		e.log.Warn("audit: drain timed out, remaining events dropped", "dropped", e.Dropped())
	}
}
