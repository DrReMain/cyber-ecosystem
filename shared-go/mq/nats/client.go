package nats

import (
	"context"
	"fmt"
	"sync"
	"time"

	natsclient "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"cyber-ecosystem/shared-go/mq"
)

// NewClient dials NATS and returns the conn + a JetStream handle + a cleanup
// that gracefully drains every active subscription before closing the
// connection (so consumer callbacks aren't pulled out from under in-flight
// messages).
func NewClient(cfg *Config) (*handle, func(), error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("%w: nil config", mq.ErrInvalidArgument)
	}
	if cfg.Endpoint == "" {
		return nil, nil, fmt.Errorf("%w: endpoint is required", mq.ErrInvalidArgument)
	}
	opts := []natsclient.Option{
		natsclient.Name("cyber-ecosystem-mq"),
		natsclient.ReconnectWait(2 * time.Second),
		natsclient.MaxReconnects(-1),
	}
	if cfg.Creds != "" {
		// Creds is a NATS credentials file path (JWT + NKey seed);
		// UserCredentials performs the challenge-response auth it requires.
		opts = append(opts, natsclient.UserCredentials(cfg.Creds))
	}
	nc, err := natsclient.Connect(cfg.Endpoint, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: nats connect: %w", mq.ErrUnavailable, err)
	}
	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, nil, fmt.Errorf("%w: jetstream init: %w", mq.ErrUnavailable, err)
	}
	h := &handle{nc: nc, js: js, cfg: *cfg}
	h.ctx, h.cancel = context.WithCancel(context.Background())
	return h, func() {
		// Cancel the lifetime ctx first so every consume callback (derived from it,
		// not from any caller ctx) stops, then drain the subscriptions, then the
		// connection. nc.Drain itself is non-blocking (server-side flush), but
		// ordering it after the subscriptions are closed avoids ack/dlq calls racing
		// a closing connection.
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		h.cancel()
		h.drainSubs(ctx)
		cancel()
		_ = nc.Drain()
	}, nil
}

// handle bundles the NATS conn + JetStream handle + resolved config, and tracks
// active subscriptions so shutdown can drain them in order.
type handle struct {
	nc  *natsclient.Conn
	js  jetstream.JetStream
	cfg Config

	// ctx is the handle lifetime context. Consume callbacks derive from it (not from
	// any Subscribe caller ctx) so long-lived subscriptions outlive the originating
	// request; cancelled by the returned cleanup.
	ctx    context.Context
	cancel context.CancelFunc

	mu   sync.Mutex
	subs map[*subscription]struct{}
}

func (h *handle) register(s *subscription) {
	h.mu.Lock()
	if h.subs == nil {
		h.subs = make(map[*subscription]struct{})
	}
	h.subs[s] = struct{}{}
	h.mu.Unlock()
}

func (h *handle) unregister(s *subscription) {
	h.mu.Lock()
	delete(h.subs, s)
	h.mu.Unlock()
}

// drainSubs gracefully closes every tracked subscription, bounded by ctx.
func (h *handle) drainSubs(ctx context.Context) {
	h.mu.Lock()
	subs := make([]*subscription, 0, len(h.subs))
	for s := range h.subs {
		subs = append(subs, s)
	}
	h.mu.Unlock()
	for _, s := range subs {
		s.drain(ctx)
	}
}
