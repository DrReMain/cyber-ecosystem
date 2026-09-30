package redis

import "testing"

// The session namespacing round-trips: user keys recovered from scanned full
// keys must equal the keys that were written.
func TestSessionKeyNamespacing(t *testing.T) {
	if got := sessionKey("sid", "field"); got != "session:sid:field" {
		t.Errorf("sessionKey = %q", got)
	}
	if got := sessionMatch("sid"); got != "session:sid:*" {
		t.Errorf("sessionMatch = %q", got)
	}
	full := sessionKey("sid", "a")
	if got := sessionUserKey("sid", full); got != "a" {
		t.Errorf("sessionUserKey round-trip = %q, want a", got)
	}
	// The match pattern must not overlap another session's keyspace.
	if sessionMatch("sid") == sessionMatch("sid2") {
		t.Error("distinct sessions share a match pattern")
	}
}
