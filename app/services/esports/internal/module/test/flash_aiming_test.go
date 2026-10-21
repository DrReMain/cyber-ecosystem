package test

import (
	"errors"
	"testing"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// The law under test: the composite formula transcribed from the legacy
// settle (FlashAimingTestPage.tsx:134-178), with hits/centerHits rebuilt
// from the fact stream — misses rides the contract's counter field.
func flashRequest(shots []*esportspb.FlashAimingShot, misses int32) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_FlashAiming{
			FlashAiming: &esportspb.FlashAimingDetail{Shots: shots, Misses: misses},
		},
	}
}

func flashShot(deviation float64, aimMs int32) *esportspb.FlashAimingShot {
	return &esportspb.FlashAimingShot{Deviation: deviation, AimTimeMs: aimMs}
}

func shots(n int, deviation float64, aimMs int32) []*esportspb.FlashAimingShot {
	out := make([]*esportspb.FlashAimingShot, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, flashShot(deviation, aimMs))
	}
	return out
}

func TestFlashAimingCompositeCapsAtTen(t *testing.T) {
	res, err := flashAimingScorer{}.recompute(flashRequest(shots(20, 10, 500), 5), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2.4 + 4 + 4 + 3 = 13.4 clamps to the composite cap.
	if res.extras[extraComposite] != 10 {
		t.Errorf("composite = %v, want 10 (cap)", res.extras[extraComposite])
	}
	if res.raw != 200 {
		t.Errorf("raw = %v, want 200", res.raw)
	}
	if res.authoritative != nil || res.grade != nil {
		t.Errorf("curve pending: authoritative/grade must stay nil, got %v/%v", res.authoritative, res.grade)
	}
}

// Deviation exactly at the threshold counts as a center hit: ≤ is the
// legacy settle's operator and the engine results face's comparison.
func TestFlashAimingCenterThresholdInclusive(t *testing.T) {
	res, err := flashAimingScorer{}.recompute(flashRequest(shots(10, 20, 800), 0), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// hitRate 3 + accuracy 4 + deviation 2 + reaction 0 = 9.
	if res.extras[extraComposite] != 9 {
		t.Errorf("composite = %v, want 9 (deviation 20 counts as center)", res.extras[extraComposite])
	}
}

// Misses feed the hit-rate denominator only — the contract's counter is the
// sole carrier for targets the shot stream never saw.
func TestFlashAimingMissesDenominator(t *testing.T) {
	res, err := flashAimingScorer{}.recompute(flashRequest(shots(10, 25, 800), 10), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// hitRate 1.5 + accuracy 0 + deviation 1 + reaction 0 = 2.5.
	if res.extras[extraComposite] != 2.5 {
		t.Errorf("composite = %v, want 2.5", res.extras[extraComposite])
	}
	if res.raw != 100 {
		t.Errorf("raw = %v, want 100", res.raw)
	}
}

// No targets and no shots at all: an honest empty result, not an error.
func TestFlashAimingZeroTotal(t *testing.T) {
	res, err := flashAimingScorer{}.recompute(flashRequest(nil, 0), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 0 || res.extras != nil || res.authoritative != nil {
		t.Errorf("zero-total result = %+v, want all-zero/nil", res)
	}
}

// Shots absent but misses present: the legacy empty-list defaults (deviation
// 30, aim 800) zero both score components — composite falls to zero.
func TestFlashAimingMissesOnly(t *testing.T) {
	res, err := flashAimingScorer{}.recompute(flashRequest(nil, 3), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.extras[extraComposite] != 0 {
		t.Errorf("composite = %v, want 0", res.extras[extraComposite])
	}
	if res.raw != 0 {
		t.Errorf("raw = %v, want 0", res.raw)
	}
}

func TestFlashAimingAbsurdAimTime(t *testing.T) {
	_, err := flashAimingScorer{}.recompute(flashRequest(shots(5, 10, 70), 0), scoringParams{})
	var invalid *invalidAttemptError
	if err == nil || !errors.As(err, &invalid) || invalid.reason != reasonAbsurdReactionTime {
		t.Fatalf("wrong error %v, want ABSURD_REACTION_TIME", err)
	}
}

func TestFlashAimingDetailMismatch(t *testing.T) {
	req := &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_RedDot{
			RedDot: &esportspb.RedDotDetail{},
		},
	}
	_, err := flashAimingScorer{}.recompute(req, scoringParams{})
	var invalid *invalidAttemptError
	if err == nil || !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("wrong error %v, want DETAIL_MISMATCH", err)
	}
}

func TestFlashAimingLinearTrackFactors(t *testing.T) {
	res, err := flashAimingScorer{}.recompute(flashRequest(shots(20, 10, 500), 5), scoringParams{difficultyFactor: 1.5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 300 {
		t.Errorf("raw with difficulty 1.5 = %v, want 300", res.raw)
	}
	res, err = flashAimingScorer{}.recompute(flashRequest(shots(20, 10, 500), 5), scoringParams{
		engineParams: map[string]string{"unit_score": "8"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 160 {
		t.Errorf("raw with unit_score 8 = %v, want 160", res.raw)
	}
}
