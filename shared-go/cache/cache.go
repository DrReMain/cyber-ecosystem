package cache

// Cache holds the cache capability interfaces. Every backend constructor (e.g.
// shared-go/cache/redis.New) MUST populate ALL fields; a nil field is undefined
// behavior (a call on it nil-dereferences deep in a request path, not at
// startup). The 10-interface surface is the full redis capability for the
// platform scaffold — adding a new interface means adding a field here plus an
// impl, not partial wiring. The underlying client's lifecycle (Close) is owned
// by the backend provider's wire-registered cleanup, not by this container.
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
