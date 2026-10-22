package test

import (
	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	// §6.8 frozen (SkillTrajectoryTestPage.tsx:332-374, Java parity): an
	// unmatched round carries the 999 sentinel and must not pollute the
	// deviation/speed means; the precision denominator counts it.
	skillTrajectoryMissDevSentinel = 999.0
	skillTrajectoryPerfectDev      = 5.0
	skillTrajectoryGoodDev         = 15.0
	skillTrajectoryPrecisionWeight = 6.0
	skillTrajectoryDevPivot        = 50.0
	skillTrajectoryDevDivisor      = 10.0
	skillTrajectorySpeedPivotMs    = 12000.0
	skillTrajectorySpeedDivisor    = 1500.0
	skillTrajectoryCompositeCap    = 10.0
	// The fit zero-point for the linear track (§四: 8 per fit percentage
	// point) shares the deviation-score pivot — 50px reads as 0% fit.
	skillTrajectoryFitZeroDev = 50.0

	skillTrajectoryUnitScoreDefault = 8.0
)

type skillTrajectoryScorer struct{}

func init() {
	scorers = append(scorers, skillTrajectoryScorer{})
}

func (skillTrajectoryScorer) testCode() string {
	return "skill_trajectory"
}

func (skillTrajectoryScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetSkillTrajectory()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	rounds := detail.GetRounds()
	if len(rounds) == 0 {
		// AFK gets the honest zero, never the deviation/speed formulas.
		return scoreResult{}, nil
	}
	prevRound, perfect, good, matched := int32(0), 0, 0, 0
	sumDev, sumPlace, fitSum := 0.0, 0.0, 0.0
	for i, r := range rounds {
		if i > 0 && r.GetRoundNumber() <= prevRound {
			// Rounds arrive in play order; a repeated or shuffled number is
			// a trimmed stream.
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		prevRound = r.GetRoundNumber()
		if !skillTrajectoryShapeKnown(r.GetShape()) {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		dev := r.GetDeviation()
		if dev < 0 {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		placed := r.GetPlayerPlacement() != nil
		if placed == (dev >= skillTrajectoryMissDevSentinel) {
			// Presence mirrors the sentinel: an unplaced round carries 999; a
			// placed one a real sub-40 deviation (legacy match cap 40px).
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		if dev < skillTrajectoryPerfectDev {
			perfect++
		} else if dev < skillTrajectoryGoodDev {
			good++
		}
		// Unmatched rounds count as zero fit — not placing pays on the linear
		// track, matching the precision denominator's no-dilution stance.
		fitSum += clamp(1-dev/skillTrajectoryFitZeroDev, 0, 1)
		if dev >= skillTrajectoryMissDevSentinel {
			continue
		}
		matched++
		sumDev += dev
		sumPlace += float64(r.GetPlacementTimeMs())
	}
	counted := float64(len(rounds))
	precisionScore := (float64(perfect) + float64(good)*0.5) / counted * skillTrajectoryPrecisionWeight
	avgDev, avgPlace := skillTrajectoryMissDevSentinel, skillTrajectorySpeedPivotMs
	if matched > 0 {
		avgDev, avgPlace = sumDev/float64(matched), sumPlace/float64(matched)
	}
	deviationScore := max(0, (skillTrajectoryDevPivot-avgDev)/skillTrajectoryDevDivisor)
	speedScore := max(0, (skillTrajectorySpeedPivotMs-avgPlace)/skillTrajectorySpeedDivisor)
	composite := clamp(precisionScore+deviationScore+speedScore, 0, skillTrajectoryCompositeCap)
	// The linear track is percentage-based (8 per fit point), not per-event —
	// the one scorer whose raw is not count×unit, so the shared tail's raw
	// half is rebuilt here with the same extras shape.
	raw := round1(fitSum / counted * 100 * params.num(unitScoreKey, skillTrajectoryUnitScoreDefault) * params.difficulty())
	return scoreResult{raw: raw, extras: map[string]float64{extraComposite: round1(composite)}}, nil
}

func skillTrajectoryShapeKnown(shape string) bool {
	switch shape {
	case "circle", "square", "triangle", "pentagon", "hexagon", "star":
		return true
	}
	return false
}
