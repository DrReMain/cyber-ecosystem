package redis

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/cache"
)

type pubsub struct{ client *redis.Client }

// NewPubSub returns the redis-backed PubSub implementation.
//
// Subscribe/PSubscribe take a context that governs the subscription lifetime;
// pass a non-request-scoped context (or context.Background) for a long-lived
// subscription — a request-scoped ctx tears the subscription down when the
// request ends. Either way, Close() stops the subscription and closes the
// delivered channel.
func NewPubSub(client *redis.Client) cache.PubSub { return &pubsub{client: client} }

func (p *pubsub) Publish(ctx context.Context, channel string, payload []byte) error {
	if err := cache.ValidateKey(channel); err != nil {
		return err
	}
	return p.client.Publish(ctx, channel, payload).Err()
}

// Subscribe args are LITERAL channel names; PSubscribe args are glob PATTERNS.
func (p *pubsub) Subscribe(ctx context.Context, channels ...string) (cache.Subscription, error) {
	if len(channels) == 0 {
		return nil, cache.ErrInvalidArgument
	}
	if err := cache.ValidateKeys(channels...); err != nil {
		return nil, err
	}
	return newSubscription(ctx, p.client.Subscribe(ctx, channels...)), nil
}

func (p *pubsub) PSubscribe(ctx context.Context, patterns ...string) (cache.Subscription, error) {
	if len(patterns) == 0 {
		return nil, cache.ErrInvalidArgument
	}
	if err := cache.ValidateKeys(patterns...); err != nil {
		return nil, err
	}
	return newSubscription(ctx, p.client.PSubscribe(ctx, patterns...)), nil
}

type subscription struct {
	ps   *redis.PubSub
	out  chan cache.Message
	done chan struct{}
	once sync.Once
}

// newSubscription starts a forwarding goroutine. The goroutine exits (and
// closes out) when the upstream channel closes (ps.Close / ctx cancel) OR Close
// signals done — so a full out buffer can never strand the goroutine. Delivery
// is ephemeral: a slow consumer back-pressures the forwarder; once out is full
// and done/ctx fire, the goroutine exits without blocking.
func newSubscription(ctx context.Context, ps *redis.PubSub) cache.Subscription {
	s := &subscription{
		ps:   ps,
		out:  make(chan cache.Message, 100),
		done: make(chan struct{}),
	}
	go func() {
		defer close(s.out)
		for msg := range ps.Channel() {
			select {
			case s.out <- cache.Message{Channel: msg.Channel, Pattern: msg.Pattern, Payload: []byte(msg.Payload)}:
			case <-s.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	return s
}

func (s *subscription) Channel() <-chan cache.Message { return s.out }

// Close is idempotent and owns the close sequence: it tears down the redis
// subscription, signals the forwarder to exit, and the forwarder closes out.
func (s *subscription) Close() error {
	var err error
	s.once.Do(func() {
		err = s.ps.Close()
		close(s.done)
	})
	return err
}
