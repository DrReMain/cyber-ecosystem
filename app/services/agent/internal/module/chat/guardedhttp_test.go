package chat

// The outbound guard's counterexamples: every reserved address class is
// refused at dial time (before any connection exists, so the probes need
// no listener and leave no network footprint), private unicast passes —
// the fully-private deployment is a first-class scenario — redirects are
// refused, and the body cap rejects over-cap streams while accepting a
// body that ends exactly at the cap.

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBlockedAddrClass(t *testing.T) {
	blocked := []string{
		"127.0.0.1", "::1", // loopback: the service's own stack
		"169.254.169.254", "fe80::1", // link-local: host gateways and cloud metadata
		"0.0.0.0", "::", // unspecified
		"224.0.0.1", "ff02::1", // multicast
	}
	allowed := []string{
		"192.168.50.140", "10.1.2.3", "172.20.1.5", // private unicast: LAN LLM endpoints
		"8.8.8.8", "2606:4700::1111", // public
	}
	for _, s := range blocked {
		if !blockedAddrClass(net.ParseIP(s)) {
			t.Errorf("%s should be blocked", s)
		}
	}
	for _, s := range allowed {
		if blockedAddrClass(net.ParseIP(s)) {
			t.Errorf("%s should be allowed", s)
		}
	}
}

func TestGuardedClientRefusesReservedDials(t *testing.T) {
	client := newGuardedHTTPClient(time.Second, 1<<20)
	for _, target := range []string{
		"http://127.0.0.1:9/models",
		"http://169.254.169.254/latest/meta-data",
		"http://[::1]:9/models",
		"http://0.0.0.0:9/models",
	} {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			t.Fatalf("dial to %s should be refused", target)
		}
		if !strings.Contains(err.Error(), errBlockedDial.Error()) {
			t.Fatalf("dial to %s: wrong refusal reason: %v", target, err)
		}
	}
}

func lanListener(t *testing.T) net.Listener {
	t.Helper()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Skipf("interface addresses unavailable: %v", err)
	}
	for _, a := range addrs {
		ipn, ok := a.(*net.IPNet)
		if !ok || ipn.IP.IsLoopback() || ipn.IP.To4() == nil || blockedAddrClass(ipn.IP) {
			continue
		}
		l, lerr := net.Listen("tcp", net.JoinHostPort(ipn.IP.String(), "0"))
		if lerr != nil {
			continue
		}
		return l
	}
	t.Skip("no non-loopback interface address — private-dial positive control unavailable")
	return nil
}

func TestGuardedClientAllowsPrivateDial(t *testing.T) {
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"test-model"}]}`)
	}))
	l := lanListener(t)
	srv.Listener = l
	srv.Start()
	defer srv.Close()

	client := newGuardedHTTPClient(5*time.Second, 1<<20)
	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("private dial should pass: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), "test-model") {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestGuardedClientRefusesRedirects(t *testing.T) {
	client := newGuardedHTTPClient(time.Second, 1<<20)
	if client.CheckRedirect == nil {
		t.Fatal("redirect policy missing")
	}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://192.168.50.140:1234/v1", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := client.CheckRedirect(req, nil); !errors.Is(err, http.ErrUseLastResponse) {
		t.Fatalf("redirect policy: %v", err)
	}
}

func TestCappedBodyRejectsOverCapAcceptsAtCap(t *testing.T) {
	over := &cappedBody{rc: io.NopCloser(strings.NewReader(strings.Repeat("a", 100))), n: 50}
	_, err := io.ReadAll(over)
	if err == nil || !strings.Contains(err.Error(), errBodyOverCap.Error()) {
		t.Fatalf("over-cap read should fail with the cap error, got: %v", err)
	}

	atCap := &cappedBody{rc: io.NopCloser(strings.NewReader(strings.Repeat("a", 50))), n: 50}
	got, err := io.ReadAll(atCap)
	if err != nil {
		t.Fatalf("at-cap body should read clean: %v", err)
	}
	if len(got) != 50 {
		t.Fatalf("at-cap body length: %d", len(got))
	}
}
