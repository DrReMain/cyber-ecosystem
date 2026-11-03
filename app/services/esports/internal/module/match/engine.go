package match

import (
	"hash/fnv"
	"strconv"

	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
	"cyber-ecosystem/app/services/esports/internal/module/game"
	"cyber-ecosystem/app/services/esports/internal/module/test"
)

const (
	genreFPS          = "fps"
	genreMOBA         = "moba"
	universalCategory = "universal"

	// Draw weights in half units, integer arithmetic end to end: a
	// preference hit weighs ×1.5 (3 halves) against the plain 1.0
	// (2 halves), so no float boundary can ever drift a pick.
	weightPlain      = 2
	weightPreference = 3
)

func arrange(order []string, tests []*test.Test, genre string, seed int64) []types.MatchStep {
	steps := make([]types.MatchStep, 0, len(order))
	for _, ability := range order {
		pool := make([]*test.Test, 0, len(tests))
		weights := make([]int, 0, len(tests))
		for _, t := range tests {
			// Pool law: primary-attribution tests whose category is the
			// genre or universal — unattributed tests never enter any pool.
			// Pool order rides the catalog's stable created_at listing, so
			// a seed addresses a stable position.
			if t.Ability == nil || *t.Ability != ability || t.Category == nil {
				continue
			}
			if *t.Category != genre && *t.Category != universalCategory {
				continue
			}
			pool = append(pool, t)
			// Preference weighting lives in the draw layer only: the
			// preference's whole footprint in this vocabulary is the
			// leaning type its entries voted in, so a test of the
			// session's own genre is the hit and draws at ×1.5 while a
			// universal test stays at 1.0. The weight bends which test a
			// slot draws, never which abilities run — coverage stays
			// match_config's word alone.
			if *t.Category == genre {
				weights = append(weights, weightPreference)
			} else {
				weights = append(weights, weightPlain)
			}
		}
		if len(pool) == 0 {
			// An undrawable slot drops out of the lineup rather than
			// blocking the session behind it.
			continue
		}
		steps = append(steps, types.MatchStep{Ability: ability, TestCode: pool[pickIndex(seed, ability, weights)].Code})
	}
	return steps
}

func majorityGenre(entries []types.GameEntry, games []*game.Game) (string, bool) {
	category := make(map[string]string, len(games))
	for _, g := range games {
		if g.Category != nil {
			category[g.Code] = *g.Category
		}
	}
	var fps, moba int
	for _, e := range entries {
		// Custom-typed names carry no category and never vote.
		if e.GameID == "" {
			continue
		}
		switch category[e.GameID] {
		case genreFPS:
			fps++
		case genreMOBA:
			moba++
		}
	}
	switch {
	case fps == 0 && moba == 0, fps == moba:
		// No classified picks or a tie hands the choice to the pick page;
		// the vote only counts the two seeded genres — other categories
		// have no session law yet.
		return "", false
	case fps > moba:
		return genreFPS, true
	default:
		return genreMOBA, true
	}
}

func drawSeed(userID, day, genre string, salt int) int64 {
	h := fnv.New64a()
	for _, part := range []string{userID, day, genre, strconv.Itoa(salt)} {
		_, _ = h.Write([]byte(part))
		// Delimiters keep adjacent empty parts from collapsing into a
		// shared prefix ("a"+"bc" == "ab"+"c").
		_, _ = h.Write([]byte{'|'})
	}
	// The top bit shifts off: storage is a signed bigint and the driver
	// refuses values past the int64 ceiling, so the seed is born in the
	// int63 domain — one bit of entropy cheaper, still uniform enough.
	return utils.ConvNum[int64](h.Sum64() >> 1)
}

func pickIndex(seed int64, ability string, weights []int) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(ability))
	// SplitMix64 over the mixed seed: a fixed integer mixer, stable across
	// Go releases where math/rand's stream is not — a session drawn under
	// one build replays identically under the next.
	z := utils.ConvNum[uint64](seed) ^ h.Sum64()
	z += 0x9e3779b97f4a7c15
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	z ^= z >> 31
	total := 0
	for _, w := range weights {
		total += w
	}
	// Weighted walk: every pool member owns a span of [0, total)
	// proportional to its weight, and the draw lands in the first span
	// that covers it. x is always below total, so the walk returns before
	// the tail.
	x := utils.ConvNum[int](z % utils.ConvNum[uint64](total))
	for i, w := range weights {
		if x < w {
			return i
		}
		x -= w
	}
	return len(weights) - 1
}
