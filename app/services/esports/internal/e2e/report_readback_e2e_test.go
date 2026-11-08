package e2e

// The read-back faces' live probe: GetSessionReport and GetAbilityDigest
// over the real database — the attempts receipt with the three numbers
// kept apart (this session's best, the previous session's same-ability
// score, the cross-session best), the retake board folded to one verdict
// (the composite-agreement case: a 0/0/100 retake must answer like the
// anchor's own report, never a 33rd-percentile average), the honest
// thin-pool lines, the cross-session digest, and both refusals. Gated:
// ENT_E2E=1 with the dev database up runs it; anything else skips.

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/bootstrap"
	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	entitlement "cyber-ecosystem/app/services/esports/internal/ent/entitlement"
	entmatchsession "cyber-ecosystem/app/services/esports/internal/ent/matchsession"
	entprofile "cyber-ecosystem/app/services/esports/internal/ent/profile"
	entreportpolicy "cyber-ecosystem/app/services/esports/internal/ent/reportpolicy"
	entreportsnapshot "cyber-ecosystem/app/services/esports/internal/ent/reportsnapshot"
	"cyber-ecosystem/app/services/esports/internal/ent/schema/local_mixins"
	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
	enttest "cyber-ecosystem/app/services/esports/internal/ent/test"
	enttestrecord "cyber-ecosystem/app/services/esports/internal/ent/testrecord"
	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/module/report"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	rbOwner    = "e2erbowner000000001"
	rbOther    = "e2erbother0000000001"
	rbPool     = "e2erbpool00000000001"
	rbTestA    = "e2erbtesta"
	rbTestB    = "e2erbtestb"
	rbTestD    = "e2erbtestd"
	rbTestM    = "e2erbtestm"
	rbAbilityA = "reactionSpeed"
	rbAbilityB = "aimingAccuracy"
	rbAbilityD = "decisionEfficiency"
	rbAbilityM = "trackingStability"
)

func TestReportReadBackE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the read-back end-to-end")
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
	pf, err := platform.NewPlatform(db, platform.NewEntErrorHandler(), nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	t.Cleanup(func() {
		rbClean(t, db)
		closeDB()
	})
	rbClean(t, db)

	uc := report.NewReportUC(logger, pf, report.NewRecordBoardRP(logger, pf), report.NewSnapshotRP(logger, pf),
		entitle.NewEntitlementRP(logger, pf), entitle.NewReportPolicyRP(logger, pf),
		profile.NewProfileRP(logger, pf), bootstrap.NewEventSubscriptions(logger))
	seedCtx := repSubject(rbOwner)
	now := time.Now()

	seedTest := func(code, category, ability string) {
		t.Helper()
		if _, err := db.Test.Create().SetCode(code).SetName("e2e read-back " + code).
			SetCategory(category).SetAbility(ability).Save(seedCtx); err != nil {
			t.Fatalf("seed test %s: %v", code, err)
		}
	}
	seedTest(rbTestA, "fps", rbAbilityA)
	seedTest(rbTestB, "fps", rbAbilityB)
	seedTest(rbTestD, "fps", rbAbilityD)
	seedTest(rbTestM, "moba", rbAbilityM)

	for _, u := range []string{rbOwner, rbPool} {
		if _, err := db.Profile.Create().SetUserID(u).
			SetBirthYear(2010).SetSex(entprofile.Sex("male")).Save(seedCtx); err != nil {
			t.Fatalf("profile %s: %v", u, err)
		}
	}

	step := func(code, ability string) types.MatchStep {
		return types.MatchStep{TestCode: code, Ability: ability}
	}
	newSession := func(user, genre, status string, steps ...types.MatchStep) *ent.MatchSession {
		t.Helper()
		s, serr := db.MatchSession.Create().SetUserID(user).SetGenre(genre).SetStatus(entmatchsession.Status(status)).
			SetSteps(steps).Save(seedCtx)
		if serr != nil {
			t.Fatalf("session %s/%s: %v", user, genre, serr)
		}
		return s
	}
	// Creation order is the timeline: sessions and their records are
	// inserted oldest-first, so the immutable created_at stamps carry the
	// story the read-back faces order by.
	record := func(user, code, session string, attempt int, score float64, minutesAgo int) *ent.TestRecord {
		t.Helper()
		r, rerr := db.TestRecord.Create().
			SetTestCode(code).SetUserID(user).SetSessionID(session).SetAttemptNo(attempt).
			SetStartedAt(now.Add(-time.Duration(minutesAgo) * time.Minute)).
			SetEndedAt(now.Add(-time.Duration(minutesAgo-1) * time.Minute)).
			SetRawDetail(map[string]any{"e2e": true}).
			SetValid(true).SetAuthoritativeScore(score).
			Save(seedCtx)
		if rerr != nil {
			t.Fatalf("record %s#%d: %v", code, attempt, rerr)
		}
		return r
	}
	// A curve-pending attempt: the engine wrote raw, the game's
	// calibration curve has not frozen, authoritative stays null.
	pendingRecord := func(user, code, session string, attempt int, raw float64, minutesAgo int) *ent.TestRecord {
		t.Helper()
		r, rerr := db.TestRecord.Create().
			SetTestCode(code).SetUserID(user).SetSessionID(session).SetAttemptNo(attempt).
			SetStartedAt(now.Add(-time.Duration(minutesAgo) * time.Minute)).
			SetEndedAt(now.Add(-time.Duration(minutesAgo-1) * time.Minute)).
			SetRawDetail(map[string]any{"e2e": true}).
			SetValid(true).SetRawScore(raw).
			Save(seedCtx)
		if rerr != nil {
			t.Fatalf("pending record %s#%d: %v", code, attempt, rerr)
		}
		return r
	}

	// The cohort pool rides the same ability the owner excels at, same
	// age and sex, well past the floor, mediocre scores — enough to judge
	// reactionSpeed and too thin for every other ability.
	for i := 1; i <= 11; i++ {
		record(rbPool, rbTestA, "", i, 20+float64(i), 600)
	}

	// Day one, fps: A retaken 72→78, B once at 70, D never played.
	s1 := newSession(rbOwner, "fps", "expired", step(rbTestA, rbAbilityA), step(rbTestB, rbAbilityB), step(rbTestD, rbAbilityD))
	record(rbOwner, rbTestA, s1.ID, 1, 72, 300)
	record(rbOwner, rbTestA, s1.ID, 2, 78, 290)
	record(rbOwner, rbTestB, s1.ID, 1, 70, 280)
	// Day two, fps: the composite-agreement case — A's 0/0/100 retake
	// with B at 60, the anchor landing on the winning retry.
	s2 := newSession(rbOwner, "fps", "expired", step(rbTestA, rbAbilityA), step(rbTestB, rbAbilityB))
	record(rbOwner, rbTestA, s2.ID, 1, 0, 200)
	record(rbOwner, rbTestA, s2.ID, 2, 0, 190)
	record(rbOwner, rbTestB, s2.ID, 1, 60, 180)
	a100 := record(rbOwner, rbTestA, s2.ID, 3, 100, 170)
	// Day three, moba for the owner; a same-day moba for another member.
	s3 := newSession(rbOwner, "moba", "active", step(rbTestM, rbAbilityM))
	record(rbOwner, rbTestM, s3.ID, 1, 81, 100)
	s4 := newSession(rbOther, "moba", "active", step(rbTestM, rbAbilityM))
	record(rbOther, rbTestM, s4.ID, 1, 50, 90)

	ownerCtx := repSubject(rbOwner)

	// Face 1 — the first session's receipt: the retake's best with its
	// attempts in order, a first-played line with no previous, the
	// unplayed step absent, and the cross-session words already true
	// (the later session's 100 owns the history best).
	r1, err := uc.GetSessionReport(ownerCtx, s1.ID)
	if err != nil {
		t.Fatalf("session report s1: %v", err)
	}
	if r1.Tested != 2 || r1.Total != 4 || len(r1.Abilities) != 2 {
		t.Fatalf("s1 headline: tested=%d total=%d lines=%d", r1.Tested, r1.Total, len(r1.Abilities))
	}
	if r1.Abilities[0].Ability != rbAbilityA || r1.Abilities[0].Best != 78 {
		t.Fatalf("s1 A line: %+v", r1.Abilities[0])
	}
	if got := r1.Abilities[0].Attempts; len(got) != 2 || got[0] != 72 || got[1] != 78 {
		t.Fatalf("s1 A attempts: %v", got)
	}
	if r1.Abilities[0].PrevSession != nil || r1.Abilities[1].PrevSession != nil {
		t.Fatalf("s1 prev must be absent on the first session: %+v", r1.Abilities)
	}
	if r1.Abilities[0].HistoryBest == nil || *r1.Abilities[0].HistoryBest != 100 {
		t.Fatalf("s1 A history best: %+v", r1.Abilities[0])
	}
	if r1.Abilities[1].Best != 70 || *r1.Abilities[1].HistoryBest != 70 {
		t.Fatalf("s1 B line: %+v", r1.Abilities[1])
	}

	// Face 2 — the composite-agreement case: the 0/0/100 board answers as
	// one 100 verdict — a per-attempt average would drag the composite
	// into the thirties and every line's percentile down with it.
	r2, err := uc.GetSessionReport(ownerCtx, s2.ID)
	if err != nil {
		t.Fatalf("session report s2: %v", err)
	}
	if got := r2.Abilities[0].Attempts; len(got) != 3 || got[0] != 0 || got[1] != 0 || got[2] != 100 {
		t.Fatalf("s2 A attempts: %v", got)
	}
	if r2.Abilities[0].Best != 100 || r2.Abilities[0].PrevSession == nil || *r2.Abilities[0].PrevSession != 78 {
		t.Fatalf("s2 A line: %+v", r2.Abilities[0])
	}
	if r2.Abilities[1].PrevSession == nil || *r2.Abilities[1].PrevSession != 70 {
		t.Fatalf("s2 B prev: %+v", r2.Abilities[1])
	}
	if r2.Abilities[0].Percentile == nil || *r2.Abilities[0].Percentile <= 90 {
		t.Fatalf("s2 A percentile: %+v", r2.Abilities[0])
	}
	if r2.Abilities[1].Percentile != nil {
		t.Fatalf("s2 B percentile must stay absent over a thin pool: %+v", r2.Abilities[1])
	}
	if r2.Composite == nil || *r2.Composite <= 70 || r2.Tier == shared.TierUnjudged || r2.Tier == "" {
		t.Fatalf("s2 verdict: tier=%q composite=%v", r2.Tier, r2.Composite)
	}
	if r2.Anchor == nil || *r2.Anchor != a100.ID {
		t.Fatalf("s2 anchor: %v want %s", r2.Anchor, a100.ID)
	}
	// And the anchor's own report page answers the same composite — two
	// faces, one number; the free read freezes nothing on its way.
	rep2, err := uc.GetReport(ownerCtx, a100.ID)
	if err != nil {
		t.Fatalf("anchor report: %v", err)
	}
	if rep2.Composite == nil || *rep2.Composite != *r2.Composite {
		t.Fatalf("composite disagreement: session=%v report=%v", r2.Composite, rep2.Composite)
	}
	if n := snapshotCount(t, db, ownerCtx, a100.ID); n != 0 {
		t.Fatalf("free read froze, snapshots=%d", n)
	}

	// Face 3 — the moba session over its own thin pool: one line, an
	// honestly unjudged verdict.
	r3, err := uc.GetSessionReport(ownerCtx, s3.ID)
	if err != nil {
		t.Fatalf("session report s3: %v", err)
	}
	if r3.Tested != 1 || r3.Total != 4 || len(r3.Abilities) != 1 || r3.Abilities[0].Best != 81 {
		t.Fatalf("s3 line: %+v", r3)
	}
	// The session face's thin-pool honesty is the empty tier (the
	// unjudged sentinel is the report face's own spelling of the same
	// answer, spelled at the service bridge).
	if r3.Tier != "" || r3.Composite != nil {
		t.Fatalf("s3 must answer unjudged over a thin pool: tier=%q composite=%v", r3.Tier, r3.Composite)
	}

	// Face 4 — the curve-pending session: a game whose calibration curve
	// has not frozen (raw landed, authoritative null) keeps every attempt
	// in the detail rows while feeding no line math — the score a member
	// earned never waits on the curve. The calibrated sibling rides the
	// same response: both states, one face.
	s5 := newSession(rbOwner, "fps", "expired", step(rbTestD, rbAbilityD), step(rbTestA, rbAbilityA))
	pendingRecord(rbOwner, rbTestD, s5.ID, 1, 420, 62)
	pendingRecord(rbOwner, rbTestD, s5.ID, 2, 400, 61)
	a88 := record(rbOwner, rbTestA, s5.ID, 1, 88, 60)
	r4, err := uc.GetSessionReport(ownerCtx, s5.ID)
	if err != nil {
		t.Fatalf("session report s5: %v", err)
	}
	if r4.Tested != 1 || len(r4.Abilities) != 1 || r4.Abilities[0].Ability != rbAbilityA || r4.Abilities[0].Best != 88 {
		t.Fatalf("s5 lines: tested=%d %+v", r4.Tested, r4.Abilities)
	}
	if len(r4.AttemptDetails) != 3 {
		t.Fatalf("s5 details: %+v", r4.AttemptDetails)
	}
	if d0 := r4.AttemptDetails[0]; d0.TestCode != rbTestD || d0.AttemptNo != 1 || d0.Raw != 420 || d0.Authoritative != nil || d0.Grade != nil {
		t.Fatalf("s5 pending row 1: %+v", d0)
	}
	if d1 := r4.AttemptDetails[1]; d1.TestCode != rbTestD || d1.AttemptNo != 2 || d1.Raw != 400 || d1.Authoritative != nil {
		t.Fatalf("s5 pending row 2: %+v", d1)
	}
	if d2 := r4.AttemptDetails[2]; d2.TestCode != rbTestA || d2.Authoritative == nil || *d2.Authoritative != 88 {
		t.Fatalf("s5 calibrated row: %+v", d2)
	}
	if r4.Anchor == nil || *r4.Anchor != a88.ID {
		t.Fatalf("s5 anchor: %v want %s", r4.Anchor, a88.ID)
	}

	// The all-pending session is the finding's own shape: nothing to
	// judge, everything still visible — the report answers the rows, not
	// an empty board.
	s6 := newSession(rbOwner, "fps", "expired", step(rbTestD, rbAbilityD))
	pendingRecord(rbOwner, rbTestD, s6.ID, 1, 410, 59)
	r4b, err := uc.GetSessionReport(ownerCtx, s6.ID)
	if err != nil {
		t.Fatalf("session report s6: %v", err)
	}
	if len(r4b.Abilities) != 0 || r4b.Tested != 0 || r4b.Anchor != nil || r4b.Tier != "" || r4b.Composite != nil {
		t.Fatalf("s6 must answer no verdict: %+v", r4b)
	}
	if len(r4b.AttemptDetails) != 1 || r4b.AttemptDetails[0].Raw != 410 {
		t.Fatalf("s6 details: %+v", r4b.AttemptDetails)
	}

	// Face 5 — the digest: five valid-attempt sessions split by genre
	// (the curve-pending-only s6 counts in the headline and owns the
	// latest fps line — both laws say "carried a valid attempt", the
	// anchor field itself carrying the honest absent), the cross-session
	// bests (the retake's 100 owns reactionSpeed, the first session's 70
	// owns aimingAccuracy), and the moba latest line carrying its anchor.
	d, err := uc.GetAbilityDigest(ownerCtx)
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	if d.SessionsTotal != 5 || d.SessionsFPS != 4 || d.SessionsMOBA != 1 {
		t.Fatalf("digest counts: total=%d fps=%d moba=%d", d.SessionsTotal, d.SessionsFPS, d.SessionsMOBA)
	}
	// The entry cards' gates ride the digest (the owner carries no grant):
	// both verdicts answer the honest closed state, so the center page's
	// unlock tags render from this one call.
	if d.ReportAllowed || d.ReportReason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_NO_ENTITLEMENT {
		t.Fatalf("digest report gate: allowed=%v reason=%v", d.ReportAllowed, d.ReportReason)
	}
	if d.TrendAdvancedAllowed || d.TrendAdvancedReason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_NO_ENTITLEMENT {
		t.Fatalf("digest trend gate: allowed=%v reason=%v", d.TrendAdvancedAllowed, d.TrendAdvancedReason)
	}
	// The member plane's two facts on a never-member: zero unlocks, no
	// window end to quote.
	if d.MembershipUnlockedReports != 0 || d.LastMembershipUntil != nil {
		t.Fatalf("digest member facts on never-member: count=%d until=%v", d.MembershipUnlockedReports, d.LastMembershipUntil)
	}
	best := make(map[string]report.DigestAbilityRow, len(d.Abilities))
	for _, a := range d.Abilities {
		best[a.Ability] = a
	}
	if a := best[rbAbilityA]; a.Best != 100 || a.Sessions != 3 {
		t.Fatalf("digest A: %+v", a)
	}
	if b := best[rbAbilityB]; b.Best != 70 || b.Sessions != 2 {
		t.Fatalf("digest B: %+v", b)
	}
	if m := best[rbAbilityM]; m.Best != 81 || m.Sessions != 1 {
		t.Fatalf("digest M: %+v", m)
	}
	if len(d.Sessions) != 2 {
		t.Fatalf("digest latest lines: %+v", d.Sessions)
	}
	for _, l := range d.Sessions {
		var want string
		switch l.Genre {
		case shared.GenreFPS:
			want = s6.ID
		case shared.GenreMOBA:
			want = s3.ID
		}
		if l.SessionID != want {
			t.Fatalf("digest latest %s: %+v", l.Genre, l)
		}
	}
	// The pending-only latest line carries its anchor honestly absent;
	// the calibrated one carries a real anchor.
	for _, l := range d.Sessions {
		if l.Genre == shared.GenreFPS && l.Anchor != nil {
			t.Fatalf("digest latest fps anchor must stay absent: %+v", l)
		}
		if l.Genre == shared.GenreMOBA && l.Anchor == nil {
			t.Fatalf("digest latest moba anchor: %+v", l)
		}
	}

	// Face 6 — the refusals: a missing session and a foreign one.
	if _, err := uc.GetSessionReport(ownerCtx, "e2e-no-such-session"); !esportspb.IsEsportsMatchSessionNotFound(err) {
		t.Fatalf("missing session: %v", err)
	}
	if _, err := uc.GetSessionReport(ownerCtx, s4.ID); !esportspb.IsEsportsReportSessionNotOwned(err) {
		t.Fatalf("foreign session: %v", err)
	}
}

func rbClean(t *testing.T, db *ent.Client) {
	t.Helper()
	// Hard deletes under SkipSoftDelete so reruns see a clean slate; the
	// deletes key on this probe's own users and test codes, never a bare
	// table wipe — the dev database carries real seed rows.
	dctx := local_mixins.SkipSoftDelete(context.Background())
	users := []string{rbOwner, rbOther, rbPool}
	if _, err := db.ReportSnapshot.Delete().Where(entreportsnapshot.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean snapshots: %v", err)
	}
	if _, err := db.Entitlement.Delete().Where(entitlement.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean grants: %v", err)
	}
	if _, err := db.TestRecord.Delete().Where(enttestrecord.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean records: %v", err)
	}
	if _, err := db.MatchSession.Delete().Where(entmatchsession.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean sessions: %v", err)
	}
	if _, err := db.Profile.Delete().Where(entprofile.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean profiles: %v", err)
	}
	for _, code := range []string{rbTestA, rbTestB, rbTestD, rbTestM} {
		if _, err := db.Test.Delete().Where(enttest.Code(code)).Exec(dctx); err != nil {
			t.Fatalf("clean test %s: %v", code, err)
		}
	}
}

const (
	rbqFull    = "e2erbqfull000000001"
	rbqPart    = "e2erbqpart000000001"
	rbqPool    = "e2erbqpool000000001"
	rbqTop     = "e2erbqtop0000000001"
	rbqTopPool = "e2erbqtoppool0000001"

	rbqRS  = "e2erbqrs"
	rbqAA  = "e2erbqaa"
	rbqTS  = "e2erbqts"
	rbqSRA = "e2erbqsra"
	rbqRA  = "e2erbqra"
	rbqDE  = "e2erbqde"
)

// The cluster faces' live probe over a degenerate cohort — every pool
// row sits at the same score, so each ability's cohort answer is the
// deterministic side-bench 50 and the golden expectations carry no
// hand-computed z-scores: all eight dimensions at 50, the composite 50,
// tier C. A partially-fed member pins the missing-dimension law (four
// of six abilities feed exactly two of eight dimensions), the trend
// chronology rides the per-point genre, and the digest's percentile
// column gets its promised seat back. Gated like its sibling: ENT_E2E=1
// with the dev database up.
func TestReportClusterE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the cluster end-to-end")
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
	pf, err := platform.NewPlatform(db, platform.NewEntErrorHandler(), nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	t.Cleanup(func() {
		rbqClean(t, db)
		closeDB()
	})
	rbqClean(t, db)

	policyRP := entitle.NewReportPolicyRP(logger, pf)
	uc := report.NewReportUC(logger, pf, report.NewRecordBoardRP(logger, pf), report.NewSnapshotRP(logger, pf),
		entitle.NewEntitlementRP(logger, pf), policyRP,
		profile.NewProfileRP(logger, pf), bootstrap.NewEventSubscriptions(logger))
	seedCtx := repSubject(rbqFull)
	now := time.Now()

	seedTest := func(code, ability string) {
		t.Helper()
		if _, err := db.Test.Create().SetCode(code).SetName("e2e cluster " + code).
			SetCategory("fps").SetAbility(ability).Save(seedCtx); err != nil {
			t.Fatalf("seed test %s: %v", code, err)
		}
	}
	six := []struct{ code, ability string }{
		{rbqRS, "reactionSpeed"}, {rbqAA, "aimingAccuracy"}, {rbqTS, "trackingStability"},
		{rbqSRA, "skillReleaseAccuracy"}, {rbqRA, "resourceAllocation"}, {rbqDE, "decisionEfficiency"},
	}
	for _, s := range six {
		seedTest(s.code, s.ability)
	}
	for _, u := range []string{rbqFull, rbqPart, rbqPool} {
		if _, err := db.Profile.Create().SetUserID(u).
			SetBirthYear(2010).SetSex(entprofile.Sex("male")).Save(seedCtx); err != nil {
			t.Fatalf("profile %s: %v", u, err)
		}
	}
	step := func(code, ability string) types.MatchStep {
		return types.MatchStep{TestCode: code, Ability: ability}
	}
	newSession := func(user, genre, status string, steps ...types.MatchStep) *ent.MatchSession {
		t.Helper()
		s, serr := db.MatchSession.Create().SetUserID(user).SetGenre(genre).SetStatus(entmatchsession.Status(status)).
			SetSteps(steps).Save(seedCtx)
		if serr != nil {
			t.Fatalf("session %s/%s: %v", user, genre, serr)
		}
		return s
	}
	record := func(user, code, session string, attempt int, score float64, minutesAgo int) {
		t.Helper()
		if _, rerr := db.TestRecord.Create().
			SetTestCode(code).SetUserID(user).SetSessionID(session).SetAttemptNo(attempt).
			SetStartedAt(now.Add(-time.Duration(minutesAgo) * time.Minute)).
			SetEndedAt(now.Add(-time.Duration(minutesAgo-1) * time.Minute)).
			SetRawDetail(map[string]any{"e2e": true}).
			SetValid(true).SetAuthoritativeScore(score).
			Save(seedCtx); rerr != nil {
			t.Fatalf("record %s#%d: %v", code, attempt, rerr)
		}
	}

	// The degenerate pool: ten rows per ability, all at 100 — past the
	// confidence floor, std zero, so every bench lands on the side's 50.
	for _, s := range six {
		for i := 1; i <= 10; i++ {
			record(rbqPool, s.code, "", i, 100, 600)
		}
	}
	// The full member's two sessions, genres mixed so a series carries
	// both genres and the per-point genre label has something to say.
	s1 := newSession(rbqFull, "fps", "expired",
		step(rbqRS, "reactionSpeed"), step(rbqAA, "aimingAccuracy"), step(rbqTS, "trackingStability"), step(rbqSRA, "skillReleaseAccuracy"))
	record(rbqFull, rbqRS, s1.ID, 1, 100, 300)
	record(rbqFull, rbqAA, s1.ID, 1, 100, 290)
	record(rbqFull, rbqTS, s1.ID, 1, 100, 280)
	record(rbqFull, rbqSRA, s1.ID, 1, 100, 270)
	s2 := newSession(rbqFull, "moba", "active",
		step(rbqRS, "reactionSpeed"), step(rbqSRA, "skillReleaseAccuracy"), step(rbqRA, "resourceAllocation"), step(rbqDE, "decisionEfficiency"))
	record(rbqFull, rbqRS, s2.ID, 1, 100, 200)
	record(rbqFull, rbqSRA, s2.ID, 1, 100, 190)
	record(rbqFull, rbqRA, s2.ID, 1, 100, 180)
	record(rbqFull, rbqDE, s2.ID, 1, 100, 170)
	// The partial member: four of six abilities only.
	sp := newSession(rbqPart, "fps", "expired",
		step(rbqRS, "reactionSpeed"), step(rbqAA, "aimingAccuracy"), step(rbqTS, "trackingStability"), step(rbqSRA, "skillReleaseAccuracy"))
	record(rbqPart, rbqRS, sp.ID, 1, 100, 250)
	record(rbqPart, rbqAA, sp.ID, 1, 100, 240)
	record(rbqPart, rbqTS, sp.ID, 1, 100, 230)
	record(rbqPart, rbqSRA, sp.ID, 1, 100, 220)

	// The top member rides its own cohort (birth 2009 — the key the
	// degradation law benches by), its pool spread so the z-score speaks:
	// the member's 100s sit far above a 15-ish average and land on a
	// deterministic 100, the anchor an at-the-line row needs. Only
	// neuralResponse's three sources are measured, so exactly one
	// dimension renders.
	for _, u := range []string{rbqTop, rbqTopPool} {
		if _, err := db.Profile.Create().SetUserID(u).
			SetBirthYear(2009).SetSex(entprofile.Sex("male")).Save(seedCtx); err != nil {
			t.Fatalf("profile %s: %v", u, err)
		}
	}
	// The pool rows hang on their own session: the record unique key is
	// (session, code, attempt) with no user in it, so a bare "" session
	// would collide with the degenerate pool's same-code rows.
	spool := newSession(rbqTopPool, "fps", "expired",
		step(rbqRS, "reactionSpeed"), step(rbqAA, "aimingAccuracy"), step(rbqTS, "trackingStability"))
	for _, s := range []struct{ code, ability string }{
		{rbqRS, "reactionSpeed"}, {rbqAA, "aimingAccuracy"}, {rbqTS, "trackingStability"},
	} {
		for i := 1; i <= 10; i++ {
			record(rbqTopPool, s.code, spool.ID, i, 10+float64(i), 600)
		}
	}
	st := newSession(rbqTop, "fps", "expired",
		step(rbqRS, "reactionSpeed"), step(rbqAA, "aimingAccuracy"), step(rbqTS, "trackingStability"))
	record(rbqTop, rbqRS, st.ID, 1, 100, 160)
	record(rbqTop, rbqAA, st.ID, 1, 100, 150)
	record(rbqTop, rbqTS, st.ID, 1, 100, 140)

	memberGrant := func(user string) {
		t.Helper()
		if _, gerr := db.Entitlement.Create().SetUserID(user).
			SetScopeKind(entitlement.ScopeKind(shared.ScopeMembership)).
			SetSource("e2e").SetExpiresAt(now.Add(24 * time.Hour)).Save(seedCtx); gerr != nil {
			t.Fatalf("membership grant %s: %v", user, gerr)
		}
	}

	fullCtx := repSubject(rbqFull)

	// Face 1 — the gate first: cognitive ships membership-gated, so with
	// no grant the numbers never travel; the covered count stays public.
	c1, err := uc.GetCognitiveReport(fullCtx)
	if err != nil {
		t.Fatalf("cognitive ungated: %v", err)
	}
	if c1.Allowed || c1.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_NO_ENTITLEMENT {
		t.Fatalf("cognitive gate: allowed=%v reason=%v", c1.Allowed, c1.Reason)
	}
	// The to-trend CTA's gate rides the same face: closed before the
	// grant, so the upsell renders; open after, so a member never meets it.
	if c1.TrendAdvancedAllowed {
		t.Fatalf("cognitive trend gate pre-grant: %+v", c1.TrendAdvancedAllowed)
	}
	if c1.Covered != 6 || c1.Tested != 0 || c1.Total != 8 || len(c1.Dimensions) != 0 || c1.Composite != nil {
		t.Fatalf("cognitive gated shape: %+v", c1)
	}

	// The membership grant opens the gate; the degenerate pool feeds all
	// eight dimensions the same 50.
	memberGrant(rbqFull)
	c2, err := uc.GetCognitiveReport(fullCtx)
	if err != nil {
		t.Fatalf("cognitive gated: %v", err)
	}
	if !c2.Allowed || c2.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE {
		t.Fatalf("cognitive open gate: allowed=%v reason=%v", c2.Allowed, c2.Reason)
	}
	if !c2.TrendAdvancedAllowed || c2.TrendAdvancedReason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE {
		t.Fatalf("cognitive trend gate post-grant: allowed=%v reason=%v", c2.TrendAdvancedAllowed, c2.TrendAdvancedReason)
	}
	if c2.Tested != 8 || len(c2.Dimensions) != 8 {
		t.Fatalf("cognitive fed dims: tested=%d rows=%d", c2.Tested, len(c2.Dimensions))
	}
	if c2.Dimensions[0].Dimension != "neuralResponse" || c2.Dimensions[7].Dimension != "creativeThinking" {
		t.Fatalf("cognitive dim order: %s..%s", c2.Dimensions[0].Dimension, c2.Dimensions[7].Dimension)
	}
	for _, d := range c2.Dimensions {
		if d.Value != 50 || d.Tier != shared.TierC {
			t.Fatalf("cognitive dim %s: %+v", d.Dimension, d)
		}
	}
	if c2.Composite == nil || *c2.Composite != 50 || c2.Tier != shared.TierC {
		t.Fatalf("cognitive composite: %+v", c2)
	}

	// Face 2 — the missing-dimension law: four abilities feed exactly
	// neuralResponse (RS+AA+TS) and motorCoordination (AA+TS+SRA); the
	// other six stay absent rows, never zero axes.
	memberGrant(rbqPart)
	cp, err := uc.GetCognitiveReport(repSubject(rbqPart))
	if err != nil {
		t.Fatalf("cognitive partial: %v", err)
	}
	if !cp.Allowed || cp.Covered != 4 || cp.Tested != 2 || len(cp.Dimensions) != 2 {
		t.Fatalf("cognitive partial shape: %+v", cp)
	}
	if cp.Dimensions[0].Dimension != "neuralResponse" || cp.Dimensions[1].Dimension != "motorCoordination" {
		t.Fatalf("cognitive partial dims: %+v", cp.Dimensions)
	}

	// Face 3 — the trend chronology: two sessions, the RS series
	// carrying one point per genre at the degenerate 50, the series
	// order the vocabulary's own.
	tr, err := uc.GetTrendReport(fullCtx)
	if err != nil {
		t.Fatalf("trend: %v", err)
	}
	if tr.Sessions != 2 || len(tr.Abilities) != 6 {
		t.Fatalf("trend headline: sessions=%d series=%d", tr.Sessions, len(tr.Abilities))
	}
	if !tr.AdvancedAllowed || tr.AdvancedReason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE {
		t.Fatalf("trend advanced gate: %+v", tr)
	}
	rs := tr.Abilities[0]
	if rs.Ability != "reactionSpeed" || len(rs.Points) != 2 {
		t.Fatalf("trend RS series: %+v", rs)
	}
	if rs.Points[0].Genre != shared.GenreFPS || rs.Points[0].Percentile != 50 ||
		rs.Points[1].Genre != shared.GenreMOBA || rs.Points[1].Percentile != 50 {
		t.Fatalf("trend RS points: %+v", rs.Points)
	}
	if tr.Abilities[5].Ability != "decisionEfficiency" || len(tr.Abilities[5].Points) != 1 {
		t.Fatalf("trend DE series: %+v", tr.Abilities[5])
	}

	// The honest empty page still knows its gate: no sessions, no
	// series, and the advanced verdict judged all the same.
	tp, err := uc.GetTrendReport(repSubject(rbqPool))
	if err != nil {
		t.Fatalf("trend empty: %v", err)
	}
	if tp.Sessions != 0 || len(tp.Abilities) != 0 || tp.AdvancedAllowed ||
		tp.AdvancedReason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_NO_ENTITLEMENT {
		t.Fatalf("trend empty shape: %+v", tp)
	}

	// The partial member's chronology carries exactly the measured four —
	// the missing-dimension law's trend face: an unmeasured ability has no
	// series, never a flat line at zero.
	trp, err := uc.GetTrendReport(repSubject(rbqPart))
	if err != nil {
		t.Fatalf("trend partial: %v", err)
	}
	if trp.Sessions != 1 || len(trp.Abilities) != 4 {
		t.Fatalf("trend partial shape: %+v", trp)
	}
	if trp.Abilities[0].Ability != "reactionSpeed" || len(trp.Abilities[0].Points) != 1 {
		t.Fatalf("trend partial RS: %+v", trp.Abilities[0])
	}

	// Face 4 — the digest's promised seat: every best now carries its
	// cohort answer, the degenerate 50.
	dg, err := uc.GetAbilityDigest(fullCtx)
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	if len(dg.Abilities) != 6 {
		t.Fatalf("digest rows: %d", len(dg.Abilities))
	}
	for _, a := range dg.Abilities {
		if a.Percentile == nil || *a.Percentile != 50 {
			t.Fatalf("digest %s percentile: %+v", a.Ability, a)
		}
	}

	// Face 5 — the career aggregate: open by the standing default, the
	// seeded gates shut on the degenerate 50 (every threshold sits above
	// it), and the prescription board carries the four weakest rendered
	// dimensions against the system's own excellence line — the same 50
	// everywhere, so the weakest four are the model's own order.
	cr, err := uc.GetCareerReport(fullCtx)
	if err != nil {
		t.Fatalf("career: %v", err)
	}
	if !cr.Allowed || cr.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_OPEN {
		t.Fatalf("career open gate: allowed=%v reason=%v", cr.Allowed, cr.Reason)
	}
	if len(cr.Recommendations) != 0 {
		t.Fatalf("career shut gates: %+v", cr.Recommendations)
	}
	if len(cr.Improvements) != 4 {
		t.Fatalf("career improvements: %d", len(cr.Improvements))
	}
	for i, want := range []string{"neuralResponse", "sustainedAttention", "motorCoordination", "stressResilience"} {
		imp := cr.Improvements[i]
		if imp.Dimension != want || imp.Current != 50 || imp.Target != 85 || imp.GapPercent != 41 {
			t.Fatalf("career improvement %s: %+v", want, imp)
		}
		if len(imp.Suggestions) != 3 || imp.Suggestions[0] == "已达优秀水平" {
			t.Fatalf("career improvement %s suggestions: %+v", want, imp.Suggestions)
		}
	}

	// The partial member's two rendered dimensions feed no area — a
	// gate with a missing basis dimension is unjudgeable, and the board
	// carries exactly the two rendered rows.
	crp, err := uc.GetCareerReport(repSubject(rbqPart))
	if err != nil {
		t.Fatalf("career partial: %v", err)
	}
	if !crp.Allowed || len(crp.Recommendations) != 0 || len(crp.Improvements) != 2 {
		t.Fatalf("career partial shape: %+v", crp)
	}
	if crp.Improvements[0].Dimension != "neuralResponse" || crp.Improvements[1].Dimension != "motorCoordination" {
		t.Fatalf("career partial board: %+v", crp.Improvements)
	}

	// A never-tested member renders the honest empty page under the open
	// gate — no bests feed any dimension, so neither the areas nor the
	// board have anything to state. (The pool member is not that case:
	// their own pool rows feed their dimensions all the same 50.)
	crx, err := uc.GetCareerReport(repSubject("e2erbqnone00000000"))
	if err != nil {
		t.Fatalf("career empty: %v", err)
	}
	if !crx.Allowed || len(crx.Recommendations) != 0 || len(crx.Improvements) != 0 {
		t.Fatalf("career empty shape: %+v", crx)
	}

	// Face 6 — the at-the-line row, the zero-gap regression line: an
	// already-excellent dimension states a zero gap and the at-the-line
	// copy. On the wire protojson drops a zero scalar's field entirely, so
	// this row is the exact one whose gap must survive a client seam's
	// row-level narrowing — a golden gap of 41 never exercises it.
	memberGrant(rbqTop)
	crt, err := uc.GetCareerReport(repSubject(rbqTop))
	if err != nil {
		t.Fatalf("career top: %v", err)
	}
	if !crt.Allowed || crt.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_OPEN {
		t.Fatalf("career top gate: allowed=%v reason=%v", crt.Allowed, crt.Reason)
	}
	if len(crt.Recommendations) != 0 { // every area's gate set needs a dimension the three measured abilities never feed
		t.Fatalf("career top gates: %+v", crt.Recommendations)
	}
	if len(crt.Improvements) != 1 {
		t.Fatalf("career top board: %d rows", len(crt.Improvements))
	}
	top := crt.Improvements[0]
	if top.Dimension != "neuralResponse" || top.Current != 100 || top.Target != 85 || top.GapPercent != 0 {
		t.Fatalf("career at-the-line row: %+v", top)
	}
	if len(top.Suggestions) != 1 || top.Suggestions[0] != "已达优秀水平" {
		t.Fatalf("career at-the-line copy: %+v", top.Suggestions)
	}
	// The same facts through the cognitive face: one dimension at 100,
	// the composite its equal-weight mean.
	ct, err := uc.GetCognitiveReport(repSubject(rbqTop))
	if err != nil {
		t.Fatalf("cognitive top: %v", err)
	}
	if !ct.Allowed || ct.Covered != 3 || ct.Tested != 1 || len(ct.Dimensions) != 1 {
		t.Fatalf("cognitive top shape: %+v", ct)
	}
	if ct.Dimensions[0].Dimension != "neuralResponse" || ct.Dimensions[0].Value != 100 ||
		ct.Composite == nil || *ct.Composite != 100 {
		t.Fatalf("cognitive top values: %+v", ct)
	}

	// Face 7 — the policy flip: an explicit closed row is the operator's
	// own word, and it outranks a membership; a redeemed family code is
	// the order's first fact and outranks the row itself. Behind a closed
	// gate the numbers never travel.
	if err := policyRP.SetPolicy(seedCtx, shared.ReportCareer, shared.PolicyClosed, "e2e"); err != nil {
		t.Fatalf("close career: %v", err)
	}
	crf, err := uc.GetCareerReport(fullCtx)
	if err != nil {
		t.Fatalf("career closed: %v", err)
	}
	if crf.Allowed || crf.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_CLOSED ||
		len(crf.Recommendations) != 0 || len(crf.Improvements) != 0 {
		t.Fatalf("career closed shape: %+v", crf)
	}
	// The redeemed code answers before the row is even read.
	if _, gerr := db.Entitlement.Create().SetUserID(rbqTop).
		SetScopeKind(entitlement.ScopeKind(shared.ScopeReportType)).
		SetScopeRef(shared.ReportCareer).
		SetSource("e2e").SetExpiresAt(now.Add(24 * time.Hour)).Save(seedCtx); gerr != nil {
		t.Fatalf("career code grant: %v", gerr)
	}
	crc, err := uc.GetCareerReport(repSubject(rbqTop))
	if err != nil {
		t.Fatalf("career coded: %v", err)
	}
	if !crc.Allowed || crc.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_CODE_REDEEMED {
		t.Fatalf("career coded gate: allowed=%v reason=%v", crc.Allowed, crc.Reason)
	}
	if len(crc.Improvements) != 1 || crc.Improvements[0].GapPercent != 0 {
		t.Fatalf("career coded board: %+v", crc.Improvements)
	}
	// Cognitive flips the same way; the covered count stays public, the
	// numbers behind the gate do not travel.
	if err := policyRP.SetPolicy(seedCtx, shared.ReportCognitive, shared.PolicyClosed, "e2e"); err != nil {
		t.Fatalf("close cognitive: %v", err)
	}
	cf, err := uc.GetCognitiveReport(fullCtx)
	if err != nil {
		t.Fatalf("cognitive closed: %v", err)
	}
	if cf.Allowed || cf.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_CLOSED ||
		cf.Covered != 6 || cf.Tested != 0 || len(cf.Dimensions) != 0 || cf.Composite != nil {
		t.Fatalf("cognitive closed shape: %+v", cf)
	}
	// The trend face's advanced verdict rides membership, never the
	// policy families — a closed cognitive row must not lock the
	// chronology or its advanced tier.
	trf, err := uc.GetTrendReport(fullCtx)
	if err != nil {
		t.Fatalf("trend under closed: %v", err)
	}
	if !trf.AdvancedAllowed || trf.AdvancedReason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE ||
		trf.Sessions != 2 || len(trf.Abilities) != 6 {
		t.Fatalf("trend under closed shape: %+v", trf)
	}
}

func rbqClean(t *testing.T, db *ent.Client) {
	t.Helper()
	dctx := local_mixins.SkipSoftDelete(context.Background())
	users := []string{rbqFull, rbqPart, rbqPool, rbqTop, rbqTopPool}
	// The flip rows this probe owns: updated_by is the probe's own marker,
	// so an operator's row never sweeps.
	if _, err := db.ReportPolicy.Delete().Where(
		entreportpolicy.ReportTypeIn(entreportpolicy.ReportType(shared.ReportCareer), entreportpolicy.ReportType(shared.ReportCognitive)),
		entreportpolicy.UpdatedByEQ("e2e"),
	).Exec(dctx); err != nil {
		t.Fatalf("clean policies: %v", err)
	}
	if _, err := db.ReportSnapshot.Delete().Where(entreportsnapshot.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean snapshots: %v", err)
	}
	if _, err := db.Entitlement.Delete().Where(entitlement.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean grants: %v", err)
	}
	if _, err := db.TestRecord.Delete().Where(enttestrecord.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean records: %v", err)
	}
	if _, err := db.MatchSession.Delete().Where(entmatchsession.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean sessions: %v", err)
	}
	if _, err := db.Profile.Delete().Where(entprofile.UserIDIn(users...)).Exec(dctx); err != nil {
		t.Fatalf("clean profiles: %v", err)
	}
	for _, code := range []string{rbqRS, rbqAA, rbqTS, rbqSRA, rbqRA, rbqDE} {
		if _, err := db.Test.Delete().Where(enttest.Code(code)).Exec(dctx); err != nil {
			t.Fatalf("clean test %s: %v", code, err)
		}
	}
}
