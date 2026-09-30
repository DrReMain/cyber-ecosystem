package cache

// Cache is the cache capability root: the full surface as ten sub-interfaces.
// Backend constructors (redis.New) populate every field — a nil field
// nil-derefs deep in a request path, not at startup. The underlying client's
// lifecycle (Close) belongs to the backend provider's wire-registered cleanup,
// not to this container.
type Cache struct {
	KV          KV
	Hash        Hash
	List        List
	Set         Set
	SortedSet   SortedSet
	Counter     Counter
	Lock        Lock
	RateLimiter RateLimiter
	PubSub      PubSub
	Session     Session
}
