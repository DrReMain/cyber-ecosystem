package test

import (
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// The law under test: friend/foe correctness is judged server-side from the
// clicked target's type, the spawned_enemy counter closes the old Tier B
// gap as the accuracy denominator, and the composite transcribes the
// legacy PRD settle (the legacy scoring module).
func targetRecognitionRequest(spawnedEnemy int32, clicks ...*esportspb.TargetRecognitionClick) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_TargetRecognition{
			TargetRecognition: &esportspb.TargetRecognitionDetail{Clicks: clicks, SpawnedEnemy: spawnedEnemy},
		},
	}
}

func trClick(targetType string, rt *int32) *esportspb.TargetRecognitionClick {
	return &esportspb.TargetRecognitionClick{TargetType: targetType, ReactionTimeMs: rt, ElapsedMs: 1000}
}

func TestTargetRecognitionCompositeFromFacts(t *testing.T) {
	res, err := targetRecognitionScorer{}.recompute(
		targetRecognitionRequest(5,
			trClick("enemy", utils.Ptr(int32(300))),
			trClick("enemy", utils.Ptr(int32(400))),
			trClick("friendly", nil),
			trClick("neutral", utils.Ptr(int32(500))),
		),
		scoringParams{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// correct 2/5 → accuracy 2; enemy avg 350 → speed 250/60 ≈ 4.167 and
	// visual (1−150/800)×2 = 1.625; penalties 2×0.4 + 3×0.2 = 1.4;
	// composite = 2 + 4.167 + 1.625 − 1.4 = 6.4.
	if res.extras[extraComposite] != 6.4 {
		t.Errorf("composite = %v, want 6.4", res.extras[extraComposite])
	}
	if res.raw != 40 {
		t.Errorf("raw = %v, want 40 (2 correct × unit 20)", res.raw)
	}
	if res.authoritative != nil || res.grade != nil {
		t.Errorf("curve pending: authoritative/grade must stay nil, got %v/%v", res.authoritative, res.grade)
	}
}

// No enemy identification is an honest zero on both tracks, whatever the
// false-positive noise around it.
func TestTargetRecognitionZeroCorrect(t *testing.T) {
	res, err := targetRecognitionScorer{}.recompute(
		targetRecognitionRequest(5, trClick("friendly", utils.Ptr(int32(300)))),
		scoringParams{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 0 || res.extras != nil || res.authoritative != nil {
		t.Errorf("zero-correct result = %+v, want the honest empty zero", res)
	}
}

// The per-click window gates every present reaction time — enemy or not —
// and an absent rt (no reaction) simply skips the gate and the mean.
func TestTargetRecognitionRtWindowGatesEveryClick(t *testing.T) {
	_, err := targetRecognitionScorer{}.recompute(
		targetRecognitionRequest(5, trClick("friendly", utils.Ptr(int32(10)))),
		scoringParams{},
	)
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonAbsurdReactionTime {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonAbsurdReactionTime)
	}
}

// The type vocabulary is open but three-word by contract; anything else is
// a detail mismatch, not a zero score.
func TestTargetRecognitionTypeVocabulary(t *testing.T) {
	_, err := targetRecognitionScorer{}.recompute(
		targetRecognitionRequest(5, trClick("hostile", utils.Ptr(int32(300)))),
		scoringParams{},
	)
	var invalid *invalidAttemptError
	if err == nil || !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

// More correct identifications than enemies spawned is a fabricated stream
// — legacy clamped the rate client-side; the server-side law rejects.
func TestTargetRecognitionCountDeviation(t *testing.T) {
	_, err := targetRecognitionScorer{}.recompute(
		targetRecognitionRequest(2,
			trClick("enemy", utils.Ptr(int32(300))),
			trClick("enemy", utils.Ptr(int32(400))),
			trClick("enemy", utils.Ptr(int32(500))),
		),
		scoringParams{},
	)
	var invalid *invalidAttemptError
	if err == nil || !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

func TestTargetRecognitionRegistered(t *testing.T) {
	if _, ok := lookupScorer("target_recognition"); !ok {
		t.Fatal("target_recognition scorer not registered")
	}
}

// The engine stamps every target tap with a reaction time; an enemy click
// without one is a trimmed stream — the absent-rt mean-zero path would
// otherwise inflate the speed terms (the review's inflation hole).
func TestTargetRecognitionEnemyClickRequiresRt(t *testing.T) {
	_, err := targetRecognitionScorer{}.recompute(
		targetRecognitionRequest(5, trClick("enemy", nil)),
		scoringParams{},
	)
	var invalid *invalidAttemptError
	if err == nil || !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}
