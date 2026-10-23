package test

import (
	"math"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	trackingMovingHitThresholdDefault = 45.0
	trackingMovingStabilityAnchor     = 40.0
	trackingMovingStabilityDivisor    = 8.0
	trackingMovingMoveDistanceCap     = 1000.0
	trackingMovingMoveBonusFloor      = 500.0
	trackingMovingReactionBonusMs     = 1500.0
	trackingMovingUnitScoreDefault    = 10.0
)

type trackingMovingScorer struct{}

func init() {
	scorers = append(scorers, trackingMovingScorer{})
}

func (trackingMovingScorer) testCode() string {
	return "tracking_moving"
}

func (trackingMovingScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetTrackingMoving()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	frames := detail.GetFrames()
	durationSec := params.effectiveDurationSec()
	hitPx := params.num("hit_threshold_px", trackingMovingHitThresholdDefault)
	type shotFrame struct {
		deviation, lifetime float64
		hit                 bool
	}
	shots := make([]shotFrame, 0, len(frames))
	var moveDistance, prevX, prevY float64
	havePrev := false
	prevMs := float64(-1)
	for _, f := range frames {
		elapsed := float64(f.GetElapsedMs())
		if prevMs >= 0 && elapsed <= prevMs {
			// The merged stream is time-sorted on the wire — a regression is
			// an edited history.
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		if f.GetDeviation() < 0 {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		if f.Hit != nil {
			// Shot frames carry both facts or they are not shot frames:
			// the lifetime backs the reaction term, and the hit verdict is
			// cross-checked against the deviation the same frame reports.
			if f.DotLifetimeMs == nil {
				return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
			}
			lifetime := float64(f.GetDotLifetimeMs())
			if lifetime < reactionTimeMin {
				return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdReactionTime}
			}
			if f.GetHit() != (f.GetDeviation() <= hitPx) {
				return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
			}
			shots = append(shots, shotFrame{deviation: f.GetDeviation(), lifetime: lifetime, hit: f.GetHit()})
		} else {
			// Movement frames never carry a lifetime — the shape is the
			// discriminator the legacy merged stream lacked.
			if f.DotLifetimeMs != nil {
				return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
			}
			if havePrev {
				dx := f.GetPlayer().GetX() - prevX
				dy := f.GetPlayer().GetY() - prevY
				moveDistance += math.Sqrt(dx*dx + dy*dy)
			}
			prevX, prevY = f.GetPlayer().GetX(), f.GetPlayer().GetY()
			havePrev = true
		}
		prevMs = elapsed
	}
	if len(frames) > 0 && prevMs > durationSec*1000+spanGraceMs {
		return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdTimeSpan}
	}
	if len(shots) == 0 {
		// Honest zero: the client scored a shotless session zero too — the
		// stability term must not pay for never shooting.
		return scoreResult{raw: 0, extras: map[string]float64{extraComposite: 0}}, nil
	}
	hits, devSum, hitLifetimeSum := 0, 0.0, 0.0
	for _, s := range shots {
		devSum += s.deviation
		if s.hit {
			hits++
			hitLifetimeSum += s.lifetime
		}
	}
	accuracy := float64(hits) / float64(len(shots))
	avgDev := devSum / float64(len(shots))
	var avgLifetime float64
	if hits > 0 {
		avgLifetime = hitLifetimeSum / float64(hits)
	}
	stability := math.Max(0, (trackingMovingStabilityAnchor-avgDev)/trackingMovingStabilityDivisor)
	movement := math.Min(1, moveDistance/trackingMovingMoveDistanceCap) * 2
	multitasking := accuracy * 0.4
	if avgLifetime < trackingMovingReactionBonusMs {
		// The legacy arm pays this bonus to a hitless session too (its mean
		// defaults to zero) — parity kept, a disclosed carry.
		multitasking += 0.3
	}
	if moveDistance > trackingMovingMoveBonusFloor {
		multitasking += 0.3
	}
	composite := clamp(accuracy*4+stability+movement+multitasking*2, 0, 10)
	// The linear track counts confirmed hits.
	return curvePendingResult(hits, params, trackingMovingUnitScoreDefault, composite), nil
}
