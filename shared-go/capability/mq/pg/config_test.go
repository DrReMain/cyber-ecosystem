package pg

import (
	"testing"
	"time"
)

// Zero (unset) values resolve to defaults; explicit values pass through.
func TestOrDefaults(t *testing.T) {
	if got := pollIntervalOrDefault(0); got != defaultPollInterval {
		t.Errorf("pollIntervalOrDefault(0) = %v", got)
	}
	if got := pollIntervalOrDefault(time.Second); got != time.Second {
		t.Errorf("pollIntervalOrDefault(1s) = %v, want passthrough", got)
	}
	if got := visibilityOrDefault(0); got != defaultVisibility {
		t.Errorf("visibilityOrDefault(0) = %v", got)
	}
	if got := maxRetriesOrDefault(0); got != defaultMaxRetries {
		t.Errorf("maxRetriesOrDefault(0) = %d", got)
	}
	if got := retentionOrDefault(0); got != defaultRetention {
		t.Errorf("retentionOrDefault(0) = %v", got)
	}
	if got := batchSizeOrDefault(0); got != defaultBatchSize {
		t.Errorf("batchSizeOrDefault(0) = %d", got)
	}
	// Negative values also resolve to defaults: the 0/negative contract is
	// "unset", not merely "not positive".
	if got := maxRetriesOrDefault(-1); got != defaultMaxRetries {
		t.Errorf("maxRetriesOrDefault(-1) = %d, want default", got)
	}
}
