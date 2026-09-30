package cache

import (
	"strings"
	"testing"
)

func TestValidateKey(t *testing.T) {
	if err := ValidateKey(""); err == nil {
		t.Error("empty key should be rejected")
	}
	if err := ValidateKey(strings.Repeat("k", maxKeyLen+1)); err == nil {
		t.Error("over-long key should be rejected")
	}
	if err := ValidateKey(strings.Repeat("k", maxKeyLen)); err != nil {
		t.Errorf("key at the bound should pass: %v", err)
	}
}

func TestValidateKeys(t *testing.T) {
	if err := ValidateKeys(); err != nil {
		t.Errorf("no keys should pass: %v", err)
	}
	if err := ValidateKeys("ok", ""); err == nil {
		t.Error("one bad key in the batch should fail the batch")
	}
}

// sessionForbidden chars break the "session:{id}:{key}" namespacing or enable
// pattern injection (Destroy("*") deleting every session).
func TestValidateSessionID(t *testing.T) {
	for _, id := range []string{"", "a:b", "a*b", "a?b", "a[b", `a\b`} {
		if err := ValidateSessionID(id); err == nil {
			t.Errorf("session id %q should be rejected", id)
		}
	}
	if err := ValidateSessionID("sess-42"); err != nil {
		t.Errorf("plain id should pass: %v", err)
	}
}

func TestValidateSessionKey(t *testing.T) {
	if err := ValidateSessionKey(""); err == nil {
		t.Error("empty session key should be rejected")
	}
	if err := ValidateSessionKey("a:b"); err == nil {
		t.Error("session key containing ':' collides across fields/sessions")
	}
	if err := ValidateSessionKey("field"); err != nil {
		t.Errorf("plain session key should pass: %v", err)
	}
}
