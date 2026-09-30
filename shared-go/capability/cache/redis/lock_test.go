package redis

import (
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/cache"
)

func TestApplyLockOpts(t *testing.T) {
	o := applyLockOpts(nil)
	if o.Tries != 0 || o.RetryDelay != 0 {
		t.Errorf("no opts: %+v, want zero value", o)
	}
	o = applyLockOpts([]cache.LockOption{cache.WithRetry(3, 250*time.Millisecond)})
	if o.Tries != 3 || o.RetryDelay != 250*time.Millisecond {
		t.Errorf("WithRetry: %+v", o)
	}
}

// resolveLockRetry applies the blocking-Lock defaults: unlimited attempts
// (bounded by ctx/ttl) and a 100ms backoff.
func TestResolveLockRetry(t *testing.T) {
	if tries, delay := resolveLockRetry(&cache.LockOptions{}); tries != 1<<30 || delay != 100*time.Millisecond {
		t.Errorf("defaults = %d, %v", tries, delay)
	}
	if tries, delay := resolveLockRetry(&cache.LockOptions{Tries: -5, RetryDelay: -time.Second}); tries != 1<<30 || delay != 100*time.Millisecond {
		t.Errorf("negative values = %d, %v, want defaults", tries, delay)
	}
	if tries, delay := resolveLockRetry(&cache.LockOptions{Tries: 4, RetryDelay: time.Second}); tries != 4 || delay != time.Second {
		t.Errorf("explicit values should pass through: %d, %v", tries, delay)
	}
}
