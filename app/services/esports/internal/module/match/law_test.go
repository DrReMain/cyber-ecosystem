package match

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
)

// The session law under test (Q1-②): Resolve binds an active session at
// any age — the test domain's submit seam must not invent an expiry the
// biz layer no longer has. The locked lineup replays verbatim.
func TestLawResumesAtAnyAge(t *testing.T) {
	ancient := &MatchSession{
		ID: "s1", UserID: "u1", Genre: "fps", Difficulty: "experience",
		Steps:        []types.MatchStep{{Ability: "reactionSpeed", TestCode: "red_dot"}},
		RetestBudget: 3, Status: statusActive,
		CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
	}
	rp := &fakeMatchRP{byID: map[string]*MatchSession{"s1": ancient}}
	law := NewSessionLaw(slog.Default(), rp)
	binding, err := law.Resolve(context.Background(), "u1", "s1")
	if err != nil {
		t.Fatal(err)
	}
	if binding.RetestBudget != 3 || len(binding.Steps) != 1 || binding.Steps[0].TestCode != "red_dot" {
		t.Fatalf("binding = %+v, want the ancient session's own lineup", binding)
	}
}
