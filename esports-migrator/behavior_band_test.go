package main

import "testing"

// The grade tiers mirror the live scoring path's frozen percentile
// cut-offs (top 5/15/40); the boundaries are the contract — a drifted
// threshold here means migrated rows break the store's grade-from-score
// invariant.
func TestBandForBoundaries(t *testing.T) {
	cases := []struct {
		score float64
		want  string
	}{
		{100, "S"},
		{95, "S"},
		{94.9, "A"},
		{85, "A"},
		{84.9, "B"},
		{60, "B"},
		{59.9, "C"},
		{0, "C"},
	}
	for _, tc := range cases {
		if got := bandFor(tc.score); got != tc.want {
			t.Errorf("bandFor(%v) = %q, want %q", tc.score, got, tc.want)
		}
	}
}
