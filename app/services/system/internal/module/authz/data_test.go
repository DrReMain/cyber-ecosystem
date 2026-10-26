package authz

// forcedDue is the forced-rebuild cap's whole decision — pinned here
// because its regression is invisible at runtime until an unknown
// staleness class outlives the cap.

import (
	"testing"
	"time"
)

func TestForcedDue(t *testing.T) {
	now := time.Now()
	if !forcedDue(0, now) {
		t.Fatal("an unset last-forced must be due")
	}
	if forcedDue(now.Add(-time.Hour).UnixNano(), now) {
		t.Fatal("one hour since the last rebuild must not be due")
	}
	if !forcedDue(now.Add(-25*time.Hour).UnixNano(), now) {
		t.Fatal("25 hours since the last rebuild must be due")
	}
}
