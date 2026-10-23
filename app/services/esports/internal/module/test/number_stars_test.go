package test

import (
	"errors"
	"testing"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

func nsClick(number, elapsedMs int32) *esportspb.NumberStarsClick {
	return &esportspb.NumberStarsClick{Number: number, ElapsedMs: elapsedMs}
}

func nsRound(roundNumber int32, numbers []int32, clicks ...*esportspb.NumberStarsClick) *esportspb.NumberStarsRound {
	return &esportspb.NumberStarsRound{
		RoundNumber: roundNumber,
		Numbers:     numbers,
		Clicks:      clicks,
	}
}

func nsRequest(rounds ...*esportspb.NumberStarsRound) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_NumberStars{
			NumberStars: &esportspb.NumberStarsDetail{Rounds: rounds},
		},
	}
}

// fiveAscending: one clean round tapped in order at the given cadence.
func fiveAscending(cadenceMs int32) []*esportspb.NumberStarsClick {
	clicks := make([]*esportspb.NumberStarsClick, 0, 5)
	for i := 1; i <= 5; i++ {
		clicks = append(clicks, nsClick(int32(i), int32(i)*cadenceMs))
	}
	return clicks
}

func TestNumberStarsFastPaceGolden(t *testing.T) {
	rounds := make([]*esportspb.NumberStarsRound, 0, 3)
	for i := 1; i <= 3; i++ {
		rounds = append(rounds, &esportspb.NumberStarsRound{
			RoundNumber: int32(i),
			Numbers:     []int32{2, 4, 1, 3, 5},
			Clicks:      fiveAscending(500),
		})
	}
	res, err := numberStarsScorer{}.recompute(nsRequest(rounds...), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 3 rounds at 500ms cadence: base 1.875 + speed 7.5 + accuracy 3 → clamp 10.
	if res.extras[extraComposite] != 10 {
		t.Errorf("composite = %v, want 10", res.extras[extraComposite])
	}
	// 15 correct taps × unit 15 = 225 (band mid).
	if res.raw != 225 {
		t.Errorf("raw = %v, want 225", res.raw)
	}
	if res.authoritative != nil {
		t.Errorf("authoritative = %v, want nil (curve pending)", *res.authoritative)
	}
}

func TestNumberStarsMidPaceGolden(t *testing.T) {
	res, err := numberStarsScorer{}.recompute(nsRequest(
		// Round 1: one mistap (8) between the first two hits.
		nsRound(1, []int32{2, 4, 1, 3, 5},
			nsClick(1, 1000), nsClick(8, 1500), nsClick(2, 2300),
			nsClick(3, 3600), nsClick(4, 4900), nsClick(5, 6200)),
		nsRound(2, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
		// Round 3: another mistap (6).
		nsRound(3, []int32{1, 2, 3, 4, 5},
			nsClick(1, 1000), nsClick(6, 1500), nsClick(2, 2300),
			nsClick(3, 3600), nsClick(4, 4900), nsClick(5, 6200)),
		nsRound(4, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
		// Round 5 unfinished at the timer: two hits, one mistap; its target
		// order is 1,3,5,7,9 so the 2 tap misses.
		nsRound(5, []int32{9, 7, 5, 3, 1},
			nsClick(1, 700), nsClick(2, 1400), nsClick(3, 2100)),
	), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 22 correct, 3 incorrect (mistaps 8/6/2), 4 completed rounds, avg
	// interval 22200/17: base 2.5 + speed (2000−1305.88)/200 + accuracy 1.5
	// = 7.4706 → 7.5.
	if res.extras[extraComposite] != 7.5 {
		t.Errorf("composite = %v, want 7.5", res.extras[extraComposite])
	}
	if res.raw != 330 {
		t.Errorf("raw = %v, want 330", res.raw)
	}
}

func TestNumberStarsUnitScoreAndDifficultyOverride(t *testing.T) {
	req := nsRequest(
		nsRound(1, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
		nsRound(2, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
		nsRound(3, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
		nsRound(4, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
		nsRound(5, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
		nsRound(6, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
		nsRound(7, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
		nsRound(8, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
	)
	res, err := numberStarsScorer{}.recompute(req, scoringParams{
		difficultyFactor: 1.5,
		engineParams:     map[string]string{unitScoreKey: "20"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 40 correct × overridden unit 20 × difficulty 1.5 = 1200.
	if res.raw != 1200 {
		t.Errorf("raw = %v, want 1200", res.raw)
	}
	// base caps at 5 (8/8 rounds), speed (2000−1300)/200 = 3.5, accuracy 3.
	if res.extras[extraComposite] != 10 {
		t.Errorf("composite = %v, want 10", res.extras[extraComposite])
	}
}

func TestNumberStarsHonestZero(t *testing.T) {
	// All-mistap round plus the empty submission: both earn nothing.
	res, err := numberStarsScorer{}.recompute(nsRequest(
		nsRound(1, []int32{1, 2, 3, 4, 5},
			nsClick(9, 1000), nsClick(8, 2000), nsClick(7, 3000)),
	), scoringParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.raw != 0 || res.extras[extraComposite] != 0 {
		t.Errorf("all-mistap: raw = %v composite = %v, want 0/0", res.raw, res.extras[extraComposite])
	}

	res, err = numberStarsScorer{}.recompute(nsRequest(), scoringParams{})
	if err != nil {
		t.Fatalf("empty rounds error: %v", err)
	}
	if res.raw != 0 || res.extras[extraComposite] != 0 {
		t.Errorf("empty rounds: raw = %v composite = %v, want 0/0", res.raw, res.extras[extraComposite])
	}
}

func TestNumberStarsRejections(t *testing.T) {
	cases := []struct {
		name string
		req  *esportspb.SubmitTestAttemptRequest
	}{
		{
			name: "detail mismatch",
			req: &esportspb.SubmitTestAttemptRequest{
				RawDetail: &esportspb.SubmitTestAttemptRequest_DualZone{},
			},
		},
		{
			name: "regressed round number",
			req: nsRequest(
				nsRound(1, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
				nsRound(1, []int32{1, 2, 3, 4, 5}, fiveAscending(1300)...),
			),
		},
		{
			name: "backwards stamp within round",
			req: nsRequest(
				nsRound(1, []int32{1, 2, 3, 4, 5},
					nsClick(1, 2000), nsClick(2, 1000)),
			),
		},
		{
			name: "same-frame double tap",
			req: nsRequest(
				nsRound(1, []int32{1, 2, 3, 4, 5},
					nsClick(1, 1000), nsClick(2, 1000)),
			),
		},
		{
			name: "empty target sequence",
			req:  nsRequest(nsRound(1, nil, nsClick(1, 1000))),
		},
	}
	for _, tc := range cases {
		_, err := numberStarsScorer{}.recompute(tc.req, scoringParams{})
		var invalid *invalidAttemptError
		if !errors.As(err, &invalid) || invalid.reason != reasonDetailMismatch {
			t.Errorf("%s: err = %v, want DETAIL_MISMATCH", tc.name, err)
		}
	}
}
