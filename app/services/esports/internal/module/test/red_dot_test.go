package test

import (
	"errors"
	"testing"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// The law under test: the frozen red_dot card — the formula, its clamp
// window, the outlier rejection, and the 217-run quantile anchors.
func redDotRequest(events ...*esportspb.RedDotEvent) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_RedDot{
			RedDot: &esportspb.RedDotDetail{Events: events},
		},
	}
}

func redDotEvent(isHit bool, rt int32, deviation float64) *esportspb.RedDotEvent {
	return &esportspb.RedDotEvent{IsHit: isHit, ReactionTimeMs: rt, Deviation: deviation}
}

func TestCalibrateRedDotAnchors(t *testing.T) {
	anchors := []struct {
		raw  float64
		want float64
	}{
		{0, 0},
		{8.8, 50},  // p50
		{14.7, 70}, // p75
		{28, 90},   // p90
		{59.6, 99}, // p99
		{120, 100}, // cap
		{200, 100}, // beyond the cap
	}
	for _, a := range anchors {
		if got := calibrateRedDot(a.raw); got != a.want {
			t.Errorf("calibrateRedDot(%v) = %v, want %v", a.raw, got, a.want)
		}
	}
}

func TestRedDotGoldenFormula(t *testing.T) {
	events := make([]*esportspb.RedDotEvent, 0, 30)
	for i := 0; i < 30; i++ {
		events = append(events, redDotEvent(true, 500, 10))
	}
	res, err := redDotScorer{}.recompute(redDotRequest(events...), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 300 {
		t.Errorf("raw = %v, want 300", res.raw)
	}
	if res.authoritative == nil || *res.authoritative != 90.6 {
		t.Errorf("authoritative = %v, want 90.6", res.authoritative)
	}
	if res.grade == nil || *res.grade != "A" {
		t.Errorf("grade = %v, want A", res.grade)
	}
	if res.extras[extraCalibrationInput] != 30 {
		t.Errorf("calibration_input = %v, want 30", res.extras[extraCalibrationInput])
	}
}

// Precision is a bonus factor: zero precision keeps the 0.3 floor instead of
// zeroing the whole attempt.
func TestRedDotPrecisionFloor(t *testing.T) {
	events := make([]*esportspb.RedDotEvent, 0, 30)
	for i := 0; i < 30; i++ {
		events = append(events, redDotEvent(true, 500, 100))
	}
	res, err := redDotScorer{}.recompute(redDotRequest(events...), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.authoritative == nil || *res.authoritative != 50.7 {
		t.Errorf("authoritative = %v, want 50.7", res.authoritative)
	}
}

// avgRt clamps to [200, 1500] — reaction+aim times below the floor or above
// the ceiling are outliers, not superhuman scores.
func TestRedDotReactionClamp(t *testing.T) {
	fast := make([]*esportspb.RedDotEvent, 0, 30)
	slow := make([]*esportspb.RedDotEvent, 0, 30)
	for i := 0; i < 30; i++ {
		fast = append(fast, redDotEvent(true, 100, 10))
		slow = append(slow, redDotEvent(true, 3000, 10))
	}
	res, err := redDotScorer{}.recompute(redDotRequest(fast...), scoringParams{})
	if err != nil {
		t.Fatalf("fast: unexpected error: %v", err)
	}
	if res.authoritative == nil || *res.authoritative != 99.3 {
		t.Errorf("fast authoritative = %v, want 99.3 (clamped avg 200)", res.authoritative)
	}
	res, err = redDotScorer{}.recompute(redDotRequest(slow...), scoringParams{})
	if err != nil {
		t.Fatalf("slow: unexpected error: %v", err)
	}
	if res.authoritative == nil || *res.authoritative != 54.1 {
		t.Errorf("slow authoritative = %v, want 54.1 (clamped avg 1500)", res.authoritative)
	}
}

// Mixed precision and a mid-band input pin both the weight arithmetic and
// the 84.5 band edge (one tenth below the A threshold).
func TestRedDotMixedPrecision(t *testing.T) {
	events := make([]*esportspb.RedDotEvent, 0, 30)
	for i := 0; i < 30; i++ {
		deviation := 10.0
		if i >= 15 {
			deviation = 100
		}
		events = append(events, redDotEvent(true, 400, deviation))
	}
	res, err := redDotScorer{}.recompute(redDotRequest(events...), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.authoritative == nil || *res.authoritative != 84.5 {
		t.Errorf("authoritative = %v, want 84.5", res.authoritative)
	}
	if res.grade == nil || *res.grade != "B" {
		t.Errorf("grade = %v, want B", res.grade)
	}
	if res.extras[extraCalibrationInput] != 24.4 {
		t.Errorf("calibration_input = %v, want 24.4", res.extras[extraCalibrationInput])
	}
}

// A reaction time outside the sanity window is fabricated data: the attempt
// is rejected with the absurd-reaction reason, never scored.
func TestRedDotAbsurdReactionTime(t *testing.T) {
	events := []*esportspb.RedDotEvent{
		redDotEvent(true, 50, 10),
		redDotEvent(true, 61000, 10),
	}
	for _, e := range events {
		_, err := redDotScorer{}.recompute(redDotRequest(e), scoringParams{})
		if err == nil {
			t.Fatalf("rt %d must be rejected", e.GetReactionTimeMs())
		}
		var invalid *invalidAttemptError
		if !errors.As(err, &invalid) || invalid.reason != reasonAbsurdReactionTime {
			t.Fatalf("rt %d: wrong error %v", e.GetReactionTimeMs(), err)
		}
	}
}

// Misses in the event stream (isHit=false) never enter the averages; an
// all-miss attempt scores an honest zero on both tracks.
func TestRedDotZeroHits(t *testing.T) {
	events := []*esportspb.RedDotEvent{
		redDotEvent(false, 500, 40),
		redDotEvent(false, 700, 60),
	}
	res, err := redDotScorer{}.recompute(redDotRequest(events...), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 0 || res.authoritative == nil || *res.authoritative != 0 || *res.grade != "C" {
		t.Errorf("zero-hit result = %+v, want raw 0 / authoritative 0 / grade C", res)
	}
}

func TestRedDotDetailMismatch(t *testing.T) {
	req := &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_FlashAiming{
			FlashAiming: &esportspb.FlashAimingDetail{},
		},
	}
	_, err := redDotScorer{}.recompute(req, scoringParams{})
	var invalid *invalidAttemptError
	if err == nil || !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("wrong error %v, want DETAIL_MISMATCH", err)
	}
}

// The linear track carries the difficulty factor and a unit-score
// override from the engine-param KV.
func TestRedDotLinearTrackFactors(t *testing.T) {
	events := make([]*esportspb.RedDotEvent, 0, 30)
	for i := 0; i < 30; i++ {
		events = append(events, redDotEvent(true, 500, 10))
	}
	res, err := redDotScorer{}.recompute(redDotRequest(events...), scoringParams{difficultyFactor: 1.5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 450 {
		t.Errorf("raw with difficulty 1.5 = %v, want 450", res.raw)
	}
	res, err = redDotScorer{}.recompute(redDotRequest(events...), scoringParams{
		engineParams: map[string]string{"unit_score": "12.5"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 375 {
		t.Errorf("raw with unit_score 12.5 = %v, want 375", res.raw)
	}
}
