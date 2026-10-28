package test

import (
	"errors"
	"testing"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// ccKey: the seeded twelve-relation matrix verbatim — Tank counters Fighter
// and Assassin, Fighter Marksman and Mage, Mage Tank and Marksman, Marksman
// Fighter and Support, Support Assassin and Mage, Assassin Marksman and Mage.
func ccKey() map[string]map[string]any {
	rels := [][2]string{
		{"Tank", "Fighter"}, {"Tank", "Assassin"},
		{"Fighter", "Marksman"}, {"Fighter", "Mage"},
		{"Mage", "Tank"}, {"Mage", "Marksman"},
		{"Marksman", "Fighter"}, {"Marksman", "Support"},
		{"Support", "Assassin"}, {"Support", "Mage"},
		{"Assassin", "Marksman"}, {"Assassin", "Mage"},
	}
	key := make(map[string]map[string]any, len(rels))
	for _, r := range rels {
		key[r[0]+"|"+r[1]] = map[string]any{"counter_score": 20}
	}
	return key
}

func ccParams() scoringParams {
	return scoringParams{durationSec: 30, answerKey: ccKey()}
}

func ccRound(n int32, enemy, sel []string, ms int32) *esportspb.CharacterCounterRound {
	return &esportspb.CharacterCounterRound{RoundNumber: n, EnemyLineup: enemy, SelectedLineup: sel, TimeSpentMs: ms}
}

func ccRequest(rounds ...*esportspb.CharacterCounterRound) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_CharacterCounter{
			CharacterCounter: &esportspb.CharacterCounterDetail{Rounds: rounds},
		},
	}
}

// ccEnemy: five of six roles, missing Assassin. ccBest (excluding Tank) is
// the optimal answer at nine hitting pairs — both Tank and Support counter
// only one enemy role each, so dropping either maximizes the cross product.
func ccEnemy() []string { return []string{"Tank", "Fighter", "Mage", "Marksman", "Support"} }
func ccBest() []string  { return []string{"Fighter", "Mage", "Marksman", "Support", "Assassin"} }
func ccAlt() []string   { return []string{"Tank", "Fighter", "Mage", "Marksman", "Support"} }

func ccSession(enemy, sel []string, ms int32) []*esportspb.CharacterCounterRound {
	return []*esportspb.CharacterCounterRound{
		ccRound(1, enemy, sel, ms), ccRound(2, enemy, sel, ms), ccRound(3, enemy, sel, ms),
	}
}

func TestCharacterCounterPerfectGolden(t *testing.T) {
	res, err := characterCounterScorer{}.recompute(ccRequest(ccSession(ccEnemy(), ccBest(), 8000)...), ccParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Nine hitting pairs a round against c_max nine: Q=1 at the 8s reference
	// speed → composite 100 exactly (speed headroom unspent).
	if res.extras[extraComposite] != 100.0 {
		t.Errorf("composite = %v, want 100.0", res.extras[extraComposite])
	}
	// 27 hitting pairs × unit 15 = 405.
	if res.raw != 405 {
		t.Errorf("raw = %v, want 405", res.raw)
	}
	if res.authoritative != nil {
		t.Errorf("authoritative = %v, want nil (curve pending)", *res.authoritative)
	}
}

func TestCharacterCounterImperfectFastGolden(t *testing.T) {
	res, err := characterCounterScorer{}.recompute(ccRequest(ccSession(ccEnemy(), ccAlt(), 4000)...), ccParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Eight pairs against nine optimal (Q=0.888…) with the speed multiplier
	// saturated at 1.5: 133.3 — imperfect-fast outruns the 100 the old
	// formula saturated everyone into.
	if res.extras[extraComposite] != 133.3 {
		t.Errorf("composite = %v, want 133.3", res.extras[extraComposite])
	}
	if res.raw != 360 {
		t.Errorf("raw = %v, want 360", res.raw)
	}
}

func TestCharacterCounterSpeedCapGolden(t *testing.T) {
	res, err := characterCounterScorer{}.recompute(ccRequest(ccSession(ccEnemy(), ccBest(), 4000)...), ccParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Perfect counter play at double the reference speed: Q×100×SF caps at
	// the rework's 150, not the family's 100.
	if res.extras[extraComposite] != 150.0 {
		t.Errorf("composite = %v, want 150.0", res.extras[extraComposite])
	}
}

func TestCharacterCounterUnitScoreAndDifficultyOverride(t *testing.T) {
	p := ccParams()
	p.difficultyFactor = 1.5
	p.engineParams = map[string]string{unitScoreKey: "20"}
	res, err := characterCounterScorer{}.recompute(ccRequest(ccSession(ccEnemy(), ccAlt(), 4000)...), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 24 hitting pairs × overridden unit 20 × difficulty 1.5 = 720.
	if res.raw != 720 {
		t.Errorf("raw = %v, want 720", res.raw)
	}
	if res.extras[extraComposite] != 133.3 {
		t.Errorf("composite = %v, want 133.3", res.extras[extraComposite])
	}
}

func TestCharacterCounterHonestZero(t *testing.T) {
	res, err := characterCounterScorer{}.recompute(ccRequest(), ccParams())
	if err != nil {
		t.Fatalf("empty rounds error: %v", err)
	}
	if res.raw != 0 || res.extras[extraComposite] != 0 {
		t.Errorf("empty rounds: raw = %v composite = %v, want 0/0", res.raw, res.extras[extraComposite])
	}
}

func TestCharacterCounterRejections(t *testing.T) {
	short := append([]string{}, ccBest()[:4]...)
	dupSel := []string{"Tank", "Tank", "Fighter", "Mage", "Marksman"}
	unknown := []string{"Carry", "Fighter", "Mage", "Marksman", "Support"}
	cases := []struct {
		name   string
		req    *esportspb.SubmitTestAttemptRequest
		reason string
	}{
		{
			name:   "detail mismatch",
			req:    &esportspb.SubmitTestAttemptRequest{RawDetail: &esportspb.SubmitTestAttemptRequest_NumberStars{}},
			reason: reasonDetailMismatch,
		},
		{
			name:   "unknown role word",
			req:    ccRequest(ccSession(ccEnemy(), unknown, 8000)...),
			reason: reasonDetailMismatch,
		},
		{
			name:   "duplicate inside a lineup",
			req:    ccRequest(ccSession(ccEnemy(), dupSel, 8000)...),
			reason: reasonDetailMismatch,
		},
		{
			name:   "selected shorter than enemy",
			req:    ccRequest(ccSession(ccEnemy(), short, 8000)...),
			reason: reasonDetailMismatch,
		},
		{
			name:   "round count short of the contract",
			req:    ccRequest(ccRound(1, ccEnemy(), ccBest(), 8000), ccRound(2, ccEnemy(), ccBest(), 8000)),
			reason: reasonDetailMismatch,
		},
		{
			name: "round numbers out of sequence",
			req: ccRequest(
				ccRound(1, ccEnemy(), ccBest(), 8000),
				ccRound(3, ccEnemy(), ccBest(), 8000),
				ccRound(2, ccEnemy(), ccBest(), 8000),
			),
			reason: reasonDetailMismatch,
		},
		{
			name:   "round decided under the floor",
			req:    ccRequest(ccRound(1, ccEnemy(), ccBest(), 50), ccRound(2, ccEnemy(), ccBest(), 8000), ccRound(3, ccEnemy(), ccBest(), 8000)),
			reason: reasonAbsurdReactionTime,
		},
		{
			name:   "round times past the session budget",
			req:    ccRequest(ccSession(ccEnemy(), ccBest(), 11000)...),
			reason: reasonAbsurdTimeSpan,
		},
		{
			name:   "empty lineups",
			req:    ccRequest(ccSession([]string{}, []string{}, 8000)...),
			reason: reasonDetailMismatch,
		},
	}
	for _, tc := range cases {
		_, err := characterCounterScorer{}.recompute(tc.req, ccParams())
		var invalid *invalidAttemptError
		if !errors.As(err, &invalid) || invalid.reason != tc.reason {
			t.Errorf("%s: err = %v, want %s", tc.name, err, tc.reason)
		}
	}
}
