package shared

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/capability/cache"
)

const (
	PolicyVersionKey     = "authz:policy:ver"
	PolicyChangedChannel = "authz:policy:changed"
)

func NotifyPolicyChanged(ctx context.Context, c *cache.Cache, log *slog.Logger) {
	if _, err := c.Counter.Incr(ctx, PolicyVersionKey, 1); err != nil {
		log.Warn("authz: version bump failed; periodic reconcile will converge", "error", err)
		return
	}
	if err := c.PubSub.Publish(ctx, PolicyChangedChannel, nil); err != nil {
		log.Warn("authz: change publish failed; periodic reconcile will converge", "error", err)
	}
}
