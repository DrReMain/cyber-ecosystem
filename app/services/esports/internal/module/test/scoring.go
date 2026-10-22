package test

import (
	"math"
	"strconv"
	"sync"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	reasonAbsurdReactionTime   = "ABSURD_REACTION_TIME"
	reasonDetailMismatch       = "DETAIL_MISMATCH"
	reasonBackgroundOverBudget = "BACKGROUND_OVER_BUDGET"

	extraCalibrationInput = "calibration_input"
	extraComposite        = "composite"

	bandSMin = 95.0
	bandAMin = 85.0
	bandBMin = 60.0

	unitScoreKey = "unit_score"

	reactionTimeMin = 80.0
	reactionTimeMax = 60000.0
)

type invalidAttemptError struct {
	reason string
}

func (e *invalidAttemptError) Error() string {
	return "scoring: invalid attempt: " + e.reason
}

type scoringParams struct {
	// difficultyFactor carries metric.md §三 (experience 1.0 / standard 1.5);
	// the zero value means unset and falls back to the experience baseline.
	difficultyFactor float64
	// engineParams is the same KV the client boot pack reads — one data
	// source for both ends (metric.md §五 consistency contract); absent or
	// non-numeric keys fall back to the scorer's metric-sheet constants.
	engineParams map[string]string
}

func (p scoringParams) num(key string, fallback float64) float64 {
	raw, ok := p.engineParams[key]
	if !ok {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(v) {
		return fallback
	}
	return v
}

func (p scoringParams) difficulty() float64 {
	if p.difficultyFactor > 0 {
		return p.difficultyFactor
	}
	return 1.0
}

type scoreResult struct {
	// raw: the uncapped linear track, 原始成绩 × 单位分 × 难度系数 (metric.md §一).
	raw float64
	// authoritative: the 0-100 calibrated track grade derives from — nil
	// while a game's calibration curve is pending.
	authoritative *float64
	grade         *string
	// extras: game-specific derived values (curve inputs, audit numbers).
	extras map[string]float64
}

type scorer interface {
	testCode() string
	// recompute takes the whole submit request because the gen oneof
	// interface is unexported (unnameable outside the gen package); each
	// scorer pulls its own detail via the typed getter and treats a nil
	// getter result as a dispatcher mismatch.
	recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error)
}

var scorers []scorer

var scorerIndex = sync.OnceValue(func() map[string]scorer {
	return buildScorerIndex(scorers)
})

func buildScorerIndex(list []scorer) map[string]scorer {
	// Duplicate codes panic — fail-fast at first use, the Java registry's
	// startup IllegalStateException parity; the oneof dispatcher guarantees
	// one detail type per code, so duplicates are a programming error.
	index := make(map[string]scorer, len(list))
	for _, s := range list {
		code := s.testCode()
		if _, dup := index[code]; dup {
			panic("scoring: duplicate scorer for " + code)
		}
		index[code] = s
	}
	return index
}

func lookupScorer(code string) (scorer, bool) {
	s, ok := scorerIndex()[code]
	return s, ok
}

func band(score float64) string {
	// The red_dot curve is quantile-frozen (p95→95), so the score reads as
	// percentile position and the thresholds mirror §二's top 5/15/40;
	// per-game curves keep the same rule.
	switch {
	case score >= bandSMin:
		return "S"
	case score >= bandAMin:
		return "A"
	case score >= bandBMin:
		return "B"
	default:
		return "C"
	}
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}

func linearRaw(count int, p scoringParams, unitDefault float64) float64 {
	return float64(count) * p.num(unitScoreKey, unitDefault) * p.difficulty()
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}
