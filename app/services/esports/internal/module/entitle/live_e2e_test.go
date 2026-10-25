package entitle

// The entitlement domain's live probe: the real database, the module's
// own faces end to end — generation, the concurrent redeem (one code, one
// use, the acceptance anchor), the idempotent re-redeem, the refusal
// arms, and the access judgement under both shipped defaults and operator
// overrides. Gated: ENT_E2E=1 with the dev database up runs it; anything
// else skips.

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

func entEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func entSubject(userID string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: userID, TenantID: "default", Population: "user"})
}

func TestEntitleLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the entitlement end-to-end")
	}
	const (
		admin1 = "entadmuser0000000001" // MaxRuneLen 20
		userA  = "ente2euser0000000001"
		userB  = "ente2euser0000000002"
	)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: entEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: entEnvOr("DB_PASS", "postgres"), DbName: entEnvOr("DB_NAME", "esports"),
		MaxOpenConns: 8,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	pf, err := platform.NewPlatform(db, platform.NewEntErrorHandler(), nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	t.Cleanup(func() {
		cleanEntE2E(t, db)
		closeDB()
	})
	cleanEntE2E(t, db)

	codeRP := NewCodeRP(logger, pf)
	entRP := NewEntitlementRP(logger, pf)
	policyRP := NewPolicyRP(logger, pf)
	uc := NewEntitleUC(logger, pf, codeRP, entRP, policyRP)
	adminCtx := entSubject(admin1)

	// Generation: five single-use codes over the look-alike-free alphabet.
	codes, err := NewCodeUC(logger, pf, codeRP).Generate(adminCtx, &GenerateIn{ReportType: reportSchool, Count: 5, ExpireDays: 30})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(codes) != 5 || len(codes[0]) != len(codePrefix)+codeBodyLen || codes[0][:2] != codePrefix {
		t.Fatalf("code shape: %v", codes)
	}

	// The acceptance anchor: ten members race one code — exactly one
	// redeem succeeds, nine answer exhausted, the ledger gains one row.
	code0 := codes[0]
	const racers = 10
	errs := make(chan error, racers)
	var wg sync.WaitGroup
	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			uid := fmt.Sprintf("ente2eracer%08d", i)
			if _, err := uc.Redeem(entSubject(uid), code0); err != nil {
				errs <- err
			} else {
				errs <- nil
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	var okCount, exhausted int
	for err := range errs {
		switch {
		case err == nil:
			okCount++
		case esportspb.IsEsportsActivationCodeExhausted(err):
			exhausted++
		default:
			t.Fatalf("race returned unexpected error: %v", err)
		}
	}
	if okCount != 1 || exhausted != racers-1 {
		t.Fatalf("race outcome: ok=%d exhausted=%d, want 1/%d", okCount, exhausted, racers-1)
	}
	if row, err := codeRP.ByCode(context.Background(), code0); err != nil || row.UsedCount != 1 {
		t.Fatalf("used_count after race: row=%+v err=%v", row, err)
	}
	rows, err := pf.GetClient(context.Background()).Entitlement.Query().All(context.Background())
	if err != nil {
		t.Fatalf("entitlement query: %v", err)
	}
	if len(rows) != 1 || rows[0].ScopeKind != shared.ScopeReportType || *rows[0].ScopeRef != reportSchool || rows[0].Source != code0 {
		t.Fatalf("race granted %+v", rows[0])
	}

	// The winner re-redeems: idempotent, and the use is not re-taken.
	winner := rows[0].UserID
	if out, err := uc.Redeem(entSubject(winner), code0); err != nil || out.ReportType != reportSchool {
		t.Fatalf("winner re-redeem: out=%+v err=%v", out, err)
	}
	if row, _ := codeRP.ByCode(context.Background(), code0); row.UsedCount != 1 {
		t.Fatalf("re-redeem took a second use: %+v", row)
	}

	// The same member's concurrent double redeem: both idempotent
	// answers, still exactly one use.
	code1 := codes[1]
	var pair sync.WaitGroup
	pairErrs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		pair.Add(1)
		go func() {
			defer pair.Done()
			if _, err := uc.Redeem(entSubject(userA), code1); err != nil {
				pairErrs <- err
			} else {
				pairErrs <- nil
			}
		}()
	}
	pair.Wait()
	close(pairErrs)
	for err := range pairErrs {
		if err != nil {
			t.Fatalf("same-member concurrent redeem errored: %v", err)
		}
	}
	if row, _ := codeRP.ByCode(context.Background(), code1); row.UsedCount != 1 {
		t.Fatalf("same-member race took two uses: %+v", row)
	}

	// Expired code refuses even with its use untaken.
	code2 := codes[2]
	if _, err := db.ExecContext(context.Background(),
		`UPDATE activation_code SET expires_at = now() - interval '1 second' WHERE code = $1`, code2); err != nil {
		t.Fatalf("age the code: %v", err)
	}
	if _, err := uc.Redeem(entSubject(userA), code2); !esportspb.IsEsportsActivationCodeExpired(err) {
		t.Fatalf("expired code: err=%v", err)
	}
	if _, err := uc.Redeem(entSubject(userA), "ACNOPE"); !esportspb.IsEsportsActivationCodeInvalid(err) {
		t.Fatalf("unknown code: err=%v", err)
	}

	// The access judgement under shipped defaults: career open,
	// cognitive member-gated (closed without one), school codes-only.
	access := func(uid, family string) *AccessOut {
		t.Helper()
		out, err := uc.GetReportAccess(entSubject(uid), family)
		if err != nil {
			t.Fatalf("access %s %s: %v", uid, family, err)
		}
		return out
	}
	if out := access(userB, reportCareer); !out.Allowed || out.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_OPEN {
		t.Fatalf("career default: %+v", out)
	}
	if out := access(userB, reportCognitive); out.Allowed {
		t.Fatalf("cognitive default without membership: %+v", out)
	}
	if out := access(winner, reportSchool); !out.Allowed || out.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_CODE_REDEEMED {
		t.Fatalf("school after redeem: %+v", out)
	}

	// A membership row opens the member plane, and an operator override
	// closes what shipped open — a redeemed code still enters.
	memberExpiry := time.Now().AddDate(0, 0, 31)
	if err := entRP.Grant(context.Background(), &EntitlementGrant{
		UserID: userB, ScopeKind: shared.ScopeMembership, ExpiresAt: &memberExpiry, Source: "e2e",
	}); err != nil {
		t.Fatalf("membership grant: %v", err)
	}
	if out := access(userB, reportCognitive); !out.Allowed || out.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE {
		t.Fatalf("cognitive with membership: %+v", out)
	}
	if err := uc.SetReportAccessPolicy(adminCtx, reportCareer, policyClosed); err != nil {
		t.Fatalf("policy set: %v", err)
	}
	if out := access(userB, reportCareer); out.Allowed || out.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_CLOSED {
		t.Fatalf("career after closed override: %+v", out)
	}
	if err := uc.SetReportAccessPolicy(adminCtx, reportSchool, policyMembership); err != nil {
		t.Fatalf("policy set school: %v", err)
	}
	if out := access(userB, reportSchool); !out.Allowed || out.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE {
		t.Fatalf("school under membership policy: %+v", out)
	}
	if err := uc.SetReportAccessPolicy(adminCtx, reportCognitive, policyTenant); !esportspb.IsEsportsReportPolicyNotReady(err) {
		t.Fatalf("tenant policy refused: err=%v", err)
	}
}

func cleanEntE2E(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := context.Background()
	for _, stmt := range []struct {
		q    string
		args []any
	}{
		{`DELETE FROM entitlement WHERE user_id LIKE 'ente2e%' OR user_id LIKE 'entadm%'`, nil},
		{`DELETE FROM activation_code WHERE created_by LIKE 'entadm%'`, nil},
		{`DELETE FROM report_policy`, nil},
	} {
		if _, err := db.ExecContext(ctx, stmt.q, stmt.args...); err != nil {
			t.Fatalf("entitle e2e clean: %v", err)
		}
	}
}
