package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
)

func wechatRow(endpoint string) *conf.Auth_Provider {
	return &conf.Auth_Provider{
		Key: "wechat", Kind: "wechat_mp", Enabled: true, Endpoint: endpoint,
		Wechat: &conf.Auth_Provider_WeChat{Appid: "wx-test", Secret: "s3cr3t"},
	}
}

func wechatReq(code string) identityPayload {
	c := code
	return &systempb.LoginByIdentityRequest{
		Payload: &systempb.LoginByIdentityRequest_Wechat{Wechat: &systempb.WeChatIdentityPayload{Code: &c}},
	}
}

// The exchange hits the configured endpoint with the OAuth contract and
// returns the openid the provider answered.
func TestWeChatExchangeContract(t *testing.T) {
	var gotQuery string
	p := wechatProvider{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = fmt.Fprint(w, `{"access_token":"t","openid":"oid-live","errcode":0}`)
	}))
	defer srv.Close()

	id, err := p.Exchange(context.Background(), wechatRow(srv.URL), wechatReq("the-code"))
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if id.Kind != "wechat_mp" || id.Identifier != "oid-live" {
		t.Fatalf("exchange: identity = %+v", id)
	}
	for _, want := range []string{"appid=wx-test", "secret=s3cr3t", "code=the-code", "grant_type=authorization_code"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("exchange: query %q missing %q", gotQuery, want)
		}
	}
}

// The provider's errcode family fails closed into the exchange error — the
// same shape for an expired, reused, or mismatched code.
func TestWeChatExchangeErrCodeFailsClosed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"errcode":40029,"errmsg":"invalid code"}`)
	}))
	defer srv.Close()

	p := wechatProvider{}
	if _, err := p.Exchange(context.Background(), wechatRow(srv.URL), wechatReq("bad")); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("errcode: want EXCHANGE_FAILED, got %v", err)
	}
}

// An answer without an openid is a failed exchange even at errcode 0.
func TestWeChatExchangeWithoutOpenIDFails(t *testing.T) {
	p := wechatProvider{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"errcode":0}`)
	}))
	defer srv.Close()

	if _, err := p.Exchange(context.Background(), wechatRow(srv.URL), wechatReq("c")); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("no openid: want EXCHANGE_FAILED, got %v", err)
	}
}

// A transport failure is this side's infra, never the provider's verdict.
func TestWeChatExchangeTransportFailureIsInfra(t *testing.T) {
	p := wechatProvider{}
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Close() // dead endpoint

	if _, err := p.Exchange(context.Background(), wechatRow(srv.URL), wechatReq("c")); !errorspb.IsGeneralErrorInternal(err) {
		t.Fatalf("transport: want INTERNAL, got %v", err)
	}
}

// The authorize URL is minted server-side: appid, redirect, and the
// one-time state ride the query, the wechat browser suffix closes it, and
// no secret appears.
func TestWeChatAuthorizeURL(t *testing.T) {
	p := wechatProvider{}
	got := p.AuthorizeURL(wechatRow(""), "https://app.example/cb?x=1", "st-1")
	for _, want := range []string{
		"https://open.weixin.qq.com/connect/oauth2/authorize?",
		"appid=wx-test",
		"redirect_uri=" + "https%3A%2F%2Fapp.example%2Fcb%3Fx%3D1",
		"response_type=code",
		"scope=snsapi_base",
		"state=st-1",
		"#wechat_redirect",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("authorize url %q missing %q", got, want)
		}
	}
	if strings.Contains(got, "s3cr3t") {
		t.Fatal("authorize url: the secret must never appear")
	}
}
