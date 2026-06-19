package mq

import "context"

// Subscription is a running consumer; Close stops consuming.
type Subscription interface {
	Close() error
}

// Consumer subscribes to a topic. group is the consumer identity: the same
// group across subscribers → competing work-queue (each message → one); a
// different group → independent durable → broadcast (each group gets all).
// handler returns nil to ack; an error to retry (and, after MaxRetries, DLQ).
//
// Delivery is at-least-once: handlers MUST be idempotent. There is NO
// per-message ordering guarantee across competing consumers within a group. A
// handler must return well within the backend's AckWait, otherwise the same
// message is redelivered while the first invocation is still running. The
// provided ctx is cancelled when the subscription closes; do not start
// long-lived work on it.
type Consumer interface {
	Subscribe(ctx context.Context, topic, group string, handler func(ctx context.Context, msg Message) error) (Subscription, error)
}
