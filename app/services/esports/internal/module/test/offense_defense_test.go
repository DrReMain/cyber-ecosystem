package test

import (
	"errors"
	"fmt"
	"testing"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

func odScenario(i int) string { return fmt.Sprintf("OD_%03d", i) }

// odKey: an answer bank mirroring the seed shape — the first ten scenarios
// key on attack, the rest on defense — surface enough for the goldens
// without hauling the real bank content into unit land.
func odKey() map[string]map[string]any {
	key := make(map[string]map[string]any, 20)
	for i := 1; i <= 10; i++ {
		key[odScenario(i)] = map[string]any{"complexity": "simple", "optimal_choice": "attack"}
	}
	for i := 11; i <= 20; i++ {
		key[odScenario(i)] = map[string]any{"complexity": "complex", "optimal_choice": "defense"}
	}
	return key
}

func odDecision(scenario, choice string, ms int32, timeout bool) *esportspb.OffenseDefenseDecision {
	d := &esportspb.OffenseDefenseDecision{
		ScenarioId:     scenario,
		DecisionTimeMs: ms,
		Timeout:        timeout,
	}
	if choice != "" {
		d.Choice = &choice
	}
	return d
}

func odRequest(decisions ...*esportspb.OffenseDefenseDecision) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_OffenseDefense{
			OffenseDefense: &esportspb.OffenseDefenseDetail{Decisions: decisions},
		},
	}
}

func TestOffenseDefenseFastPaceGolden(t *testing.T) {
	decisions := make([]*esportspb.OffenseDefenseDecision, 0, 10)
	for i := 1; i <= 10; i++ {
		// Uppercase choices pin the legacy equalsIgnoreCase parity.
		decisions = append(decisions, odDecision(odScenario(i), "ATTACK", 1200, false))
	}
	res, err := offenseDefenseScorer{}.recompute(odRequest(decisions...), scoringParams{answerKey: odKey()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 10/10 correct at 1200ms: quality 6 + speed saturated 2 + 0.5 = 8.5.
	if res.extras[extraComposite] != 8.5 {
		t.Errorf("composite = %v, want 8.5", res.extras[extraComposite])
	}
	// 10 correct decisions × unit 25 = 250 (band low).
	if res.raw != 250 {
		t.Errorf("raw = %v, want 250", res.raw)
	}
	if res.authoritative != nil {
		t.Errorf("authoritative = %v, want nil (curve pending)", *res.authoritative)
	}
}

func TestOffenseDefenseMidPaceGolden(t *testing.T) {
	decisions := []*esportspb.OffenseDefenseDecision{
		odDecision(odScenario(1), "attack", 6000, false),
		odDecision(odScenario(2), "attack", 6000, false),
		odDecision(odScenario(3), "attack", 6000, false),
		odDecision(odScenario(4), "attack", 6000, false),
		odDecision(odScenario(5), "attack", 6000, false),
		odDecision(odScenario(6), "attack", 6000, false),
		odDecision(odScenario(7), "attack", 6000, false),
		// A wrong call, then two unanswered timeouts riding window-length.
		odDecision(odScenario(8), "retreat", 6000, false),
		odDecision(odScenario(9), "", 8500, true),
		odDecision(odScenario(10), "", 8500, true),
	}
	res, err := offenseDefenseScorer{}.recompute(odRequest(decisions...), scoringParams{answerKey: odKey()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 7/10 correct, avg 6500ms, 2 timeouts: quality 4.2 + speed 1.5 +
	// 0.5 − penalty 1.0 = 5.2.
	if res.extras[extraComposite] != 5.2 {
		t.Errorf("composite = %v, want 5.2", res.extras[extraComposite])
	}
	if res.raw != 175 {
		t.Errorf("raw = %v, want 175", res.raw)
	}
}

func TestOffenseDefenseUnitScoreAndDifficultyOverride(t *testing.T) {
	decisions := make([]*esportspb.OffenseDefenseDecision, 0, 8)
	for i := 1; i <= 8; i++ {
		decisions = append(decisions, odDecision(odScenario(i), "attack", 1500, false))
	}
	res, err := offenseDefenseScorer{}.recompute(odRequest(decisions...), scoringParams{
		answerKey:        odKey(),
		difficultyFactor: 1.5,
		engineParams:     map[string]string{unitScoreKey: "30"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 8 correct × overridden unit 30 × difficulty 1.5 = 360.
	if res.raw != 360 {
		t.Errorf("raw = %v, want 360", res.raw)
	}
	if res.extras[extraComposite] != 8.5 {
		t.Errorf("composite = %v, want 8.5", res.extras[extraComposite])
	}
}

func TestOffenseDefenseHonestZero(t *testing.T) {
	res, err := offenseDefenseScorer{}.recompute(odRequest(), scoringParams{answerKey: odKey()})
	if err != nil {
		t.Fatalf("empty decisions error: %v", err)
	}
	if res.raw != 0 || res.extras[extraComposite] != 0 {
		t.Errorf("empty decisions: raw = %v composite = %v, want 0/0", res.raw, res.extras[extraComposite])
	}
}

func TestOffenseDefenseRejections(t *testing.T) {
	cases := []struct {
		name string
		req  *esportspb.SubmitTestAttemptRequest
		// The reason each case must land on.
		reason string
	}{
		{
			name:   "detail mismatch",
			req:    &esportspb.SubmitTestAttemptRequest{RawDetail: &esportspb.SubmitTestAttemptRequest_NumberStars{}},
			reason: reasonDetailMismatch,
		},
		{
			name:   "unknown scenario",
			req:    odRequest(odDecision("OD_999", "attack", 1200, false)),
			reason: reasonDetailMismatch,
		},
		{
			name:   "instant decision under the floor",
			req:    odRequest(odDecision(odScenario(1), "attack", 50, false)),
			reason: reasonAbsurdReactionTime,
		},
	}
	for _, tc := range cases {
		_, err := offenseDefenseScorer{}.recompute(tc.req, scoringParams{answerKey: odKey()})
		var invalid *invalidAttemptError
		if !errors.As(err, &invalid) || invalid.reason != tc.reason {
			t.Errorf("%s: err = %v, want %s", tc.name, err, tc.reason)
		}
	}
}
