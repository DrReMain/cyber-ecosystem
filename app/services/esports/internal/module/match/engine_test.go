package match

import (
	"reflect"
	"testing"

	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
	"cyber-ecosystem/app/services/esports/internal/module/game"
	"cyber-ecosystem/app/services/esports/internal/module/test"
)

func seedCatalog() []*test.Test {
	// Category and ability mirror the seeded catalog verbatim — the pool
	// law's real input shape.
	mk := func(code, category, ability string) *test.Test {
		return &test.Test{Code: code, Category: utils.Ptr(category), Ability: utils.Ptr(ability)}
	}
	return []*test.Test{
		mk("red_dot", "universal", "reactionSpeed"),
		mk("number_stars", "universal", "reactionSpeed"),
		mk("dual_zone", "universal", "reactionSpeed"),
		mk("target_recognition", "universal", "reactionSpeed"),
		mk("flash_aiming", "fps", "aimingAccuracy"),
		mk("multi_target", "fps", "aimingAccuracy"),
		mk("moving_target", "fps", "trackingStability"),
		mk("tracking_moving", "fps", "trackingStability"),
		mk("skill_trajectory", "universal", "skillReleaseAccuracy"),
		mk("offense_defense", "moba", "decisionEfficiency"),
		mk("resource_match", "moba", "resourceAllocation"),
		mk("character_counter", "moba", "resourceAllocation"),
		mk("economic_sim", "moba", "resourceAllocation"),
	}
}

var fpsOrder = []string{"reactionSpeed", "aimingAccuracy", "trackingStability", "skillReleaseAccuracy"}
var mobaOrder = []string{"resourceAllocation", "decisionEfficiency", "reactionSpeed", "skillReleaseAccuracy"}

func TestArrangePoolLaw(t *testing.T) {
	// A nil-ability row rides along to prove unattributed tests never
	// enter any pool.
	catalog := append(seedCatalog(), &test.Test{Code: "phantom", Category: utils.Ptr("fps")})
	pools := map[string]map[string]bool{
		"reactionSpeed":        {"red_dot": true, "number_stars": true, "dual_zone": true, "target_recognition": true},
		"aimingAccuracy":       {"flash_aiming": true, "multi_target": true},
		"trackingStability":    {"moving_target": true, "tracking_moving": true},
		"skillReleaseAccuracy": {"skill_trajectory": true},
		"resourceAllocation":   {"resource_match": true, "character_counter": true, "economic_sim": true},
		"decisionEfficiency":   {"offense_defense": true},
	}
	for _, genre := range []string{"fps", "moba"} {
		order := fpsOrder
		if genre == "moba" {
			order = mobaOrder
		}
		steps := arrange(order, catalog, genre, 1)
		if len(steps) != len(order) {
			t.Fatalf("%s: %d steps, want %d", genre, len(steps), len(order))
		}
		for i, st := range steps {
			if st.Ability != order[i] {
				t.Errorf("%s step %d: ability %q, want %q (render_order fidelity)", genre, i, st.Ability, order[i])
			}
			if !pools[st.Ability][st.TestCode] {
				t.Errorf("%s step %d: %q outside its ability pool", genre, i, st.TestCode)
			}
		}
	}
	// A slot no test can fill drops out instead of blocking the lineup.
	withGhost := append([]string{"nonexistentAbility"}, fpsOrder...)
	steps := arrange(withGhost, catalog, "fps", 1)
	if len(steps) != len(fpsOrder) {
		t.Errorf("ghost slot leaked %d steps, want %d", len(steps), len(fpsOrder))
	}
}

func TestArrangeDeterminismAndJitter(t *testing.T) {
	catalog := seedCatalog()
	a := arrange(fpsOrder, catalog, "fps", 42)
	b := arrange(fpsOrder, catalog, "fps", 42)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("same seed redrew differently: %v vs %v", a, b)
	}
	// The four-wide reactionSpeed pool must actually jitter across seeds —
	// a draw that never moves would make the anti-memorization law dead.
	rsPicks := map[string]bool{}
	for seed := int64(0); seed < 8; seed++ {
		steps := arrange(fpsOrder, catalog, "fps", seed)
		rsPicks[steps[0].TestCode] = true
		// Forced pools stay invariant: one drawable SRA test means every
		// seed lands on it.
		if steps[3].TestCode != "skill_trajectory" {
			t.Errorf("seed %d: forced SRA pick %q, want skill_trajectory", seed, steps[3].TestCode)
		}
	}
	if len(rsPicks) < 2 {
		t.Errorf("reactionSpeed pool never jittered across 8 seeds: %v", rsPicks)
	}
}

func TestMajorityGenre(t *testing.T) {
	games := []*game.Game{
		{Code: "lol", Category: utils.Ptr("moba")},
		{Code: "csgo", Category: utils.Ptr("fps")},
		{Code: "apex", Category: utils.Ptr("fps")},
		{Code: "nocat", Category: nil},
	}
	cases := []struct {
		name    string
		entries []types.GameEntry
		want    string
		ok      bool
	}{
		{"fps majority", []types.GameEntry{{GameID: "lol"}, {GameID: "csgo"}, {GameID: "apex"}}, "fps", true},
		{"moba single vote", []types.GameEntry{{GameID: "lol"}}, "moba", true},
		{"tie routes to pick page", []types.GameEntry{{GameID: "lol"}, {GameID: "csgo"}}, "", false},
		{"no picks", nil, "", false},
		{"custom names never vote", []types.GameEntry{{CustomName: "whatever"}}, "", false},
		{"uncataloged id never votes", []types.GameEntry{{GameID: "zzz"}}, "", false},
		{"categoryless game never votes", []types.GameEntry{{GameID: "nocat"}}, "", false},
	}
	for _, tc := range cases {
		got, ok := majorityGenre(tc.entries, games)
		if got != tc.want || ok != tc.ok {
			t.Errorf("%s: got (%q, %v), want (%q, %v)", tc.name, got, ok, tc.want, tc.ok)
		}
	}
}

func TestDrawSeedLaws(t *testing.T) {
	base := drawSeed("u1", "20261023", "fps", 0)
	if again := drawSeed("u1", "20261023", "fps", 0); again != base {
		t.Errorf("same inputs redrew: %d vs %d", again, base)
	}
	flips := map[string]int64{
		"day":    drawSeed("u1", "20261024", "fps", 0),
		"salt":   drawSeed("u1", "20261023", "fps", 1),
		"genre":  drawSeed("u1", "20261023", "moba", 0),
		"entity": drawSeed("u2", "20261023", "fps", 0),
	}
	for name, v := range flips {
		if v == base {
			t.Errorf("%s flip produced the same seed", name)
		}
	}
}
