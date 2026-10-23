package test

import (
	"math"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	resourceMatchMinionCapDefault    = 25.0
	resourceMatchMonsterCapDefault   = 12.0
	resourceMatchDragonCapDefault    = 5.0
	resourceMatchCollisionCapDefault = 10.0
	resourceMatchOptimalDefault      = resourceMatchMinionCapDefault + resourceMatchMonsterCapDefault + resourceMatchDragonCapDefault
	// The legacy formula's flat tactical term — a constant every attempt
	// earns; carried for curve parity, a disclosed carry.
	resourceMatchTacticalDiversity  = 1.0
	resourceMatchEfficiencySaturate = 2.0
	resourceMatchSpanFloorMs        = 500.0
	resourceMatchUnitScoreDefault   = 20.0
)

type resourceMatchScorer struct{}

func init() {
	scorers = append(scorers, resourceMatchScorer{})
}

func (resourceMatchScorer) testCode() string {
	return "resource_match"
}

func (resourceMatchScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetResourceMatch()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	// Spawn caps are engine params (minion_count/monster_count/dragon_count)
	// so a retuned field economy recalibrates without a code change; the
	// legacy formula hardcoded 42.
	caps := map[string]float64{
		"minion":  params.num("minion_count", resourceMatchMinionCapDefault),
		"monster": params.num("monster_count", resourceMatchMonsterCapDefault),
		"dragon":  params.num("dragon_count", resourceMatchDragonCapDefault),
	}
	optimal := caps["minion"] + caps["monster"] + caps["dragon"]
	if optimal <= 0 {
		optimal = resourceMatchOptimalDefault
	}
	counts := map[string]int{}
	lastCollectMs := float64(-1)
	for _, c := range detail.GetCollects() {
		if _, known := caps[c.GetKind()]; !known {
			// Kinds are judged against the spawn economy's vocabulary — an
			// unknown kind is an edited history, not a degraded client bank.
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		elapsed := float64(c.GetElapsedMs())
		if lastCollectMs >= 0 {
			if elapsed <= lastCollectMs {
				return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
			}
			if elapsed-lastCollectMs < reactionTimeMin {
				// Per-collect floor: two picks closer than any movement can
				// bridge is a fabricated density, the diff-gate family.
				return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdReactionTime}
			}
		}
		counts[c.GetKind()]++
		if float64(counts[c.GetKind()]) > caps[c.GetKind()] {
			// Collecting past the spawn cap fails the Java bounds check
			// server-side; the cap itself rides the engine params.
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		lastCollectMs = elapsed
	}
	lastCollisionMs := float64(-1)
	for _, c := range detail.GetCollisions() {
		elapsed := float64(c.GetElapsedMs())
		if lastCollisionMs >= 0 && elapsed <= lastCollisionMs {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		lastCollisionMs = elapsed
	}
	// The collision cap is the death economy: health 100 at −10 a hit means
	// the eleventh collision cannot exist because the game already ended.
	collisionCap := params.num("collision_cap", resourceMatchCollisionCapDefault)
	if float64(len(detail.GetCollisions())) > collisionCap {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	durationSec := float64(params.durationSec)
	if durationSec <= 0 {
		durationSec = sessionDurationFallbackSec
	}
	spanMs := math.Max(lastCollectMs, lastCollisionMs)
	var spanSec float64
	if spanMs < 0 {
		// No events at all — the Java cold arm scored a timestamp-less
		// request against the full session span.
		spanSec = durationSec
	} else {
		if spanMs > durationSec*1000+spanGraceMs {
			// The span is bounded by the config mirror, never the client's
			// clock (a fabrication closure).
			return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdTimeSpan}
		}
		if spanMs < resourceMatchSpanFloorMs {
			// Ten separate hits cannot land inside half a second — an event
			// stream squeezed into an instant is fabricated.
			return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdTimeSpan}
		}
		spanSec = math.Max(spanMs/1000, 0.001)
	}
	collected := len(detail.GetCollects())
	resourceEfficiency := float64(collected) / optimal * 4
	pathEfficiency := math.Min(1, float64(collected)/spanSec/resourceMatchEfficiencySaturate)
	riskControl := math.Max(0, 2-float64(len(detail.GetCollisions()))*0.1)
	composite := clamp(resourceEfficiency+pathEfficiency*3+riskControl+resourceMatchTacticalDiversity, 0, 10)
	return curvePendingResult(collected, params, resourceMatchUnitScoreDefault, round1(composite)), nil
}
