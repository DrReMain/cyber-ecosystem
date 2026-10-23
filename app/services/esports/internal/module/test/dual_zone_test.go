package test

import (
	"errors"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// The law under test: the survival composite transcribed from the legacy
// settle (Java DualZoneScorer), with dodges/collisions/survival rebuilt
// from the behavior stream against the pause-carried net clock.
func dualZoneRequest(spanMs int64, behaviors ...*esportspb.DualZoneBehavior) *esportspb.SubmitTestAttemptRequest {
	start := time.Unix(1700000000, 0)
	return &esportspb.SubmitTestAttemptRequest{
		StartedAt: timestamppb.New(start),
		EndedAt:   timestamppb.New(start.Add(time.Duration(spanMs) * time.Millisecond)),
		RawDetail: &esportspb.SubmitTestAttemptRequest_DualZone{
			DualZone: &esportspb.DualZoneDetail{Behaviors: behaviors},
		},
	}
}

func dualZoneDodge(elapsedMs, obstacleId int32) *esportspb.DualZoneBehavior {
	return &esportspb.DualZoneBehavior{
		ElapsedMs:  elapsedMs,
		Zone:       "left",
		Event:      "dodge",
		ObstacleId: &obstacleId,
	}
}

func dualZoneTenDodges() []*esportspb.DualZoneBehavior {
	bs := make([]*esportspb.DualZoneBehavior, 0, 10)
	for i := 1; i <= 10; i++ {
		bs = append(bs, dualZoneDodge(int32(i*1000), int32(i)))
	}
	return bs
}

func TestDualZoneSurvivalScore(t *testing.T) {
	res, err := dualZoneScorer{}.recompute(
		dualZoneRequest(30000, dualZoneTenDodges()...),
		scoringParams{durationSec: 30},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 30s of 30s at weight 1.15 caps serverScore at 100 → composite 10.0.
	if res.extras[extraComposite] != 10 {
		t.Errorf("composite = %v, want 10", res.extras[extraComposite])
	}
	if res.raw != 150 {
		t.Errorf("raw = %v, want 150 (10 dodges × unit 15)", res.raw)
	}
	if res.authoritative != nil || res.grade != nil {
		t.Errorf("curve pending: authoritative/grade must stay nil, got %v/%v", res.authoritative, res.grade)
	}
}

// The base_score_time override retunes the survival reference without a
// deploy (the tier_value_* precedent).
func TestDualZoneBaseTimeOverride(t *testing.T) {
	res, err := dualZoneScorer{}.recompute(
		dualZoneRequest(18000,
			dualZoneDodge(5000, 1),
			dualZoneDodge(10000, 2),
		),
		scoringParams{durationSec: 30, engineParams: map[string]string{"base_score_time": "60"}},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 18/60 × 100 × 1.15 = 34.5 → composite 3.5.
	if res.extras[extraComposite] != 3.5 {
		t.Errorf("composite = %v, want 3.5", res.extras[extraComposite])
	}
	if res.raw != 30 {
		t.Errorf("raw = %v, want 30", res.raw)
	}
}

// A paused run keeps its play time: the wall clock grew by the pause
// segment, the net clock did not — the tolerance compares against play
// time (the Java gate would have rejected this shape).
func TestDualZonePauseCarry(t *testing.T) {
	req := dualZoneRequest(40000, dualZoneTenDodges()...)
	req.Backgrounds = []*esportspb.BackgroundSpan{{ElapsedMs: 15000, DurationMs: 10000}}
	res, err := dualZoneScorer{}.recompute(req, scoringParams{durationSec: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.extras[extraComposite] != 10 || res.raw != 150 {
		t.Errorf("paused run = %v/%v, want composite 10 and raw 150", res.extras[extraComposite], res.raw)
	}
}

// Collision ends the run: the score rides the collision moment, not the
// un-played remainder.
func TestDualZoneCollisionEndsRun(t *testing.T) {
	res, err := dualZoneScorer{}.recompute(
		dualZoneRequest(6000,
			dualZoneDodge(1000, 1),
			dualZoneDodge(3000, 2),
			dualZoneDodge(5000, 3),
			&esportspb.DualZoneBehavior{ElapsedMs: 6000, Zone: "right", Event: "collision"},
		),
		scoringParams{durationSec: 30},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 6/30 × 100 × 1.15 = 23 → composite 2.3; raw counts the dodges only.
	if res.extras[extraComposite] != 2.3 {
		t.Errorf("composite = %v, want 2.3", res.extras[extraComposite])
	}
	if res.raw != 45 {
		t.Errorf("raw = %v, want 45", res.raw)
	}
}

// An event-free stream is an idle-but-alive run: zero raw, full survival —
// the score's numerator is staying alive, not the dodge count.
func TestDualZoneIdleRunScoresZero(t *testing.T) {
	res, err := dualZoneScorer{}.recompute(dualZoneRequest(30000), scoringParams{durationSec: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 0 {
		t.Errorf("raw = %v, want 0", res.raw)
	}
}

func TestDualZoneAbsurdTimeSpan(t *testing.T) {
	_, err := dualZoneScorer{}.recompute(
		dualZoneRequest(300, dualZoneDodge(100, 1)),
		scoringParams{durationSec: 30},
	)
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonAbsurdTimeSpan {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonAbsurdTimeSpan)
	}
}

// The submitted span cannot overrun the server-configured length (a
// duration fabrication closure).
func TestDualZoneSpanOverrunsConfig(t *testing.T) {
	bs := dualZoneTenDodges()
	bs = append(bs, dualZoneDodge(32500, 11))
	_, err := dualZoneScorer{}.recompute(dualZoneRequest(33000, bs...), scoringParams{durationSec: 30})
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonAbsurdTimeSpan {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonAbsurdTimeSpan)
	}
}

// A stream may end before the clock (idle tail), never overrun it.
func TestDualZoneStreamOverrunsClock(t *testing.T) {
	bs := dualZoneTenDodges()
	bs = append(bs, dualZoneDodge(32000, 11))
	_, err := dualZoneScorer{}.recompute(dualZoneRequest(30000, bs...), scoringParams{durationSec: 30})
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

// The vocabulary is lowercased by contract (legacy LEFT/RIGHT → left/right).
func TestDualZoneEventVocabulary(t *testing.T) {
	_, err := dualZoneScorer{}.recompute(
		dualZoneRequest(30000, &esportspb.DualZoneBehavior{ElapsedMs: 1000, Zone: "left", Event: "DODGE"}),
		scoringParams{durationSec: 30},
	)
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

func TestDualZoneZoneVocabulary(t *testing.T) {
	_, err := dualZoneScorer{}.recompute(
		dualZoneRequest(30000, &esportspb.DualZoneBehavior{ElapsedMs: 1000, Zone: "middle", Event: "dodge"}),
		scoringParams{durationSec: 30},
	)
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

func TestDualZoneBackwardsStamp(t *testing.T) {
	_, err := dualZoneScorer{}.recompute(
		dualZoneRequest(30000, dualZoneDodge(5000, 1), dualZoneDodge(4000, 2)),
		scoringParams{durationSec: 30},
	)
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

func TestDualZoneEventsAfterCollision(t *testing.T) {
	_, err := dualZoneScorer{}.recompute(
		dualZoneRequest(3000,
			&esportspb.DualZoneBehavior{ElapsedMs: 1000, Zone: "left", Event: "collision"},
			dualZoneDodge(2000, 1),
		),
		scoringParams{durationSec: 30},
	)
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

// A trailing collision must land on the net clock — a collision stream with
// a padded span is the fabrication face the alignment gate closes.
func TestDualZoneCollisionClockMisaligned(t *testing.T) {
	_, err := dualZoneScorer{}.recompute(
		dualZoneRequest(30000,
			dualZoneDodge(1000, 1),
			&esportspb.DualZoneBehavior{ElapsedMs: 2000, Zone: "left", Event: "collision"},
		),
		scoringParams{durationSec: 30},
	)
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

// One obstacle records one event; a repeated id double-counts a dodge.
func TestDualZoneRepeatedObstacle(t *testing.T) {
	_, err := dualZoneScorer{}.recompute(
		dualZoneRequest(30000, dualZoneDodge(1000, 7), dualZoneDodge(2000, 7)),
		scoringParams{durationSec: 30},
	)
	var invalid *invalidAttemptError
	if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
		t.Fatalf("err = %v, want invalid attempt %s", err, reasonDetailMismatch)
	}
}

func TestDualZoneRegistered(t *testing.T) {
	if _, ok := lookupScorer("dual_zone"); !ok {
		t.Fatal("dual_zone scorer not registered")
	}
}
