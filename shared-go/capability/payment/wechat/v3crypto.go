package wechat

// The WeChat Pay APIv3 cryptographic core, mirrored as pure functions with
// no IO and no channel state: WeChat protocol knowledge lives with the
// WeChat backend, not in the contract package.

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	stderrors "errors"
	"fmt"
	"strings"
)

// BuildRequestMessage assembles the five-part string the merchant signs for
// every API call: method, path with query, timestamp, nonce and body, each
// followed by a newline — a GET's empty body leaves two trailing newlines,
// which is part of the contract, not an accident.
func BuildRequestMessage(method, pathWithQuery string, timestamp int64, nonce string, body []byte) string {
	return fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", method, pathWithQuery, timestamp, nonce, body)
}

// SignSHA256RSA signs the message with RSA-SHA256 (PKCS#1 v1.5) and returns
// the base64 signature — the only signature scheme APIv3 defines.
func SignSHA256RSA(key *rsa.PrivateKey, message string) (string, error) {
	digest := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("payment: sign: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// BuildAuthorizationHeader renders the Authorization header value APIv3
// expects; the quotes are part of the wire format, and the token order is
// the documented one.
func BuildAuthorizationHeader(mchID, serialNo string, timestamp int64, nonce, b64Signature string) string {
	return fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%d",serial_no="%s"`,
		mchID, nonce, b64Signature, timestamp, serialNo)
}

// BuildVerifyMessage assembles the three-part string the channel signs on
// every response and notification: timestamp, nonce and body,
// newline-terminated.
func BuildVerifyMessage(timestamp int64, nonce string, body []byte) string {
	return fmt.Sprintf("%d\n%s\n%s\n", timestamp, nonce, body)
}

// VerifySHA256RSA checks a channel signature (base64, RSA-SHA256 PKCS#1
// v1.5) over the message with the platform public key in hand. The caller
// maps the failure onto the sentinel (ErrInvalidSignature) — the crypto
// layer stays primitive-shaped.
func VerifySHA256RSA(pub *rsa.PublicKey, message, b64Signature string) error {
	sig, err := base64.StdEncoding.DecodeString(b64Signature)
	if err != nil {
		return fmt.Errorf("payment: bad signature encoding: %w", err)
	}
	digest := sha256.Sum256([]byte(message))
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], sig)
}

// DecryptAES256GCM decrypts a notification or certificate resource: key is
// the 32-byte APIv3 key, nonce and associatedData arrive inside the
// resource object, ciphertext is base64. Empty associatedData (platform
// certificates use "") is a legal value distinct from absent.
func DecryptAES256GCM(key []byte, nonce, associatedData, b64Ciphertext string) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	ct, err := base64.StdEncoding.DecodeString(b64Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("payment: bad ciphertext encoding: %w", err)
	}
	plaintext, err := gcm.Open(nil, []byte(nonce), ct, []byte(associatedData))
	if err != nil {
		return nil, fmt.Errorf("payment: gcm open: %w", err)
	}
	return plaintext, nil
}

// EncryptAES256GCM is DecryptAES256GCM's inverse — the mock provider's
// minting side. A notification forged here is byte-compatible with the real
// channel's shape, which is exactly why dev parsing equals production
// parsing.
func EncryptAES256GCM(key, nonce []byte, associatedData string, plaintext []byte) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(gcm.Seal(nil, nonce, plaintext, []byte(associatedData))), nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("payment: gcm cipher: %w", err)
	}
	return cipher.NewGCM(block)
}

// ParsePrivateKeyPEM reads a merchant private key PEM — PKCS#1 first (the
// format WeChat issues), PKCS#8 as the operator-convenience fallback.
func ParsePrivateKeyPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, stderrors.New("payment: no PEM block in private key")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("payment: parse private key: %w", err)
	}
	rk, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, stderrors.New("payment: private key is not RSA")
	}
	return rk, nil
}

// ParsePublicKeyPEM reads a bare RSA public key PEM (SubjectPublicKeyInfo
// block) — the shape the channel's downloadable wechatpay public key ships
// in, which unlike a certificate carries no serial or expiry of its own.
func ParsePublicKeyPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, stderrors.New("payment: no PEM block in public key")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("payment: parse public key: %w", err)
	}
	rk, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, stderrors.New("payment: public key is not RSA")
	}
	return rk, nil
}

// ParseCertificatePEM reads an X.509 certificate PEM (merchant or platform).
func ParseCertificatePEM(pemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, stderrors.New("payment: no PEM block in certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("payment: parse certificate: %w", err)
	}
	return cert, nil
}

// CertificateSerial renders the certificate serial as the uppercase hex the
// protocol carries in serial_no / Wechatpay-Serial headers.
func CertificateSerial(cert *x509.Certificate) string {
	return strings.ToUpper(cert.SerialNumber.Text(16))
}
