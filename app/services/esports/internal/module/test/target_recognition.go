package test

import (
	"math"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	targetRecognitionUnitScoreDefault     = 20.0
	targetRecognitionAccuracyWeight       = 5.0
	targetRecognitionSpeedPivotMs         = 600.0
	targetRecognitionSpeedDivisor         = 60.0
	targetRecognitionVisualFloorMs        = 200.0
	targetRecognitionVisualSpanMs         = 800.0
	targetRecognitionVisualWeight         = 2.0
	targetRecognitionFalsePositivePenalty = 0.4
	targetRecognitionMissedPenalty        = 0.2
	targetRecognitionCompositeCap         = 10.0
)

type targetRecognitionScorer struct{}

func init() {
	scorers = append(scorers, targetRecognitionScorer{})
}

func (targetRecognitionScorer) testCode() string {
	return "target_recognition"
}

func (targetRecognitionScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetTargetRecognition()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	spawnedEnemy := detail.GetSpawnedEnemy()
	correct := 0
	falsePositives := 0
	rts := make([]float64, 0, len(detail.GetClicks()))
	for _, c := range detail.GetClicks() {
		if rt := c.ReactionTimeMs; rt != nil {
			v := float64(*rt)
			if v < reactionTimeMin || v > reactionTimeMax {
				return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdReactionTime}
			}
			if c.GetTargetType() == "enemy" {
				rts = append(rts, v)
			}
		} else if c.GetTargetType() == "enemy" {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		switch c.GetTargetType() {
		case "enemy":
			correct++
		case "friendly", "neutral":
			falsePositives++
		default:
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
	}
	// Correct identifications above the spawned count is a fabricated
	// stream — legacy clamped the rate client-side; the server rejects.
	if correct > int(spawnedEnemy) {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	if correct == 0 || spawnedEnemy == 0 {
		return scoreResult{}, nil
	}
	avgRt := mean(rts)
	accuracy := float64(correct) / float64(spawnedEnemy) * targetRecognitionAccuracyWeight
	speed := math.Max(0, (targetRecognitionSpeedPivotMs-avgRt)/targetRecognitionSpeedDivisor)
	visual := clamp(1-(avgRt-targetRecognitionVisualFloorMs)/targetRecognitionVisualSpanMs, 0, 1) * targetRecognitionVisualWeight
	missed := float64(spawnedEnemy-int32(correct)) * targetRecognitionMissedPenalty
	composite := clamp(
		accuracy+speed+visual-
			float64(falsePositives)*targetRecognitionFalsePositivePenalty-missed,
		0, targetRecognitionCompositeCap)
	return curvePendingResult(correct, params, targetRecognitionUnitScoreDefault, composite), nil
}
