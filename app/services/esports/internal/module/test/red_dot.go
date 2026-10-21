package test

import (
	"math"

	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	// §6.1 frozen (Java RedDotScorer, 217-clean-run quantile anchors — an
	// absolute standard, not population-drifting); only the unit score is
	// calibratable (metric.md §四).
	redDotRTRef              = 500.0 // avgRt=500 → speed factor 1.0
	redDotRTFloor            = 200.0
	redDotRTCeil             = 1500.0
	redDotPreciseDeviationPx = 15.0 // strictly below — Java uses <, not ≤
	redDotPrecisionFloor     = 0.3  // zero precision never zeroes the score
	redDotPrecisionWeight    = 0.7
	redDotUnitScoreDefault   = 10.0
	redDotUnitScoreKey       = "unit_score"
)

type redDotScorer struct{}

func init() {
	scorers = append(scorers, redDotScorer{})
}

func (redDotScorer) testCode() string {
	return "red_dot"
}

func (redDotScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetRedDot()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	hits := 0
	precise := 0
	rts := make([]float64, 0, len(detail.GetEvents()))
	for _, e := range detail.GetEvents() {
		if !e.GetIsHit() {
			continue
		}
		hits++
		rt := float64(e.GetReactionTimeMs())
		// Per-hit sanity window: a value outside it is fabricated data —
		// the attempt is rejected, never scored zero (§6.1 outlier guard).
		if rt < reactionTimeMin || rt > reactionTimeMax {
			return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdReactionTime}
		}
		rts = append(rts, rt)
		if e.GetDeviation() < redDotPreciseDeviationPx {
			precise++
		}
	}
	if hits == 0 {
		// No hit carries no outlier risk; both tracks are an honest zero.
		return scoreResult{raw: 0, authoritative: utils.Ptr(0.0), grade: utils.Ptr("C")}, nil
	}
	raw := linearRaw(hits, params, redDotUnitScoreDefault)
	avgRt := clamp(mean(rts), redDotRTFloor, redDotRTCeil)
	precision := float64(precise) / float64(hits)
	calibrationInput := float64(hits) * (redDotRTRef / avgRt) *
		(redDotPrecisionFloor + redDotPrecisionWeight*precision)
	authoritative := round1(calibrateRedDot(calibrationInput))
	return scoreResult{
		raw:           round1(raw),
		authoritative: &authoritative,
		grade:         utils.Ptr(band(authoritative)),
		extras:        map[string]float64{extraCalibrationInput: round1(calibrationInput)},
	}, nil
}

func calibrateRedDot(raw float64) float64 {
	// Frozen piecewise-linear anchors (p50/p75/p90/p99 of 217 clean runs);
	// above p99 the final segment's slope runs to the 120→100 cap.
	switch {
	case raw <= 0:
		return 0
	case raw <= 8.8:
		return raw / 8.8 * 50
	case raw <= 14.7:
		return 50 + (raw-8.8)/(14.7-8.8)*20
	case raw <= 28:
		return 70 + (raw-14.7)/(28-14.7)*20
	case raw <= 59.6:
		return 90 + (raw-28)/(59.6-28)*9
	default:
		return math.Min(100, 99+(raw-59.6)/(120-59.6))
	}
}
