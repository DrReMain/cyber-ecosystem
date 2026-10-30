package e2e

// The wechat login chain, end to end over the live stack (postgres + redis)
// with the provider's OAuth endpoint override pointing at an in-test
// stand-in — the injection seam that closes the E6b lesson: the begin
// answers an authorize URL on the stand-in's base (the same base the token
// exchange honours), the stand-in's authorize hop answers the browser
// redirect WeChat itself would issue (the code plus the state echoed
// verbatim), and the exchange mints the openid the auto-registration keys.
// Credential, member seat, verifiable session — no WeChat, no human in the
// loop. Every leg rides the service handlers, the mapping the wire exposes.
// Gated: ENT_E2E=1 with the dev stack (postgres + redis) up runs it.

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	entpopulation "cyber-ecosystem/app/services/system/internal/ent/population"
	authmod "cyber-ecosystem/app/services/system/internal/module/auth"
	authzmod "cyber-ecosystem/app/services/system/internal/module/authz"
	populationmod "cyber-ecosystem/app/services/system/internal/module/population"
	tenantmod "cyber-ecosystem/app/services/system/internal/module/tenant"
	usermod "cyber-ecosystem/app/services/system/internal/module/user"
	"cyber-ecosystem/app/services/system/internal/platform"
)

const (
	e2eWechatOpenid = "e2e-openid-wxlogin"
	e2eWechatCode   = "e2e-wx-code"
)

func TestWechatLoginLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev stack (postgres + redis) up to run the wechat login end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	bare := context.Background()

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: tenEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: tenEnvOr("DB_PASS", "postgres"), DbName: tenEnvOr("DB_NAME", "system"),
		MaxOpenConns: 8,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	cacheClient, closeCache, err := platform.NewCache(&conf.Data{Redis: &conf.Data_Redis{
		Addr: tenEnvOr("REDIS_ADDR", "localhost:6379"), Db: 0,
	}}, logger)
	if err != nil {
		t.Fatalf("cache: %v", err)
	}
	cacheErrHandler, err := platform.NewCacheErrorHandler()
	if err != nil {
		t.Fatalf("cache error handler: %v", err)
	}
	pf, err := platform.NewPlatform(cacheClient, cacheErrHandler, db, platform.NewEntErrorHandler(), nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}

	memberSeen, err := db.Population.Query().Where(entpopulation.CodeEQ("member")).Exist(bare)
	if err != nil {
		t.Fatalf("population exist: %v", err)
	}
	if !memberSeen {
		if _, err = db.Population.Create().SetCode("member").SetName("Member").SetPlane(security.PlaneMember).Save(bare); err != nil {
			t.Fatalf("population: %v", err)
		}
	}

	// The wechat stand-in mirrors the interface boundary: the authorize hop
	// answers the redirect WeChat itself would issue (the code, and the
	// state echoed verbatim onto the redirect target), the token endpoint
	// answers the openid for the code it issued.
	wx := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/connect/oauth2/authorize":
			q := r.URL.Query()
			sep := "?"
			if strings.Contains(q.Get("redirect_uri"), "?") {
				sep = "&"
			}
			// #nosec G710 — the stand-in mirrors WeChat's authorize contract:
			// the redirect target is this test's own redirect_uri echoed
			// back to its own callback, no open surface exists.
			http.Redirect(w, r,
				q.Get("redirect_uri")+sep+"code="+url.QueryEscape(e2eWechatCode)+"&state="+url.QueryEscape(q.Get("state")),
				http.StatusFound)
		case "/sns/oauth2/access_token":
			if r.URL.Query().Get("code") != e2eWechatCode {
				_, _ = fmt.Fprint(w, `{"errcode":40029,"errmsg":"invalid code"}`)
				return
			}
			_, _ = fmt.Fprint(w, `{"access_token":"t","openid":"`+e2eWechatOpenid+`","errcode":0}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer wx.Close()

	userRP := usermod.NewUserRP(logger, pf)
	authzRP, err := authzmod.NewAuthzRP(logger, pf)
	if err != nil {
		t.Fatalf("authz rp: %v", err)
	}
	tenantRP := tenantmod.NewTenantRP(logger, pf)
	populationRP := populationmod.NewPopulationRP(logger, pf)
	tokenRP := authmod.NewTokenRP(logger, pf)
	providers, _ := e2eProviders(t, logger, cacheClient, tokenRP)
	authConf := &conf.Auth{Providers: []*conf.Auth_Provider{{
		Key: "wechat", Kind: "wechat_mp", Enabled: true, AutoRegister: true,
		Endpoint: wx.URL, Population: "member",
		Wechat: &conf.Auth_Provider_WeChat{Appid: "wx-e2e", Secret: "s3cr3t"},
	}}}
	uc, err := authmod.NewAuthUC(logger, pf, authConf, providers, userRP, authzRP, tenantRP, populationRP, tokenRP)
	if err != nil {
		t.Fatalf("auth uc: %v", err)
	}
	t.Cleanup(func() {
		sweepBornUsers(t, db, "wechat_mp", e2eWechatOpenid)
		closeCache()
		closeDB()
	})
	sweepBornUsers(t, db, "wechat_mp", e2eWechatOpenid)

	svc := authmod.NewAuthService(logger, uc)

	// The begin answers an authorize URL on the stand-in's base — the
	// injection seam's browser-facing half, without which the hop leaves
	// for open.weixin.qq.com and no mock can carry it.
	redirectURI := "https://ctr.local/login/wechat/callback?from=e2e"
	begin, err := svc.BeginIdentityLogin(bare, &systempb.BeginIdentityLoginRequest{
		Provider: sp("wechat"), RedirectUri: sp(redirectURI)})
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if !strings.HasPrefix(begin.AuthorizeUrl, wx.URL+"/connect/oauth2/authorize?") {
		t.Fatalf("authorize url %q not on the stand-in base %s", begin.AuthorizeUrl, wx.URL)
	}

	// The browser leg: a non-following client reads the redirect WeChat's
	// consent page would issue — the code and the state echoed verbatim.
	hop := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := hop.Get(begin.AuthorizeUrl)
	if err != nil {
		t.Fatalf("authorize hop: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("authorize hop: status %d, want 302", resp.StatusCode)
	}
	loc, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		t.Fatalf("authorize hop location: %v", err)
	}
	code, state := loc.Query().Get("code"), loc.Query().Get("state")
	beginURL, err := url.Parse(begin.AuthorizeUrl)
	if err != nil {
		t.Fatalf("begin url parse: %v", err)
	}
	if code != e2eWechatCode || state == "" || state != beginURL.Query().Get("state") {
		t.Fatalf("authorize hop: code=%q state=%q (begin state %q)", code, state, beginURL.Query().Get("state"))
	}

	// The callback leg exchanges the code for the openid and lands the
	// auto-registered member.
	loggedIn, err := svc.LoginByIdentity(bare, &systempb.LoginByIdentityRequest{
		Provider: sp("wechat"), State: &state,
		Payload: &systempb.LoginByIdentityRequest_Wechat{Wechat: &systempb.WeChatIdentityPayload{Code: &code}}})
	if err != nil {
		t.Fatalf("login by identity: %v", err)
	}
	subject, err := uc.AuthenticateSession(bare, loggedIn.SessionToken)
	if err != nil {
		t.Fatalf("verify session: %v", err)
	}
	if subject.Population != "member" {
		t.Fatalf("session population = %q, want member", subject.Population)
	}
	assertBornMemberChain(t, db, "wechat_mp", e2eWechatOpenid)
}
