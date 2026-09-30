package redis

import (
	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/capability/cache"
)

// New assembles a fully-populated *cache.Cache; it is the only assembly path,
// so the container's every-field invariant holds by construction.
func New(client *redis.Client) *cache.Cache {
	return &cache.Cache{
		KV:          newKV(client),
		Hash:        newHash(client),
		List:        newList(client),
		Set:         newSet(client),
		SortedSet:   newSortedSet(client),
		Counter:     newCounter(client),
		Lock:        newLock(client),
		RateLimiter: newRateLimiter(client),
		PubSub:      newPubSub(client),
		Session:     newSession(client),
	}
}
