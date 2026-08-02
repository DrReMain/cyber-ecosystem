package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken returns a 256-bit opaque token (64 lowercase hex chars) from
// crypto/rand. Strong entropy (2^256), unpredictable, collision-negligible.
func GenerateToken() (string, error) {
	b := make([]byte, 32) // 256-bit
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
