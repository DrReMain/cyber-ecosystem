package report

// The report domain's live probe: the real database, the freeze contract
// end to end — the free tier's live judgement, the lazy freeze on first
// unlocked read, the pin surviving a rubric that grew afterwards, the
// event-driven freeze, and the free tier re-answering live once the grant
// is gone. Gated: ENT_E2E=1 with the dev database up runs it; anything
// else skips.

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"

	"cyber-ecosystem/app/services/esports/internal/bootstrap"
	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	entitlement "cyber-ecosystem/app/services/esports/internal/ent/entitlement"
	entprofile "cyber-ecosystem/app/services/esports/internal/ent/profile"
	entreportsnapshot "cyber-ecosystem/app/services/esports/internal/ent/reportsnapshot"
	"cyber-ecosystem/app/services/esports/internal/ent/schema/local_mixins"
	enttest "cyber-ecosystem/app/services/esports/internal/ent/test"
	enttestrecord "cyber-ecosystem/app/services/esports/internal/ent/testrecord"
	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	repOwner    = "e2erepowner000000001"
	repCohort   = "e2erepcohort00000001"
	repTestCode = "e2ereptest01"
)

func repEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func repSubject(userID string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: userID, TenantID: "default", Population: "user"})
}

func TestReportFreezeE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the report freeze end-to-end")
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
		cleanReportE2E(t, db)
		closeDB()
	})
	cleanReportE2E(t, db)

	bus := bootstrap.NewEventSubscriptions(logger)
	uc := NewReportUC(logger, pf, NewRecordBoardRP(logger, pf), NewSnapshotRP(logger, pf),
		entitle.NewEntitlementRP(logger, pf), profile.NewProfileRP(logger, pf), bus)
	entRP := entitle.NewEntitlementRP(logger, pf)
	snapRP := NewSnapshotRP(logger, pf)

	seedCtx := repSubject(repOwner)
	testRow, err := db.Test.Create().
		SetCode(repTestCode).
		SetName("e2e freeze probe").
		SetCategory("fps").
		SetAbility("reactionSpeed").
		Save(seedCtx)
	if err != nil {
		t.Fatalf("seed test: %v", err)
	}
	now := time.Now()
	record, err := db.TestRecord.Create().
		SetTestCode(testRow.Code).
		SetUserID(repOwner).
		SetAttemptNo(1).
		SetStartedAt(now.Add(-time.Minute)).
		SetEndedAt(now).
		SetRawDetail(map[string]any{"e2e": true}).
		SetAuthoritativeScore(88.5).
		Save(seedCtx)
	if err != nil {
		t.Fatalf("seed record: %v", err)
	}
	ownerCtx := repSubject(repOwner)

	// Act 1 — no grant: the free tier answers live over a pool too thin
	// to say anything, and nothing freezes.
	r1, err := uc.GetReport(ownerCtx, record.ID)
	if err != nil {
		t.Fatalf("free read: %v", err)
	}
	if r1.Unlocked || r1.Composite != nil || r1.Tier != "" {
		t.Fatalf("free read shape: unlocked=%v tier=%q composite=%v", r1.Unlocked, r1.Tier, r1.Composite)
	}
	if n := snapshotCount(t, db, ownerCtx, record.ID); n != 0 {
		t.Fatalf("free read must not freeze, snapshots=%d", n)
	}

	// Act 2 — the trial grant lands; the first unlocked read freezes
	// lazily. No profile, no cohort key: the honest pin is unjudged over
	// a none basis.
	if err := entRP.Grant(seedCtx, &entitle.EntitlementGrant{
		UserID: repOwner, ScopeKind: shared.ScopeTrial, ScopeRef: &record.ID, Source: "e2e",
	}); err != nil {
		t.Fatalf("grant: %v", err)
	}
	r2, err := uc.GetReport(ownerCtx, record.ID)
	if err != nil {
		t.Fatalf("unlocked read: %v", err)
	}
	if !r2.Unlocked || r2.Tier != shared.TierUnjudged || r2.Composite != nil {
		t.Fatalf("frozen shape: unlocked=%v tier=%q composite=%v", r2.Unlocked, r2.Tier, r2.Composite)
	}
	snap2, err := snapRP.FindByRecord(ownerCtx, record.ID)
	if err != nil || snap2 == nil {
		t.Fatalf("lazy freeze missing: %v %v", snap2, err)
	}
	if snap2.CohortBasis != basisNone || snap2.Tier != shared.TierUnjudged || len(snap2.Abilities) != 1 {
		t.Fatalf("snapshot row: %+v", snap2)
	}

	// Act 3 — the rubric grows: a profile and a same-age same-sex pool
	// now exist, the live judgement would carry a real tier — the frozen
	// answer must not move.
	if _, err := db.Profile.Create().
		SetUserID(repOwner).
		SetBirthYear(2010).
		SetSex(entprofile.Sex("male")).
		Save(seedCtx); err != nil {
		t.Fatalf("owner profile: %v", err)
	}
	if _, err := db.Profile.Create().
		SetUserID(repCohort).
		SetBirthYear(2010).
		SetSex(entprofile.Sex("male")).
		Save(seedCtx); err != nil {
		t.Fatalf("cohort profile: %v", err)
	}
	for i := 1; i <= 11; i++ {
		if _, err := db.TestRecord.Create().
			SetTestCode(testRow.Code).
			SetUserID(repCohort).
			SetAttemptNo(i).
			SetStartedAt(now.Add(-time.Hour)).
			SetEndedAt(now.Add(-time.Hour + time.Minute)).
			SetRawDetail(map[string]any{"e2e": true}).
			SetAuthoritativeScore(50 + float64(i)).
			Save(seedCtx); err != nil {
			t.Fatalf("cohort record %d: %v", i, err)
		}
	}
	r3, err := uc.GetReport(ownerCtx, record.ID)
	if err != nil {
		t.Fatalf("post-rubric read: %v", err)
	}
	if r3.Tier != shared.TierUnjudged || r3.Composite != nil {
		t.Fatalf("the pin drifted with the rubric: tier=%q composite=%v", r3.Tier, r3.Composite)
	}
	if n := snapshotCount(t, db, ownerCtx, record.ID); n != 1 {
		t.Fatalf("freeze not idempotent, snapshots=%d", n)
	}

	// Act 4 — the event path: a second record's unlock publishes, and the
	// freeze lands without any read.
	record2, err := db.TestRecord.Create().
		SetTestCode(testRow.Code).
		SetUserID(repOwner).
		SetAttemptNo(2).
		SetStartedAt(now).
		SetEndedAt(now.Add(time.Minute)).
		SetRawDetail(map[string]any{"e2e": true}).
		SetAuthoritativeScore(91).
		Save(seedCtx)
	if err != nil {
		t.Fatalf("seed record2: %v", err)
	}
	if err := entRP.Grant(seedCtx, &entitle.EntitlementGrant{
		UserID: repOwner, ScopeKind: shared.ScopeTrial, ScopeRef: &record2.ID, Source: "e2e",
	}); err != nil {
		t.Fatalf("grant2: %v", err)
	}
	bus.PublishReportUnlocked(seedCtx, shared.ReportUnlocked{UserID: repOwner, RecordID: record2.ID})
	snap4, err := snapRP.FindByRecord(ownerCtx, record2.ID)
	if err != nil || snap4 == nil {
		t.Fatalf("event freeze missing: %v %v", snap4, err)
	}
	if snap4.Tier == shared.TierUnjudged || snap4.CohortBasis != "same-age-same-sex" {
		t.Fatalf("event freeze shape: %+v", snap4)
	}

	// Act 5 — the grant gone, the same record answers free-tier live
	// again: a real tier over the grown rubric, proving the frozen
	// unjudged and the live judgement are two different answers to the
	// same record.
	gone, err := db.Entitlement.Query().
		Where(entitlement.UserIDEQ(repOwner), entitlement.ScopeRef(record.ID)).
		Only(seedCtx)
	if err != nil {
		t.Fatalf("grant lookup: %v", err)
	}
	if err := db.Entitlement.DeleteOneID(gone.ID).Exec(seedCtx); err != nil {
		t.Fatalf("grant revoke: %v", err)
	}
	r5, err := uc.GetReport(ownerCtx, record.ID)
	if err != nil {
		t.Fatalf("post-revoke read: %v", err)
	}
	if r5.Unlocked || r5.Tier == shared.TierUnjudged || r5.Tier == "" || r5.Composite == nil {
		t.Fatalf("live re-answer: unlocked=%v tier=%q composite=%v", r5.Unlocked, r5.Tier, r5.Composite)
	}
}

func snapshotCount(t *testing.T, db *ent.Client, ctx context.Context, recordID string) int {
	t.Helper()
	n, err := db.ReportSnapshot.Query().
		Where(entreportsnapshot.RecordIDEQ(recordID)).
		Count(ctx)
	if err != nil {
		t.Fatalf("snapshot count: %v", err)
	}
	return n
}

func cleanReportE2E(t *testing.T, db *ent.Client) {
	t.Helper()
	// Hard deletes under SkipSoftDelete so reruns see a clean slate; the
	// deletes key on this probe's own users and test code, never a bare
	// table wipe — the dev database carries real seed rows.
	dctx := local_mixins.SkipSoftDelete(context.Background())
	users := []string{repOwner, repCohort}
	if _, err := db.ReportSnapshot.Delete().Where(entreportsnapshot.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean snapshots: %v", err)
	}
	if _, err := db.Entitlement.Delete().Where(entitlement.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean grants: %v", err)
	}
	if _, err := db.TestRecord.Delete().Where(enttestrecord.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean records: %v", err)
	}
	if _, err := db.Profile.Delete().Where(entprofile.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean profiles: %v", err)
	}
	if _, err := db.Test.Delete().Where(enttest.Code(repTestCode)).Exec(dctx); err != nil {
		t.Fatalf("clean test: %v", err)
	}
}
