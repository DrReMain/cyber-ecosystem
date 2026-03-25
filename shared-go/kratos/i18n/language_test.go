package i18n

import (
	"context"
	"net/http"
	"testing"

	connecttransport "github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect"
	"github.com/go-kratos/kratos/v2/transport"
)

func TestHeaderLanguageProvider_RespectsQValuesWithMatcher(t *testing.T) {
	provider := NewHeaderLanguageProvider("", "zh-Hans", "en")
	ctx := newServerContextWithAcceptLanguage("en;q=0.8, zh-Hans;q=0.9")
	got := provider.GetLanguage(ctx)
	if got != "zh-Hans" {
		t.Fatalf("language = %q, want %q", got, "zh-Hans")
	}
}

func TestHeaderLanguageProvider_RespectsQValuesWithoutMatcher(t *testing.T) {
	provider := NewHeaderLanguageProvider("")
	ctx := newServerContextWithAcceptLanguage("en;q=0.8, zh-Hans;q=0.9")
	got := provider.GetLanguage(ctx)
	if got != "zh-Hans" {
		t.Fatalf("language = %q, want %q", got, "zh-Hans")
	}
}

func TestHeaderLanguageProvider_Fallback(t *testing.T) {
	provider := NewHeaderLanguageProvider("en", "zh-Hans", "en")
	got := provider.GetLanguage(context.Background())
	if got != "en" {
		t.Fatalf("language = %q, want %q", got, "en")
	}
}

type testTransport struct {
	req   transport.Header
	reply transport.Header
}

func (t *testTransport) Kind() transport.Kind            { return transport.KindHTTP }
func (t *testTransport) Endpoint() string                { return "http://test" }
func (t *testTransport) Operation() string               { return "/test.v1.TestService/Test" }
func (t *testTransport) RequestHeader() transport.Header { return t.req }
func (t *testTransport) ReplyHeader() transport.Header   { return t.reply }

func newServerContextWithAcceptLanguage(acceptLanguage string) context.Context {
	req := connecttransport.NewHeader(http.Header{})
	reply := connecttransport.NewHeader(http.Header{})
	req.Set(HeaderAcceptLanguage, acceptLanguage)
	tr := &testTransport{req: req, reply: reply}
	return transport.NewServerContext(context.Background(), tr)
}
