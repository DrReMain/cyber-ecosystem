package test

import (
	"sort"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	numberStarsRoundsBase       = 8.0
	numberStarsSpeedAnchorMs    = 2000.0
	numberStarsSpeedDivisor     = 200.0
	numberStarsAccuracyCap      = 3.0
	numberStarsAccuracyWeight   = 0.5
	numberStarsBaselineLen      = 5
	numberStarsComplexityStep   = 0.3
	numberStarsUnitScoreDefault = 15.0
)

type numberStarsScorer struct{}

func init() {
	scorers = append(scorers, numberStarsScorer{})
}

func (numberStarsScorer) testCode() string {
	return "number_stars"
}

func (numberStarsScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetNumberStars()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}

	correctTotal, incorrectTotal, roundsCompleted := 0, 0, 0
	intervalSum, intervalCount := 0.0, 0
	maxSeqLen := numberStarsBaselineLen
	prevRound, haveRound := int32(0), false
	for _, round := range detail.GetRounds() {
		roundNumber := round.GetRoundNumber()
		if haveRound && roundNumber <= prevRound {
			// A repeated or regressed round number is an edited history —
			// the array is the only truth for roundsCompleted.
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		prevRound, haveRound = roundNumber, true

		numbers := round.GetNumbers()
		if len(numbers) == 0 {
			// The engine always spawns five numbers per round; an empty
			// target sequence would game roundsCompleted for free.
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		expected := append([]int32(nil), numbers...)
		sort.Slice(expected, func(i, j int) bool { return expected[i] < expected[j] })

		// Tap-ascending semantics: a click is correct when it names the
		// smallest not-yet-tapped number — the progress pointer advances
		// only on a hit, so a mistap retries the same expectation.
		progress := 0
		prevElapsed, correctElapsed := -1.0, []float64(nil)
		for _, click := range round.GetClicks() {
			elapsed := float64(click.GetElapsedMs())
			if elapsed <= prevElapsed {
				// Strictly increasing: an equal stamp is a same-frame double
				// tap, whose zero interval would inflate the speed track.
				return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
			}
			prevElapsed = elapsed
			if progress < len(expected) && click.GetNumber() == expected[progress] {
				correctElapsed = append(correctElapsed, elapsed)
				progress++
			}
		}
		// Decision intervals ride the correct taps only, within the round —
		// rounds never chain head-to-tail.
		for i := 1; i < len(correctElapsed); i++ {
			intervalSum += correctElapsed[i] - correctElapsed[i-1]
			intervalCount++
		}

		correctTotal += progress
		incorrectTotal += len(round.GetClicks()) - progress
		if progress == len(numbers) {
			roundsCompleted++
		}
		maxSeqLen = max(maxSeqLen, len(numbers))
	}

	base := min(5.0, float64(roundsCompleted)/numberStarsRoundsBase*5)
	speed := 0.0
	if intervalCount > 0 {
		speed = max(0, (numberStarsSpeedAnchorMs-intervalSum/float64(intervalCount))/numberStarsSpeedDivisor)
	}
	accuracy := max(0, numberStarsAccuracyCap-float64(incorrectTotal)*numberStarsAccuracyWeight)
	complexity := max(0, float64(maxSeqLen-numberStarsBaselineLen)*numberStarsComplexityStep)
	if correctTotal == 0 {
		// Honest zero: idling or blind tapping earns nothing (legacy parity) —
		// the legacy formula's empty-attempt 3.0 accuracy floor is dropped.
		return curvePendingResult(0, params, numberStarsUnitScoreDefault, 0), nil
	}
	composite := clamp(base+speed+accuracy+complexity, 0, 10)
	// Authoritative stays nil: the calibration curve is pending.
	return curvePendingResult(correctTotal, params, numberStarsUnitScoreDefault, round1(composite)), nil
}
