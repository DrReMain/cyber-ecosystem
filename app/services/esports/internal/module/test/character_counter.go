package test

import (
	"math"
	"slices"
	"sort"
	"strings"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	characterCounterRoundsDefault    = 3.0
	characterCounterSpeedRefSec      = 8.0
	characterCounterSpeedFloor       = 0.5
	characterCounterSpeedCap         = 1.5
	characterCounterRawCap           = 150.0
	characterCounterUnitScoreDefault = 15.0
)

type characterCounterScorer struct{}

func init() {
	scorers = append(scorers, characterCounterScorer{})
}

func (characterCounterScorer) testCode() string {
	return "character_counter"
}

type characterCounterBank struct {
	counter map[string]float64
	roster  []string
}

func buildCharacterCounterBank(params scoringParams) characterCounterBank {
	// The bank rides the generic question bank as relation rows —
	// scenario_code "<role>|<countersRole>" with answer counter_score — so
	// the roster is the union of roles the relations mention (the legacy
	// matrix covers all six).
	counter := make(map[string]float64, len(params.answerKey))
	roles := make(map[string]bool, len(params.answerKey))
	for key, ans := range params.answerKey {
		l, r, ok := strings.Cut(key, "|")
		if !ok {
			continue
		}
		// Roles join the roster whatever the row scores — a future
		// zero-score relation still names legal lineup words.
		roles[l] = true
		roles[r] = true
		if score := answerNumber(ans, "counter_score"); score > 0 {
			counter[l+"|"+r] = score
		}
	}
	roster := make([]string, 0, len(roles))
	for role := range roles {
		roster = append(roster, role)
	}
	sort.Strings(roster)
	return characterCounterBank{counter: counter, roster: roster}
}

func bestLineupScore(bank characterCounterBank, lineup int, enemy []string) float64 {
	// c_max(E): every lineup-sized roster combination enumerated, best
	// achievable counter score against the enemy — the rework's denominator
	// that replaced the legacy flat 100.
	best := 0.0
	var walk func(start, depth int, score float64)
	walk = func(start, depth int, score float64) {
		if depth == lineup {
			best = math.Max(best, score)
			return
		}
		for j := start; j <= len(bank.roster)-(lineup-depth); j++ {
			s := score
			for _, e := range enemy {
				s += bank.counter[bank.roster[j]+"|"+e]
			}
			walk(j+1, depth+1, s)
		}
	}
	walk(0, 0, 0)
	return best
}

func (characterCounterScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetCharacterCounter()
	if detail == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	rounds := detail.GetRounds()
	if len(rounds) == 0 {
		// Honest zero: an attempt with no rounds earns nothing (the legacy
		// scorer returned ZERO for the empty list).
		return curvePendingResult(0, params, characterCounterUnitScoreDefault, 0), nil
	}
	roundCount := int(params.num("round_count", characterCounterRoundsDefault))
	if len(rounds) != roundCount {
		// The round count is the contract — the legacy formula divided by a
		// hardcoded 3 whatever arrived, so a trimmed history self-lowered
		// instead of failing; here it hard-rejects.
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	bank := buildCharacterCounterBank(params)
	durationSec := float64(params.durationSec)
	if durationSec <= 0 {
		durationSec = sessionDurationFallbackSec
	}
	hits, qSum, spanMs := 0, 0.0, 0.0
	for i, rd := range rounds {
		if rd.GetRoundNumber() != int32(i+1) {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		enemy, sel := rd.GetEnemyLineup(), rd.GetSelectedLineup()
		if len(enemy) == 0 || len(sel) != len(enemy) {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		// Mirrored lineups of distinct roster words: overlap across sides is
		// the game's law (five of six roles on both teams), repeats inside
		// one lineup are an edited history.
		seenEnemy := make(map[string]bool, len(enemy))
		seenSel := make(map[string]bool, len(sel))
		for _, w := range enemy {
			if seenEnemy[w] {
				return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
			}
			seenEnemy[w] = true
		}
		for _, w := range sel {
			if seenSel[w] {
				return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
			}
			seenSel[w] = true
		}
		for _, w := range enemy {
			if !slices.Contains(bank.roster, w) {
				// Lineup words are judged against the bank's roster — an
				// unknown role is an edited history, not a degraded client
				// (the legacy fallback treated the raw string as a role).
				return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
			}
		}
		for _, w := range sel {
			if !slices.Contains(bank.roster, w) {
				return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
			}
		}
		cRound := 0.0
		for _, s := range sel {
			for _, e := range enemy {
				if score := bank.counter[s+"|"+e]; score > 0 {
					cRound += score
					hits++
				}
			}
		}
		cMax := bestLineupScore(bank, len(enemy), enemy)
		if cMax <= 0 {
			// A degenerate bank where nothing counters the lineup would
			// divide by zero; the round then carries no signal either way.
			cMax = 1
		}
		qSum += cRound / cMax
		elapsed := float64(rd.GetTimeSpentMs())
		if elapsed < reactionTimeMin {
			return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdReactionTime}
		}
		spanMs += elapsed
	}
	if spanMs > durationSec*1000+spanGraceMs {
		// Rounds share one session clock — their sum is bounded by the
		// config mirror (a fabrication closure).
		return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdTimeSpan}
	}
	avgSec := spanMs / 1000 / float64(len(rounds))
	speed := clamp(params.num("speed_ref_sec", characterCounterSpeedRefSec)/avgSec, characterCounterSpeedFloor, characterCounterSpeedCap)
	q := qSum / float64(len(rounds))
	// The rework scale rides raw capped at 150 — the one curve input outside
	// the family's 0-10 habit, because the 1.5x speed headroom above a
	// perfect counter is discrimination-bearing (the 96.3%-saturation cure).
	composite := math.Min(q*100*speed, characterCounterRawCap)
	return curvePendingResult(hits, params, characterCounterUnitScoreDefault, round1(composite)), nil
}
