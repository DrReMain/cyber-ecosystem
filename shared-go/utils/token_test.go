package utils

import (
	"encoding/hex"
	"testing"
)

func TestRandomToken(t *testing.T) {
	token, err := RandomToken()
	if err != nil {
		t.Fatalf("RandomToken() error: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("RandomToken() length = %d, want 64 (32 bytes hex-encoded)", len(token))
	}
	if _, err := hex.DecodeString(token); err != nil {
		t.Errorf("RandomToken() = %q, not valid hex: %v", token, err)
	}
}

func TestRandomToken_Unique(t *testing.T) {
	first, err := RandomToken()
	if err != nil {
		t.Fatalf("RandomToken() error: %v", err)
	}
	for range 8 {
		other, err := RandomToken()
		if err != nil {
			t.Fatalf("RandomToken() error: %v", err)
		}
		if other == first {
			t.Fatalf("RandomToken() returned the same value twice: %q", first)
		}
	}
}
