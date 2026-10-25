package wechat

// A fake channel exercises the HTTP core's full verdict surface: every
// response the real edge can produce — signed, unsigned, stale, rotated,
// refusing, congested — is minted here with self-generated keys, so the
// verification ladder and the error folding are pinned without touching the
// network.

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

const (
	testMchID     = "1900000109"
	testMchSerial = "TESTMCHSERIAL"
	testAppID     = "wx1234567890"
	testPubKeyID  = "PUB_KEY_ID_TEST"
)

func pubKeyPEM(t *testing.T, key *rsa.PrivateKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal pkix public key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

func freshNonce() string { // runtime value, 12 ASCII digits: arbitrary to AES-GCM, off the constant paths
	return fmt.Sprintf("%012d", time.Now().UnixNano()%1e12)
}

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

// fakeChannel is a stand-in APIv3 edge: it answers the certificate endpoint
// with an AES-GCM-sealed platform certificate, serves per-path routes when
// registered, and falls back to a configurable signed response.
type fakeChannel struct {
	t        *testing.T
	srv      *httptest.Server
	key      *rsa.PrivateKey
	serial   string
	certPEM  []byte
	apiv3Key []byte

	mu            sync.Mutex
	status        int    // data-plane status; 0 means 200
	body          string // data-plane body
	certBody      string // certificates endpoint body override; empty builds the sealed cert
	routes        map[string]fakeRoute
	requests      []capturedRequest
	unsigned      bool // drop the signature headers
	staleTS       bool // timestamp outside the replay window
	malformedTS   bool // timestamp not parseable as an integer
	wrongKey      bool // sign with a foreign key
	foreignSerial bool // announce a serial never registered
	checkAuth     bool // assert the request's Authorization shape
	certCalls     int
}

type fakeRoute struct {
	status int
	body   string
	ct     string
}

type capturedRequest struct {
	Method string
	Path   string
	Query  string
	Body   string
}

func newFakeChannel(t *testing.T) *fakeChannel {
	t.Helper()
	key := mustRSAKey(t)
	certPEM := mustCertPEM(t, key, big.NewInt(0x515800DEAD))
	cert, err := ParseCertificatePEM(certPEM)
	if err != nil {
		t.Fatalf("parse self-signed cert: %v", err)
	}
	f := &fakeChannel{
		t:        t,
		key:      key,
		serial:   CertificateSerial(cert),
		certPEM:  certPEM,
		apiv3Key: []byte("0123456789abcdef0123456789abcdef"),
	}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v3/certificates" {
			f.mu.Lock()
			f.certCalls++
			certBody := f.certBody
			f.mu.Unlock()
			if certBody == "" {
				nonce := freshNonce()
				ct, err := EncryptAES256GCM(f.apiv3Key, []byte(nonce), "", f.certPEM)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				certBody = mustMarshal(t, map[string]any{
					"data": []map[string]any{{
						"serial_no": f.serial,
						"encrypt_certificate": map[string]string{
							"nonce":           nonce,
							"associated_data": "",
							"ciphertext":      ct,
						},
					}},
				})
			}
			for k, vs := range f.signHeaders([]byte(certBody), time.Now().Unix()) {
				for _, v := range vs {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(certBody))
			return
		}
		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = io.ReadAll(r.Body)
		}
		f.mu.Lock()
		f.requests = append(f.requests, capturedRequest{r.Method, r.URL.Path, r.URL.RawQuery, string(reqBody)})
		status, body := f.status, f.body
		routeCT := ""
		if rt, ok := f.routes[r.Method+" "+r.URL.Path]; ok {
			status, body, routeCT = rt.status, rt.body, rt.ct
		}
		ts := time.Now().Unix()
		if f.staleTS {
			ts -= int64(maxClockSkew.Seconds()) * 2
		}
		unsigned, wrongKey, foreign, checkAuth := f.unsigned, f.wrongKey, f.foreignSerial, f.checkAuth
		f.mu.Unlock()
		// Mirror the protocol's signer selection: a request pinning
		// Wechatpay-Serial is answered under that same id, otherwise the
		// platform certificate's serial signs.
		signer := f.serial
		if keyID := r.Header.Get("Wechatpay-Serial"); keyID != "" {
			signer = keyID
		}
		if checkAuth {
			auth := r.Header.Get("Authorization")
			wantPrefix := fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s"`, testMchID)
			if len(auth) < len(wantPrefix) || auth[:len(wantPrefix)] != wantPrefix {
				f.t.Errorf("Authorization shape: want prefix %q, got %q", wantPrefix, auth)
			}
		}
		if status == 0 {
			status = http.StatusOK
		}
		if !unsigned {
			h := f.signHeaders([]byte(body), ts)
			if f.malformedTS {
				h.Set("Wechatpay-Timestamp", "not-a-timestamp")
			}
			h.Set("Wechatpay-Serial", signer)
			if foreign {
				h.Set("Wechatpay-Serial", "DEADBEEFDEADBEEF")
			}
			if wrongKey {
				other := mustRSAKey(f.t)
				msg := BuildVerifyMessage(ts, "testnonce123456", []byte(body))
				sig, err := SignSHA256RSA(other, msg)
				if err != nil {
					f.t.Errorf("foreign sign: %v", err)
				}
				h.Set("Wechatpay-Signature", sig)
			}
			for k, vs := range h {
				for _, v := range vs {
					w.Header().Add(k, v)
				}
			}
		}
		if routeCT != "" {
			w.Header().Set("Content-Type", routeCT)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeChannel) signHeaders(body []byte, ts int64) http.Header {
	msg := BuildVerifyMessage(ts, "testnonce123456", body)
	sig, err := SignSHA256RSA(f.key, msg)
	if err != nil {
		f.t.Errorf("channel sign: %v", err)
	}
	h := http.Header{}
	h.Set("Wechatpay-Timestamp", strconv.FormatInt(ts, 10))
	h.Set("Wechatpay-Nonce", "testnonce123456")
	h.Set("Wechatpay-Signature", sig)
	h.Set("Wechatpay-Serial", f.serial)
	return h
}

// pushHeaders signs a notification the way a public-key-mode channel pushes
// one: no request exists to pin the signer through, so the push itself
// arrives under the public key id.
func (f *fakeChannel) pushHeaders(body []byte, ts int64) http.Header {
	h := f.signHeaders(body, ts)
	h.Set("Wechatpay-Serial", testPubKeyID)
	h.Set("Wechatpay-Signature-Type", "WECHATPAY2-SHA256-RSA2048")
	return h
}

func (f *fakeChannel) publicKeyConfig(t *testing.T) Config {
	return Config{
		MchID:                 testMchID,
		MchSerial:             testMchSerial,
		PrivateKeyPEM:         pkcs1PEM(mustRSAKey(t)),
		APIv3Key:              f.apiv3Key,
		AppID:                 testAppID,
		BaseURL:               f.srv.URL,
		WechatpayPublicKeyID:  testPubKeyID,
		WechatpayPublicKeyPEM: pubKeyPEM(t, f.key),
	}
}

func (f *fakeChannel) route(method, path string, status int, body string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.routes == nil {
		f.routes = make(map[string]fakeRoute)
	}
	f.routes[method+" "+path] = fakeRoute{status: status, body: body}
}

// routeCT registers a route with an explicit Content-Type, the shape the
// bill-download path needs (gzip payload, or a JSON late error).
func (f *fakeChannel) routeCT(method, path string, status int, body, ct string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.routes == nil {
		f.routes = make(map[string]fakeRoute)
	}
	f.routes[method+" "+path] = fakeRoute{status: status, body: body, ct: ct}
}

func (f *fakeChannel) lastRequest(t *testing.T, method, path string) capturedRequest {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := len(f.requests) - 1; i >= 0; i-- {
		if f.requests[i].Method == method && f.requests[i].Path == path {
			return f.requests[i]
		}
	}
	t.Fatalf("no captured request for %s %s", method, path)
	return capturedRequest{}
}

func (f *fakeChannel) certificateConfig(t *testing.T) Config {
	cfg := f.publicKeyConfig(t)
	cfg.WechatpayPublicKeyID = ""
	cfg.WechatpayPublicKeyPEM = nil
	return cfg
}

func doJSON(ctx context.Context, c *client) (map[string]string, error) {
	var out map[string]string
	err := c.do(ctx, http.MethodGet, "/v3/anything", nil, nil, &out)
	return out, err
}

func TestNewClientConfig(t *testing.T) {
	f := newFakeChannel(t)
	good := f.publicKeyConfig(t)

	mutate := func(f func(*Config)) Config {
		cfg := good
		f(&cfg)
		return cfg
	}
	cases := []struct {
		name string
		cfg  Config
	}{
		{"missing_mchid", mutate(func(c *Config) { c.MchID = "" })},
		{"missing_mch_serial", mutate(func(c *Config) { c.MchSerial = "" })},
		{"missing_appid", mutate(func(c *Config) { c.AppID = "" })},
		{"short_apiv3_key", mutate(func(c *Config) { c.APIv3Key = []byte("short") })},
		{"id_without_pem", mutate(func(c *Config) { c.WechatpayPublicKeyPEM = nil })},
		{"pem_without_id", mutate(func(c *Config) { c.WechatpayPublicKeyID = "" })},
		{"bad_private_key", mutate(func(c *Config) { c.PrivateKeyPEM = []byte("not a pem") })},
		{"bad_public_key", mutate(func(c *Config) { c.WechatpayPublicKeyPEM = []byte("not a pem") })},
		{"ecdsa_public_key", mutate(func(c *Config) {
			ec, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			der, _ := x509.MarshalPKIXPublicKey(&ec.PublicKey)
			c.WechatpayPublicKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
		})},
		{"corrupt_public_key_der", mutate(func(c *Config) {
			c.WechatpayPublicKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte("not a key")})
		})},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := newClient(c.cfg); !errors.Is(err, payment.ErrInvalidArgument) {
				t.Fatalf("want ErrInvalidArgument, got %v", err)
			}
		})
	}

	c, err := newClient(mutate(func(c *Config) { c.BaseURL = "" }))
	if err != nil {
		t.Fatalf("empty BaseURL: %v", err)
	}
	if c.cfg.BaseURL != defaultBaseURL {
		t.Errorf("BaseURL default: want %s, got %s", defaultBaseURL, c.cfg.BaseURL)
	}
	if pub, ok := c.verifier(testPubKeyID); !ok || pub == nil {
		t.Errorf("public key not registered under %s", testPubKeyID)
	}
}

func TestClientHappyPathPublicKeyMode(t *testing.T) {
	f := newFakeChannel(t)
	f.checkAuth = true
	f.body = `{"prepay_id":"wx123"}`
	c, err := newClient(f.publicKeyConfig(t))
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	out, err := doJSON(context.Background(), c)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if out["prepay_id"] != "wx123" {
		t.Errorf("decoded body: %v", out)
	}
}

func TestClientNoContentResponse(t *testing.T) {
	f := newFakeChannel(t)
	f.status = http.StatusNoContent
	f.body = ""
	c, err := newClient(f.publicKeyConfig(t))
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	if err := c.do(context.Background(), http.MethodPost, "/v3/anything", nil, nil, nil); err != nil {
		t.Fatalf("204 round trip: %v", err)
	}
}

func TestClientVerificationFailures(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*fakeChannel)
	}{
		{"unsigned_response", func(f *fakeChannel) { f.unsigned = true }},
		{"foreign_signature", func(f *fakeChannel) { f.wrongKey = true }},
		{"stale_timestamp", func(f *fakeChannel) { f.staleTS = true }},
		{"unknown_serial_no_refresh", func(f *fakeChannel) { f.foreignSerial = true }},
		{"malformed_timestamp", func(f *fakeChannel) {
			f.malformedTS = true
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFakeChannel(t)
			c.mut(f)
			f.body = `{}`
			cl, err := newClient(f.publicKeyConfig(t))
			if err != nil {
				t.Fatalf("newClient: %v", err)
			}
			if _, err := doJSON(context.Background(), cl); !errors.Is(err, payment.ErrInvalidSignature) {
				t.Fatalf("want ErrInvalidSignature, got %v", err)
			}
			f.mu.Lock()
			calls := f.certCalls
			f.mu.Unlock()
			if calls != 0 {
				t.Errorf("public-key mode must not fetch certificates, %d calls", calls)
			}
		})
	}
}

func TestClientCertificateBootstrap(t *testing.T) {
	f := newFakeChannel(t)
	f.body = `{"ok":"true"}`
	c, err := newClient(f.certificateConfig(t))
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	out, err := doJSON(context.Background(), c)
	if err != nil {
		t.Fatalf("do with lazy bootstrap: %v", err)
	}
	if out["ok"] != "true" {
		t.Errorf("decoded body: %v", out)
	}
	f.mu.Lock()
	calls := f.certCalls
	f.mu.Unlock()
	if calls == 0 {
		t.Error("unknown serial should have triggered one certificate fetch")
	}
	if pub, ok := c.verifier(f.serial); !ok || pub == nil {
		t.Errorf("platform certificate not registered under %s", f.serial)
	}
}

func TestClientAPIErrorClassification(t *testing.T) {
	sentinel := func(want error) func(*testing.T, error) {
		return func(t *testing.T, err error) {
			if !errors.Is(err, want) {
				t.Fatalf("want %v, got %v", want, err)
			}
		}
	}
	cases := []struct {
		name   string
		status int
		body   string
		check  func(*testing.T, error)
	}{
		{"400_param_error", 400, `{"code":"PARAM_ERROR","message":"invalid amount"}`, sentinel(payment.ErrInvalidArgument)},
		{"401_sign_error", 401, `{"code":"SIGN_ERROR","message":"signature mismatch"}`, sentinel(payment.ErrInvalidArgument)},
		{"403_trade_no_used", 403, `{"code":"OUT_TRADE_NO_USED","message":"duplicate"}`, sentinel(payment.ErrDuplicate)},
		{"403_no_auth", 403, `{"code":"NO_AUTH","message":"no permission"}`, func(t *testing.T, err error) {
			var ce *payment.ChannelError
			if !errors.As(err, &ce) || ce.Code != "NO_AUTH" {
				t.Fatalf("want ChannelError NO_AUTH, got %v", err)
			}
			if !errors.Is(err, payment.ErrChannelRejected) {
				t.Fatalf("ChannelError must unwrap to ErrChannelRejected, got %v", err)
			}
		}},
		{"404_order_not_exists", 404, `{"code":"ORDER_NOT_EXISTS","message":"no order"}`, sentinel(payment.ErrNotFound)},
		{"429_frequency_limited", 429, `{"code":"FREQUENCY_LIMITED","message":"slow down"}`, sentinel(payment.ErrUnavailable)},
		{"500_system_error", 500, `{"code":"SYSTEM_ERROR","message":"channel busy"}`, sentinel(payment.ErrUnavailable)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFakeChannel(t)
			f.status, f.body = c.status, c.body
			cl, err := newClient(f.publicKeyConfig(t))
			if err != nil {
				t.Fatalf("newClient: %v", err)
			}
			_, err = doJSON(context.Background(), cl)
			if err == nil {
				t.Fatal("want error, got nil")
			}
			c.check(t, err)
		})
	}
}

func TestClientTransportFault(t *testing.T) {
	f := newFakeChannel(t)
	c, err := newClient(f.publicKeyConfig(t))
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	f.srv.Close()
	if _, err := doJSON(context.Background(), c); !errors.Is(err, payment.ErrUnavailable) {
		t.Fatalf("want ErrUnavailable, got %v", err)
	}
}

func TestClientUndecodable2xx(t *testing.T) {
	f := newFakeChannel(t)
	f.body = "not json at all"
	c, err := newClient(f.publicKeyConfig(t))
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	if _, err := doJSON(context.Background(), c); !errors.Is(err, payment.ErrUnavailable) {
		t.Fatalf("want ErrUnavailable, got %v", err)
	}
}

func TestAPIErrorMessage(t *testing.T) {
	a := &APIError{Status: 400, Code: "PARAM_ERROR", Message: "invalid amount"}
	if s := a.Error(); !strings.Contains(s, "400") || !strings.Contains(s, "PARAM_ERROR") {
		t.Errorf("Error(): %s", s)
	}
	a.RequestID = "req-1"
	if s := a.Error(); !strings.Contains(s, "req-1") {
		t.Errorf("Error() with request id: %s", s)
	}
}

// TestClientRefreshFailureLadder walks the certificate bootstrap's failure
// shapes in certificate mode: a body that is not the certificate list, a
// payload sealed under a different APIv3 key, and a well-formed but empty
// list.
func TestClientRefreshFailureLadder(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*fakeChannel, *Config)
		want  error
	}{
		{"undecodable_certificates", func(f *fakeChannel, c *Config) { f.certBody = "not json" }, payment.ErrUnavailable},
		{"wrong_apiv3_key", func(f *fakeChannel, c *Config) {
			c.APIv3Key = []byte("ffffffffffffffffffffffffffffffff")
		}, payment.ErrDecryptFailed},
		{"empty_certificate_list", func(f *fakeChannel, c *Config) { f.certBody = `{"data":[]}` }, payment.ErrUnavailable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFakeChannel(t)
			f.body = `{}`
			cfg := f.certificateConfig(t)
			c.setup(f, &cfg)
			cl, err := newClient(cfg)
			if err != nil {
				t.Fatalf("newClient: %v", err)
			}
			if _, err := doJSON(context.Background(), cl); !errors.Is(err, c.want) {
				t.Fatalf("want %v, got %v", c.want, err)
			}
		})
	}

	// A seal that opens into something that is not a certificate refuses
	// without registering anything.
	t.Run("decrypted_non_certificate", func(t *testing.T) {
		f := newFakeChannel(t)
		f.body = `{}`
		nonce := freshNonce()
		garbage, err := EncryptAES256GCM(f.apiv3Key, []byte(nonce), "", []byte("garbage"))
		if err != nil {
			t.Fatalf("seal garbage: %v", err)
		}
		f.certBody = `{"data":[{"serial_no":"` + f.serial + `","encrypt_certificate":{"nonce":"` + nonce + `","associated_data":"","ciphertext":"` + garbage + `"}}]}`
		cl, err := newClient(f.certificateConfig(t))
		if err != nil {
			t.Fatalf("newClient: %v", err)
		}
		if _, err := doJSON(context.Background(), cl); !errors.Is(err, payment.ErrUnavailable) {
			t.Fatalf("want ErrUnavailable, got %v", err)
		}
	})
}

// A refreshed table that still cannot name the response's signer stays a
// refusal: the bootstrap ran and the serial remains unknown.
func TestClientRefreshLeavesUnknownSigner(t *testing.T) {
	f := newFakeChannel(t)
	f.body = `{}`
	f.foreignSerial = true
	cl, err := newClient(f.certificateConfig(t))
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	if _, err := doJSON(context.Background(), cl); !errors.Is(err, payment.ErrInvalidSignature) {
		t.Fatalf("want ErrInvalidSignature, got %v", err)
	}
	f.mu.Lock()
	calls := f.certCalls
	f.mu.Unlock()
	if calls == 0 {
		t.Error("unknown serial should still have triggered the certificate fetch")
	}
}

func TestEncryptShortKey(t *testing.T) {
	if _, err := EncryptAES256GCM([]byte("short"), []byte(freshNonce()), "", []byte("x")); err == nil {
		t.Fatal("want error for a key of wrong length")
	}
}
