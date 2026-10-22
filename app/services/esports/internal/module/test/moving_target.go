package test

import (
	"math"
	"slices"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	movingTargetIntervalDefault   = 50.0
	movingTargetPerfectDevDefault = 20.0
	movingTargetGoodDevDefault    = 30.0
	movingTargetPatternDevDefault = 80.0
	movingTargetStabilityAnchor   = 50.0
	movingTargetStabilityDivisor  = 10.0
	movingTargetLossPenalty       = 0.3
	movingTargetUnitScoreDefault  = 10.0
)

var movingTargetPatterns = [6]string{"straight", "curve", "zigzag", "circle", "random", "variable"}

type movingTargetScorer struct{}

func init() {
	scorers = append(scorers, movingTargetScorer{})
}

func (movingTargetScorer) testCode() string {
	return "moving_target"
}

func (movingTargetScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetMovingTarget()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	shots := detail.GetShots()
	if len(shots) == 0 {
		// Honest zero: the legacy formula's stability term handed an empty
		// stream five free points — the E3-f empty-attempt wart, fixed here
		// the same way.
		return scoreResult{raw: 0, extras: map[string]float64{extraComposite: 0}}, nil
	}
	intervalMs := params.num("sample_interval_ms", movingTargetIntervalDefault)
	perfectPx := params.num("perfect_deviation_px", movingTargetPerfectDevDefault)
	goodPx := params.num("good_deviation_px", movingTargetGoodDevDefault)
	patternPx := params.num("pattern_deviation_px", movingTargetPatternDevDefault)
	durationSec := params.effectiveDurationSec()
	perfect, good, losses := 0, 0, 0
	patternSec := make(map[string]float64, 6)
	sum, sumSq, prev, firstMs := 0.0, 0.0, float64(-1), 0.0
	// The state machine mirrors the client loop verbatim: it starts lost
	// (a first lost sample is not a loss transition) and only the
	// non-lost → lost edge counts, so runs of loss charge once.
	state := "lost"
	for _, s := range shots {
		pattern := s.GetPattern()
		if !slices.Contains(movingTargetPatterns[:], pattern) {
			// Patterns are judged against the six-phase rotation's closed
			// vocabulary — an unknown phase is an edited history.
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		dev := s.GetDeviation()
		if dev < 0 {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		elapsed := float64(s.GetElapsedMs())
		if prev >= 0 && elapsed <= prev {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		switch {
		case dev <= perfectPx:
			perfect++
			good++
			if state == "lost" {
				state = "perfect"
			}
		case dev <= goodPx:
			good++
			if state == "lost" {
				state = "good"
			}
		default:
			if state != "lost" {
				losses++
			}
			state = "lost"
		}
		if dev <= patternPx {
			patternSec[pattern] += intervalMs / 1000
		}
		sum += dev
		sumSq += dev * dev
		if prev < 0 {
			firstMs = elapsed
		}
		prev = elapsed
	}
	if prev > durationSec*1000+spanGraceMs {
		// The last sample sits inside the config mirror (§6.5 fabrication
		// closure).
		return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdTimeSpan}
	}
	streamSpanMs := prev - firstMs
	if float64(len(shots))*intervalMs > streamSpanMs+spanGraceMs {
		// Counts times the sampling interval cannot exceed the stream's own
		// span — a late-starting dense burst would otherwise mint tracking
		// time out of thin air (the gate's denominator is last minus first,
		// never the absolute clock).
		return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdTimeSpan}
	}
	n := float64(len(shots))
	mean := sum / n
	variance := sumSq/n - mean*mean
	if variance < 0 {
		// Floating-point cancellation can shave a true zero negative.
		variance = 0
	}
	stdDev := math.Sqrt(variance)
	stepSec := intervalMs / 1000
	perfectSec := float64(perfect) * stepSec
	goodSec := float64(good) * stepSec
	coverage := goodSec / durationSec * 3
	perfectScore := perfectSec / durationSec * 4
	stability := math.Max(0, (movingTargetStabilityAnchor-stdDev)/movingTargetStabilityDivisor)
	patternSum := 0.0
	for _, p := range movingTargetPatterns {
		patternSum += patternSec[p]
	}
	// Six phases always initialized client-side, so the average divides by
	// the vocabulary size, not by the phases the stream happened to carry.
	adaptation := patternSum / float64(len(movingTargetPatterns)) * 2
	composite := clamp(coverage+perfectScore+stability+adaptation-float64(losses)*movingTargetLossPenalty, 0, 10)
	// The linear track counts perfect-tracking seconds (§四: 10/秒完美追踪).
	raw := round1(perfectSec * params.num(unitScoreKey, movingTargetUnitScoreDefault) * params.difficulty())
	return scoreResult{raw: raw, extras: map[string]float64{extraComposite: round1(composite)}}, nil
}
