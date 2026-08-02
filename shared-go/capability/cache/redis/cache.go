package redis

import (
	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/capability/cache"
)

// New wires redis-backed interface implementations onto a *cache.Cache. One
// line is added per interface as its slice lands.
func New(client *redis.Client) *cache.Cache {
	return &cache.Cache{
		KV:          NewKV(client),
		Hash:        NewHash(client),
		List:        NewList(client),
		Set:         NewSet(client),
		SortedSet:   NewSortedSet(client),
		Counter:     NewCounter(client),
		Lock:        NewLock(client),
		RateLimiter: NewRateLimiter(client),
		PubSub:      NewPubSub(client),
		Session:     NewSession(client),
	}
}
