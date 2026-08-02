package mq

import "context"

// Publisher publishes a message to a topic (one topic → one backend stream).
// Delivery is at-least-once; producers needing publish-side dedup should carry
// an idempotency key in msg.Headers.
type Publisher interface {
	Publish(ctx context.Context, topic string, msg *Message) (id string, err error)
}
