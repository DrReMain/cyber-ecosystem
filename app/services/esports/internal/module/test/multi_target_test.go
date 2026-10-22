package test

import (
	"errors"
	"testing"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// The law under test: the composite formula transcribed from the legacy
// settle (MultiTargetTestPage.tsx:177-210), with hits/tier aggregates and
// decision times rebuilt from the fact stream — decision time is the
// pick-to-pick delta because the frozen pick carries no spawn stamp (§6.5).
func multiTargetRequest(missedHigh int32, picks ...*esportspb.MultiTargetPick) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_MultiTarget{
			MultiTarget: &esportspb.MultiTargetDetail{Picks: picks, MissedHigh: missedHigh},
		},
	}
}

func multiTargetPick(tier string, elapsedMs int32) *esportspb.MultiTargetPick {
	return &esportspb.MultiTargetPick{Tier: tier, IsHit: true, ElapsedMs: elapsedMs}
}

func TestMultiTargetCompositeFromFacts(t *testing.T) {
	res, err := multiTargetScorer{}.recompute(
		multiTargetRequest(2,
			multiTargetPick("high", 700),
			multiTargetPick("trap", 1400),
			multiTargetPick("medium", 2200),
		),
		scoringParams{durationSec: 30},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// rawTotal 100−50+50=100 → value 100/3000×5 = 1/6; priority 1/3×3 = 1;
	// deltas 700/700/800 → avg 733.3 → speed (1500−733.3)/150 ≈ 5.11;
	// composite = 1/6 + 1 + 5.11 − 0.3 (trap) − 0.4 (2 missed high) = 5.6.
	if res.extras[extraComposite] != 5.6 {
		t.Errorf("composite = %v, want 5.6", res.extras[extraComposite])
	}
	if res.raw != 30 {
		t.Errorf("raw = %v, want 30 (3 hits × unit 10)", res.raw)
	}
	if res.authoritative != nil || res.grade != nil {
		t.Errorf("curve pending: authoritative/grade must stay nil, got %v/%v", res.authoritative, res.grade)
	}
}

// Zero picks is an honest zero on both tracks — no decision-speed points
// without touching targets (the legacy zero guard).
func TestMultiTargetZeroHits(t *testing.T) {
	res, err := multiTargetScorer{}.recompute(multiTargetRequest(0), scoringParams{durationSec: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 0 || res.extras != nil || res.authoritative != nil {
		t.Errorf("zero-hit result = %+v, want the honest empty zero", res)
	}
}

// The pick-to-pick delta gate: a sub-80ms switch is fabricated data — the
// attempt is rejected, never scored (red_dot's per-point law, §6.5).
func TestMultiTargetAbsurdSwitch(t *testing.T) {
	_, err := multiTargetScorer{}.recompute(
		multiTargetRequest(0, multiTargetPick("high", 700), multiTargetPick("medium", 730)),
		scoringParams{durationSec: 30},
	)
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonAbsurdReactionTime {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonAbsurdReactionTime)
	}
}

// The tier vocabulary is open but four-word by contract; anything else is a
// detail mismatch, not a zero score.
func TestMultiTargetTierVocabulary(t *testing.T) {
	_, err := multiTargetScorer{}.recompute(
		multiTargetRequest(0, multiTargetPick("ultra", 700)),
		scoringParams{durationSec: 30},
	)
	var invalid *invalidAttemptError
	if err == nil || !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

// Tier values read the tier_value_* engine-param keys — ops retunes the
// value track without a deploy (§6.5, the flash unit_score precedent).
func TestMultiTargetTierValueOverride(t *testing.T) {
	res, err := multiTargetScorer{}.recompute(
		multiTargetRequest(0, multiTargetPick("high", 700), multiTargetPick("medium", 1500)),
		scoringParams{durationSec: 30, engineParams: map[string]string{"tier_value_high": "200"}},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// rawTotal 200+50=250 → value 250/3000×5 ≈ 0.417; priority 1/2×3 = 1.5;
	// avg 750 → speed 5.0; composite = 0.417+1.5+5 = 6.9.
	if res.extras[extraComposite] != 6.9 {
		t.Errorf("composite = %v, want 6.9", res.extras[extraComposite])
	}
	if res.raw != 20 {
		t.Errorf("raw = %v, want 20", res.raw)
	}
}

func TestMultiTargetRegistered(t *testing.T) {
	if _, ok := lookupScorer("multi_target"); !ok {
		t.Fatal("multi_target scorer not registered")
	}
}

// The vocabulary gate binds misses too: a non-hit pick riding an unknown
// tier is a trimmed stream, not a scoreable fact (the review's gate hole).
func TestMultiTargetTierVocabularyBindsMisses(t *testing.T) {
	miss := &esportspb.MultiTargetPick{Tier: "ultra", IsHit: false, ElapsedMs: 700}
	_, err := multiTargetScorer{}.recompute(multiTargetRequest(0, miss), scoringParams{durationSec: 30})
	var invalid *invalidAttemptError
	if err == nil || !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}
