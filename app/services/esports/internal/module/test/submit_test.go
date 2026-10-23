package test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
)

// The submit seam under test: facts up, verdicts down — the response carries
// the recomputed dual-track scores and the ability receipt from the catalog
// row, and the record persists the fact stream verbatim under its oneof kind.
func TestSubmitScoresRedDot(t *testing.T) {
	rp, _, uc := submitFixture()
	out, err := uc.Submit(memberCtx(), redDotReq("red_dot", 1, hitEvent(300, 5), hitEvent(400, 20)))
	if err != nil {
		t.Fatal(err)
	}
	if !out.Valid || out.InvalidReason != nil {
		t.Fatalf("valid = %v, reason = %v, want valid with no reason", out.Valid, out.InvalidReason)
	}
	// 2 hits × unit 10 × experience 1.0; the calibration input
	// 2×(500/350)×(0.3+0.7×0.5) rides the frozen curve to 10.6, band C.
	if out.RawScore == nil || *out.RawScore != 20 {
		t.Fatalf("raw score = %v, want 20", out.RawScore)
	}
	if out.Authoritative == nil || *out.Authoritative != 10.6 {
		t.Fatalf("authoritative score = %v, want 10.6", out.Authoritative)
	}
	if out.Grade == nil || *out.Grade != "C" {
		t.Fatalf("grade = %v, want C", out.Grade)
	}
	if out.Ability == nil || *out.Ability != "reactionSpeed" {
		t.Fatalf("ability = %v, want reactionSpeed", out.Ability)
	}
	if out.RetestLeft != 0 {
		t.Fatalf("retest left = %d, want 0 (practice rides outside the session budget)", out.RetestLeft)
	}
	if len(rp.created) != 1 {
		t.Fatalf("records created = %d, want 1", len(rp.created))
	}
	rec := rp.created[0]
	if rec.UserID != "u1" {
		t.Fatalf("record user = %s, want u1", rec.UserID)
	}
	if _, ok := rec.RawDetail["red_dot"].(map[string]any); !ok {
		t.Fatalf("raw detail = %v, want a red_dot map", rec.RawDetail)
	}
}

// The validity law under test: a backgrounding span past the config budget
// invalidates the attempt — the record still persists (facts are never
// lost), both score tracks stay null, and the reason rides as ops vocabulary.
func TestSubmitInvalidOnBackgroundOverBudget(t *testing.T) {
	rp, _, uc := submitFixture()
	req := redDotReq("red_dot", 1, hitEvent(300, 5))
	req.Backgrounds = []*esportspb.BackgroundSpan{{ElapsedMs: 5000, DurationMs: 11000}}
	out, err := uc.Submit(memberCtx(), req)
	if err != nil {
		t.Fatal(err)
	}
	if out.Valid || out.InvalidReason == nil || *out.InvalidReason != reasonBackgroundOverBudget {
		t.Fatalf("valid = %v, reason = %v, want invalid with %s", out.Valid, out.InvalidReason, reasonBackgroundOverBudget)
	}
	if out.RawScore != nil || out.Authoritative != nil || out.Grade != nil {
		t.Fatalf("scores on invalid attempt = %v/%v/%v, want all nil", out.RawScore, out.Authoritative, out.Grade)
	}
	if len(rp.created) != 1 || rp.created[0].Valid {
		t.Fatalf("stored records = %+v, want one invalid record", rp.created)
	}
}

// The scorer's rejection law under test: a fabricated reaction time arrives
// as an invalid attempt carrying the scorer's reason, never as a zero score.
func TestSubmitInvalidOnScorerRejection(t *testing.T) {
	rp, _, uc := submitFixture()
	out, err := uc.Submit(memberCtx(), redDotReq("red_dot", 1, hitEvent(10, 5)))
	if err != nil {
		t.Fatal(err)
	}
	if out.Valid || out.InvalidReason == nil || *out.InvalidReason != reasonAbsurdReactionTime {
		t.Fatalf("valid = %v, reason = %v, want invalid with %s", out.Valid, out.InvalidReason, reasonAbsurdReactionTime)
	}
	if len(rp.created) != 1 || rp.created[0].RawScore != nil {
		t.Fatalf("stored records = %+v, want one unscored record", rp.created)
	}
}

// The catalog law under test: an unknown code resolves to the domain's
// not-found, while a catalogued game without a scorer is the server's gap,
// not the client's — an internal error, never an invalid attempt.
func TestSubmitUnknownCode(t *testing.T) {
	_, _, uc := submitFixture()
	_, err := uc.Submit(memberCtx(), redDotReq("nope", 1, hitEvent(300, 5)))
	if !esportspb.IsEsportsTestNotFound(err) {
		t.Fatalf("err = %v, want ESPORTS_TEST_NOT_FOUND", err)
	}
}

func TestSubmitUnscoredGame(t *testing.T) {
	rp, _, uc := submitFixture()
	// Every seeded game carries a scorer now, so the unscored arm rides a
	// code the catalog never knew — the rejection contract is the same.
	rp.tests["ghost_game"] = &Test{ID: "t13", Code: "ghost_game", Ability: utils.Ptr("resourceAllocation")}
	rp.configs["ghost_game"] = &TestConfig{BackgroundInvalidSec: 10}
	_, err := uc.Submit(memberCtx(), redDotReq("ghost_game", 1))
	if !esportspb.IsEsportsUnspecified(err) {
		t.Fatalf("err = %v, want ESPORTS_UNSPECIFIED", err)
	}
	if len(rp.created) != 0 {
		t.Fatalf("records created = %d, want 0", len(rp.created))
	}
}

func TestSubmitRequiresSession(t *testing.T) {
	_, _, uc := submitFixture()
	_, err := uc.Submit(context.Background(), redDotReq("red_dot", 1, hitEvent(300, 5)))
	if !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("err = %v, want GENERAL_ERROR_UNAUTHENTICATED", err)
	}
}

// The ordinal law under test: the wire ordinal is ignored — a practice
// attempt stores the vacuous ordinal 1 no matter what the client carried.
func TestSubmitPracticeOrdinalTruth(t *testing.T) {
	rp, _, uc := submitFixture()
	out, err := uc.Submit(memberCtx(), redDotReq("red_dot", 5, hitEvent(300, 5)))
	if err != nil {
		t.Fatal(err)
	}
	if len(rp.created) != 1 || rp.created[0].AttemptNo != 1 || rp.created[0].SessionID != nil {
		t.Fatalf("stored record = %+v, want ordinal 1 with no session", rp.created)
	}
	if out.RetestLeft != 0 {
		t.Fatalf("practice retest left = %d, want 0", out.RetestLeft)
	}
}

// The session budget law under test: ordinals count prior records in the
// session for this test plus one, the receipt walks the budget down to
// zero, and the attempt past the budget is refused outright.
func TestSubmitSessionBudget(t *testing.T) {
	rp, law, uc := submitFixture()
	law.binding = &SessionBinding{
		Difficulty:   "experience",
		RetestBudget: 3,
		Steps:        []types.MatchStep{{Ability: "reactionSpeed", TestCode: "red_dot"}},
	}
	for i, want := range []int32{2, 1, 0} {
		out, err := uc.Submit(memberCtx(), withSession(redDotReq("red_dot", int32(i+1), hitEvent(300, 5)), "sess-1"))
		if err != nil {
			t.Fatal(err)
		}
		if out.RetestLeft != want {
			t.Fatalf("attempt %d retest left = %d, want %d", i+1, out.RetestLeft, want)
		}
		if rp.created[i].AttemptNo != i+1 || *rp.created[i].SessionID != "sess-1" {
			t.Fatalf("record %d = %+v, want ordinal %d bound to sess-1", i, rp.created[i], i+1)
		}
	}
	_, err := uc.Submit(memberCtx(), withSession(redDotReq("red_dot", 4, hitEvent(300, 5)), "sess-1"))
	if !esportspb.IsEsportsMatchRetestExhausted(err) {
		t.Fatalf("err = %v, want ESPORTS_MATCH_RETEST_EXHAUSTED", err)
	}
	if len(rp.created) != 3 {
		t.Fatalf("records created = %d, want 3 (the refused attempt never stores)", len(rp.created))
	}
}

// The session miss law under test: a foreign, closed, or unknown session
// refuses the submit before anything scores.
func TestSubmitSessionNotFound(t *testing.T) {
	rp, law, uc := submitFixture()
	law.miss = true
	_, err := uc.Submit(memberCtx(), withSession(redDotReq("red_dot", 1, hitEvent(300, 5)), "sess-x"))
	if !esportspb.IsEsportsMatchSessionNotFound(err) {
		t.Fatalf("err = %v, want ESPORTS_MATCH_SESSION_NOT_FOUND", err)
	}
	if len(rp.created) != 0 {
		t.Fatalf("records created = %d, want 0", len(rp.created))
	}
}

// The lineup law under test: the session draws its own steps — a submit
// naming a test outside them is a client picking its own game.
func TestSubmitOutsideSessionSteps(t *testing.T) {
	_, law, uc := submitFixture()
	law.binding = &SessionBinding{
		Difficulty:   "experience",
		RetestBudget: 3,
		Steps:        []types.MatchStep{{Ability: "skillReleaseAccuracy", TestCode: "skill_trajectory"}},
	}
	_, err := uc.Submit(memberCtx(), withSession(redDotReq("red_dot", 1, hitEvent(300, 5)), "sess-1"))
	if !esportspb.IsEsportsMatchTestNotInSession(err) {
		t.Fatalf("err = %v, want ESPORTS_MATCH_TEST_NOT_IN_SESSION", err)
	}
}

// The mode-axis law under test: a standard session multiplies the linear
// track by the standard factor — 20 raw on experience, 30 on standard.
func TestSubmitSessionDifficulty(t *testing.T) {
	_, law, uc := submitFixture()
	law.binding = &SessionBinding{
		Difficulty:   "standard",
		RetestBudget: 1,
		Steps:        []types.MatchStep{{Ability: "reactionSpeed", TestCode: "red_dot"}},
	}
	out, err := uc.Submit(memberCtx(), withSession(redDotReq("red_dot", 1, hitEvent(300, 5), hitEvent(400, 20)), "sess-1"))
	if err != nil {
		t.Fatal(err)
	}
	if out.RawScore == nil || *out.RawScore != 30 {
		t.Fatalf("standard raw score = %v, want 30", out.RawScore)
	}
}

// The claim race under test: a concurrent submit stores the same ordinal
// between this request's count and its insert — the constraint failure
// triggers one recount pass that claims the next slot, and the receipt
// follows the stored ordinal, not the count this request read.
func TestSubmitSessionCollisionRecounts(t *testing.T) {
	rp, law, uc := submitFixture()
	law.binding = &SessionBinding{
		Difficulty:   "experience",
		RetestBudget: 3,
		Steps:        []types.MatchStep{{Ability: "reactionSpeed", TestCode: "red_dot"}},
	}
	rp.collideOnce = &AttemptRecord{ID: "rival", TestCode: "red_dot", UserID: "u1", AttemptNo: 1, SessionID: utils.Ptr("sess-1")}
	out, err := uc.Submit(memberCtx(), withSession(redDotReq("red_dot", 1, hitEvent(300, 5)), "sess-1"))
	if err != nil {
		t.Fatal(err)
	}
	if out.RetestLeft != 1 {
		t.Fatalf("retest left = %d, want 1 (the rival spent one)", out.RetestLeft)
	}
	if len(rp.created) != 2 || rp.created[1].AttemptNo != 2 || *rp.created[1].SessionID != "sess-1" {
		t.Fatalf("stored = %+v, want the rival at 1 then mine at 2 in sess-1", rp.created)
	}
}

// The exhausted-by-race arm: when the rival's record spends the last
// chance, the recount pass refuses rather than claims.
func TestSubmitSessionCollisionExhausted(t *testing.T) {
	rp, law, uc := submitFixture()
	law.binding = &SessionBinding{
		Difficulty:   "experience",
		RetestBudget: 1,
		Steps:        []types.MatchStep{{Ability: "reactionSpeed", TestCode: "red_dot"}},
	}
	rp.collideOnce = &AttemptRecord{ID: "rival", TestCode: "red_dot", UserID: "u1", AttemptNo: 1, SessionID: utils.Ptr("sess-1")}
	_, err := uc.Submit(memberCtx(), withSession(redDotReq("red_dot", 1, hitEvent(300, 5)), "sess-1"))
	if !esportspb.IsEsportsMatchRetestExhausted(err) {
		t.Fatalf("err = %v, want ESPORTS_MATCH_RETEST_EXHAUSTED", err)
	}
	if len(rp.created) != 1 {
		t.Fatalf("records created = %d, want 1 (only the rival stored)", len(rp.created))
	}
}

func withSession(req *esportspb.SubmitTestAttemptRequest, id string) *esportspb.SubmitTestAttemptRequest {
	req.MatchSessionId = utils.Ptr(id)
	return req
}

func memberCtx() context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: "u1"})
}

type fakeSubmitRP struct {
	TestRP
	tests   map[string]*Test
	configs map[string]*TestConfig
	created []*AttemptRecord
	// collideOnce makes the next CreateAttempt simulate a lost ordinal
	// race: the rival's record stores first, then this insert surfaces the
	// constraint envelope.
	collideOnce *AttemptRecord
}

func (rp *fakeSubmitRP) GetByCode(_ context.Context, code string) (*Test, error) {
	t, ok := rp.tests[code]
	if !ok {
		return nil, esportspb.ErrorEsportsTestNotFound("")
	}
	return t, nil
}

func (rp *fakeSubmitRP) GetConfigByCode(_ context.Context, code string) (*TestConfig, error) {
	return rp.configs[code], nil
}

func (rp *fakeSubmitRP) ListAnswers(_ context.Context, _ string) (map[string]map[string]any, error) {
	// The fixture's fact-stream games carry no question bank.
	return nil, nil
}

func (rp *fakeSubmitRP) CreateAttempt(_ context.Context, rec *AttemptRecord) (*AttemptRecord, error) {
	if rp.collideOnce != nil {
		rival := rp.collideOnce
		rp.collideOnce = nil
		rp.created = append(rp.created, rival)
		return nil, errorspb.ErrorInfraErrorDbConstraint("").WithCause(fmt.Errorf("duplicate key value violates unique constraint"))
	}
	rec.ID = "rec-1"
	rp.created = append(rp.created, rec)
	return rec, nil
}

func (rp *fakeSubmitRP) CountBySessionCodes(_ context.Context, sessionID string) (map[string]int, error) {
	counts := map[string]int{}
	for _, rec := range rp.created {
		if rec.SessionID != nil && *rec.SessionID == sessionID {
			counts[rec.TestCode]++
		}
	}
	return counts, nil
}

func submitFixture() (*fakeSubmitRP, *fakeSessionLaw, *TestUC) {
	rp := &fakeSubmitRP{
		tests: map[string]*Test{
			"red_dot": {ID: "t1", Code: "red_dot", Ability: utils.Ptr("reactionSpeed")},
		},
		configs: map[string]*TestConfig{
			"red_dot": {BackgroundInvalidSec: 10},
		},
	}
	law := &fakeSessionLaw{}
	return rp, law, NewTestUC(slog.Default(), nil, rp, law)
}

type fakeSessionLaw struct {
	binding *SessionBinding
	miss    bool
}

func (l *fakeSessionLaw) Resolve(_ context.Context, _, _ string) (*SessionBinding, error) {
	if l.miss || l.binding == nil {
		return nil, esportspb.ErrorEsportsMatchSessionNotFound("")
	}
	return l.binding, nil
}

func redDotReq(code string, attemptNo int32, events ...*esportspb.RedDotEvent) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		TestCode:  utils.Ptr(code),
		AttemptNo: utils.Ptr(attemptNo),
		StartedAt: timestamppb.New(time.Unix(1770000000, 0)),
		EndedAt:   timestamppb.New(time.Unix(1770000030, 0)),
		RawDetail: &esportspb.SubmitTestAttemptRequest_RedDot{RedDot: &esportspb.RedDotDetail{Events: events}},
	}
}

func hitEvent(rt int32, deviation float64) *esportspb.RedDotEvent {
	return &esportspb.RedDotEvent{
		Target:         &esportspb.Point{X: 100, Y: 100},
		Click:          &esportspb.Point{X: 101, Y: 100},
		ReactionTimeMs: rt,
		Deviation:      deviation,
		IsHit:          true,
	}
}
