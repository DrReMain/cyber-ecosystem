package cache

import "context"

// Message is a delivered pub/sub event.
type Message struct {
	Channel string
	Pattern string // set for pattern subscriptions, empty otherwise
	Payload []byte
}

// Subscription is a live pub/sub subscription. Channel delivers messages until
// Close is called, after which the channel is closed and the forwarding
// goroutine exits (no leak).
type Subscription interface {
	Channel() <-chan Message
	Close() error
}

// PubSub is a real-time broadcast. Delivery is ephemeral — not durable; use the
// mq layer for reliable delivery. Subscribe/PSubscribe are long-lived until
// Subscription.Close.
type PubSub interface {
	Publish(ctx context.Context, channel string, payload []byte) error
	Subscribe(ctx context.Context, channels ...string) (Subscription, error)
	PSubscribe(ctx context.Context, patterns ...string) (Subscription, error)
}
