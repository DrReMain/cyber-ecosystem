package connect

import (
	"net/http"
	"reflect"
	"sort"
	"testing"
)

func TestTransportKind(t *testing.T) {
	tr := &Transport{}
	if got := tr.Kind(); got != KindConnect {
		t.Fatalf("Kind() = %v, want %v", got, KindConnect)
	}
}

func TestTransportFields(t *testing.T) {
	reqHeader := NewHeader(http.Header{"X-Req": {"1"}})
	replyHeader := NewHeader(http.Header{"X-Reply": {"2"}})
	tr := &Transport{
		endpoint:    "connect://127.0.0.1:8080",
		operation:   "/acme.echo.v1.EchoService/Ping",
		reqHeader:   reqHeader,
		replyHeader: replyHeader,
	}
	if got := tr.Endpoint(); got != "connect://127.0.0.1:8080" {
		t.Fatalf("Endpoint() = %q", got)
	}
	if got := tr.Operation(); got != "/acme.echo.v1.EchoService/Ping" {
		t.Fatalf("Operation() = %q", got)
	}
	if got := tr.RequestHeader().Get("X-Req"); got != "1" {
		t.Fatalf("RequestHeader().Get() = %q", got)
	}
	if got := tr.ReplyHeader().Get("X-Reply"); got != "2" {
		t.Fatalf("ReplyHeader().Get() = %q", got)
	}
}

func TestHeaderKeys(t *testing.T) {
	h := NewHeader(http.Header{
		"Abb": {"1"},
		"Bcc": {"2"},
	})
	keys := h.Keys()
	sort.Strings(keys)
	want := []string{"Abb", "Bcc"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("Keys() = %v, want %v", keys, want)
	}
}
