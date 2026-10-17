package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
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

func TestHash_SaltUniqueness(t *testing.T) {
	first, err := Hash("cyber-ecosystem")
	if err != nil {
		t.Fatalf("Hash error: %v", err)
	}
	second, err := Hash("cyber-ecosystem")
	if err != nil {
		t.Fatalf("Hash error: %v", err)
	}
	if first == second {
		t.Error("two hashes of the same plaintext are identical (salt reuse)")
	}
}

func TestVerify_LegacyBcrypt(t *testing.T) {
	const password = "cyber-ecosystem"
	legacy, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword error: %v", err)
	}
	if !Verify(password, string(legacy)) {
		t.Errorf("Verify(password, legacy bcrypt digest) = false, want true")
	}
	if Verify("wrong-password", string(legacy)) {
		t.Errorf("Verify(wrong, legacy bcrypt digest) = true, want false")
	}
}

func TestNeedsRehash(t *testing.T) {
	fresh, err := Hash("cyber-ecosystem")
	if err != nil {
		t.Fatalf("Hash error: %v", err)
	}
	if NeedsRehash(fresh) {
		t.Errorf("NeedsRehash(fresh argon2id digest) = true, want false")
	}

	legacy, err := bcrypt.GenerateFromPassword([]byte("cyber-ecosystem"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword error: %v", err)
	}
	if !NeedsRehash(string(legacy)) {
		t.Errorf("NeedsRehash(legacy bcrypt digest) = false, want true")
	}

	// Editing the parameter section without recomputing the digest must NOT
	// verify — the parameters are authenticated content, not free metadata.
	tampered := strings.Replace(fresh, "m=19456,t=2,p=1", "m=8192,t=2,p=1", 1)
	if Verify("cyber-ecosystem", tampered) {
		t.Errorf("Verify(.., parameter-tampered digest) = true, want false")
	}
	if !NeedsRehash(tampered) {
		t.Errorf("NeedsRehash(parameter-tampered digest) = false, want true")
	}

	// A digest genuinely minted with non-current parameters still verifies
	// (verification uses the recorded parameters) and flags for upgrade.
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		t.Fatalf("read salt: %v", err)
	}
	driftedKey := argon2.IDKey([]byte("cyber-ecosystem"), salt, 2, 8192, 1, argon2KeyLen)
	drifted := fmt.Sprintf("%sv=%d$m=8192,t=2,p=1$%s$%s", argon2idPrefix, argon2.Version,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(driftedKey))
	if !Verify("cyber-ecosystem", drifted) {
		t.Errorf("Verify with recorded (drifted) parameters = false, want true")
	}
	if !NeedsRehash(drifted) {
		t.Errorf("NeedsRehash(drifted-parameters digest) = false, want true")
	}
}

func TestVerify_Malformed(t *testing.T) {
	for _, hash := range []string{
		"",
		"plaintext",
		"$argon2id$garbage",
		"$argon2id$v=19$m=19456,t=2,p=1$not-base64$AAA",
		"$argon2id$v=19$m=19456,t=2,p=1$c2FsdA", // missing key field
		"$5$rounds=535000$sha",
	} {
		if Verify("cyber-ecosystem", hash) {
			t.Errorf("Verify(.., %q) = true, want false", hash)
		}
		if !NeedsRehash(hash) {
			t.Errorf("NeedsRehash(%q) = false, want true", hash)
		}
	}
}
