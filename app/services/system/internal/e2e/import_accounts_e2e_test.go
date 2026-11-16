package e2e

// The operator's bulk registration face, end to end over the live stack
// (postgres + redis): each phone births the exact triple the sms
// auto-registration writes (xid account + empty-hash password plane + sms
// credential + member seat), a phone the bank already names skips, the
// same phone twice within a file answers invalid, the +86 spelling folds
// onto the bare number, and a non-mainland shape answers invalid without
// touching the bank. Every leg rides the service handler — the mapping
// the wire exposes. Gated: ENT_E2E=1 with the dev stack up.

import (
	"context"
	"log/slog"
	"os"
	"testing"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/ent"
	entcredential "cyber-ecosystem/app/services/system/internal/ent/credential"
	entmembership "cyber-ecosystem/app/services/system/internal/ent/membership"
	entuser "cyber-ecosystem/app/services/system/internal/ent/user"
	authmod "cyber-ecosystem/app/services/system/internal/module/auth"
	authzmod "cyber-ecosystem/app/services/system/internal/module/authz"
	populationmod "cyber-ecosystem/app/services/system/internal/module/population"
	tenantmod "cyber-ecosystem/app/services/system/internal/module/tenant"
	usermod "cyber-ecosystem/app/services/system/internal/module/user"
	"cyber-ecosystem/app/services/system/internal/platform"
)

// The probe's own number range — absent from every seed and every other
// e2e's constants so the clean never touches a foreign row.
const (
	iaPhoneA   = "13900007701"
	iaPhoneBad = "+44123456789" // shape-legal, not a mainland mobile
)

func TestImportAccountsE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev stack (postgres + redis) up to run the import-accounts end-to-end")
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
	providers, _ := e2eProviders(t, logger, cacheClient, tokenRP)
	authConf := &conf.Auth{Providers: []*conf.Auth_Provider{{
		Key: "sms", Kind: "sms", Enabled: true, AutoRegister: true, Population: "member",
		Sms: &conf.Auth_Provider_Sms{SignName: "曦光未来", TemplateCode: "SMS_E2E"},
	}}}
	uc, err := authmod.NewAuthUC(logger, pf, authConf, providers, userRP, authzRP, tenantRP, populationRP, tokenRP)
	if err != nil {
		t.Fatalf("auth uc: %v", err)
	}
	t.Cleanup(func() {
		iaClean(t, db)
		closeCache()
		closeDB()
	})
	iaClean(t, db)

	svc := authmod.NewAuthService(logger, uc)
	row := func(phone string) *systempb.ImportAccountRow {
		s := phone
		return &systempb.ImportAccountRow{Phone: &s}
	}

	// The mixed file: one fresh mainland number, the same number again
	// (a file error, not a bank duplicate), and a shape-legal foreign
	// number the channel law refuses — the ledger closes at 1/0/2.
	out, err := svc.ImportAccounts(bare, &systempb.ImportAccountsRequest{Rows: []*systempb.ImportAccountRow{
		row(iaPhoneA), row(iaPhoneA), row(iaPhoneBad),
	}})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if out.Created != 1 || out.Skipped != 0 || out.Invalid != 2 {
		t.Fatalf("counts: created=%d skipped=%d invalid=%d", out.Created, out.Skipped, out.Invalid)
	}
	if out.Results[1].Status != systempb.AccountImportStatus_ACCOUNT_IMPORT_STATUS_INVALID ||
		out.Results[2].Status != systempb.AccountImportStatus_ACCOUNT_IMPORT_STATUS_INVALID {
		t.Fatalf("verdicts: %+v", out.Results)
	}
	created := out.Results[0]
	if created.UserId == "" {
		t.Fatalf("created carries no user id: %+v", created)
	}

	// The landed triple is the auto-registration's own shape: xid handle,
	// a permanently closed password plane, the sms credential on the
	// canonical spelling, and the member seat where a later login lands.
	u, err := db.User.Get(bare, created.UserId)
	if err != nil {
		t.Fatalf("user row: %v", err)
	}
	if u.Account == "" || u.PasswordHash != "" || !u.Enabled {
		t.Fatalf("user shape: account=%q hash-set=%v enabled=%v", u.Account, u.PasswordHash != "", u.Enabled)
	}
	cred, err := db.Credential.Query().Where(
		entcredential.KindEQ("sms"), entcredential.IdentifierEQ(iaPhoneA)).Only(bare)
	if err != nil {
		t.Fatalf("credential row: %v", err)
	}
	if cred.UserID != created.UserId {
		t.Fatalf("credential join: %s != %s", cred.UserID, created.UserId)
	}
	seat, err := db.Membership.Query().Where(entmembership.UserIDEQ(created.UserId)).Only(bare)
	if err != nil {
		t.Fatalf("seat row: %v", err)
	}
	if seat.Population != "member" {
		t.Fatalf("seat population: %s", seat.Population)
	}

	// The +86 spelling folds onto the canonical number — the second file
	// answers skipped, never a second account.
	reout, err := svc.ImportAccounts(bare, &systempb.ImportAccountsRequest{Rows: []*systempb.ImportAccountRow{
		row("+86" + iaPhoneA),
	}})
	if err != nil {
		t.Fatalf("reimport: %v", err)
	}
	if reout.Created != 0 || reout.Skipped != 1 {
		t.Fatalf("reimport counts: created=%d skipped=%d", reout.Created, reout.Skipped)
	}
}

func iaClean(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := context.Background()
	// The probe's own rows only — keyed on its phone and its created user,
	// never a bare wipe and never a foreign credential.
	creds, err := db.Credential.Query().Where(
		entcredential.KindEQ("sms"), entcredential.IdentifierEQ(iaPhoneA)).All(ctx)
	if err != nil {
		t.Fatalf("clean query: %v", err)
	}
	for _, c := range creds {
		if _, err := db.Membership.Delete().Where(entmembership.UserIDEQ(c.UserID)).Exec(ctx); err != nil {
			t.Fatalf("clean seat: %v", err)
		}
		if _, err := db.User.Delete().Where(entuser.IDEQ(c.UserID)).Exec(ctx); err != nil {
			t.Fatalf("clean user: %v", err)
		}
		if _, err := db.Credential.Delete().Where(entcredential.IDEQ(c.ID)).Exec(ctx); err != nil {
			t.Fatalf("clean credential: %v", err)
		}
	}
}
