package test

import (
	"strings"
	"testing"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// The band law under test: metric.md §二 top 5/15/40 map onto the C-end
// four-band vocabulary through the quantile-frozen score scale.
func TestBandEdges(t *testing.T) {
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
	for _, c := range cases {
		if got := band(c.score); got != c.want {
			t.Errorf("band(%v) = %q, want %q", c.score, got, c.want)
		}
	}
}

// The plugin contract under test: registered scorers are reachable by code,
// unimplemented games are an honest miss, and a duplicate code fails fast.
func TestScorerRegistryLookup(t *testing.T) {
	if _, ok := lookupScorer("red_dot"); !ok {
		t.Error("red_dot scorer not registered")
	}
	if _, ok := lookupScorer("flash_aiming"); !ok {
		t.Error("flash_aiming scorer not registered")
	}
	if _, ok := lookupScorer("economic_sim"); ok {
		t.Error("economic_sim has no scorer yet; lookup must miss")
	}
}

type dupScorer struct{}

func (dupScorer) testCode() string { return "red_dot" }

func (dupScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	return scoreResult{}, nil
}

type dupScorerTwo struct{}

func (dupScorerTwo) testCode() string { return "red_dot" }

func (dupScorerTwo) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	return scoreResult{}, nil
}

func TestScorerRegistryDuplicateFailsFast(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("duplicate registration must panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "duplicate") {
			t.Fatalf("panic message %v lacks the duplicate marker", r)
		}
	}()
	buildScorerIndex([]scorer{redDotScorer{}, dupScorer{}, dupScorerTwo{}})
}
