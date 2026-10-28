package test

import (
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

func rmCollect(kind string, ms int32) *esportspb.ResourceMatchCollect {
	return &esportspb.ResourceMatchCollect{Kind: kind, ElapsedMs: ms}
}

func rmCollision(ms int32) *esportspb.ResourceMatchCollision {
	return &esportspb.ResourceMatchCollision{ElapsedMs: ms}
}

func rmRequest(collects []*esportspb.ResourceMatchCollect, collisions []*esportspb.ResourceMatchCollision) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_ResourceMatch{
			ResourceMatch: &esportspb.ResourceMatchDetail{Collects: collects, Collisions: collisions},
		},
	}
}

// rmMixed: a session collecting nMin minions, nMon monsters and nDrag dragons
// at even spacing — kind blocks in field order, timestamps strictly rising.
func rmMixed(nMin, nMon, nDrag, stepMs int) []*esportspb.ResourceMatchCollect {
	total := nMin + nMon + nDrag
	collects := make([]*esportspb.ResourceMatchCollect, 0, total)
	for i := 0; i < total; i++ {
		kind := "dragon"
		if i < nMin {
			kind = "minion"
		} else if i < nMin+nMon {
			kind = "monster"
		}
		collects = append(collects, rmCollect(kind, utils.ConvNum[int32](i*stepMs)))
	}
	return collects
}

func TestResourceMatchPerfectClearGolden(t *testing.T) {
	// All 42 spawns collected across 20.5s: efficiency saturates the path
	// term, the economy is clean, composite caps at 10.
	res, err := resourceMatchScorer{}.recompute(
		rmRequest(rmMixed(25, 12, 5, 500), nil), scoringParams{durationSec: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.extras[extraComposite] != 10.0 {
		t.Errorf("composite = %v, want 10.0", res.extras[extraComposite])
	}
	// 42 resources × unit 20 = 840.
	if res.raw != 840 {
		t.Errorf("raw = %v, want 840", res.raw)
	}
	if res.authoritative != nil {
		t.Errorf("authoritative = %v, want nil (curve pending)", *res.authoritative)
	}
}

func TestResourceMatchMidPaceGolden(t *testing.T) {
	collects := rmMixed(10, 4, 1, 2000)
	// (i+1)*2000 shifts the stream to 2000..30000 so the span lands exactly
	// on the config duration.
	for _, c := range collects {
		c.ElapsedMs += 2000
	}
	res, err := resourceMatchScorer{}.recompute(
		rmRequest(collects, []*esportspb.ResourceMatchCollision{rmCollision(5000), rmCollision(15000), rmCollision(25000)}),
		scoringParams{durationSec: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 15 collected over 30s with 3 collisions: 15/42×4 + 0.25×3 + 1.7 + 1
	// = 4.878… → 4.9.
	if res.extras[extraComposite] != 4.9 {
		t.Errorf("composite = %v, want 4.9", res.extras[extraComposite])
	}
	if res.raw != 300 {
		t.Errorf("raw = %v, want 300", res.raw)
	}
}

func TestResourceMatchUnitScoreAndDifficultyOverride(t *testing.T) {
	collects := rmMixed(10, 4, 1, 2000)
	for _, c := range collects {
		c.ElapsedMs += 2000
	}
	res, err := resourceMatchScorer{}.recompute(
		rmRequest(collects, []*esportspb.ResourceMatchCollision{rmCollision(5000), rmCollision(15000), rmCollision(25000)}),
		scoringParams{
			durationSec:      30,
			difficultyFactor: 1.5,
			engineParams:     map[string]string{unitScoreKey: "25"},
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 15 collected × overridden unit 25 × difficulty 1.5 = 562.5.
	if res.raw != 562.5 {
		t.Errorf("raw = %v, want 562.5", res.raw)
	}
	if res.extras[extraComposite] != 4.9 {
		t.Errorf("composite = %v, want 4.9", res.extras[extraComposite])
	}
}

func TestResourceMatchHonestZero(t *testing.T) {
	res, err := resourceMatchScorer{}.recompute(rmRequest(nil, nil), scoringParams{durationSec: 30})
	if err != nil {
		t.Fatalf("empty streams error: %v", err)
	}
	// Raw earns nothing; the composite keeps the Java parity floor — risk
	// 2 plus tactical 1 for surviving without a single collect.
	if res.raw != 0 || res.extras[extraComposite] != 3.0 {
		t.Errorf("empty streams: raw = %v composite = %v, want 0/3.0", res.raw, res.extras[extraComposite])
	}
}

func TestResourceMatchRejections(t *testing.T) {
	overSpawn := make([]*esportspb.ResourceMatchCollect, 0, 26)
	for i := 0; i < 26; i++ {
		overSpawn = append(overSpawn, rmCollect("minion", utils.ConvNum[int32](i*100)))
	}
	overCap := make([]*esportspb.ResourceMatchCollision, 0, 11)
	for i := 0; i < 11; i++ {
		overCap = append(overCap, rmCollision(utils.ConvNum[int32](i*1500)))
	}
	cases := []struct {
		name   string
		req    *esportspb.SubmitTestAttemptRequest
		reason string
	}{
		{
			name:   "detail mismatch",
			req:    &esportspb.SubmitTestAttemptRequest{RawDetail: &esportspb.SubmitTestAttemptRequest_NumberStars{}},
			reason: reasonDetailMismatch,
		},
		{
			name:   "unknown kind",
			req:    rmRequest([]*esportspb.ResourceMatchCollect{rmCollect("buff", 1000)}, nil),
			reason: reasonDetailMismatch,
		},
		{
			name:   "collect past the spawn cap",
			req:    rmRequest(overSpawn, nil),
			reason: reasonDetailMismatch,
		},
		{
			name:   "collects out of order",
			req:    rmRequest([]*esportspb.ResourceMatchCollect{rmCollect("minion", 5000), rmCollect("minion", 4000)}, nil),
			reason: reasonDetailMismatch,
		},
		{
			name:   "collect gap under the floor",
			req:    rmRequest([]*esportspb.ResourceMatchCollect{rmCollect("minion", 1000), rmCollect("minion", 1060)}, nil),
			reason: reasonAbsurdReactionTime,
		},
		{
			name:   "span past the config duration",
			req:    rmRequest([]*esportspb.ResourceMatchCollect{rmCollect("minion", 33000)}, nil),
			reason: reasonAbsurdTimeSpan,
		},
		{
			name:   "collisions past the death cap",
			req:    rmRequest(nil, overCap),
			reason: reasonDetailMismatch,
		},
		{
			name:   "collisions out of order",
			req:    rmRequest(nil, []*esportspb.ResourceMatchCollision{rmCollision(9000), rmCollision(8000)}),
			reason: reasonDetailMismatch,
		},
	}
	for _, tc := range cases {
		_, err := resourceMatchScorer{}.recompute(tc.req, scoringParams{durationSec: 30})
		var invalid *invalidAttemptError
		if !errors.As(err, &invalid) || invalid.reason != tc.reason {
			t.Errorf("%s: err = %v, want %s", tc.name, err, tc.reason)
		}
	}
}
