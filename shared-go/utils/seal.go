package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

func Seal(plaintext, masterKey string) (string, error) {
	aead, err := newGCM(masterKey)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}
	// base64(nonce||ciphertext): the prefix is recoverable at open time
	// without an out-of-band nonce channel.
	return base64.StdEncoding.EncodeToString(aead.Seal(nonce, nonce, []byte(plaintext), nil)), nil
}

func Open(sealed, masterKey string) (string, error) {
	aead, err := newGCM(masterKey)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(sealed)
	if err != nil {
		return "", fmt.Errorf("decode sealed value: %w", err)
	}
	ns := aead.NonceSize()
	if len(raw) < ns {
		return "", fmt.Errorf("sealed value shorter than a nonce")
	}
	plaintext, err := aead.Open(nil, raw[:ns], raw[ns:], nil)
	if err != nil {
		return "", fmt.Errorf("open sealed value: %w", err)
	}
	return string(plaintext), nil
}

func newGCM(masterKey string) (cipher.AEAD, error) {
	if masterKey == "" {
		return nil, fmt.Errorf("seal master key is required")
	}
	// Derived, not used raw: the config accepts any non-empty string.
	sum := sha256.Sum256([]byte(masterKey))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, fmt.Errorf("derive master key: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("init gcm: %w", err)
	}
	return aead, nil
}
