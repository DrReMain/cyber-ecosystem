package e2e

// The operator faces' live probe: the report-policy table answering its
// defaults and its overrides as one settled read, the mint-and-list pair
// over activation codes, the catalog and the one config write this plane
// owns (absent dials untouched, the engine bag replaced whole), the banked
// scenarios with their answer keys, and the stored attempts behind the
// verdicts — the same rows the member faces cut theirs from. Gated:
// ENT_E2E=1 with the dev database up runs it; anything else skips.

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/utils"

	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"

	"cyber-ecosystem/app/services/esports/internal/bootstrap"
	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	entactivationbatch "cyber-ecosystem/app/services/esports/internal/ent/activationbatch"
	entactivation "cyber-ecosystem/app/services/esports/internal/ent/activationcode"
	entreportpolicy "cyber-ecosystem/app/services/esports/internal/ent/reportpolicy"
	"cyber-ecosystem/app/services/esports/internal/ent/schema/local_mixins"
	enttestrecord "cyber-ecosystem/app/services/esports/internal/ent/testrecord"
	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/module/test"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	abAdmin = "e2eabadmin0000000001"
	abUser  = "e2eabuser00000000001"
)

func TestAdminReadBackE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the admin read-back end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: repEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: repEnvOr("DB_PASS", "postgres"), DbName: repEnvOr("DB_NAME", "esports"),
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
		abClean(t, db)
		closeDB()
	})
	abClean(t, db)

	adminCtx := repSubject(abAdmin)
	entUC := entitle.NewEntitleUC(logger, nil, nil, nil,
		entitle.NewReportPolicyRP(logger, pf), bootstrap.NewEventSubscriptions(logger))
	codeUC := entitle.NewCodeUC(logger, pf, entitle.NewActivationCodeRP(logger, pf), profile.NewProfileRP(logger, pf))
	testUC := test.NewTestUC(logger, nil, test.NewTestRP(logger, pf), nil, nil)

	// Face 1 — the policy table: the three families answer as one settled
	// read, defaults for the untouched, an override row with its audit
	// fields for the one this probe set (school flipped closed→open, then
	// back so later probes meet the shipped world).
	pols, err := entUC.ListReportAccessPolicies(adminCtx)
	if err != nil {
		t.Fatalf("list policies: %v", err)
	}
	if len(pols) != 3 {
		t.Fatalf("policies = %d rows, want 3", len(pols))
	}
	wantDefault := map[string]string{
		shared.ReportCognitive: shared.PolicyMembership,
		shared.ReportCareer:    shared.PolicyOpen,
		shared.ReportSchool:    shared.PolicyClosed,
	}
	for _, p := range pols {
		if p.Policy != wantDefault[p.Family] || p.Overridden {
			t.Fatalf("family %s: policy=%q overridden=%v, want default %q", p.Family, p.Policy, p.Overridden, wantDefault[p.Family])
		}
	}
	if err := entUC.SetReportAccessPolicy(adminCtx, shared.ReportSchool, shared.PolicyOpen); err != nil {
		t.Fatalf("set school policy: %v", err)
	}
	pols, err = entUC.ListReportAccessPolicies(adminCtx)
	if err != nil {
		t.Fatalf("relist policies: %v", err)
	}
	var school *entitle.PolicyRow
	for _, p := range pols {
		if p.Family == shared.ReportSchool {
			school = p
		}
	}
	if school == nil || !school.Overridden || school.Policy != shared.PolicyOpen || school.UpdatedBy != abAdmin {
		t.Fatalf("school override: %+v", school)
	}

	// Face 2 — mint and list: generation answers what was minted, the list
	// keeps showing it, and the code shape is the operator's AC prefix.
	gen, err := codeUC.Generate(adminCtx, &entitle.GenerateIn{ReportType: shared.ReportCognitive, Count: 2, ExpireDays: 7})
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	codes := gen.Codes
	if len(codes) != 2 {
		t.Fatalf("minted %d codes, want 2", len(codes))
	}
	for _, c := range codes {
		if !strings.HasPrefix(c, "AC") || len(c) != 2+16 {
			t.Fatalf("code shape %q", c)
		}
	}
	listed, err := codeUC.List(adminCtx, &entitle.ListCodesIn{PageRequest: &commonv1.PageRequest{}, ReportType: utils.Ptr(shared.ReportCognitive)})
	if err != nil {
		t.Fatalf("list codes: %v", err)
	}
	byCode := make(map[string]bool, len(listed.Codes))
	for _, row := range listed.Codes {
		byCode[row.Code] = true
	}
	for _, c := range codes {
		if !byCode[c] {
			t.Fatalf("minted code %s missing from list", c)
		}
	}

	// Face 3 — the catalog and the one write: the banked test resolves,
	// the config write touches only the dial it carried, and the engine
	// bag replaces whole.
	tests, err := testUC.List(adminCtx)
	if err != nil {
		t.Fatalf("list tests: %v", err)
	}
	var odID string
	for _, tt := range tests {
		if tt.Code == "offense_defense" {
			odID = tt.ID
		}
	}
	if odID == "" {
		t.Fatalf("offense_defense missing from the catalog")
	}
	orig, err := testUC.GetConfig(adminCtx, odID)
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	origBg, origParams := orig.BackgroundInvalidSec, orig.EngineParams
	t.Cleanup(func() {
		_, _ = testUC.UpdateConfig(context.Background(), &test.UpdateConfigIn{
			TestID: odID, BackgroundInvalidSec: &origBg, EngineParams: origParams,
		})
	})
	bumped := origBg + 1
	updated, err := testUC.UpdateConfig(adminCtx, &test.UpdateConfigIn{TestID: odID, BackgroundInvalidSec: &bumped})
	if err != nil {
		t.Fatalf("update config: %v", err)
	}
	if updated.BackgroundInvalidSec != bumped || updated.DurationSec != orig.DurationSec {
		t.Fatalf("config update: %+v", updated)
	}
	bagged, err := testUC.UpdateConfig(adminCtx, &test.UpdateConfigIn{
		TestID: odID, EngineParams: map[string]string{"spawn_interval_ms": "1500"},
	})
	if err != nil {
		t.Fatalf("update params: %v", err)
	}
	// The bag replaced whole — replace-what-you-saw, never a merge.
	if bagged.EngineParams["spawn_interval_ms"] != "1500" || len(bagged.EngineParams) != 1 {
		t.Fatalf("engine bag: %+v (was %+v)", bagged.EngineParams, origParams)
	}

	// Face 4 — the bank: the seeded scenarios answer with their answer
	// keys — operator visibility into what correctness is judged against.
	bank, err := testUC.ListQuestionBank(adminCtx, "offense_defense", nil)
	if err != nil {
		t.Fatalf("list bank: %v", err)
	}
	if len(bank.Rows) == 0 {
		t.Fatalf("offense_defense bank is empty")
	}
	for _, row := range bank.Rows {
		if len(row.Answer) == 0 {
			t.Fatalf("scenario %s carries no answer key", row.ScenarioCode)
		}
	}

	// Face 5 — the attempts: this probe's own row lists under its filters
	// and reads back whole, raw facts included.
	now := time.Now()
	rec, rerr := db.TestRecord.Create().
		SetTestCode("offense_defense").SetUserID(abUser).SetAttemptNo(1).
		SetStartedAt(now.Add(-time.Minute)).SetEndedAt(now).
		SetRawDetail(map[string]any{"offense_defense": map[string]any{"decisions": []any{}}}).
		SetValid(true).SetAuthoritativeScore(75).Save(repSubject(abUser))
	if rerr != nil {
		t.Fatalf("seed record: %v", rerr)
	}
	attempts, err := testUC.ListAttempts(adminCtx, &test.AttemptsQueryIn{TestCode: utils.Ptr("offense_defense"), UserID: utils.Ptr(abUser)})
	if err != nil {
		t.Fatalf("list attempts: %v", err)
	}
	if len(attempts.Rows) != 1 || attempts.Rows[0].ID != rec.ID {
		t.Fatalf("attempts rows: %+v", attempts.Rows)
	}
	full, err := testUC.GetAttempt(adminCtx, rec.ID)
	if err != nil {
		t.Fatalf("get attempt: %v", err)
	}
	if full == nil || full.Authoritative == nil || *full.Authoritative != 75 || full.RawDetail["offense_defense"] == nil {
		t.Fatalf("attempt detail: %+v", full)
	}
}

func abClean(t *testing.T, db *ent.Client) {
	t.Helper()
	// Hard deletes under SkipSoftDelete so reruns see a clean slate; the
	// deletes key on this probe's own identities, never a bare table wipe.
	dctx := local_mixins.SkipSoftDelete(context.Background())
	if _, err := db.ReportPolicy.Delete().Where(entreportpolicy.UpdatedBy(abAdmin)).Exec(dctx); err != nil {
		t.Fatalf("clean policies: %v", err)
	}
	if _, err := db.ActivationCode.Delete().Where(entactivation.CreatedBy(abAdmin)).Exec(dctx); err != nil {
		t.Fatalf("clean codes: %v", err)
	}
	if _, err := db.ActivationBatch.Delete().Where(entactivationbatch.CreatedBy(abAdmin)).Exec(dctx); err != nil {
		t.Fatalf("clean batches: %v", err)
	}
	if _, err := db.TestRecord.Delete().Where(enttestrecord.UserID(abUser)).Exec(dctx); err != nil {
		t.Fatalf("clean records: %v", err)
	}
}
