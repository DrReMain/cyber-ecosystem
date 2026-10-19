package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// sealV1 prefixes v1 ciphertexts (HKDF-derived key). The dot is the
// discriminator: it never occurs in the legacy base64(nonce||ct) shape, so
// prefix detection is unambiguous.
const sealV1 = "v1."

// kdfInfo domain-separates the derived AES key from any other future use of
// the same master-key material.
const kdfInfo = "cyber-ecosystem/utils.Seal/aes-256-gcm"

const aesKeyLen = 32

func Seal(plaintext, masterKey string) (string, error) {
	aead, err := newGCM(masterKey, hkdfKey)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}
	// The nonce doubles as AAD, integrity-binding the stored prefix too;
	// base64(nonce||ciphertext) keeps recovery free of an out-of-band channel.
	return sealV1 + base64.StdEncoding.EncodeToString(aead.Seal(nonce, nonce, []byte(plaintext), nil)), nil
}

func Open(sealed, masterKey string) (string, error) {
	if strings.HasPrefix(sealed, sealV1) {
		return open(sealed[len(sealV1):], masterKey, hkdfKey)
	}
	// Legacy pre-v1 ciphertexts (undecorated base64) keep opening under the
	// sha256 derivation; they age out as owners re-save their secrets.
	return open(sealed, masterKey, legacyKey)
}

func open(b64, masterKey string, derive func(string) ([]byte, error)) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("decode sealed value: %w", err)
	}
	aead, err := newGCM(masterKey, derive)
	if err != nil {
		return "", err
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

// hkdfKey derives the AES-256 key. The master key must already be
// high-entropy — this is domain separation, not password stretching; no slow
// KDF by design, Open sits on the chat dial hot path.
func hkdfKey(masterKey string) ([]byte, error) {
	if masterKey == "" {
		return nil, fmt.Errorf("seal master key is required")
	}
	key, err := hkdf.Key(sha256.New, []byte(masterKey), nil, kdfInfo, aesKeyLen)
	if err != nil {
		return nil, fmt.Errorf("derive master key: %w", err)
	}
	return key, nil
}

// legacyKey reproduces the pre-v1 derivation (single sha256) so existing
// stored ciphertexts stay readable.
func legacyKey(masterKey string) ([]byte, error) {
	if masterKey == "" {
		return nil, fmt.Errorf("seal master key is required")
	}
	sum := sha256.Sum256([]byte(masterKey))
	return sum[:], nil
}

func newGCM(masterKey string, derive func(string) ([]byte, error)) (cipher.AEAD, error) {
	key, err := derive(masterKey)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("init cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("init gcm: %w", err)
	}
	return aead, nil
}
