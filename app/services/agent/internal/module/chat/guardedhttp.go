package chat

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"
)

const (
	modelsBodyCap  = 4 << 20  // a model list is a few KB; MBs mean hostility
	streamBodyCap  = 32 << 20 // one SSE conversation; deltas are tiny, so this is generous
	dialTimeout    = 10 * time.Second
	keepAliveIdle  = 30 * time.Second
	idleConnWindow = 90 * time.Second
	tlsHandshake   = 10 * time.Second
	expectContinue = 1 * time.Second
)

var (
	errBlockedDial = errors.New("outbound dial refused: reserved address class")
	errBodyOverCap = errors.New("response body exceeds the outbound cap")
)

type cappedBody struct {
	rc io.ReadCloser
	n  int64
}

func (b *cappedBody) Read(p []byte) (int, error) {
	if b.n <= 0 {
		var probe [1]byte // exactly-at-cap is legitimate: one probe byte tells a finished stream from an over-cap one
		if m, perr := b.rc.Read(probe[:]); m == 0 && perr == io.EOF {
			return 0, io.EOF
		}
		return 0, errBodyOverCap
	}
	if int64(len(p)) > b.n {
		p = p[:b.n]
	}
	n, err := b.rc.Read(p)
	b.n -= int64(n)
	return n, err
}

func (b *cappedBody) Close() error { return b.rc.Close() }

type bodyCappingTransport struct {
	base    http.RoundTripper
	maxBody int64
}

func (t *bodyCappingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// The cap rides the body, not the buffer: a hostile provider's
	// unbounded read dies at this seam, never in the caller's heap.
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		resp.Body = &cappedBody{rc: resp.Body, n: t.maxBody}
	}
	return resp, nil
}

func blockedAddrClass(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsUnspecified()
}

func refuseReservedControl(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("%w: %s", errBlockedDial, address)
	}
	ip := net.ParseIP(host)
	if ip == nil || blockedAddrClass(ip) {
		return fmt.Errorf("%w: %s", errBlockedDial, host)
	}
	return nil
}

func newGuardedHTTPClient(timeout time.Duration, maxBody int64) *http.Client {
	// base_url is member input, so every dial is attacker-influenced: the
	// dialer refuses exactly the address classes the network stack itself
	// reserves, checked on the address actually connected — which is what
	// makes DNS rebinding irrelevant and IP-literal targets identical.
	// Deployment topology (which LAN, which IP, DNS or not) is never this
	// code's business, so nothing else is blocked: private unicast is a
	// first-class scenario.
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   dialTimeout,
			KeepAlive: keepAliveIdle,
			Control:   refuseReservedControl,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       idleConnWindow,
		TLSHandshakeTimeout:   tlsHandshake,
		ExpectContinueTimeout: expectContinue,
	}
	return &http.Client{
		Timeout: timeout,
		Transport: &bodyCappingTransport{
			base:    tr,
			maxBody: maxBody,
		},
		// A provider endpoint is not a redirect chain: a hop would carry
		// the Authorization header to a host the user never named.
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
