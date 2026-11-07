package e2e

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
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"

	"cyber-ecosystem/app/services/esports/internal/bootstrap"
	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/entitlement"
	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
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

	codeRP := entitle.NewActivationCodeRP(logger, pf)
	entRP := entitle.NewEntitlementRP(logger, pf)
	policyRP := entitle.NewReportPolicyRP(logger, pf)
	profileRP := profile.NewProfileRP(logger, pf)
	bus := bootstrap.NewEventSubscriptions(logger)
	uc := entitle.NewEntitleUC(logger, pf, codeRP, entRP, policyRP, bus)
	codeUC := entitle.NewCodeUC(logger, pf, codeRP, profileRP)
	adminCtx := entSubject(admin1)

	// Generation: five single-use codes over the look-alike-free alphabet,
	// one batch holding them.
	gen, err := codeUC.Generate(adminCtx, &entitle.GenerateIn{ReportType: shared.ReportSchool, Count: 5, ExpireDays: 30})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	codes := gen.Codes
	if len(codes) != 5 || len(codes[0]) != 2+16 || codes[0][:2] != "AC" {
		t.Fatalf("code shape: %v", codes)
	}
	if len(gen.BatchNo) == 0 || gen.BatchNo[0] != 'B' {
		t.Fatalf("batch number shape: %q", gen.BatchNo)
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
	rows, err := pf.GetClient(context.Background()).Entitlement.Query().
		Where(entitlement.ScopeKindEQ(shared.ScopeReportType), entitlement.SourceEQ(code0)).
		All(context.Background())
	if err != nil {
		t.Fatalf("entitlement query: %v", err)
	}
	if len(rows) != 1 || *rows[0].ScopeRef != shared.ReportSchool {
		t.Fatalf("race granted %+v", rows)
	}

	// The winner re-redeems: idempotent, and the use is not re-taken.
	winner := rows[0].UserID
	if out, err := uc.Redeem(entSubject(winner), code0); err != nil || out.ReportType != shared.ReportSchool {
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

	// Expired code refuses a member with no standing grant, use untaken.
	code2 := codes[2]
	if _, err := db.ExecContext(context.Background(),
		`UPDATE activation_code SET expires_at = now() - interval '1 second' WHERE code = $1`, code2); err != nil {
		t.Fatalf("age the code: %v", err)
	}
	if _, err := uc.Redeem(entSubject("ente2egrantless00001"), code2); !esportspb.IsEsportsActivationCodeExpired(err) {
		t.Fatalf("expired code: err=%v", err)
	}
	// The precedence the service contract promises: a member whose family
	// grant stands re-presenting their own now-expired code gets the
	// idempotent answer — the grant outranks the code's expiry exactly as
	// it outranks a disabled batch, and the aged code keeps its one use.
	if _, err := db.ExecContext(context.Background(),
		`UPDATE activation_code SET expires_at = now() - interval '1 second' WHERE code = $1`, code1); err != nil {
		t.Fatalf("age the used code: %v", err)
	}
	if out, err := uc.Redeem(entSubject(userA), code1); err != nil || out.ExpiresAt == nil || !out.ExpiresAt.After(time.Now()) {
		t.Fatalf("expired-code re-present must answer the standing grant: out=%+v err=%v", out, err)
	}
	if row, _ := codeRP.ByCode(context.Background(), code1); row.UsedCount != 1 {
		t.Fatalf("expired-code re-present took a use: %+v", row)
	}
	if _, err := uc.Redeem(entSubject(userA), "ACNOPE"); !esportspb.IsEsportsActivationCodeInvalid(err) {
		t.Fatalf("unknown code: err=%v", err)
	}

	// The lapsed grant does not strand the member: after the family row
	// expires, a later code for the same family redeems again and the row
	// ratchets forward — expiry and source become the new code's, while a
	// shorter later code neither shortens the row nor burns its use.
	userC := "ente2euser0000000003"
	code3 := codes[3]
	if out, err := uc.Redeem(entSubject(userC), code3); err != nil || out.ExpiresAt == nil {
		t.Fatalf("first family redeem: out=%+v err=%v", out, err)
	}
	if _, err := db.ExecContext(context.Background(),
		`UPDATE entitlement SET expires_at = now() - interval '1 hour' WHERE user_id = $1 AND scope_kind = $2`,
		userC, shared.ScopeReportType); err != nil {
		t.Fatalf("lapse the grant: %v", err)
	}
	code4 := codes[4]
	codeRow4, err := codeRP.ByCode(context.Background(), code4)
	if err != nil || codeRow4 == nil {
		t.Fatalf("code4 read: %+v err=%v", codeRow4, err)
	}
	out4, err := uc.Redeem(entSubject(userC), code4)
	if err != nil || out4.ExpiresAt == nil || !out4.ExpiresAt.Equal(codeRow4.ExpiresAt) {
		t.Fatalf("post-lapse redeem: out=%+v err=%v want expiry %v", out4, err, codeRow4.ExpiresAt)
	}
	familySchool := shared.ReportSchool
	rowC, err := entRP.ByScope(entSubject(userC), userC, shared.ScopeReportType, &familySchool)
	if err != nil || rowC == nil || rowC.Source != code4 || !rowC.ExpiresAt.Equal(codeRow4.ExpiresAt) {
		t.Fatalf("ratcheted row: %+v err=%v", rowC, err)
	}
	for _, c := range []string{code3, code4} {
		if row, _ := codeRP.ByCode(context.Background(), c); row.UsedCount != 1 {
			t.Fatalf("use count for %s: %+v", c, row)
		}
	}
	shortGen, err := codeUC.Generate(adminCtx, &entitle.GenerateIn{ReportType: shared.ReportSchool, Count: 1, ExpireDays: 1})
	if err != nil || len(shortGen.Codes) != 1 {
		t.Fatalf("short generate: %+v err=%v", shortGen, err)
	}
	short := shortGen.Codes
	outS, err := uc.Redeem(entSubject(userC), short[0])
	if err != nil || outS.ExpiresAt == nil || !outS.ExpiresAt.Equal(codeRow4.ExpiresAt) {
		t.Fatalf("short redeem must stay idempotent on the longer row: out=%+v err=%v", outS, err)
	}
	if row, _ := codeRP.ByCode(context.Background(), short[0]); row.UsedCount != 0 {
		t.Fatalf("short code burned a use: %+v", row)
	}

	// The two-code race the family lock closes: the same member, grant
	// lapsed again, redeems two different fresh codes of the family at
	// once. The sequential shape (proven above) answers the second code
	// idempotent and keeps its use in the member's hand; the lock makes
	// the racing shape equal — exactly one use taken between the two
	// codes, the grant ratcheted to the codes' expiry.
	raceGen, err := codeUC.Generate(adminCtx, &entitle.GenerateIn{ReportType: shared.ReportSchool, Count: 2, ExpireDays: 30})
	if err != nil || len(raceGen.Codes) != 2 {
		t.Fatalf("race generate: %+v err=%v", raceGen, err)
	}
	if _, err := db.ExecContext(context.Background(),
		`UPDATE entitlement SET expires_at = now() - interval '1 hour' WHERE user_id = $1 AND scope_kind = $2`,
		userC, shared.ScopeReportType); err != nil {
		t.Fatalf("lapse the grant again: %v", err)
	}
	var two sync.WaitGroup
	twoErrs := make(chan error, 2)
	for _, code := range raceGen.Codes {
		two.Add(1)
		go func(code string) {
			defer two.Done()
			if _, err := uc.Redeem(entSubject(userC), code); err != nil {
				twoErrs <- err
			} else {
				twoErrs <- nil
			}
		}(code)
	}
	two.Wait()
	close(twoErrs)
	for err := range twoErrs {
		if err != nil {
			t.Fatalf("two-code race errored: %v", err)
		}
	}
	used := 0
	for _, code := range raceGen.Codes {
		row, err := codeRP.ByCode(context.Background(), code)
		if err != nil {
			t.Fatalf("race code readback: %v", err)
		}
		used += row.UsedCount
	}
	if used != 1 {
		t.Fatalf("two-code race burned %d uses, want 1 — the sequential shape's answer", used)
	}
	if g, gerr := entRP.ByScope(context.Background(), userC, shared.ScopeReportType, &familySchool); gerr != nil || g == nil || g.ExpiresAt == nil || !g.ExpiresAt.After(time.Now()) {
		t.Fatalf("two-code race left the grant lapsed: g=%+v err=%v", g, gerr)
	}

	// The batch plane reads as one story: the list folds counts over the
	// code rows, the code table carries each redeem's trail, a disable
	// voids only the unredeemed remainder, and the export answers the
	// whole filter in one payload.
	batchesOut, err := codeUC.ListBatches(adminCtx, &entitle.ListBatchesIn{PageRequest: &commonv1.PageRequest{}})
	if err != nil {
		t.Fatalf("list batches: %v", err)
	}
	var firstBatch *entitle.ActivationBatch
	for _, b := range batchesOut.Batches {
		if b.BatchNo == gen.BatchNo {
			firstBatch = b
		}
	}
	if firstBatch == nil || firstBatch.Count != 5 || firstBatch.UsedCount != 4 || firstBatch.DisabledAt != nil {
		t.Fatalf("batch fold: %+v", firstBatch)
	}
	if batchesOut.TotalBatches < 2 || batchesOut.TotalCodes < 6 || batchesOut.UsedCodes < 4 {
		t.Fatalf("footer fold: %+v", batchesOut)
	}
	redeemedOnly := entitle.CodeStatusRedeemed
	trailOut, err := codeUC.List(adminCtx, &entitle.ListCodesIn{PageRequest: &commonv1.PageRequest{}, BatchNo: &gen.BatchNo, Status: &redeemedOnly})
	if err != nil {
		t.Fatalf("redeemed trail: %v", err)
	}
	if len(trailOut.Codes) != 4 {
		t.Fatalf("redeemed trail: n=%d, want 4", len(trailOut.Codes))
	}
	// Each trail row names its own taker — the map is this run's redeem
	// ledger, so a swapped or stranger identity fails here.
	redeemers := map[string]string{code0: winner, code1: userA, code3: userC, code4: userC}
	for _, row := range trailOut.Codes {
		want, known := redeemers[row.Code]
		if !known || row.RedeemedBy == nil || row.RedeemedAt == nil || *row.RedeemedBy != want {
			t.Fatalf("redeem trail wrong: row=%+v want=%s", row, want)
		}
	}

	userD := "ente2euser0000000004"
	// The short batch still holds one unspent, unexpired code — the right
	// target for the disable flip. Disabling the first batch instead
	// would prove nothing: codes[2] is aged past its deadline, and the
	// other four are spent.
	if _, err := codeUC.SetBatchStatus(adminCtx, shortGen.BatchNo, true); err != nil {
		t.Fatalf("disable batch: %v", err)
	}
	if _, err := uc.Redeem(entSubject(userD), short[0]); !esportspb.IsEsportsActivationCodeDisabled(err) {
		t.Fatalf("disabled batch redeem: err=%v", err)
	}
	// code0's own batch is still live — plain idempotency: the answer
	// must be the standing grant, expiry and all.
	if out, err := uc.Redeem(entSubject(winner), code0); err != nil || out.ExpiresAt == nil || !out.ExpiresAt.Equal(*rows[0].ExpiresAt) {
		t.Fatalf("idempotent re-redeem must re-answer the standing grant: out=%+v err=%v", out, err)
	}
	// The true disable-vs-idempotent case: userC already holds short[0]'s
	// grant and a disabled batch cannot void a landed use — the answer is
	// the standing (longer) grant, not a fresh mint from short[0].
	standingC, err := entRP.ByScope(context.Background(), userC, shared.ScopeReportType, &familySchool)
	if err != nil || standingC == nil || standingC.ExpiresAt == nil {
		t.Fatalf("standing grant read: %+v err=%v", standingC, err)
	}
	if out, err := uc.Redeem(entSubject(userC), short[0]); err != nil || out.ExpiresAt == nil || !out.ExpiresAt.Equal(*standingC.ExpiresAt) {
		t.Fatalf("idempotent under disabled must re-answer the standing grant: out=%+v err=%v", out, err)
	}
	if _, err := codeUC.SetBatchStatus(adminCtx, shortGen.BatchNo, false); err != nil {
		t.Fatalf("re-enable batch: %v", err)
	}
	if _, err := uc.Redeem(entSubject(userD), short[0]); err != nil {
		t.Fatalf("redeem after re-enable: %v", err)
	}
	unusedOnly := entitle.CodeStatusUnused
	exported, err := codeUC.Export(adminCtx, &entitle.ListCodesIn{BatchNo: &gen.BatchNo, Status: &unusedOnly})
	if err != nil || len(exported) != 1 || exported[0].Code != codes[2] {
		t.Fatalf("export unused of first batch: n=%d err=%v", len(exported), err)
	}
	if exported[0].RedeemedBy != nil || exported[0].RedeemedAt != nil {
		t.Fatalf("unused row carries a redeem trail: %+v", exported[0])
	}

	// The detail page's find-one search: the exact code narrows the
	// batch's table to its own row.
	codeHit := codes[2]
	hitOut, err := codeUC.List(adminCtx, &entitle.ListCodesIn{PageRequest: &commonv1.PageRequest{}, BatchNo: &gen.BatchNo, Code: &codeHit})
	if err != nil || len(hitOut.Codes) != 1 || hitOut.Codes[0].Code != codeHit {
		t.Fatalf("code filter: n=%d err=%v", len(hitOut.Codes), err)
	}

	// The member's own view: userA redeemed exactly one code across every
	// batch, so the filter answers that row and only that row — the user
	// detail's report-unlock facts read from here.
	userFilter := userA
	userOut, err := codeUC.List(adminCtx, &entitle.ListCodesIn{PageRequest: &commonv1.PageRequest{}, User: &userFilter})
	if err != nil || len(userOut.Codes) != 1 || userOut.Codes[0].Code != codes[1] {
		t.Fatalf("user filter: n=%d err=%v", len(userOut.Codes), err)
	}
	if userOut.Codes[0].RedeemedBy == nil || *userOut.Codes[0].RedeemedBy != userA {
		t.Fatalf("user filter returned a foreign row: %+v", userOut.Codes[0])
	}

	// The access judgement under shipped defaults: career open,
	// cognitive member-gated (closed without one), school codes-only.
	access := func(uid, family string) *entitle.AccessOut {
		t.Helper()
		out, err := uc.GetReportAccess(entSubject(uid), family)
		if err != nil {
			t.Fatalf("access %s %s: %v", uid, family, err)
		}
		return out
	}
	if out := access(userB, shared.ReportCareer); !out.Allowed || out.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_OPEN {
		t.Fatalf("career default: %+v", out)
	}
	if out := access(userB, shared.ReportCognitive); out.Allowed {
		t.Fatalf("cognitive default without membership: %+v", out)
	}
	if out := access(winner, shared.ReportSchool); !out.Allowed || out.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_CODE_REDEEMED {
		t.Fatalf("school after redeem: %+v", out)
	}

	// A membership row opens the member plane, and an operator override
	// closes what shipped open — a redeemed code still enters.
	memberExpiry := time.Now().AddDate(0, 0, 31)
	if err := entRP.Grant(context.Background(), &entitle.EntitlementGrant{
		UserID: userB, ScopeKind: shared.ScopeMembership, ExpiresAt: &memberExpiry, Source: "e2e",
	}); err != nil {
		t.Fatalf("membership grant: %v", err)
	}
	if out := access(userB, shared.ReportCognitive); !out.Allowed || out.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE {
		t.Fatalf("cognitive with membership: %+v", out)
	}
	if err := uc.SetReportAccessPolicy(adminCtx, shared.ReportCareer, shared.PolicyClosed); err != nil {
		t.Fatalf("policy set: %v", err)
	}
	if out := access(userB, shared.ReportCareer); out.Allowed || out.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_CLOSED {
		t.Fatalf("career after closed override: %+v", out)
	}
	if err := uc.SetReportAccessPolicy(adminCtx, shared.ReportSchool, shared.PolicyMembership); err != nil {
		t.Fatalf("policy set school: %v", err)
	}
	if out := access(userB, shared.ReportSchool); !out.Allowed || out.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE {
		t.Fatalf("school under membership policy: %+v", out)
	}
	if err := uc.SetReportAccessPolicy(adminCtx, shared.ReportCognitive, shared.PolicyTenant); !esportspb.IsEsportsReportPolicyNotReady(err) {
		t.Fatalf("tenant policy refused: err=%v", err)
	}

	// The completion hook: userB's membership is active so their trial
	// mints a permanent grant; the member-less racer stays grant-less, and
	// a repeat call stays one row.
	recA := "ente2erecord0000001"
	bus.PublishTrialCompleted(context.Background(), shared.TrialCompleted{UserID: userB, RecordID: recA})
	ref := recA
	if row, err := pf.GetClient(context.Background()).Entitlement.Query().
		Where(entitlement.UserIDEQ(userB), entitlement.ScopeRefEQ(ref)).Only(context.Background()); err != nil {
		t.Fatalf("membership trial grant: %v", err)
	} else if row.ScopeKind != shared.ScopeTrial || row.ExpiresAt != nil || row.Source != "membership" {
		t.Fatalf("membership trial grant shape: %+v", row)
	}
	bus.PublishTrialCompleted(context.Background(), shared.TrialCompleted{UserID: winner, RecordID: recA})
	n, err := pf.GetClient(context.Background()).Entitlement.Query().
		Where(entitlement.ScopeRefEQ(ref)).Count(context.Background())
	if err != nil {
		t.Fatalf("hook memberless count: %v", err)
	}
	if n != 1 {
		t.Fatalf("hook granted %d rows, want 1 (memberless must not grant)", n)
	}
	bus.PublishTrialCompleted(context.Background(), shared.TrialCompleted{UserID: userB, RecordID: recA})
	if n, err := pf.GetClient(context.Background()).Entitlement.Query().
		Where(entitlement.ScopeRefEQ(ref)).Count(context.Background()); err != nil || n != 1 {
		t.Fatalf("hook repeat: n=%d err=%v, want 1", n, err)
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
		{`DELETE FROM activation_batch WHERE created_by LIKE 'entadm%'`, nil},
		{`DELETE FROM report_policy WHERE report_type IN ('cognitive', 'career', 'school')`, nil},
	} {
		if _, err := db.ExecContext(ctx, stmt.q, stmt.args...); err != nil {
			t.Fatalf("entitle e2e clean: %v", err)
		}
	}
}
