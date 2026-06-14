// Tests in this file exercise the TLS path of the Connect transport.
//
// The main integration_test.go harness is h2c-only (cleartext HTTP/2). These
// tests bring up their OWN server+client pair configured for real TLS:
//   - server: connect.NewServer(connect.TLSConfig(serverConf), ...)
//     -> Start() does ServeTLS (server.go: tlsConf != nil branch).
//   - client: connect.Dial(WithEndpoint, WithTLSConfig(clientConf)) WITHOUT
//     WithH2C -> defaultRoundTripper builds a secure http.Transport with
//     TLSClientConfig (HTTP/2 is negotiated via ALPN, NOT h2c).
//
// This proves the full server+client TLS path works for both unary and
// streaming RPCs.

package integration

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	mobilepb "cyber-ecosystem/gen/go/cyber/mobile/v1"
	mobilev1connect "cyber-ecosystem/gen/go/cyber/mobile/v1/v1connect"
	v1 "cyber-ecosystem/gen/go/cyber/transfer/v1"
)

// tlsSelfSigned generates a self-signed certificate valid for 127.0.0.1 and
// localhost, returning two *tls.Config:
//   - serverConf: carries the certificate (used by connect.TLSConfig).
//   - clientConf: carries a RootCAs pool containing the cert so the client
//     trusts the self-signed server (used by connect.WithTLSConfig).
//
// Standard test-cert pattern: ECDSA key + x509 template with IP/DNS SANs,
// CreateCertificate to self-sign, tls.X509KeyPair to load it.
func tlsSelfSigned(t *testing.T) (serverConf, clientConf *tls.Config) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatalf("serial: %v", err)
	}

	now := time.Now()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "connect-tls-test"},
		NotBefore:    now.Add(-1 * time.Hour),
		NotAfter:     now.Add(1 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:     []string{"localhost"},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("key pair: %v", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(certPEM) {
		t.Fatalf("failed to append cert to pool")
	}

	serverConf = &tls.Config{Certificates: []tls.Certificate{pair}}
	clientConf = &tls.Config{RootCAs: pool}
	return serverConf, clientConf
}

// startTLSServer brings up a loopback Connect server serving real TLS (NOT h2c)
// with the test service registered, and returns a typed v1connect client built
// from connect.Dial WITH WithTLSConfig and WITHOUT WithH2C. The stop func tears
// everything down.
func startTLSServer(t *testing.T) (mobilev1connect.MobileTransferServiceClient, func()) {
	t.Helper()

	serverConf, clientConf := tlsSelfSigned(t)

	srv := connect.NewServer(
		connect.Address("127.0.0.1:0"),
		connect.TLSConfig(serverConf),
		connect.Timeout(0),
	)
	mobilepb.RegisterMobileTransferServiceConnectServer(srv, testService{})

	ep, err := srv.Endpoint()
	if err != nil {
		t.Fatalf("endpoint: %v", err)
	}

	ctx := context.Background()
	go func() { _ = srv.Start(ctx) }()
	waitReady(t, ep.Host)

	// NO WithH2C: a TLS client negotiates HTTP/2 via ALPN on a secure
	// http.Transport (see defaultRoundTripper: insecure && enableH2C is
	// false here, so it builds the TLS http.Transport branch).
	cli, err := connect.Dial(ctx,
		connect.WithEndpoint(ep.Host),
		connect.WithTLSConfig(clientConf),
	)
	if err != nil {
		_ = srv.Stop(ctx)
		t.Fatalf("dial: %v", err)
	}
	client := mobilev1connect.NewMobileTransferServiceClient(cli.HTTPClient(), cli.BaseURL(), cli.ClientOptions()...)

	stop := func() {
		_ = cli.Close()
		_ = srv.Stop(ctx)
	}
	return client, stop
}

// TestTLSUnary verifies the full TLS server+client path for a unary RPC: Raw
// over TLS succeeds and returns the correct data and content type.
func TestTLSUnary(t *testing.T) {
	cli, stop := startTLSServer(t)
	defer stop()

	resp, err := cli.Raw(context.Background(), connectrpc.NewRequest(&v1.RawRequest{
		ContentType: "text/plain",
		Data:        []byte("hello-tls"),
	}))
	if err != nil {
		t.Fatalf("Raw over TLS: %v", err)
	}
	if got := string(resp.Msg.Data); got != "hello-tls" {
		t.Fatalf("data = %q, want %q", got, "hello-tls")
	}
	if got := resp.Msg.GetContentType(); got != "text/plain" {
		t.Fatalf("content_type = %q, want %q", got, "text/plain")
	}
}

// TestTLSServerStream verifies the TLS path for a server-streaming RPC:
// Subscribe over TLS receives all 5 events. Streaming requires HTTP/2, which
// over TLS is negotiated via ALPN (not h2c) — proving the secure transport
// carries multiplexed streams correctly.
func TestTLSServerStream(t *testing.T) {
	cli, stop := startTLSServer(t)
	defer stop()

	stream, err := cli.Subscribe(context.Background(), connectrpc.NewRequest(&v1.SubscribeRequest{Topic: "tls"}))
	if err != nil {
		t.Fatalf("Subscribe over TLS: %v", err)
	}

	var got int
	wantFirst := "tls-1"
	var firstOK bool
	for stream.Receive() {
		got++
		if got == 1 && stream.Msg().GetEventId() == wantFirst {
			firstOK = true
		}
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream error over TLS: %v", err)
	}
	if got != 5 {
		t.Fatalf("received %d events over TLS, want 5", got)
	}
	if !firstOK {
		t.Fatalf("first event id not %q", wantFirst)
	}
}
