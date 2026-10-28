package e2e

// The sms login chain, end to end over the live stack (postgres + redis):
// the send answers a one-shot state and records the code the channel would
// have delivered, the resend inside the window throttles, a wrong code
// burns both the code and the state and answers the uniform invalid class,
// the burned state refuses a replay, and the full chain lands an
// auto-registered member — credential, member seat, verifiable session.
// Every leg rides the service handlers, the mapping the wire exposes, so a
// broken handler field wiring cannot survive green.
// Gated: ENT_E2E=1 with the dev stack (postgres + redis) up runs it.

import (
	"context"
	"log/slog"
	"os"
	"testing"

	smsmock "cyber-ecosystem/shared-go/capability/sms/mock"
	"cyber-ecosystem/shared-go/kratos/security"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/ent"
	entcredential "cyber-ecosystem/app/services/system/internal/ent/credential"
	entmembership "cyber-ecosystem/app/services/system/internal/ent/membership"
	entpopulation "cyber-ecosystem/app/services/system/internal/ent/population"
	entuser "cyber-ecosystem/app/services/system/internal/ent/user"
	authmod "cyber-ecosystem/app/services/system/internal/module/auth"
	authzmod "cyber-ecosystem/app/services/system/internal/module/authz"
	populationmod "cyber-ecosystem/app/services/system/internal/module/population"
	tenantmod "cyber-ecosystem/app/services/system/internal/module/tenant"
	usermod "cyber-ecosystem/app/services/system/internal/module/user"
	"cyber-ecosystem/app/services/system/internal/platform"
)

// Two phones, one per leg: the throttle probes ride the first phone's
// window; the full chain rides the second phone's fresh window — no
// implementation reaching into the limiter's storage keys to fake time.
const (
	e2eSmsPhoneA = "13900001111"
	e2eSmsPhoneB = "13900002222"
)

func smsClean(t *testing.T, db *ent.Client, trp authmod.TokenRP) {
	t.Helper()
	ctx := context.Background()
	for _, phone := range []string{e2eSmsPhoneA, e2eSmsPhoneB} {
		cs, err := db.Credential.Query().Where(
			entcredential.KindEQ("sms"), entcredential.IdentifierEQ(phone)).All(ctx)
		if err != nil {
			t.Fatalf("clean credential query: %v", err)
		}
		for _, c := range cs {
			if _, err := db.Membership.Delete().Where(entmembership.UserIDEQ(c.UserID)).Exec(ctx); err != nil {
				t.Fatalf("clean membership: %v", err)
			}
			if _, err := db.User.Delete().Where(entuser.IDEQ(c.UserID)).Exec(ctx); err != nil {
				t.Fatalf("clean user: %v", err)
			}
			if _, err := db.Credential.Delete().Where(entcredential.IDEQ(c.ID)).Exec(ctx); err != nil {
				t.Fatalf("clean credential: %v", err)
			}
		}
		// The code and throttle keys reset so the next run starts from
		// zero; orphaned states die on their own TTL. The send-window key
		// carries redis_rate's own "rate:" storage prefix — the one place
		// this harness reaches past the capability surface, without which
		// a re-run inside 60s would inherit the prior run's window.
		for _, k := range []string{
			"auth:web:sms-code:" + phone,
			"auth:sms:send-daily:" + phone,
			"rate:auth:sms:send-window:" + phone,
		} {
			if err := trp.Del(ctx, k); err != nil {
				t.Fatalf("clean cache key %s: %v", k, err)
			}
		}
	}
}

func TestSmsLoginLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev stack (postgres + redis) up to run the sms login end-to-end")
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

	userRP := usermod.NewUserRP(logger, pf)
	authzRP, err := authzmod.NewAuthzRP(logger, pf)
	if err != nil {
		t.Fatalf("authz rp: %v", err)
	}
	tenantRP := tenantmod.NewTenantRP(logger, pf)
	populationRP := populationmod.NewPopulationRP(logger, pf)
	tokenRP := authmod.NewTokenRP(logger, pf)
	providers, ctrl := e2eProviders(t, logger, cacheClient, tokenRP)
	authConf := &conf.Auth{Providers: []*conf.Auth_Provider{{
		Key: "sms", Kind: "sms", Enabled: true, AutoRegister: true, Population: "member",
		Sms: &conf.Auth_Provider_Sms{SignName: "曦光未来", TemplateCode: "SMS_E2E"},
	}}}
	uc, err := authmod.NewAuthUC(logger, pf, authConf, providers, userRP, authzRP, tenantRP, populationRP, tokenRP)
	if err != nil {
		t.Fatalf("auth uc: %v", err)
	}
	// The closes ride the cleanup, after the sweep — defers would run
	// before t.Cleanup and strand the sweep on a closed client.
	t.Cleanup(func() {
		smsClean(t, db, tokenRP)
		closeCache()
		closeDB()
	})

	smsClean(t, db, tokenRP)

	svc := authmod.NewAuthService(logger, uc)

	// The send answers the one-shot state and the mock backend records the
	// delivered code.
	sent1, err := svc.SendSmsCode(bare, &systempb.SendSmsCodeRequest{Provider: sp("sms"), Phone: sp(e2eSmsPhoneA)})
	if err != nil {
		t.Fatalf("send code: %v", err)
	}
	code1 := lastSmsCode(t, ctrl, e2eSmsPhoneA)
	if len(code1) != 6 {
		t.Fatalf("recorded code %q not 6 digits", code1)
	}

	// The resend inside the window throttles before the channel.
	if _, err = svc.SendSmsCode(bare, &systempb.SendSmsCodeRequest{Provider: sp("sms"), Phone: sp(e2eSmsPhoneA)}); !systempb.IsSystemSmsRateLimited(err) {
		t.Fatalf("resend inside window: want rate limited, got %v", err)
	}

	// A wrong code answers the uniform invalid class and burns both the
	// code and the state.
	wrong := "000000"
	if wrong == code1 {
		wrong = "999999"
	}
	if _, err = svc.LoginByIdentity(bare, smsIdentityReq(sent1.State, e2eSmsPhoneA, wrong)); !systempb.IsSystemSmsCodeInvalid(err) {
		t.Fatalf("wrong code: want code invalid, got %v", err)
	}
	if _, err = svc.LoginByIdentity(bare, smsIdentityReq(sent1.State, e2eSmsPhoneA, code1)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("replay of burned state: want exchange failed, got %v", err)
	}

	// The full chain on the second phone's fresh window: a send exchanges
	// into an auto-registered member with credential, member seat, and a
	// verifiable session.
	sent2, err := svc.SendSmsCode(bare, &systempb.SendSmsCodeRequest{Provider: sp("sms"), Phone: sp(e2eSmsPhoneB)})
	if err != nil {
		t.Fatalf("second phone send: %v", err)
	}
	code2 := lastSmsCode(t, ctrl, e2eSmsPhoneB)
	loggedIn, err := svc.LoginByIdentity(bare, smsIdentityReq(sent2.State, e2eSmsPhoneB, code2))
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
	smsAssertMemberChain(t, db)
}

// lastSmsCode reads the newest recorded send's code — the harness's window
// into what the channel would have delivered.
func lastSmsCode(t *testing.T, ctrl *smsmock.Controller, phone string) string {
	t.Helper()
	sent := ctrl.Sent()
	var last *smsmock.Entry
	for i := range sent {
		if sent[i].To == phone {
			last = &sent[i]
		}
	}
	if last == nil {
		t.Fatalf("log backend recorded no send to %s", phone)
	}
	if last.SignName != "曦光未来" || last.TemplateCode != "SMS_E2E" {
		t.Fatalf("recorded entry not the issued message: %+v", last)
	}
	return last.Params["code"]
}

// smsAssertMemberChain lands the auto-registration contract on real rows:
// the credential keys a born user (xid account, empty-hash password plane,
// enabled) holding exactly the member seat.
func smsAssertMemberChain(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := context.Background()
	c, err := db.Credential.Query().Where(
		entcredential.KindEQ("sms"), entcredential.IdentifierEQ(e2eSmsPhoneB)).Only(ctx)
	if err != nil {
		t.Fatalf("credential row: %v", err)
	}
	u, err := db.User.Get(ctx, c.UserID)
	if err != nil {
		t.Fatalf("born user: %v", err)
	}
	if !u.Enabled || u.Account == "" || u.PasswordHash != "" {
		t.Fatalf("born user shape wrong: enabled=%v account=%q password_set=%v", u.Enabled, u.Account, u.PasswordHash != "")
	}
	ms, err := db.Membership.Query().Where(entmembership.UserIDEQ(u.ID)).All(ctx)
	if err != nil {
		t.Fatalf("seat query: %v", err)
	}
	if len(ms) != 1 || ms[0].Population != "member" || ms[0].TenantID != security.DefaultTenant {
		t.Fatalf("member seat = %+v", ms)
	}
}

func sp(s string) *string { return &s }

func smsIdentityReq(state, phone, code string) *systempb.LoginByIdentityRequest {
	p, c, s, prov := phone, code, state, "sms"
	return &systempb.LoginByIdentityRequest{
		Provider: &prov,
		State:    &s,
		Payload:  &systempb.LoginByIdentityRequest_Sms{Sms: &systempb.SmsIdentityPayload{Phone: &p, Code: &c}},
	}
}
