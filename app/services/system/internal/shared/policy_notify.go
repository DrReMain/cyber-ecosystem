package shared

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/capability/cache"
)

const PolicyChangedChannel = "authz:policy:changed"

// The wake-only half of the protocol: the engine's version truth is the
// authz_stamps row bumped inside each change transaction, so this call
// only shortens the wait — a failed publish costs timeliness, never
// correctness, and the periodic reconcile plus the forced-rebuild cap
// bound the staleness either way.
func NotifyPolicyChanged(ctx context.Context, c *cache.Cache, log *slog.Logger) {
	if err := c.PubSub.Publish(ctx, PolicyChangedChannel, nil); err != nil {
		log.Warn("authz: change publish failed; periodic reconcile will converge", "error", err)
	}
}
