package test

import (
	"errors"
	"testing"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// The law under test: the composite transcribed from the legacy settle
// (the legacy page settle), with the unmatched-round sentinel
// kept out of the deviation/speed means and inside the precision
// denominator.
func skillRound(n int32, dev float64, placeMs int32) *esportspb.SkillTrajectoryRound {
	r := &esportspb.SkillTrajectoryRound{
		RoundNumber:     n,
		Shape:           "circle",
		Deviation:       dev,
		PlacementTimeMs: placeMs,
		Placed:          true,
	}
	if dev < 999 {
		r.PlayerPlacement = &esportspb.Point{X: 100, Y: 100}
	}
	return r
}

func skillTrajectoryRequest(rounds ...*esportspb.SkillTrajectoryRound) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_SkillTrajectory{
			SkillTrajectory: &esportspb.SkillTrajectoryDetail{Rounds: rounds},
		},
	}
}

func TestSkillTrajectoryCompositeFromFacts(t *testing.T) {
	res, err := skillTrajectoryScorer{}.recompute(
		skillTrajectoryRequest(
			skillRound(1, 3, 9000),
			skillRound(2, 7, 9000),
			skillRound(3, 12, 9000),
			skillRound(4, 20, 9000),
			skillRound(5, 45, 9000),
		),
		scoringParams{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// precision (1 + 2×0.5)/5×6 = 2.4; avgDev 17.4 → (50−17.4)/10 = 3.26;
	// avgPlace 9000 → (12000−9000)/1500 = 2.0; composite 7.66 → 7.7.
	if res.extras[extraComposite] != 7.7 {
		t.Errorf("composite = %v, want 7.7", res.extras[extraComposite])
	}
	// fit [0.94, 0.86, 0.76, 0.6, 0.1] avg 0.652 → 65.2 points × 8 = 521.6.
	if res.raw != 521.6 {
		t.Errorf("raw = %v, want 521.6", res.raw)
	}
	if res.authoritative != nil || res.grade != nil {
		t.Errorf("curve pending: authoritative/grade must stay nil, got %v/%v", res.authoritative, res.grade)
	}
}

// Unmatched rounds (sentinel 999, no placement) count in the precision
// denominator and as zero fit, but stay out of the deviation/speed means —
// one missed round must not drag the two averages to death.
func TestSkillTrajectorySentinelRounds(t *testing.T) {
	res, err := skillTrajectoryScorer{}.recompute(
		skillTrajectoryRequest(
			skillRound(1, 4, 5000),
			skillRound(2, 999, 0),
			skillRound(3, 10, 6000),
			skillRound(4, 999, 0),
			skillRound(5, 30, 7000),
		),
		scoringParams{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// precision (1 + 0.5)/5×6 = 1.8; matched avgDev 44/3 → (50−14.67)/10 =
	// 3.53; avgPlace 6000 → speed 4.0; composite 9.33 → 9.3.
	if res.extras[extraComposite] != 9.3 {
		t.Errorf("composite = %v, want 9.3", res.extras[extraComposite])
	}
	// fit [0.92, 0, 0.8, 0, 0.4] avg 0.424 → 42.4 points × 8 = 339.2.
	if res.raw != 339.2 {
		t.Errorf("raw = %v, want 339.2", res.raw)
	}
}

// The unit_score override rides the percentage track like every other
// scorer's — one config feeds client and server alike.
func TestSkillTrajectoryUnitScoreOverride(t *testing.T) {
	res, err := skillTrajectoryScorer{}.recompute(
		skillTrajectoryRequest(
			skillRound(1, 4, 5000),
			skillRound(2, 999, 0),
			skillRound(3, 10, 6000),
			skillRound(4, 999, 0),
			skillRound(5, 30, 7000),
		),
		scoringParams{engineParams: map[string]string{"unit_score": "10"}},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// avg fit 0.424 → 42.4 points × 10 = 424.
	if res.raw != 424 {
		t.Errorf("raw = %v, want 424", res.raw)
	}
}

// Zero rounds is an honest zero on both tracks (the legacy AFK guard).
func TestSkillTrajectoryZeroRounds(t *testing.T) {
	res, err := skillTrajectoryScorer{}.recompute(skillTrajectoryRequest(), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 0 || res.extras != nil || res.authoritative != nil {
		t.Errorf("zero-round result = %+v, want the honest empty zero", res)
	}
}

func TestSkillTrajectoryShapeVocabulary(t *testing.T) {
	r := skillRound(1, 4, 5000)
	r.Shape = "octagon"
	_, err := skillTrajectoryScorer{}.recompute(skillTrajectoryRequest(r), scoringParams{})
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

// Rounds arrive in play order; a repeated number is a trimmed stream.
func TestSkillTrajectoryRepeatedRoundNumber(t *testing.T) {
	_, err := skillTrajectoryScorer{}.recompute(
		skillTrajectoryRequest(skillRound(1, 4, 5000), skillRound(1, 10, 6000)),
		scoringParams{},
	)
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

// Placement presence mirrors the sentinel both ways: a placed round with a
// sentinel deviation and an unplaced round with a real one are both
// fabricated shapes.
func TestSkillTrajectorySentinelMisaligned(t *testing.T) {
	placed := skillRound(1, 999, 0)
	placed.PlayerPlacement = &esportspb.Point{X: 1, Y: 1}
	_, err := skillTrajectoryScorer{}.recompute(skillTrajectoryRequest(placed), scoringParams{})
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}

	unplaced := skillRound(1, 500, 5000)
	unplaced.PlayerPlacement = nil
	_, err = skillTrajectoryScorer{}.recompute(skillTrajectoryRequest(unplaced), scoringParams{})
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

func TestSkillTrajectoryNegativeDeviation(t *testing.T) {
	_, err := skillTrajectoryScorer{}.recompute(skillTrajectoryRequest(skillRound(1, -1, 5000)), scoringParams{})
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

func TestSkillTrajectoryRegistered(t *testing.T) {
	if _, ok := lookupScorer("skill_trajectory"); !ok {
		t.Fatal("skill_trajectory scorer not registered")
	}
}
