package shared

import (
	"context"
	"time"

	"cyber-ecosystem/shared-go/capability/cache"
	"cyber-ecosystem/shared-go/utils"
)

// TTL derived from SessionMax.
const RevokedMarkerPrefix = "auth:web:revoked:"

const SessionMax = 30 * 24 * time.Hour

func RevokeSessions(ctx context.Context, c *cache.Cache, userID string) error {
	return c.KV.Set(ctx, RevokedMarkerPrefix+userID, utils.MustMarshal(time.Now()), SessionMax)
}
