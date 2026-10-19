package pg

import (
	"testing"
	"time"
)

// pollBackoff backs off exponentially across consecutive fetch failures,
// capped at 10s, and returns the base interval when healthy.
func TestPollBackoff(t *testing.T) {
	base := 500 * time.Millisecond
	if got := pollBackoff(base, 0); got != base {
		t.Errorf("healthy = %v, want base", got)
	}
	if got := pollBackoff(base, 1); got != base*2 {
		t.Errorf("1 failure = %v, want %v", got, base*2)
	}
	if got := pollBackoff(base, 3); got != base*8 {
		t.Errorf("3 failures = %v, want %v", got, base*8)
	}
	if got := pollBackoff(base, 100); got != 10*time.Second {
		t.Errorf("many failures = %v, want 10s cap", got)
	}
}

// nakBackoff is linear in attempt number with the 1m ceiling NATS's Nak
// delay also uses, so both backends retry in lockstep (parity invariant).
func TestNakBackoff(t *testing.T) {
	if got := nakBackoff(1); got != 100*time.Millisecond {
		t.Errorf("nakBackoff(1) = %v, want 100ms", got)
	}
	if got := nakBackoff(5); got != 500*time.Millisecond {
		t.Errorf("nakBackoff(5) = %v, want 500ms (linear)", got)
	}
	if got := nakBackoff(1000); got != time.Minute {
		t.Errorf("nakBackoff(1000) = %v, want 1m cap", got)
	}
}
