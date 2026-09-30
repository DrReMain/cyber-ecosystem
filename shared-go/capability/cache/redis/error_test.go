package redis

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/capability/cache"
)

// mapErr is the single error boundary of the redis adapter.
func TestMapErr(t *testing.T) {
	if err := mapErr(nil); err != nil {
		t.Errorf("nil: got %v", err)
	}

	// Already-classified sentinels (incl. wrapped) pass through unchanged.
	for _, sentinel := range []error{
		cache.ErrCacheMiss, cache.ErrKeyNotFound, cache.ErrSessionNotFound,
		cache.ErrQuotaExceeded, cache.ErrInvalidArgument, cache.ErrLockNotAcquired,
	} {
		wrapped := fmt.Errorf("op: %w", sentinel)
		if !errors.Is(mapErr(wrapped), sentinel) {
			t.Errorf("sentinel %v should pass through", sentinel)
		}
	}

	// Context errors belong to the caller — pass through so they can react to
	// their own cancellation/deadline rather than seeing ErrUnavailable.
	for _, cerr := range []error{context.Canceled, context.DeadlineExceeded} {
		if got := mapErr(cerr); !errors.Is(got, cerr) || errors.Is(got, cache.ErrUnavailable) {
			t.Errorf("ctx error %v should pass through unwrapped, got %v", cerr, got)
		}
	}

	// Everything else (network, WRONGTYPE, pool exhaustion, raw redis.Nil from
	// an impl that forgot to map it) wraps as ErrUnavailable.
	for _, raw := range []error{redis.Nil, errors.New("dial tcp: refused"), redis.ErrClosed} {
		if !errors.Is(mapErr(raw), cache.ErrUnavailable) {
			t.Errorf("unknown error %v should wrap as ErrUnavailable", raw)
		}
	}
}
