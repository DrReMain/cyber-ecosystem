package utils

import (
	"testing"
)

func TestHash(t *testing.T) {
	const password = "cyber-ecosystem"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash(%q) error: %v", password, err)
	}
	if hash == "" {
		t.Fatalf("Hash(%q) returned empty string", password)
	}
	if hash == password {
		t.Fatalf("Hash(%q) returned plaintext", password)
	}
}

func TestHash_Verify(t *testing.T) {
	const password = "cyber-ecosystem"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash error: %v", err)
	}

	if !Verify(password, hash) {
		t.Errorf("Verify(%q, hash) = false, want true", password)
	}
	if Verify("wrong-password", hash) {
		t.Errorf("Verify(wrong, hash) = true, want false")
	}
}

func TestHash_EmptyString(t *testing.T) {
	hash, err := Hash("")
	if err != nil {
		t.Fatalf("Hash('') error: %v", err)
	}
	if !Verify("", hash) {
		t.Errorf("Verify('', hash) = false, want true")
	}
}
