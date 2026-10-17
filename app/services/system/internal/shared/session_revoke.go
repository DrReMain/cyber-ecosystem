package shared

import (
	"context"
	"time"

	"cyber-ecosystem/shared-go/capability/cache"
	"cyber-ecosystem/shared-go/utils"
)

const RevokedMarkerPrefix = "auth:web:revoked:"

// SessionMax is the absolute session lifetime cap. It lives here — beside
// the revocation marker — because the marker's TTL derives from it: one
// home, one derivation, never two constants to keep in sync.
const SessionMax = 30 * 24 * time.Hour

func RevokeSessions(ctx context.Context, c *cache.Cache, userID string) error {
	return c.KV.Set(ctx, RevokedMarkerPrefix+userID, utils.MustMarshal(time.Now()), SessionMax)
}
