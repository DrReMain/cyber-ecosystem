package e2e

// The S2 concurrency probes, live over the dev stack. Two shapes:
//
// The unique-constraint race: the operator's bulk import and a concurrent
// sms login both try to birth the same phone's account. Whichever writer
// seats the credential first wins the unique; a loser that had already
// passed the not-found check unwinds its whole transaction (no orphan
// user), and the row family lands exactly once. The interleaving is not
// forced — several interleavings have both writers succeed (the import
// skips a taken phone by design) — the invariant is the converged end
// state, never which writer won.
//
// ReplaceGrants convergence: two concurrent diff-replaces on one role.
// The S1 fix swallowed the DoNothing conflict shape (sql.ErrNoRows) and
// fell through to the re-query, so neither writer errors out on the
// clash and every op that lands carries its own policy links. Which ops
// survive is interleaving-dependent (a writer may remove an op another
// just seated — both are authorized intents, the documented union
// semantics); the pinned laws are: no error, no deadlock, the shared
// anchor op survives with its links, and every present op's links match
// the intent that seated it.
// Gated: ENT_E2E=1 with the dev stack (postgres + redis) up runs them.

import (
	"context"
	"log/slog"
	"os"
	"testing"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/ent"
	entauthzpolicy "cyber-ecosystem/app/services/system/internal/ent/authzpolicy"
	entcredential "cyber-ecosystem/app/services/system/internal/ent/credential"
	entmembership "cyber-ecosystem/app/services/system/internal/ent/membership"
	entrole "cyber-ecosystem/app/services/system/internal/ent/role"
	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
	enttenant "cyber-ecosystem/app/services/system/internal/ent/tenant"
	entuser "cyber-ecosystem/app/services/system/internal/ent/user"
	authmod "cyber-ecosystem/app/services/system/internal/module/auth"
	authzmod "cyber-ecosystem/app/services/system/internal/module/authz"
	populationmod "cyber-ecosystem/app/services/system/internal/module/population"
	rolemod "cyber-ecosystem/app/services/system/internal/module/role"
	tenantmod "cyber-ecosystem/app/services/system/internal/module/tenant"
	usermod "cyber-ecosystem/app/services/system/internal/module/user"
	"cyber-ecosystem/app/services/system/internal/platform"
)

const (
	racePhone  = "13900007777" // its own range — no other e2e or seed names it
	raceTenant = "e2e-race-tenant"
)

func raceClean(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := local_mixins.SkipSoftDelete(context.Background())
	creds, err := db.Credential.Query().Where(
		entcredential.KindEQ("sms"), entcredential.IdentifierEQ(racePhone)).All(ctx)
	if err != nil {
		t.Fatalf("race clean credential query: %v", err)
	}
	for _, c := range creds {
		if _, err := db.Membership.Delete().Where(entmembership.UserIDEQ(c.UserID)).Exec(ctx); err != nil {
			t.Fatalf("race clean seat: %v", err)
		}
		if _, err := db.User.Delete().Where(entuser.IDEQ(c.UserID)).Exec(ctx); err != nil {
			t.Fatalf("race clean user: %v", err)
		}
		if _, err := db.Credential.Delete().Where(entcredential.IDEQ(c.ID)).Exec(ctx); err != nil {
			t.Fatalf("race clean credential row: %v", err)
		}
	}
	if _, err := db.Role.Delete().Where(entrole.CodeHasPrefix("e2e-race-")).Exec(ctx); err != nil {
		t.Fatalf("race clean role: %v", err)
	}
	if _, err := db.AuthzPolicy.Delete().Where(entauthzpolicy.NameHasPrefix("e2e-race-")).Exec(ctx); err != nil {
		t.Fatalf("race clean policy: %v", err)
	}
	if _, err := db.Membership.Delete().Where(entmembership.TenantIDEQ(raceTenant)).Exec(ctx); err != nil {
		t.Fatalf("race clean tenant seats: %v", err)
	}
	if _, err := db.Tenant.Delete().Where(enttenant.CodeEQ(raceTenant)).Exec(ctx); err != nil {
		t.Fatalf("race clean tenant: %v", err)
	}
}

// TestUniqueConstraintRaceLiveE2E walks the two birth writers into each
// other: the import's row and a concurrent sms login's auto-registration
// target the same fresh phone. The converged family is exactly one
// credential keying exactly one enabled user holding exactly one member
// seat — the loser's unwound transaction leaves no orphans behind.
func TestUniqueConstraintRaceLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev stack (postgres + redis) up to run the constraint-race end-to-end")
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

	userRP := usermod.NewUserRP(logger, pf)
	tokenRP := authmod.NewTokenRP(logger, pf)
	authzRP, err := authzmod.NewAuthzRP(logger, pf)
	if err != nil {
		t.Fatalf("authz rp: %v", err)
	}
	tenantRP := tenantmod.NewTenantRP(logger, pf)
	populationRP := populationmod.NewPopulationRP(logger, pf)
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
		for _, k := range []string{
			"auth:web:sms-code:" + racePhone,
			"auth:sms:send-daily:" + racePhone,
			"rate:auth:sms:send-window:" + racePhone,
		} {
			if err := tokenRP.Del(context.Background(), k); err != nil {
				t.Fatalf("clean cache key %s: %v", k, err)
			}
		}
		raceClean(t, db)
		closeCache()
		closeDB()
	})
	raceClean(t, db)

	svc := authmod.NewAuthService(logger, uc)

	// The login leg's proof is minted before the gun: both writers enter
	// with everything they need, so the race is between the two birth
	// writes, not the send.
	sent, err := svc.SendSmsCode(bare, &systempb.SendSmsCodeRequest{Provider: sp("sms"), Phone: sp(racePhone)})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	code := lastSmsCode(t, ctrl, racePhone)
	phone := racePhone
	start := make(chan struct{})
	outcome := make(chan error, 2)
	go func() {
		<-start
		_, lerr := svc.LoginByIdentity(bare, smsIdentityReq(sent.State, racePhone, code))
		outcome <- lerr
	}()
	go func() {
		<-start
		_, ierr := svc.ImportAccounts(bare, &systempb.ImportAccountsRequest{Rows: []*systempb.ImportAccountRow{{Phone: &phone}}})
		outcome <- ierr
	}()
	close(start)
	errs := [2]error{<-outcome, <-outcome}

	// A loser is legal (both passed the not-found check, the unique broke
	// the tie); both failing would be the disease this probe hunts.
	failed := 0
	for _, e := range errs {
		if e != nil {
			failed++
		}
	}
	if failed > 1 {
		t.Fatalf("both writers failed: %v / %v", errs[0], errs[1])
	}

	creds, cerr := db.Credential.Query().Where(
		entcredential.KindEQ("sms"), entcredential.IdentifierEQ(racePhone)).All(bare)
	if cerr != nil {
		t.Fatalf("credential query: %v", cerr)
	}
	if len(creds) != 1 {
		t.Fatalf("credentials = %d, want exactly 1", len(creds))
	}
	u, uerr := db.User.Get(bare, creds[0].UserID)
	if uerr != nil {
		t.Fatalf("winner user: %v", uerr)
	}
	if !u.Enabled || u.PasswordHash != "" {
		t.Fatalf("winner shape: enabled=%v password-set=%v", u.Enabled, u.PasswordHash != "")
	}
	seats, serr := db.Membership.Query().Where(entmembership.UserIDEQ(u.ID)).All(bare)
	if serr != nil {
		t.Fatalf("seat query: %v", serr)
	}
	if len(seats) != 1 || seats[0].Population != "member" {
		t.Fatalf("winner seats: %+v", seats)
	}
}

// TestReplaceGrantsConvergenceLiveE2E races two diff-replaces on one
// role. The anchor op is pre-seated; both intents keep it and add one op
// of their own. Neither writer may error on the clash (the swallowed
// DoNothing conflict), the anchor survives with its links, and every
// added op that survives carries exactly the policy links of the intent
// that seated it — the fall-through the fix added; a loser's op shipping
// without its links is the bug this pins.
func TestReplaceGrantsConvergenceLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the replace-grants convergence end-to-end")
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
	pf, err := platform.NewPlatform(nil, nil, db, platform.NewEntErrorHandler(), nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	t.Cleanup(func() {
		raceClean(t, db)
		closeDB()
	})
	raceClean(t, db)

	if _, err = db.Tenant.Create().SetCode(raceTenant).SetName("e2e race").Save(bare); err != nil {
		t.Fatalf("tenant: %v", err)
	}
	ctxR := tenSubject(raceTenant)

	windows := map[string]any{"windows": []any{map[string]any{"start": "09:00", "end": "18:00"}}}
	p1, err := db.AuthzPolicy.Create().SetKind("time_window").SetName("e2e-race-p1").SetParams(windows).SetEnabled(true).Save(bare)
	if err != nil {
		t.Fatalf("policy p1: %v", err)
	}
	p2, err := db.AuthzPolicy.Create().SetKind("time_window").SetName("e2e-race-p2").SetParams(windows).SetEnabled(true).Save(bare)
	if err != nil {
		t.Fatalf("policy p2: %v", err)
	}
	roleR, err := db.Role.Create().SetCode("e2e-race-role").SetName("e2e race role").Save(ctxR)
	if err != nil {
		t.Fatalf("role: %v", err)
	}

	rp := rolemod.NewRoleRP(logger, pf)
	grant := func(op string, policies ...string) *rolemod.Permission {
		return &rolemod.Permission{Operation: op, ScopeKind: kauthz.ScopeKindAll, PolicyIDs: policies}
	}

	if err = rp.ReplaceGrants(ctxR, roleR.ID, []*rolemod.Permission{grant("/race/a", p1.ID)}); err != nil {
		t.Fatalf("seed grants: %v", err)
	}
	intentB := []*rolemod.Permission{grant("/race/a", p1.ID), grant("/race/b", p2.ID)}
	intentC := []*rolemod.Permission{grant("/race/a", p1.ID), grant("/race/c", p1.ID)}

	start := make(chan struct{})
	outcome := make(chan error, 2)
	go func() { <-start; outcome <- rp.ReplaceGrants(ctxR, roleR.ID, intentB) }()
	go func() { <-start; outcome <- rp.ReplaceGrants(ctxR, roleR.ID, intentC) }()
	close(start)
	for i := 0; i < 2; i++ {
		if rerr := <-outcome; rerr != nil {
			t.Fatalf("concurrent replace %d: %v", i, rerr)
		}
	}

	gs, err := rp.ListGrants(ctxR, roleR.ID)
	if err != nil {
		t.Fatalf("list grants: %v", err)
	}
	links := map[string][]string{}
	for _, g := range gs {
		links[g.Operation] = g.PolicyIDs
	}
	if len(gs) < 2 || len(gs) > 3 {
		t.Fatalf("grant count: %d (%v)", len(gs), links)
	}
	// The anchor survives every interleaving with its links intact.
	if got := links["/race/a"]; len(got) != 1 || got[0] != p1.ID {
		t.Fatalf("anchor links: %v", got)
	}
	if got, ok := links["/race/b"]; ok && (len(got) != 1 || got[0] != p2.ID) {
		t.Fatalf("opB links: %v", got)
	}
	if got, ok := links["/race/c"]; ok && (len(got) != 1 || got[0] != p1.ID) {
		t.Fatalf("opC links: %v", got)
	}
}
