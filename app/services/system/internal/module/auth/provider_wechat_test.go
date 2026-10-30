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
// returns the openid the provider answered. The userinfo leg finds no
// endpoint here and degrades to a nil profile — the login it decorates
// never fails for want of a nickname.
func TestWeChatExchangeContract(t *testing.T) {
	var gotQuery string
	p := wechatProvider{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wechatExchangePath {
			http.NotFound(w, r)
			return
		}
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
	if id.Profile != nil {
		t.Fatalf("exchange: degraded userinfo must leave the profile nil, got %+v", id.Profile)
	}
	for _, want := range []string{"appid=wx-test", "secret=s3cr3t", "code=the-code", "grant_type=authorization_code"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("exchange: query %q missing %q", gotQuery, want)
		}
	}
}

// The userinfo leg carries the person behind the openid: the exchange spends
// the access token it just earned, and wechat's raw nickname/headimgurl land
// on the identity verbatim — sanitising is the consumer's side of the wire.
func TestWeChatExchangeFetchesUserinfo(t *testing.T) {
	var gotQuery string
	p := wechatProvider{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case wechatExchangePath:
			_, _ = fmt.Fprint(w, `{"access_token":"t-ok","openid":"oid-live","errcode":0}`)
		case wechatUserinfoPath:
			gotQuery = r.URL.RawQuery
			_, _ = fmt.Fprint(w, `{"nickname":"昵称","headimgurl":"https://qlogo.cn/a.png","errcode":0}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	id, err := p.Exchange(context.Background(), wechatRow(srv.URL), wechatReq("the-code"))
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if id.Profile == nil || id.Profile.Nickname != "昵称" || id.Profile.AvatarURL != "https://qlogo.cn/a.png" {
		t.Fatalf("userinfo: profile = %+v, want wechat's raw person", id.Profile)
	}
	for _, want := range []string{"access_token=t-ok", "openid=oid-live", "lang=zh_CN"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("userinfo: query %q missing %q", gotQuery, want)
		}
	}
}

// An errcode from the userinfo endpoint is a degraded leg, not a failed
// exchange: the openid already authenticated, the login stands, and the
// profile reads as absent.
func TestWeChatUserinfoErrCodeDegradesToNil(t *testing.T) {
	p := wechatProvider{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case wechatExchangePath:
			_, _ = fmt.Fprint(w, `{"access_token":"t","openid":"oid-live","errcode":0}`)
		case wechatUserinfoPath:
			_, _ = fmt.Fprint(w, `{"errcode":48001,"errmsg":"api unauthorized"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	id, err := p.Exchange(context.Background(), wechatRow(srv.URL), wechatReq("the-code"))
	if err != nil {
		t.Fatalf("exchange must survive a degraded userinfo: %v", err)
	}
	if id.Identifier != "oid-live" || id.Profile != nil {
		t.Fatalf("degraded userinfo: identity = %+v, want openid standing with a nil profile", id)
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
// one-time state ride the query, the userinfo scope is what the nickname
// prefill needs, the wechat browser suffix closes it, and no secret appears.
func TestWeChatAuthorizeURL(t *testing.T) {
	p := wechatProvider{}
	got := p.AuthorizeURL(wechatRow(""), "https://app.example/cb?x=1", "st-1")
	for _, want := range []string{
		"https://open.weixin.qq.com/connect/oauth2/authorize?",
		"appid=wx-test",
		"redirect_uri=" + "https%3A%2F%2Fapp.example%2Fcb%3Fx%3D1",
		"response_type=code",
		"scope=snsapi_userinfo",
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

// An endpoint override sends the browser's authorize hop to the same
// stand-in the token exchange already honours — one base serves both OAuth
// legs, which is what lets a mock carry the whole chain without WeChat or a
// human in the loop.
func TestWeChatAuthorizeURLOverridden(t *testing.T) {
	p := wechatProvider{}
	got := p.AuthorizeURL(wechatRow("http://127.0.0.1:18443"), "https://app.example/cb", "st-9")
	for _, want := range []string{
		"http://127.0.0.1:18443/connect/oauth2/authorize?",
		"appid=wx-test",
		"redirect_uri=" + "https%3A%2F%2Fapp.example%2Fcb",
		"response_type=code",
		"scope=snsapi_userinfo",
		"state=st-9",
		"#wechat_redirect",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("authorize url %q missing %q", got, want)
		}
	}
	if strings.Contains(got, "open.weixin.qq.com") {
		t.Fatal("authorize url: an overridden row must not fall back to the official host")
	}
}
