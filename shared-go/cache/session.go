package cache

import (
	"context"
	"time"
)

// Session is KV scoped by a sessionID with convenience helpers (TTL refresh,
// enumerate/destroy). Internally keys are namespaced as "session:{id}:{key}".
// Refresh on a missing/expired session returns ErrSessionNotFound.
type Session interface {
	Get(ctx context.Context, sessionID, key string) ([]byte, error)
	Set(ctx context.Context, sessionID, key string, val []byte, ttl time.Duration) error
	Del(ctx context.Context, sessionID, key string) error
	Exists(ctx context.Context, sessionID string) (bool, error)
	Refresh(ctx context.Context, sessionID string, ttl time.Duration) error
	Destroy(ctx context.Context, sessionID string) error
	Keys(ctx context.Context, sessionID string) ([]string, error)
}
