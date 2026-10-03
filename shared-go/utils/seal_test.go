package utils

import (
	"strings"
	"testing"
)

func TestSeal(t *testing.T) {
	const (
		secret    = "sk-agent-0123456789"
		masterKey = "agent-insecure-dev-key"
	)

	sealed, err := Seal(secret, masterKey)
	if err != nil {
		t.Fatalf("Seal error: %v", err)
	}
	if sealed == "" || sealed == secret || strings.Contains(sealed, secret) {
		t.Fatalf("Seal returned empty or plaintext-bearing value %q", sealed)
	}
}

func TestSeal_Open_Roundtrip(t *testing.T) {
	const (
		secret    = "sk-agent-0123456789"
		masterKey = "agent-insecure-dev-key"
	)

	sealed, err := Seal(secret, masterKey)
	if err != nil {
		t.Fatalf("Seal error: %v", err)
	}

	opened, err := Open(sealed, masterKey)
	if err != nil {
		t.Fatalf("Open error: %v", err)
	}
	if opened != secret {
		t.Errorf("Open = %q, want %q", opened, secret)
	}
}

func TestSeal_NonceUniqueness(t *testing.T) {
	const (
		secret    = "sk-agent-0123456789"
		masterKey = "agent-insecure-dev-key"
	)

	first, err := Seal(secret, masterKey)
	if err != nil {
		t.Fatalf("Seal error: %v", err)
	}
	second, err := Seal(secret, masterKey)
	if err != nil {
		t.Fatalf("Seal error: %v", err)
	}
	if first == second {
		t.Error("two seals of the same plaintext are identical (nonce reuse)")
	}
}

func TestOpen_WrongKey(t *testing.T) {
	sealed, err := Seal("sk-agent-0123456789", "agent-insecure-dev-key")
	if err != nil {
		t.Fatalf("Seal error: %v", err)
	}
	if _, err := Open(sealed, "another-master-key"); err == nil {
		t.Error("Open with a wrong master key succeeded")
	}
}

func TestSeal_EmptyMasterKey(t *testing.T) {
	if _, err := Seal("secret", ""); err == nil {
		t.Error("Seal with empty master key succeeded")
	}
	if _, err := Open("", ""); err == nil {
		t.Error("Open with empty master key succeeded")
	}
}
