package test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// The submit seam under test: facts up, verdicts down — the response carries
// the recomputed dual-track scores and the ability receipt from the catalog
// row, and the record persists the fact stream verbatim under its oneof kind.
func TestSubmitScoresRedDot(t *testing.T) {
	rp, uc := submitFixture()
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
	if out.RetestLeft != 3 {
		t.Fatalf("retest left = %d, want 3", out.RetestLeft)
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
	rp, uc := submitFixture()
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
	rp, uc := submitFixture()
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
	_, uc := submitFixture()
	_, err := uc.Submit(memberCtx(), redDotReq("nope", 1, hitEvent(300, 5)))
	if !esportspb.IsEsportsTestNotFound(err) {
		t.Fatalf("err = %v, want ESPORTS_TEST_NOT_FOUND", err)
	}
}

func TestSubmitUnscoredGame(t *testing.T) {
	rp, uc := submitFixture()
	rp.tests["economic_sim"] = &Test{ID: "t13", Code: "economic_sim", Ability: utils.Ptr("resourceAllocation")}
	rp.configs["economic_sim"] = &TestConfig{BackgroundInvalidSec: 10}
	_, err := uc.Submit(memberCtx(), redDotReq("economic_sim", 1))
	if !esportspb.IsEsportsUnspecified(err) {
		t.Fatalf("err = %v, want ESPORTS_UNSPECIFIED", err)
	}
	if len(rp.created) != 0 {
		t.Fatalf("records created = %d, want 0", len(rp.created))
	}
}

func TestSubmitRequiresSession(t *testing.T) {
	_, uc := submitFixture()
	_, err := uc.Submit(context.Background(), redDotReq("red_dot", 1, hitEvent(300, 5)))
	if !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("err = %v, want GENERAL_ERROR_UNAUTHENTICATED", err)
	}
}

func TestSubmitRejectsZeroOrdinal(t *testing.T) {
	_, uc := submitFixture()
	_, err := uc.Submit(memberCtx(), redDotReq("red_dot", 0))
	if !errorspb.IsGeneralErrorValidationFailed(err) {
		t.Fatalf("err = %v, want GENERAL_ERROR_VALIDATION_FAILED", err)
	}
}

// The ordinal law under test: attempt numbers start at 1 and the receipt
// consumes the first-session budget — 3 left after the first play, clamped
// at zero once spent.
func TestSubmitRetestBudget(t *testing.T) {
	_, uc := submitFixture()
	for _, c := range []struct {
		attempt int32
		want    int32
	}{{1, 3}, {2, 2}, {4, 0}, {9, 0}} {
		out, err := uc.Submit(memberCtx(), redDotReq("red_dot", c.attempt, hitEvent(300, 5)))
		if err != nil {
			t.Fatal(err)
		}
		if out.RetestLeft != c.want {
			t.Fatalf("attempt %d retest left = %d, want %d", c.attempt, out.RetestLeft, c.want)
		}
	}
}

func memberCtx() context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: "u1"})
}

type fakeSubmitRP struct {
	TestRP
	tests   map[string]*Test
	configs map[string]*TestConfig
	created []*AttemptRecord
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
	rec.ID = "rec-1"
	rp.created = append(rp.created, rec)
	return rec, nil
}

func submitFixture() (*fakeSubmitRP, *TestUC) {
	rp := &fakeSubmitRP{
		tests: map[string]*Test{
			"red_dot": {ID: "t1", Code: "red_dot", Ability: utils.Ptr("reactionSpeed")},
		},
		configs: map[string]*TestConfig{
			"red_dot": {BackgroundInvalidSec: 10},
		},
	}
	return rp, NewTestUC(slog.Default(), nil, rp)
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
