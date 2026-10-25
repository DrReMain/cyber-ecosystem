package wechat

// The APIv3 HTTP core: every request leaves signed by the merchant key and
// every response arrives verified against a channel key held in one
// serial→public-key table — the configured wechatpay public key under its
// PUB_KEY_ID, platform certificates under their serials, both at once
// through a mode switch. The channel names its signer in Wechatpay-Serial,
// so the table makes either mode verifiable without configuration churn.

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

// maxClockSkew bounds how far a signed response's timestamp may sit from
// local time — the replay window the protocol's anti-replay rule asks for.
const maxClockSkew = 5 * time.Minute

// requestTimeout is the transport's own ceiling so a missing context
// deadline can never hang a call; a shorter context deadline still wins.
const requestTimeout = 30 * time.Second

type client struct {
	cfg    Config
	mchKey *rsa.PrivateKey
	hc     *http.Client

	mu   sync.RWMutex
	pubs map[string]*rsa.PublicKey
}

func newClient(cfg Config) (*client, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", payment.ErrInvalidArgument, err)
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	key, err := ParsePrivateKeyPEM(cfg.PrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("%w: merchant key: %w", payment.ErrInvalidArgument, err)
	}
	c := &client{
		cfg:    cfg,
		mchKey: key,
		hc:     &http.Client{Timeout: requestTimeout},
		pubs:   make(map[string]*rsa.PublicKey),
	}
	if len(cfg.WechatpayPublicKeyPEM) > 0 {
		pub, err := ParsePublicKeyPEM(cfg.WechatpayPublicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("%w: wechatpay public key: %w", payment.ErrInvalidArgument, err)
		}
		// Stored under the upper-cased form the verification lookup uses —
		// a configured id in any case must still match its own headers.
		c.pubs[strings.ToUpper(cfg.WechatpayPublicKeyID)] = pub
	}
	return c, nil
}

// do runs one signed request: execute, verify the signed response, decode
// the body into out. A 2xx response without signature headers is refused —
// every APIv3 response is signed, so an unsigned one did not come from the
// channel's edge.
func (c *client) do(ctx context.Context, method, path string, query url.Values, body []byte, out any) error {
	h, respBody, err := c.execute(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	if err := c.verifySigned(ctx, h, respBody); err != nil {
		return err
	}
	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("%w: decode response: %w", payment.ErrUnavailable, err)
	}
	return nil
}

// execute performs the signed round trip and returns the response headers
// and raw body. Transport faults are unavailability; non-2xx becomes a
// classified sentinel.
func (c *client) execute(ctx context.Context, method, path string, query url.Values, body []byte) (http.Header, []byte, error) {
	full := path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+full, rd)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: build request: %w", payment.ErrInvalidArgument, err)
	}
	ts := time.Now().Unix()
	nonce, err := newNonce()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: nonce: %w", payment.ErrUnavailable, err)
	}
	sig, err := SignSHA256RSA(c.mchKey, BuildRequestMessage(method, full, ts, nonce, body))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: sign: %w", payment.ErrInvalidArgument, err)
	}
	req.Header.Set("Authorization", BuildAuthorizationHeader(c.cfg.MchID, c.cfg.MchSerial, ts, nonce, sig))
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.WechatpayPublicKeyID != "" {
		// Ask the channel to sign this response with the public key we
		// hold; without it a mid-switch merchant can be answered by a key
		// it has no way to learn.
		req.Header.Set("Wechatpay-Serial", c.cfg.WechatpayPublicKeyID)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", payment.ErrUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: read response: %w", payment.ErrUnavailable, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		a := &APIError{Status: resp.StatusCode, RequestID: resp.Header.Get("Request-ID")}
		var e struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &e) == nil {
			a.Code, a.Message = e.Code, e.Message
		}
		return nil, nil, classifyAPIError(a)
	}
	return resp.Header, respBody, nil
}

// verifySigned checks the Wechatpay-* signature headers over the response
// body: timestamp inside the replay window, the signer's serial known to
// the verifier table, signature valid. In certificate mode an unknown
// serial means the channel rotated — the table refreshes once and the
// lookup retries; in public-key mode an unknown serial is an anomaly and
// refuses outright.
func (c *client) verifySigned(ctx context.Context, h http.Header, body []byte) error {
	sig := h.Get("Wechatpay-Signature")
	if sig == "" {
		return fmt.Errorf("%w: response carries no signature", payment.ErrInvalidSignature)
	}
	ts, err := strconv.ParseInt(h.Get("Wechatpay-Timestamp"), 10, 64)
	if err != nil {
		return fmt.Errorf("%w: malformed timestamp", payment.ErrInvalidSignature)
	}
	if skew := time.Since(time.Unix(ts, 0)); skew > maxClockSkew || skew < -maxClockSkew {
		return fmt.Errorf("%w: timestamp outside replay window", payment.ErrInvalidSignature)
	}
	nonce := h.Get("Wechatpay-Nonce")
	serial := strings.ToUpper(h.Get("Wechatpay-Serial"))
	pub, ok := c.verifier(serial)
	if !ok {
		if c.cfg.WechatpayPublicKeyID != "" {
			return fmt.Errorf("%w: unknown signer %s", payment.ErrInvalidSignature, serial)
		}
		if err := c.refreshCertificates(ctx); err != nil {
			return err
		}
		pub, ok = c.verifier(serial)
		if !ok {
			return fmt.Errorf("%w: unknown signer %s", payment.ErrInvalidSignature, serial)
		}
	}
	if err := VerifySHA256RSA(pub, BuildVerifyMessage(ts, nonce, body), sig); err != nil {
		return fmt.Errorf("%w: %w", payment.ErrInvalidSignature, err)
	}
	return nil
}

func (c *client) verifier(serial string) (*rsa.PublicKey, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	pub, ok := c.pubs[serial]
	return pub, ok
}

// refreshCertificates pulls the channel's platform-certificate list and
// registers every verifier key. The response itself is verified after
// registration — the protocol's bootstrap answer to the chicken-and-egg of
// the certificate endpoint: decrypt and register first, then check the
// signature of the very bytes that carried the keys.
func (c *client) refreshCertificates(ctx context.Context) error {
	h, body, err := c.execute(ctx, http.MethodGet, "/v3/certificates", nil, nil)
	if err != nil {
		return err
	}
	var list struct {
		Data []struct {
			SerialNo           string `json:"serial_no"`
			EncryptCertificate struct {
				Nonce          string `json:"nonce"`
				AssociatedData string `json:"associated_data"`
				Ciphertext     string `json:"ciphertext"`
			} `json:"encrypt_certificate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &list); err != nil {
		return fmt.Errorf("%w: decode certificates: %w", payment.ErrUnavailable, err)
	}
	fresh := make(map[string]*rsa.PublicKey, len(list.Data))
	for _, d := range list.Data {
		pemBytes, err := DecryptAES256GCM(c.cfg.APIv3Key,
			d.EncryptCertificate.Nonce, d.EncryptCertificate.AssociatedData, d.EncryptCertificate.Ciphertext)
		if err != nil {
			return fmt.Errorf("%w: decrypt certificate %s: %w", payment.ErrDecryptFailed, d.SerialNo, err)
		}
		cert, err := ParseCertificatePEM(pemBytes)
		if err != nil {
			return fmt.Errorf("%w: parse certificate %s: %w", payment.ErrUnavailable, d.SerialNo, err)
		}
		pub, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("%w: certificate %s is not RSA", payment.ErrUnavailable, d.SerialNo)
		}
		fresh[strings.ToUpper(d.SerialNo)] = pub
	}
	if len(fresh) == 0 {
		return fmt.Errorf("%w: certificate list empty", payment.ErrUnavailable)
	}
	// The response carrying these certificates must verify under the table
	// that now includes them before any is trusted for later notifications;
	// a failed verification rolls back exactly what this refresh added (a
	// serial names one certificate, so a key the table already held is
	// restored, not dropped).
	c.mu.Lock()
	saved := make(map[string]*rsa.PublicKey, len(fresh))
	for k := range fresh {
		if old, ok := c.pubs[k]; ok {
			saved[k] = old
		}
	}
	maps.Copy(c.pubs, fresh)
	c.mu.Unlock()
	if err := c.verifySigned(ctx, h, body); err != nil {
		c.mu.Lock()
		for k := range fresh {
			if old, ok := saved[k]; ok {
				c.pubs[k] = old
			} else {
				delete(c.pubs, k)
			}
		}
		c.mu.Unlock()
		return err
	}
	return nil
}

func newNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
