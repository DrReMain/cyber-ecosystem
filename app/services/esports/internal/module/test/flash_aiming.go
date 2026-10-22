package test

import (
	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	// Composite formula transcribed from the legacy frontend settle
	// (FlashAimingTestPage.tsx:134-178, PRD 2.9): frozen shape, except the
	// center threshold, which shares the client boot pack's engine-param key.
	flashCenterDeviationPxDefault = 20.0
	flashCenterDeviationPxKey     = "center_hit_deviation_px"
	flashUnitScoreDefault         = 10.0

	flashHitRateWeight    = 3.0
	flashAccuracyWeight   = 4.0
	flashDeviationPivot   = 30.0
	flashDeviationDivisor = 5.0
	flashAimPivot         = 800.0
	flashAimDivisor       = 100.0
	flashCompositeCap     = 10.0
)

type flashAimingScorer struct{}

func init() {
	scorers = append(scorers, flashAimingScorer{})
}

func (flashAimingScorer) testCode() string {
	return "flash_aiming"
}

func (flashAimingScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetFlashAiming()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	shots := detail.GetShots()
	// The contract's misses field is the only carrier for expired-or-replaced
	// targets — the shot stream has no event for them (E0 increment 4).
	misses := int(detail.GetMisses())
	hits := len(shots)
	total := hits + misses
	if total == 0 {
		return scoreResult{}, nil
	}
	centerThreshold := params.num(flashCenterDeviationPxKey, flashCenterDeviationPxDefault)
	centerHits := 0
	devSum, aimSum := 0.0, 0.0
	for _, s := range shots {
		deviation := s.GetDeviation()
		devSum += deviation
		aimSum += float64(s.GetAimTimeMs())
		// ≤ matches the legacy settle and the engine's local results face.
		if deviation <= centerThreshold {
			centerHits++
		}
	}
	avgDeviation, avgAimTime := flashDeviationPivot, flashAimPivot
	if hits > 0 {
		avgDeviation = devSum / float64(hits)
		avgAimTime = aimSum / float64(hits)
	}
	if avgAimTime < reactionTimeMin {
		return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdReactionTime}
	}
	hitRateScore := float64(hits) / float64(total) * flashHitRateWeight
	accuracyScore := 0.0
	if hits > 0 {
		accuracyScore = float64(centerHits) / float64(hits) * flashAccuracyWeight
	}
	deviationScore := max(0, (flashDeviationPivot-avgDeviation)/flashDeviationDivisor)
	reactionScore := max(0, (flashAimPivot-avgAimTime)/flashAimDivisor)
	composite := clamp(hitRateScore+accuracyScore+deviationScore+reactionScore, 0, flashCompositeCap)
	return curvePendingResult(hits, params, flashUnitScoreDefault, composite), nil
}
