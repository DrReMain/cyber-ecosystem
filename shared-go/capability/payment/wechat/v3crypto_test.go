package wechat

// Self-generated RSA keypairs and self-signed certificates as fixtures: the
// cryptographic core is protocol-shaped, not instance-shaped — any RSA-2048
// key and any X.509 certificate exercise it identically. Round-trips prove
// the framing bytes; the negative cases prove the rejections a real peer
// (or an attacker) produces.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	return key
}

func mustCertPEM(t *testing.T, key *rsa.PrivateKey, serial *big.Int) []byte {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "payment-conformance"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func pkcs1PEM(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}

func pkcs8PEM(t *testing.T, key any) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal pkcs8: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

// TestBuildRequestMessage pins the exact framing: five newline-terminated
// parts, with a GET's empty body leaving two trailing newlines — the
// detail a wrong implementation gets wrong first.
func TestBuildRequestMessage(t *testing.T) {
	got := BuildRequestMessage("GET", "/v3/certificates", 1639731960, "abc123", nil)
	want := "GET\n/v3/certificates\n1639731960\nabc123\n\n"
	if got != want {
		t.Errorf("GET empty body:\n want %q\n got  %q", want, got)
	}
	got = BuildRequestMessage("POST", "/v3/pay/transactions/jsapi?x=1", 1639731960, "n", []byte(`{"a":1}`))
	want = "POST\n/v3/pay/transactions/jsapi?x=1\n1639731960\nn\n" + `{"a":1}` + "\n"
	if got != want {
		t.Errorf("POST body:\n want %q\n got  %q", want, got)
	}
}

func TestBuildVerifyMessage(t *testing.T) {
	got := BuildVerifyMessage(1639731960, "nonce", []byte(`{"code":"SUCCESS"}`))
	want := "1639731960\nnonce\n" + `{"code":"SUCCESS"}` + "\n"
	if got != want {
		t.Errorf("verify message:\n want %q\n got  %q", want, got)
	}
}

// TestBuildAuthorizationHeader pins the wire format including the quotes and
// the documented token order.
func TestBuildAuthorizationHeader(t *testing.T) {
	got := BuildAuthorizationHeader("1900000001", "ABCDEF123", 1639731960, "n0nce", "c2ln==")
	want := `WECHATPAY2-SHA256-RSA2048 mchid="1900000001",nonce_str="n0nce",signature="c2ln==",timestamp="1639731960",serial_no="ABCDEF123"`
	if got != want {
		t.Errorf("authorization header:\n want %q\n got  %q", want, got)
	}
}

func TestSignVerifyRoundTrip(t *testing.T) {
	key := mustRSAKey(t)
	msg := BuildVerifyMessage(1, "n", []byte("body"))
	sig, err := SignSHA256RSA(key, msg)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if err := VerifySHA256RSA(&key.PublicKey, msg, sig); err != nil {
		t.Fatalf("verify round trip: %v", err)
	}
	if err := VerifySHA256RSA(&key.PublicKey, msg+"x", sig); err == nil {
		t.Error("tampered message must fail verification")
	}
	if err := VerifySHA256RSA(&key.PublicKey, msg, sig[:len(sig)-4]+"AAAA"); err == nil {
		t.Error("tampered signature must fail verification")
	}
	other := mustRSAKey(t)
	if err := VerifySHA256RSA(&other.PublicKey, msg, sig); err == nil {
		t.Error("wrong public key must fail verification")
	}
	if err := VerifySHA256RSA(&key.PublicKey, msg, "not base64 !!"); err == nil {
		t.Error("non-base64 signature must fail")
	}
}

func TestGCMRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes
	for _, aad := range []string{"", "transaction"} {
		nonce := freshNonce()
		ct, err := EncryptAES256GCM(key, []byte(nonce), aad, []byte(`{"out_trade_no":"ord-1"}`))
		if err != nil {
			t.Fatalf("encrypt (aad=%q): %v", aad, err)
		}
		pt, err := DecryptAES256GCM(key, nonce, aad, ct)
		if err != nil {
			t.Fatalf("decrypt (aad=%q): %v", aad, err)
		}
		if string(pt) != `{"out_trade_no":"ord-1"}` {
			t.Errorf("plaintext mismatch (aad=%q): %s", aad, pt)
		}
	}

	nonce := freshNonce()
	ct, err := EncryptAES256GCM(key, []byte(nonce), "transaction", []byte("x"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := DecryptAES256GCM(key, nonce, "other", ct); err == nil {
		t.Error("wrong associated data must fail")
	}
	if _, err := DecryptAES256GCM(key, "nonce9999999", "transaction", ct); err == nil {
		t.Error("wrong nonce must fail")
	}
	if _, err := DecryptAES256GCM([]byte("another-32-byte-key-another-32-byt"), "nonce1234567", "transaction", ct); err == nil {
		t.Error("wrong key must fail")
	}
	if _, err := DecryptAES256GCM(key, "nonce1234567", "transaction", "not base64 !!"); err == nil {
		t.Error("non-base64 ciphertext must fail")
	}
	raw, err := base64.StdEncoding.DecodeString(ct)
	if err != nil {
		t.Fatalf("decode self-produced ciphertext: %v", err)
	}
	raw[0] ^= 0xFF
	tampered := base64.StdEncoding.EncodeToString(raw)
	if _, err := DecryptAES256GCM(key, "nonce1234567", "transaction", tampered); err == nil {
		t.Error("tampered ciphertext must fail (GCM authentication)")
	}
	if _, err := DecryptAES256GCM([]byte("short"), "nonce1234567", "transaction", ct); err == nil {
		t.Error("short key must fail")
	}
}

func TestParsePrivateKeyPEM(t *testing.T) {
	key := mustRSAKey(t)
	parsed, err := ParsePrivateKeyPEM(pkcs1PEM(key))
	if err != nil {
		t.Fatalf("pkcs1 round trip: %v", err)
	}
	if parsed.Equal(key) != true {
		t.Error("pkcs1 parsed key differs from original")
	}
	parsed, err = ParsePrivateKeyPEM(pkcs8PEM(t, key))
	if err != nil {
		t.Fatalf("pkcs8 round trip: %v", err)
	}
	if !parsed.Equal(key) {
		t.Error("pkcs8 parsed key differs from original")
	}
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ecdsa key: %v", err)
	}
	if _, err := ParsePrivateKeyPEM(pkcs8PEM(t, ecKey)); err == nil {
		t.Error("non-RSA key must fail")
	}
	if _, err := ParsePrivateKeyPEM([]byte("not a pem")); err == nil {
		t.Error("garbage input must fail")
	}
	if _, err := ParsePrivateKeyPEM(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: []byte("garbage")})); err == nil {
		t.Error("garbage PEM body must fail")
	}
}

func TestParseCertificateAndSerial(t *testing.T) {
	key := mustRSAKey(t)
	serial := big.NewInt(0xABCDEF)
	pemBytes := mustCertPEM(t, key, serial)
	cert, err := ParseCertificatePEM(pemBytes)
	if err != nil {
		t.Fatalf("certificate round trip: %v", err)
	}
	// Uppercase hex — the shape serial_no headers carry.
	if got := CertificateSerial(cert); got != "ABCDEF" {
		t.Errorf("serial: want %q, got %q", "ABCDEF", got)
	}
	if _, err := ParseCertificatePEM([]byte("not a pem")); err == nil {
		t.Error("garbage certificate input must fail")
	}
	if _, err := ParseCertificatePEM(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("garbage")})); err == nil {
		t.Error("garbage certificate body must fail")
	}
}
