package bootstrap

import (
	"context"
	"log/slog"
	"sync"

	"cyber-ecosystem/app/services/esports/internal/shared"
)

type EventSubscriptions struct { // the in-process domain event bus: subscriber lists under one mutex, synchronous delivery; sits beside the Lifecycle registry as process-boot machinery
	log           *slog.Logger
	mu            sync.Mutex
	trialHandlers []func(context.Context, shared.TrialCompleted)
}

func NewEventSubscriptions(logger *slog.Logger) *EventSubscriptions {
	return &EventSubscriptions{log: logger.With("module", "bootstrap/eventbus")}
}

func (s *EventSubscriptions) SubscribeTrialCompleted(handler func(context.Context, shared.TrialCompleted)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trialHandlers = append(s.trialHandlers, handler)
}

func (s *EventSubscriptions) PublishTrialCompleted(ctx context.Context, e shared.TrialCompleted) {
	// Snapshot the handler list under the lock; delivery itself is
	// best-effort by contract — the publisher's own transaction is the
	// truth, a subscriber's failure is its own log line, never the
	// publisher's error.

	s.mu.Lock()
	handlers := make([]func(context.Context, shared.TrialCompleted), len(s.trialHandlers))
	copy(handlers, s.trialHandlers)
	s.mu.Unlock()
	for _, h := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					s.log.Error("trial-completed handler panicked", "record_id", e.RecordID, "panic", r)
				}
			}()
			h(ctx, e)
		}()
	}
}
