package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// Argon2id parameters — OWASP baseline (19 MiB, t=2, p=1). Login-time memory
// budget is concurrency × 19 MiB, sized for back-office QPS. Changing these
// upgrades lazily: NeedsRehash flags old digests, the login flow rewrites
// them — never a batch migration.
const (
	argon2Time    uint32 = 2
	argon2Memory  uint32 = 19456 // KiB
	argon2Threads uint8  = 1
	argon2SaltLen        = 16
	argon2KeyLen         = 32
)

const (
	argon2idPrefix = "$argon2id$"
	bcryptPrefix   = "$2"
)

// Hash returns the argon2id digest of the given plaintext as a PHC string.
func Hash(plaintext string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}
	key := argon2.IDKey([]byte(plaintext), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return fmt.Sprintf("%sv=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2idPrefix, argon2.Version, argon2Memory, argon2Time, argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

// Verify reports whether the plaintext matches the stored digest. Legacy
// bcrypt digests keep verifying (rows written before the argon2id switch);
// callers upgrade them via NeedsRehash after a successful verification.
func Verify(plaintext, hash string) bool {
	if strings.HasPrefix(hash, argon2idPrefix) {
		return verifyArgon2id(plaintext, hash)
	}
	if strings.HasPrefix(hash, bcryptPrefix) {
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
	}
	return false
}

// NeedsRehash reports whether a verified digest predates the current
// algorithm or parameters — the login flow re-hashes the plaintext there.
func NeedsRehash(hash string) bool {
	_, _, p, ok := decodeArgon2id(hash)
	return !ok || p.memory != argon2Memory || p.time != argon2Time || p.threads != argon2Threads
}

type argon2Params struct {
	memory  uint32
	time    uint32
	threads uint8
}

// verifyArgon2id recomputes with the parameters recorded in the digest, not
// the current constants — parameter drift never breaks old rows.
func verifyArgon2id(plaintext, hash string) bool {
	salt, key, p, ok := decodeArgon2id(hash)
	if !ok {
		return false
	}
	got := argon2.IDKey([]byte(plaintext), salt, p.time, p.memory, p.threads, ConvNum[uint32](len(key)))
	return subtle.ConstantTimeCompare(got, key) == 1
}

// decodeArgon2id splits a PHC string "$argon2id$v=N$m=M,t=T,p=P$<salt>$<key>"
// with unpadded standard-base64 fields.
func decodeArgon2id(hash string) (salt, key []byte, p argon2Params, ok bool) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || !strings.HasPrefix(parts[2], "v=") {
		return nil, nil, p, false
	}
	kv := make(map[string]string, 3)
	for f := range strings.SplitSeq(parts[3], ",") {
		k, v, found := strings.Cut(f, "=")
		if !found {
			return nil, nil, p, false
		}
		kv[k] = v
	}
	m, err1 := strconv.ParseUint(kv["m"], 10, 32)
	t, err2 := strconv.ParseUint(kv["t"], 10, 32)
	th, err3 := strconv.ParseUint(kv["p"], 10, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return nil, nil, p, false
	}
	var err error
	if salt, err = base64.RawStdEncoding.DecodeString(parts[4]); err != nil || len(salt) == 0 {
		return nil, nil, p, false
	}
	// An empty key field would make any plaintext verify (IDKey(0) yields an
	// empty digest and equal-empty compares true) — reject it here.
	if key, err = base64.RawStdEncoding.DecodeString(parts[5]); err != nil || len(key) == 0 {
		return nil, nil, p, false
	}
	p = argon2Params{memory: ConvNum[uint32](m), time: ConvNum[uint32](t), threads: ConvNum[uint8](th)}
	return salt, key, p, true
}
