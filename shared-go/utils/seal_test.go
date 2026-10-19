package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
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

func TestSeal_V1Prefix(t *testing.T) {
	sealed, err := Seal("sk-agent-0123456789", "agent-master-key")
	if err != nil {
		t.Fatalf("Seal error: %v", err)
	}
	if !strings.HasPrefix(sealed, "v1.") {
		t.Fatalf("Seal output lacks the v1 version prefix: %q", sealed)
	}
}

// Pre-v1 ciphertexts (undecorated base64, sha256-derived key) must keep
// opening after the HKDF switch — this pins the legacy shape.
func TestOpen_LegacyCiphertext(t *testing.T) {
	const (
		secret    = "sk-agent-0123456789"
		masterKey = "agent-insecure-dev-key"
	)
	key, err := legacyKey(masterKey)
	if err != nil {
		t.Fatalf("legacyKey error: %v", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes.NewCipher error: %v", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("cipher.NewGCM error: %v", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("read nonce: %v", err)
	}
	legacy := base64.StdEncoding.EncodeToString(aead.Seal(nonce, nonce, []byte(secret), nil))

	opened, err := Open(legacy, masterKey)
	if err != nil {
		t.Fatalf("Open(legacy ciphertext) error: %v", err)
	}
	if opened != secret {
		t.Errorf("Open(legacy ciphertext) = %q, want %q", opened, secret)
	}
	if _, err := Open(legacy, "another-master-key"); err == nil {
		t.Error("Open(legacy ciphertext) with a wrong master key succeeded")
	}
}
