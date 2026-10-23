package test

import (
	"math"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	// Frozen (legacy engine TIER_SPECS); tier values stay ops-tunable
	// through the tier_value_* engine-param keys without a deploy.
	multiTargetHighValueDefault   = 100.0
	multiTargetMediumValueDefault = 50.0
	multiTargetLowValueDefault    = 25.0
	multiTargetTrapValueDefault   = -50.0
	multiTargetHighValueKey       = "tier_value_high"
	multiTargetMediumValueKey     = "tier_value_medium"
	multiTargetLowValueKey        = "tier_value_low"
	multiTargetTrapValueKey       = "tier_value_trap"
	multiTargetUnitScoreDefault   = 10.0

	multiTargetOptimalFloorSec     = 1.0
	multiTargetOptimalPerSec       = 100.0
	multiTargetValueWeight         = 5.0
	multiTargetPriorityWeight      = 3.0
	multiTargetSpeedPivotMs        = 1500.0
	multiTargetSpeedDivisor        = 150.0
	multiTargetTrapPenalty         = 0.3
	multiTargetMissedHighPenalty   = 0.2
	multiTargetCompositeCap        = 10.0
	multiTargetDurationFallbackSec = 30.0
)

type multiTargetScorer struct{}

func init() {
	scorers = append(scorers, multiTargetScorer{})
}

func (multiTargetScorer) testCode() string {
	return "multi_target"
}

func (multiTargetScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetMultiTarget()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	picks := detail.GetPicks()
	hits := 0
	highHits := 0
	trapHits := 0
	rawTotal := 0.0
	deltas := make([]float64, 0, len(picks))
	prev := 0.0
	for _, p := range picks {
		elapsed := float64(p.GetElapsedMs())
		delta := elapsed - prev
		prev = elapsed
		// Pick-to-pick switching cost: the frozen pick carries no
		// spawn stamp, so the delta is the rebuildable form of legacy's
		// click-minus-spawn reaction time.
		if delta < reactionTimeMin {
			return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdReactionTime}
		}
		deltas = append(deltas, delta)
		// The vocabulary gate binds every pick — a miss riding an unknown
		// tier is still a trimmed stream, not a scoreable fact.
		value, known := multiTargetTierValue(p.GetTier(), params)
		if !known {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		if !p.GetIsHit() {
			continue
		}
		hits++
		rawTotal += value
		switch p.GetTier() {
		case "high":
			highHits++
		case "trap":
			trapHits++
		}
	}
	if hits == 0 {
		return scoreResult{}, nil
	}
	durationSec := float64(params.durationSec)
	if durationSec <= 0 {
		// Java parity fallback for a config-less call path; the seeded
		// schema default keeps this arm cold in practice.
		durationSec = multiTargetDurationFallbackSec
	}
	optimal := math.Max(durationSec, multiTargetOptimalFloorSec) * multiTargetOptimalPerSec
	valueScore := rawTotal / optimal * multiTargetValueWeight
	priorityScore := float64(highHits) / float64(hits) * multiTargetPriorityWeight
	speedScore := math.Max(0, (multiTargetSpeedPivotMs-mean(deltas))/multiTargetSpeedDivisor)
	missedHigh := float64(detail.GetMissedHigh())
	composite := clamp(
		valueScore+priorityScore+speedScore-
			float64(trapHits)*multiTargetTrapPenalty-missedHigh*multiTargetMissedHighPenalty,
		0, multiTargetCompositeCap)
	// Authoritative stays nil: the calibration curve is pending.
	return curvePendingResult(hits, params, multiTargetUnitScoreDefault, composite), nil
}

func multiTargetTierValue(tier string, p scoringParams) (float64, bool) {
	switch tier {
	case "high":
		return p.num(multiTargetHighValueKey, multiTargetHighValueDefault), true
	case "medium":
		return p.num(multiTargetMediumValueKey, multiTargetMediumValueDefault), true
	case "low":
		return p.num(multiTargetLowValueKey, multiTargetLowValueDefault), true
	case "trap":
		return p.num(multiTargetTrapValueKey, multiTargetTrapValueDefault), true
	}
	return 0, false
}
