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
	pf, err := platform.NewPlatform(db, platform.NewEntErrorHandler(), nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	t.Cleanup(func() {
		rbClean(t, db)
		closeDB()
	})
	rbClean(t, db)

	uc := report.NewReportUC(logger, pf, report.NewRecordBoardRP(logger, pf), report.NewSnapshotRP(logger, pf),
		entitle.NewEntitlementRP(logger, pf), profile.NewProfileRP(logger, pf), bootstrap.NewEventSubscriptions(logger))
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

	// Face 4 — the digest: three valid sessions split by genre, the
	// cross-session bests (the retake's 100 owns reactionSpeed, the
	// first session's 70 owns aimingAccuracy), and the latest line per
	// genre carrying the anchor.
	d, err := uc.GetAbilityDigest(ownerCtx)
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	if d.SessionsTotal != 3 || d.SessionsFPS != 2 || d.SessionsMOBA != 1 {
		t.Fatalf("digest counts: total=%d fps=%d moba=%d", d.SessionsTotal, d.SessionsFPS, d.SessionsMOBA)
	}
	best := make(map[string]report.DigestAbilityRow, len(d.Abilities))
	for _, a := range d.Abilities {
		best[a.Ability] = a
	}
	if a := best[rbAbilityA]; a.Best != 100 || a.Sessions != 2 {
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
			want = s2.ID
		case shared.GenreMOBA:
			want = s3.ID
		}
		if l.SessionID != want || l.Anchor == nil {
			t.Fatalf("digest latest %s: %+v", l.Genre, l)
		}
	}

	// Face 5 — the refusals: a missing session and a foreign one.
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
