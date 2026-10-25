package mock

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"cyber-ecosystem/shared-go/capability/payment"
)

// The mini wire protocol mirrors the interface-boundary behavior of a real
// channel notification, not its internal cryptography: a body carrying an
// encrypted payload, a signature header authenticating the body, and the
// three failure shapes (bad signature, undecryptable payload, malformed
// envelope) the contract's sentinels name. Mock notifications therefore
// exercise a consumer's full parse path — verify, decrypt, decode, event.
const (
	signatureHeader = "X-Mockpay-Signature"
	wireVersion     = 1
)

type wireEnvelope struct {
	V       int    `json:"v"`
	Nonce   string `json:"nonce"`
	Payload string `json:"payload"`
}

func (b *backend) sign(body []byte) string {
	mac := hmac.New(sha256.New, b.hmacKey)
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (b *backend) seal(event *payment.NotificationEvent) (*payment.NotificationInput, error) {
	plaintext, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("mock: marshal event: %w", err)
	}
	block, err := aes.NewCipher(b.aead)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	envelope := wireEnvelope{
		V:       wireVersion,
		Nonce:   base64.StdEncoding.EncodeToString(nonce),
		Payload: base64.StdEncoding.EncodeToString(gcm.Seal(nil, nonce, plaintext, nil)),
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	header := http.Header{}
	header.Set(signatureHeader, b.sign(body))
	return &payment.NotificationInput{Headers: header, Body: body}, nil
}

// Parse verifies then decrypts then decodes — the same ladder a real
// channel's Notifier walks, mapping each failure onto the contract's
// sentinels.
func (b *backend) Parse(_ context.Context, in *payment.NotificationInput) (*payment.NotificationEvent, error) {
	if err := payment.ValidateNotificationInput(in); err != nil {
		return nil, err
	}
	if in.Headers.Get(signatureHeader) == "" {
		return nil, fmt.Errorf("%w: missing %s", payment.ErrInvalidSignature, signatureHeader)
	}
	if !hmac.Equal([]byte(in.Headers.Get(signatureHeader)), []byte(b.sign(in.Body))) {
		return nil, payment.ErrInvalidSignature
	}
	var envelope wireEnvelope
	if err := json.Unmarshal(in.Body, &envelope); err != nil {
		return nil, fmt.Errorf("%w: malformed envelope: %w", payment.ErrInvalidArgument, err)
	}
	nonce, err := base64.StdEncoding.DecodeString(envelope.Nonce)
	if err != nil {
		return nil, fmt.Errorf("%w: malformed nonce: %w", payment.ErrInvalidArgument, err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		return nil, fmt.Errorf("%w: malformed payload: %w", payment.ErrInvalidArgument, err)
	}
	block, err := aes.NewCipher(b.aead)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, payment.ErrDecryptFailed
	}
	var event payment.NotificationEvent
	if err := json.Unmarshal(plaintext, &event); err != nil {
		return nil, fmt.Errorf("%w: malformed event: %w", payment.ErrInvalidArgument, err)
	}
	return &event, nil
}
