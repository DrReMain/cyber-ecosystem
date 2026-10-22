package test

import (
	"strings"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	offenseDefenseQualityWeight = 6.0
	offenseDefenseSpeedAnchorMs = 8000.0
	offenseDefenseSpeedDivisor  = 2000.0
	offenseDefenseSpeedCap      = 2.0
	// The legacy formula's flat consistency term — a constant every attempt
	// earns; carried for curve parity and disclosed in §6.10.
	offenseDefenseConsistency       = 0.5
	offenseDefenseTimeoutWeight     = 0.5
	offenseDefenseTimeoutPenaltyCap = 2.0
	offenseDefenseUnitScoreDefault  = 25.0
)

type offenseDefenseScorer struct{}

func init() {
	scorers = append(scorers, offenseDefenseScorer{})
}

func (offenseDefenseScorer) testCode() string {
	return "offense_defense"
}

func (offenseDefenseScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetOffenseDefense()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	decisions := detail.GetDecisions()
	if len(decisions) == 0 {
		// Honest zero: an attempt with no decisions earns nothing (the
		// legacy scorer returned ZERO for the empty list).
		return curvePendingResult(0, params, offenseDefenseUnitScoreDefault, 0), nil
	}
	correct, timeouts := 0, 0
	times := make([]float64, 0, len(decisions))
	for _, d := range decisions {
		optimal := answerString(params.answerKey[d.GetScenarioId()], "optimal_choice")
		if optimal == "" {
			// Decisions are judged against the seeded key — an unknown
			// scenario id is an edited history, not a degraded client bank
			// (the legacy fallback trusted the client's isCorrect flag).
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		// An absent choice is an unanswered round (a timeout landing with no
		// tap); a present choice matches the key case-insensitively, the
		// legacy equalsIgnoreCase parity — garbage vocabulary simply misses.
		if choice := d.GetChoice(); choice != "" && strings.EqualFold(choice, optimal) {
			correct++
		}
		if d.GetTimeout() {
			timeouts++
		}
		elapsed := float64(d.GetDecisionTimeMs())
		if elapsed < reactionTimeMin {
			// Per-decision floor: the legacy gate watched only the mean, so a
			// single fabricated instant could hide under slow siblings.
			return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdReactionTime}
		}
		times = append(times, elapsed)
	}
	// The denominator is the decision array itself — the legacy formula
	// trusted the client's declared totalRounds, an inflate-by-shrink face.
	quality := float64(correct) / float64(len(decisions)) * offenseDefenseQualityWeight
	speed := clamp(
		(offenseDefenseSpeedAnchorMs-mean(times))/offenseDefenseSpeedDivisor*offenseDefenseSpeedCap,
		0, offenseDefenseSpeedCap)
	penalty := min(offenseDefenseTimeoutPenaltyCap, float64(timeouts)*offenseDefenseTimeoutWeight)
	composite := clamp(quality+speed+offenseDefenseConsistency-penalty, 0, 10)
	// The legacy scorer computed a complex-scenario accuracy it never used;
	// dead aggregate dropped. Authoritative stays nil (§6.10 curve pending).
	return curvePendingResult(correct, params, offenseDefenseUnitScoreDefault, round1(composite)), nil
}
