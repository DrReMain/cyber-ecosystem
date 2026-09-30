package shared

import (
	"context"
	"time"

	"cyber-ecosystem/shared-go/capability/cache"
	"cyber-ecosystem/shared-go/utils"
)

const RevokedMarkerPrefix = "auth:web:revoked:"

const revocationTTL = 30 * 24 * time.Hour

func RevokeSessions(ctx context.Context, c *cache.Cache, userID string) error {
	// The TTL must outlive every session the marker can target: past 30 days
	// each has hit its own absolute cap, so the marker has nothing left to kill.
	return c.KV.Set(ctx, RevokedMarkerPrefix+userID, utils.MustMarshal(time.Now()), revocationTTL)
}
