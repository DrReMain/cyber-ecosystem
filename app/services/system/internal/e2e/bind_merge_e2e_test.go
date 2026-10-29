package e2e

// The bind + merge chains, end to end over the live stack (postgres +
// redis): the bind twin's mint carries the calling session (an anonymous
// public mint never opens a bind), a self-bind is idempotent, a foreign
// credential intercepts, the merge moves the credential, disables the
// absorbed account, writes the one-shot ledger row, and kills the absorbed
// account's live session, and system-side weight (a set password) refuses
// the absorption. Every leg rides the service handlers, the mapping the
// wire exposes. Gated: ENT_E2E=1 with the dev stack (postgres + redis) up.

import (
	"context"
	"log/slog"
	"os"
	"testing"

	smsmock "cyber-ecosystem/shared-go/capability/sms/mock"
	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/ent"
	entaccountmerge "cyber-ecosystem/app/services/system/internal/ent/accountmerge"
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

const (
	e2eBindPhonePoor    = "13900003333"
	e2eBindPhoneRich    = "13900004444"
	e2eBindPhoneHeavy   = "13900005555"
	e2eBindHeavyAccount = "e2e-merge-heavy"
)

func bindMergeClean(t *testing.T, db *ent.Client, trp authmod.TokenRP) {
	t.Helper()
	ctx := context.Background()
	for _, phone := range []string{e2eBindPhonePoor, e2eBindPhoneRich, e2eBindPhoneHeavy} {
		cs, err := db.Credential.Query().Where(
			entcredential.KindEQ("sms"), entcredential.IdentifierEQ(phone)).All(ctx)
		if err != nil {
			t.Fatalf("clean credential query: %v", err)
		}
		for _, c := range cs {
			if _, err := db.AccountMerge.Delete().Where(entaccountmerge.PoorIDEQ(c.UserID)).Exec(ctx); err != nil {
				t.Fatalf("clean ledger: %v", err)
			}
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
		if _, err := db.User.Delete().Where(entuser.AccountEQ(e2eBindHeavyAccount)).Exec(ctx); err != nil {
			t.Fatalf("clean heavy user: %v", err)
		}
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

// bindSendAndCode clears the send window, sends the bind-twin code from a
// session, and returns the bound state plus the code the channel recorded.
func bindSendAndCode(t *testing.T, svc *authmod.AuthService, ctrl *smsmock.Controller, trp authmod.TokenRP, ctx context.Context, phone string) (string, string) {
	t.Helper()
	if err := trp.Del(ctx, "rate:auth:sms:send-window:"+phone); err != nil {
		t.Fatalf("clear window: %v", err)
	}
	sent, err := svc.SendBindSmsCode(ctx, &systempb.SendBindSmsCodeRequest{Provider: sp("sms"), Phone: sp(phone)})
	if err != nil {
		t.Fatalf("bind send to %s: %v", phone, err)
	}
	return sent.State, lastSmsCode(t, ctrl, phone)
}

func TestBindMergeLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev stack (postgres + redis) up to run the bind + merge end-to-end")
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
	t.Cleanup(func() {
		bindMergeClean(t, db, tokenRP)
		closeCache()
		closeDB()
	})
	bindMergeClean(t, db, tokenRP)

	svc := authmod.NewAuthService(logger, uc)

	// The two accounts: the poor side births through a public sms login on
	// its phone (keeping the session the merge must kill), the rich side
	// the same on its own phone.
	poorLogin := e2eSmsPublicSend(t, svc, ctrl, tokenRP, e2eBindPhonePoor)
	poorSubject, err := uc.AuthenticateSession(bare, poorLogin)
	if err != nil {
		t.Fatalf("poor session: %v", err)
	}
	richLogin := e2eSmsPublicSend(t, svc, ctrl, tokenRP, e2eBindPhoneRich)
	richSubject, err := uc.AuthenticateSession(bare, richLogin)
	if err != nil {
		t.Fatalf("rich session: %v", err)
	}
	richCtx := security.WithSubject(bare, richSubject)

	// The session-tie gap stays closed: a public anonymous mint never
	// opens a bind, even with the live code in hand.
	if err = tokenRP.Del(bare, "rate:auth:sms:send-window:"+e2eBindPhonePoor); err != nil {
		t.Fatalf("clear window: %v", err)
	}
	pubSent, err := svc.SendSmsCode(bare, &systempb.SendSmsCodeRequest{Provider: sp("sms"), Phone: sp(e2eBindPhonePoor)})
	if err != nil {
		t.Fatalf("public send: %v", err)
	}
	pubCode := lastSmsCode(t, ctrl, e2eBindPhonePoor)
	if _, err = svc.BindCredential(richCtx, bindReq(pubSent.State, e2eBindPhonePoor, pubCode)); !systempb.IsSystemIdentityExchangeFailed(err) {
		t.Fatalf("anonymous state at bind: want EXCHANGE_FAILED, got %v", err)
	}

	// Self-bind is idempotent: the rich side's own phone answers success
	// and leaves exactly one credential row.
	selfState, selfCode := bindSendAndCode(t, svc, ctrl, tokenRP, richCtx, e2eBindPhoneRich)
	if _, err = svc.BindCredential(richCtx, bindReq(selfState, e2eBindPhoneRich, selfCode)); err != nil {
		t.Fatalf("self bind: %v", err)
	}
	if n := bindMergeCredentialCount(t, db, e2eBindPhoneRich); n != 1 {
		t.Fatalf("self bind: %d credential rows, want the one", n)
	}

	// A foreign credential intercepts: the poor side's phone belongs to
	// another account and the bind refuses without moving anything.
	foreignState, foreignCode := bindSendAndCode(t, svc, ctrl, tokenRP, richCtx, e2eBindPhonePoor)
	if _, err = svc.BindCredential(richCtx, bindReq(foreignState, e2eBindPhonePoor, foreignCode)); !systempb.IsSystemIdentityAlreadyBound(err) {
		t.Fatalf("foreign bind: want ALREADY_BOUND, got %v", err)
	}
	if owner := bindMergeCredential(t, db, e2eBindPhonePoor).UserID; owner != poorSubject.UserID {
		t.Fatalf("foreign bind: credential owner moved to %s", owner)
	}

	// Rich absorbs poor: the proof is the poor phone's live code on a
	// session-bound state; the credential moves, the poor account
	// disables, the ledger row lands, and the poor session dies.
	mergeState, mergeCode := bindSendAndCode(t, svc, ctrl, tokenRP, richCtx, e2eBindPhonePoor)
	if _, err = svc.MergeAccounts(richCtx, mergeReq(mergeState, e2eBindPhonePoor, mergeCode)); err != nil {
		t.Fatalf("merge: %v", err)
	}
	moved := bindMergeCredential(t, db, e2eBindPhonePoor)
	if moved.UserID != richSubject.UserID {
		t.Fatalf("merge: credential owner = %s, want the rich side", moved.UserID)
	}
	poorRow, err := db.User.Get(bare, poorSubject.UserID)
	if err != nil {
		t.Fatalf("poor row: %v", err)
	}
	if poorRow.Enabled {
		t.Fatal("merge: the absorbed account stayed enabled")
	}
	ledger, err := db.AccountMerge.Query().Where(entaccountmerge.PoorIDEQ(poorSubject.UserID)).Only(bare)
	if err != nil {
		t.Fatalf("ledger row: %v", err)
	}
	if ledger.RichID != richSubject.UserID || ledger.Kind != "sms" || ledger.Identifier != e2eBindPhonePoor {
		t.Fatalf("ledger row = %+v", ledger)
	}
	if _, err = uc.AuthenticateSession(bare, poorLogin); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("absorbed session: want 401, got %v", err)
	}

	// System-side weight refuses: a password-holding account is never
	// empty enough to absorb.
	hash, herr := utils.Hash("e2e-heavy-pw")
	if herr != nil {
		t.Fatalf("heavy hash: %v", herr)
	}
	heavy, err := db.User.Create().SetAccount(e2eBindHeavyAccount).SetPasswordHash(hash).SetEnabled(true).Save(bare)
	if err != nil {
		t.Fatalf("heavy user: %v", err)
	}
	if _, err = db.Credential.Create().SetKind("sms").SetIdentifier(e2eBindPhoneHeavy).SetUserID(heavy.ID).Save(bare); err != nil {
		t.Fatalf("heavy credential: %v", err)
	}
	heavyState, heavyCode := bindSendAndCode(t, svc, ctrl, tokenRP, richCtx, e2eBindPhoneHeavy)
	if _, err = svc.MergeAccounts(richCtx, mergeReq(heavyState, e2eBindPhoneHeavy, heavyCode)); !systempb.IsSystemMergeNotMergeable(err) {
		t.Fatalf("heavy merge: want NOT_MERGEABLE, got %v", err)
	}
	if owner := bindMergeCredential(t, db, e2eBindPhoneHeavy).UserID; owner != heavy.ID {
		t.Fatalf("heavy merge: credential owner moved to %s", owner)
	}
}

// e2eSmsPublicSend runs the public send + login pair on a fresh window and
// returns the minted session token.
func e2eSmsPublicSend(t *testing.T, svc *authmod.AuthService, ctrl *smsmock.Controller, trp authmod.TokenRP, phone string) string {
	t.Helper()
	ctx := context.Background()
	if err := trp.Del(ctx, "rate:auth:sms:send-window:"+phone); err != nil {
		t.Fatalf("clear window: %v", err)
	}
	sent, err := svc.SendSmsCode(ctx, &systempb.SendSmsCodeRequest{Provider: sp("sms"), Phone: sp(phone)})
	if err != nil {
		t.Fatalf("public send to %s: %v", phone, err)
	}
	loggedIn, err := svc.LoginByIdentity(ctx, smsIdentityReq(sent.State, phone, lastSmsCode(t, ctrl, phone)))
	if err != nil {
		t.Fatalf("login by identity on %s: %v", phone, err)
	}
	return loggedIn.SessionToken
}

func bindReq(state, phone, code string) *systempb.BindCredentialRequest {
	p, c, s, prov := phone, code, state, "sms"
	return &systempb.BindCredentialRequest{
		Provider: &prov,
		State:    &s,
		Payload:  &systempb.BindCredentialRequest_Sms{Sms: &systempb.SmsIdentityPayload{Phone: &p, Code: &c}},
	}
}

func mergeReq(state, phone, code string) *systempb.MergeAccountsRequest {
	p, c, s, prov := phone, code, state, "sms"
	return &systempb.MergeAccountsRequest{
		Provider: &prov,
		State:    &s,
		Payload:  &systempb.MergeAccountsRequest_Sms{Sms: &systempb.SmsIdentityPayload{Phone: &p, Code: &c}},
	}
}

func bindMergeCredential(t *testing.T, db *ent.Client, phone string) *ent.Credential {
	t.Helper()
	c, err := db.Credential.Query().Where(
		entcredential.KindEQ("sms"), entcredential.IdentifierEQ(phone)).Only(context.Background())
	if err != nil {
		t.Fatalf("credential row for %s: %v", phone, err)
	}
	return c
}

func bindMergeCredentialCount(t *testing.T, db *ent.Client, phone string) int {
	t.Helper()
	n, err := db.Credential.Query().Where(
		entcredential.KindEQ("sms"), entcredential.IdentifierEQ(phone)).Count(context.Background())
	if err != nil {
		t.Fatalf("credential count for %s: %v", phone, err)
	}
	return n
}
