package wechat

// Live handshake against the real APIv3 edge; skipped unless the merchant
// variables are set. The probe is the CP1 gate: one fully signed round trip
// proves both directions — the channel accepted the merchant signature (any
// answer but 401), and the answer itself verifies against the channel key
// in hand. The verdict tree treats a verifiable 404 as a pass: a
// public-key-mode merchant owns no platform certificates, so the certificate
// endpoint answering "none exist" is a healthy round trip, not a failure.
//
//	WECHAT_PAY_CONFORMANCE_MCHID=... \
//	WECHAT_PAY_CONFORMANCE_MCH_SERIAL=... \
//	WECHAT_PAY_CONFORMANCE_APPID=wx... \
//	WECHAT_PAY_CONFORMANCE_APIV3_KEY=<32 bytes> \
//	WECHAT_PAY_CONFORMANCE_PRIVATE_KEY_FILE=/path/apiclient_key.pem \
//	WECHAT_PAY_CONFORMANCE_PUBLIC_KEY_ID=PUB_KEY_ID_... \
//	WECHAT_PAY_CONFORMANCE_PUBLIC_KEY_FILE=/path/pub_key.pem \
//	go test ./shared-go/capability/payment/wechat/ -run TestLiveConformance -count=1 -v
//
// The public-key pair is the mode new merchants get; without it the probe
// runs certificate mode, where the bootstrap itself is the handshake.
// Values come from the merchant console — never from the repo.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

func conformanceConfig(t *testing.T) (Config, bool) {
	t.Helper()
	cfg := Config{
		MchID:     os.Getenv("WECHAT_PAY_CONFORMANCE_MCHID"),
		MchSerial: os.Getenv("WECHAT_PAY_CONFORMANCE_MCH_SERIAL"),
		AppID:     os.Getenv("WECHAT_PAY_CONFORMANCE_APPID"),
		APIv3Key:  []byte(os.Getenv("WECHAT_PAY_CONFORMANCE_APIV3_KEY")),
		BaseURL:   os.Getenv("WECHAT_PAY_CONFORMANCE_BASE_URL"),
	}
	if cfg.MchID == "" {
		return Config{}, false
	}
	keyPath := os.Getenv("WECHAT_PAY_CONFORMANCE_PRIVATE_KEY_FILE")
	if cfg.MchSerial == "" || cfg.AppID == "" || len(cfg.APIv3Key) == 0 || keyPath == "" {
		t.Fatal("WECHAT_PAY_CONFORMANCE_* incomplete: MCH_SERIAL, APPID, APIV3_KEY and PRIVATE_KEY_FILE are required when MCHID is set")
	}
	//nolint:gosec // the key path is operator-supplied conformance input, never request data
	privateKey, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("read private key %s: %v", keyPath, err)
	}
	cfg.PrivateKeyPEM = privateKey
	if pubID := os.Getenv("WECHAT_PAY_CONFORMANCE_PUBLIC_KEY_ID"); pubID != "" {
		pubPath := os.Getenv("WECHAT_PAY_CONFORMANCE_PUBLIC_KEY_FILE")
		if pubPath == "" {
			t.Fatal("WECHAT_PAY_CONFORMANCE_PUBLIC_KEY_FILE is required when PUBLIC_KEY_ID is set")
		}
		//nolint:gosec // same operator-supplied conformance path
		pub, err := os.ReadFile(pubPath)
		if err != nil {
			t.Fatalf("read public key %s: %v", pubPath, err)
		}
		cfg.WechatpayPublicKeyID = pubID
		cfg.WechatpayPublicKeyPEM = pub
	}
	return cfg, true
}

func TestLiveConformance(t *testing.T) {
	cfg, ok := conformanceConfig(t)
	if !ok {
		t.Skip("WECHAT_PAY_CONFORMANCE_MCHID not set; live conformance is opt-in")
	}
	c, err := newClient(cfg)
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("signature-handshake", func(t *testing.T) {
		if cfg.WechatpayPublicKeyID == "" {
			// Certificate mode: the bootstrap is the handshake — a signed
			// GET, the sealed list decrypted, registered, and the very
			// response verified.
			if err := c.refreshCertificates(ctx); err != nil {
				t.Fatalf("certificate bootstrap (the handshake): %v", err)
			}
			return
		}
		// Public-key mode: sign one GET by hand and judge the tree —
		// unreachable fails, unverifiable fails, verifiable-but-401 fails
		// (the channel refused our signature), anything else passes.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+"/v3/certificates", nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}
		ts := time.Now().Unix()
		nonce, err := newNonce()
		if err != nil {
			t.Fatalf("nonce: %v", err)
		}
		sig, err := SignSHA256RSA(c.mchKey, BuildRequestMessage(http.MethodGet, "/v3/certificates", ts, nonce, nil))
		if err != nil {
			t.Fatalf("sign: %v", err)
		}
		req.Header.Set("Authorization", BuildAuthorizationHeader(c.cfg.MchID, c.cfg.MchSerial, ts, nonce, sig))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Wechatpay-Serial", c.cfg.WechatpayPublicKeyID)
		resp, err := c.hc.Do(req)
		if err != nil {
			t.Fatalf("round trip: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read response: %v", err)
		}
		t.Logf("handshake response: http %d, headers %v, body %s",
			resp.StatusCode, resp.Header, string(body[:min(len(body), 400)]))
		if err := c.verifySigned(ctx, resp.Header, body); err != nil {
			t.Fatalf("response does not verify against the channel key: %v", err)
		}
		if resp.StatusCode == http.StatusUnauthorized {
			t.Fatalf("channel refused the merchant signature (http 401) — check mchid, serial and key")
		}
		t.Logf("handshake verified: http %d from %s", resp.StatusCode, c.cfg.BaseURL)
	})

	t.Run("absent-order-folds-notfound", func(t *testing.T) {
		absent := fmt.Sprintf("conf%d", time.Now().UnixNano())
		err := c.do(ctx, http.MethodGet,
			"/v3/pay/transactions/out-trade-no/"+absent,
			url.Values{"mchid": {c.cfg.MchID}}, nil, nil)
		if !errors.Is(err, payment.ErrNotFound) {
			t.Fatalf("absent order: want ErrNotFound, got %v", err)
		}
	})
}
