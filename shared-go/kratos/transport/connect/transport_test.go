package connect

import (
	"net/http"
	"testing"

	"github.com/go-kratos/kratos/v3/transport"
)

func TestTransportKind(t *testing.T) {
	tr := &Transport{operation: "/svc/Method"}
	if got := tr.Kind(); got != KindConnect {
		t.Fatalf("Kind = %v, want %v", got, KindConnect)
	}
}

func TestHeaderCarrierRoundtrip(t *testing.T) {
	h := headerCarrier(http.Header{})
	h.Set("x-a", "1")
	h.Add("x-a", "2")
	if got := h.Get("x-a"); got != "1" {
		t.Fatalf("Get = %q", got)
	}
	if vals := h.Values("x-a"); len(vals) != 2 {
		t.Fatalf("Values len = %d", len(vals))
	}
	if k := h.Keys(); len(k) != 1 {
		t.Fatalf("Keys len = %d", len(k))
	}
}

func TestKindConnectIsString(t *testing.T) {
	if transport.Kind(KindConnect).String() != "connect" {
		t.Fatalf("KindConnect string mismatch")
	}
}

func TestTransportAccessors(t *testing.T) {
	// Fully-populated Transport: every accessor returns the set field value,
	// and RequestHeader/ReplyHeader expose the carriers we constructed.
	tr := &Transport{
		endpoint:    "https://example.test/svc.Foo/Bar",
		operation:   "/svc.Foo/Bar",
		reqHeader:   headerCarrier(http.Header{"X-Req-A": {"rv"}}),
		replyHeader: headerCarrier(http.Header{"X-Rep-B": {"pv"}}),
		httpMethod:  "POST",
		remoteAddr:  "1.2.3.4:5678",
	}

	if got := tr.Endpoint(); got != "https://example.test/svc.Foo/Bar" {
		t.Errorf("Endpoint = %q, want the set endpoint", got)
	}
	if got := tr.Operation(); got != "/svc.Foo/Bar" {
		t.Errorf("Operation = %q, want the set operation", got)
	}
	if got := tr.HTTPMethod(); got != "POST" {
		t.Errorf("HTTPMethod = %q, want POST", got)
	}
	if got := tr.RemoteAddr(); got != "1.2.3.4:5678" {
		t.Errorf("RemoteAddr = %q, want the set address", got)
	}

	// RequestHeader returns the set carrier and supports Set/Get on it.
	req := tr.RequestHeader()
	if got := req.Get("X-Req-A"); got != "rv" {
		t.Errorf("RequestHeader().Get(X-Req-A) = %q, want rv", got)
	}
	req.Set("X-Req-C", "added")
	if got := req.Get("X-Req-C"); got != "added" {
		t.Errorf("RequestHeader().Get(X-Req-C) = %q, want added", got)
	}

	// ReplyHeader is a SEPARATE carrier from RequestHeader: writes to one
	// must not leak into the other.
	rep := tr.ReplyHeader()
	if got := rep.Get("X-Rep-B"); got != "pv" {
		t.Errorf("ReplyHeader().Get(X-Rep-B) = %q, want pv", got)
	}
	req.Set("X-Leak-Check", "only-req")
	if got := rep.Get("X-Leak-Check"); got != "" {
		t.Errorf("ReplyHeader saw RequestHeader value: got %q, want empty", got)
	}
	rep.Set("X-Leak-Check-Rep", "only-rep")
	if got := req.Get("X-Leak-Check-Rep"); got != "" {
		t.Errorf("RequestHeader saw ReplyHeader value: got %q, want empty", got)
	}

	// Lazy init: a zero-value Transport with nil header fields must
	// auto-create usable (non-nil) carriers on first access.
	empty := &Transport{}
	if got := empty.RequestHeader(); got == nil {
		t.Fatal("RequestHeader() on zero Transport returned nil, want auto-created")
	} else {
		got.Set("X-Lazy", "yes")
		if v := got.Get("X-Lazy"); v != "yes" {
			t.Errorf("lazy RequestHeader Get = %q, want yes", v)
		}
	}
	if got := empty.ReplyHeader(); got == nil {
		t.Fatal("ReplyHeader() on zero Transport returned nil, want auto-created")
	} else {
		got.Set("X-Lazy-Rep", "yes")
		if v := got.Get("X-Lazy-Rep"); v != "yes" {
			t.Errorf("lazy ReplyHeader Get = %q, want yes", v)
		}
	}
}
