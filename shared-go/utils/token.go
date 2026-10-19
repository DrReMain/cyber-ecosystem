package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// RandomToken returns a fresh opaque credential: 32 random bytes hex-encoded
// (256 bits of entropy, 64 chars).
func RandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
