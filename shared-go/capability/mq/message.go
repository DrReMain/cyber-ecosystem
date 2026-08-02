package mq

import "time"

// Message is a queue message. Publishers set Topic/Payload/Headers; the backend
// fills ID/Timestamp on consume.
//
// ID is opaque and backend-local (e.g. a backend message sequence). It is NOT a
// durable idempotency key: it may reset across backend lifecycle events (stream
// purge/recreate). Under at-least-once delivery, dedup with an application
// idempotency key carried in Headers, not ID. Handlers must not retain Payload
// past return without copying — the abstraction is backend-agnostic and a
// backend may recycle the backing buffer.
type Message struct {
	Topic     string
	Payload   []byte
	Headers   map[string]string
	ID        string
	Timestamp time.Time
}
