package test

import (
	"time"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	// Frozen (Java DualZoneScorer + the skillPrediction page settle):
	// the survival reference and the preset weight stay ops-tunable through
	// the base_score_time / difficulty_weight engine-param keys.
	dualZoneBaseTimeSecDefault      = 30.0
	dualZoneBaseTimeSecKey          = "base_score_time"
	dualZoneDifficultyWeightDefault = 1.15
	dualZoneDifficultyWeightKey     = "difficulty_weight"
	dualZoneUnitScoreDefault        = 15.0

	// The legacy ±1500ms wall-clock tolerance, applied one-way per gate —
	// see the two call sites for which side each gate bounds.
	dualZoneWallClockToleranceMs = 1500.0
	dualZoneMinSurvivalMs        = 500.0
	dualZoneCompositeCap         = 100.0
)

type dualZoneScorer struct{}

func init() {
	scorers = append(scorers, dualZoneScorer{})
}

func (dualZoneScorer) testCode() string {
	return "dual_zone"
}

func (dualZoneScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetDualZone()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	netMs := netPlayMs(req)
	if netMs < dualZoneMinSurvivalMs {
		return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdTimeSpan}
	}
	behaviors := detail.GetBehaviors()
	dodged := 0
	lastCollision := -1
	prevElapsed, maxElapsed := 0.0, -1.0
	seen := map[int32]struct{}{}
	for i, b := range behaviors {
		elapsed := float64(b.GetElapsedMs())
		if elapsed < prevElapsed {
			// The stream is game-loop push order; a backwards stamp is an
			// edited stream, not a slow game.
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		prevElapsed, maxElapsed = elapsed, max(maxElapsed, elapsed)
		if b.GetZone() != "left" && b.GetZone() != "right" {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		switch b.GetEvent() {
		case "dodge":
			dodged++
		case "collision":
			lastCollision = i
		default:
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		// One obstacle records one event — a repeated id double-counts.
		if b.ObstacleId != nil {
			if _, dup := seen[b.GetObstacleId()]; dup {
				return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
			}
			seen[b.GetObstacleId()] = struct{}{}
		}
	}
	if lastCollision >= 0 {
		// Collision ends the run (legacy triggerCollision parity): nothing
		// may follow it, and the net clock must land on it.
		collisionMs := float64(behaviors[lastCollision].GetElapsedMs())
		if lastCollision != len(behaviors)-1 ||
			absF(collisionMs-netMs) > dualZoneWallClockToleranceMs {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
	}
	if maxElapsed-netMs > dualZoneWallClockToleranceMs {
		// A stream may end before the clock (idle tail to the timer), but
		// never overrun the net clock — events past it are fabricated time.
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	if durationMs := float64(params.durationSec) * 1000; durationMs > 0 &&
		netMs-durationMs > dualZoneWallClockToleranceMs {
		// The submitted span cannot exceed the server-configured length
		// (a duration fabrication closure); the cold arm keeps a
		// config-less call path scoreable.
		return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdTimeSpan}
	}
	baseTimeSec := params.num(dualZoneBaseTimeSecKey, dualZoneBaseTimeSecDefault)
	weight := params.num(dualZoneDifficultyWeightKey, dualZoneDifficultyWeightDefault)
	serverScore := max(0, min(netMs/1000/baseTimeSec*100*weight, dualZoneCompositeCap))
	// Authoritative stays nil: the calibration curve is pending.
	return curvePendingResult(dodged, params, dualZoneUnitScoreDefault, round1(serverScore/10)), nil
}

func netPlayMs(req *esportspb.SubmitTestAttemptRequest) float64 {
	// The pause-carried span: the submitted wall clock minus the background
	// segments — the server-side dual of the engine's clock-pause carry, so
	// the stream tolerance compares against play time, not wall time.
	started, ended := req.GetStartedAt(), req.GetEndedAt()
	if started == nil || ended == nil {
		return 0
	}
	span := ended.AsTime().Sub(started.AsTime())
	for _, b := range req.GetBackgrounds() {
		span -= time.Duration(b.GetDurationMs()) * time.Millisecond
	}
	return float64(span.Milliseconds())
}

func absF(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
