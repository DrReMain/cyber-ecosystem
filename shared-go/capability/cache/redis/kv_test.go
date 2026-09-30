package redis

import (
	"errors"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/cache"
)

// normalizeTTL maps go-redis TTL sentinels (raw unscaled nanoseconds) onto the
// cache contract.
func TestNormalizeTTL(t *testing.T) {
	if _, err := normalizeTTL(-2 * time.Nanosecond); !errors.Is(err, cache.ErrCacheMiss) {
		t.Errorf("absent sentinel (-2ns): %v, want ErrCacheMiss", err)
	}
	if ttl, err := normalizeTTL(-1 * time.Nanosecond); ttl != 0 || err != nil {
		t.Errorf("no-expiry sentinel (-1ns) = %v, %v; want 0, nil", ttl, err)
	}
	if ttl, err := normalizeTTL(5 * time.Second); ttl != 5*time.Second || err != nil {
		t.Errorf("real ttl = %v, %v; want passthrough", ttl, err)
	}
}
