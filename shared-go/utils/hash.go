package utils

import "golang.org/x/crypto/bcrypt"

// Hash returns the bcrypt hash of the given plaintext.
func Hash(plaintext string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Verify reports whether the plaintext matches the bcrypt hash.
func Verify(plaintext, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}
