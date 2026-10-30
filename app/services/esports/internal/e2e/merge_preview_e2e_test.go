package e2e

// The merge preview's E6c acceptance anchor over the real database: the
// corpus bar refuses the both-weighted pair — no automatic merge when both
// sides carry assessment weight — while the empty-other pair answers
// KEEP_CALLER, the direction the confirm page executes. The system side of
// the pair (credential occupancy and the hard-gate weight) stands in as a
// fake: the seam under test is this service's emptiness bar and verdict,
// not system's own answering, which its e2e family already pins.
// Gated: ENT_E2E=1 with the dev database up runs it.

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	entmatchsession "cyber-ecosystem/app/services/esports/internal/ent/matchsession"
	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
	enttestrecord "cyber-ecosystem/app/services/esports/internal/ent/testrecord"
	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/module/report"
	"cyber-ecosystem/app/services/esports/internal/platform"
)

const (
	mergeE2ECaller = "mrgcaller000000001"
	mergeE2EOtherB = "mrgotherb000000001" // carries weight — the blocked pair
	mergeE2EOtherC = "mrgotherc000000001" // empty on this side — the kept pair
)

// fakeMergeSystem answers the two system questions the preview asks with
// the facts the scenario needs; nil resolved facts would read FREE.
type fakeMergeSystem struct {
	status   *profile.SecurityFacts
	resolved *profile.CredentialFacts
}

func (f *fakeMergeSystem) ResolveCredential(ctx context.Context, cred *profile.MergeCredential) (*profile.CredentialFacts, error) {
	return f.resolved, nil
}

func (f *fakeMergeSystem) SecurityStatus(ctx context.Context) (*profile.SecurityFacts, error) {
	return f.status, nil
}

func mergeE2EClean(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := context.Background()
	for _, uid := range []string{mergeE2ECaller, mergeE2EOtherB, mergeE2EOtherC} {
		if _, err := db.TestRecord.Delete().Where(enttestrecord.UserIDEQ(uid)).Exec(ctx); err != nil {
			t.Fatalf("clean test_record: %v", err)
		}
		if _, err := db.MatchSession.Delete().Where(entmatchsession.UserIDEQ(uid)).Exec(ctx); err != nil {
			t.Fatalf("clean match_session: %v", err)
		}
	}
}

// weighMember lands one valid session attempt — the emptiness bar's
// business weight, the exact shape SessionCounts counts.
func weighMember(t *testing.T, db *ent.Client, userID string) {
	t.Helper()
	ctx := context.Background()
	ms, err := db.MatchSession.Create().
		SetUserID(userID).
		SetGenre("fps").
		SetSteps([]types.MatchStep{}).
		Save(ctx)
	if err != nil {
		t.Fatalf("match session %s: %v", userID, err)
	}
	if _, err := db.TestRecord.Create().
		SetTestCode("e2e_merge_probe").
		SetUserID(userID).
		SetAttemptNo(1).
		SetSessionID(ms.ID).
		SetStartedAt(time.Now()).
		SetEndedAt(time.Now()).
		SetRawDetail(map[string]any{"e2e": map[string]any{"probe": true}}).
		Save(ctx); err != nil {
		t.Fatalf("test record %s: %v", userID, err)
	}
}

func TestMergePreviewLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the merge preview end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	bare := context.Background()

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
		mergeE2EClean(t, db)
		closeDB()
	})
	mergeE2EClean(t, db)

	weighMember(t, db, mergeE2ECaller)
	weighMember(t, db, mergeE2EOtherB)
	// mergeE2EOtherC stays empty on this side.

	maskCaller, maskB, maskC := "139****0001", "139****0002", "139****0003"
	sys := &fakeMergeSystem{status: &profile.SecurityFacts{PhoneMask: &maskCaller, CredentialCount: 1}}
	mergeUC := report.NewMergeUC(logger, pf, sys,
		report.NewRecordBoardRP(logger, pf), entitle.NewEntitlementRP(logger, pf))
	callerCtx := security.WithSubject(bare, &security.Subject{
		UserID: mergeE2ECaller, TenantID: "default", Population: "member"})

	// The acceptance refusal: both sides carry assessment weight — no
	// automatic merge, whatever system's own gate would say.
	sys.resolved = &profile.CredentialFacts{
		UserID: mergeE2EOtherB, Enabled: true, Mergeable: true, PhoneMask: &maskB}
	blocked, err := mergeUC.Preview(callerCtx, &profile.MergeCredential{Phone: &maskB})
	if err != nil {
		t.Fatalf("preview both-weighted: %v", err)
	}
	if blocked.Status != esportspb.MergePreviewStatus_MERGE_PREVIEW_STATUS_BLOCKED {
		t.Fatalf("both-weighted preview = %v, want BLOCKED", blocked.Status)
	}
	if blocked.Caller.Sessions != 1 || blocked.Other.Sessions != 1 {
		t.Fatalf("both-weighted weights: caller=%d other=%d, want 1/1",
			blocked.Caller.Sessions, blocked.Other.Sessions)
	}

	// The empty other answers KEEP_CALLER — the merge direction the
	// caller's own session executes.
	sys.resolved = &profile.CredentialFacts{
		UserID: mergeE2EOtherC, Enabled: true, Mergeable: true, PhoneMask: &maskC}
	kept, err := mergeUC.Preview(callerCtx, &profile.MergeCredential{Phone: &maskC})
	if err != nil {
		t.Fatalf("preview empty-other: %v", err)
	}
	if kept.Status != esportspb.MergePreviewStatus_MERGE_PREVIEW_STATUS_KEEP_CALLER {
		t.Fatalf("empty-other preview = %v, want KEEP_CALLER", kept.Status)
	}
	if kept.Other.Sessions != 0 {
		t.Fatalf("empty other sessions = %d, want 0", kept.Other.Sessions)
	}
}
