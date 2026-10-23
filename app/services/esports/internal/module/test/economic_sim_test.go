package test

import (
	"errors"
	"testing"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

func esParams() scoringParams {
	return scoringParams{durationSec: 30}
}

// esHit: a tap with its reported facts — gold and hit are spelled out as
// literals so the fixtures never borrow the scorer's own arithmetic.
func esHit(ms int32, minionType string, health int32, hit bool, gold int32) *esportspb.EconomicSimLastHit {
	return &esportspb.EconomicSimLastHit{
		ElapsedMs: ms, MinionType: minionType, MinionHealth: health,
		IsLastHit: hit, GoldEarned: gold,
	}
}

func esBuy(ms, cost int32) *esportspb.EconomicSimPurchase {
	return &esportspb.EconomicSimPurchase{
		ElapsedMs: ms, ItemName: "长剑", Cost: cost,
	}
}

func esRequest(hits []*esportspb.EconomicSimLastHit, buys []*esportspb.EconomicSimPurchase, missed, finalHp int32) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_EconomicSim{
			EconomicSim: &esportspb.EconomicSimDetail{
				LastHits: hits, Purchases: buys,
				MissedMinions: &missed, FinalHealthHp: &finalHp,
			},
		},
	}
}

// esStrongSession: twelve perfect melee (21 each) plus three perfect
// cannons (90 each), one early tap (6), one too-early tap (0), two minions
// missed, one 400 purchase, terminal health 95% → 1425 on the hp wire.
func esStrongSession() *esportspb.SubmitTestAttemptRequest {
	hits := make([]*esportspb.EconomicSimLastHit, 0, 17)
	ms := int32(0)
	for i := 0; i < 12; i++ {
		hits = append(hits, esHit(ms, "melee", 3, true, 21))
		ms += 2500
	}
	for i := 0; i < 3; i++ {
		hits = append(hits, esHit(ms, "cannon", 5, true, 90))
		ms += 2500
	}
	hits = append(hits, esHit(ms, "melee", 25, false, 6))
	ms += 2500
	hits = append(hits, esHit(ms, "melee", 60, false, 0))
	return esRequest(hits, []*esportspb.EconomicSimPurchase{esBuy(ms+1000, 400)}, 2, 1425)
}

func TestEconomicSimStrongGolden(t *testing.T) {
	req := esStrongSession()
	res, err := economicSimScorer{}.recompute(req, esParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// cs 15 over 30s → 1.5; 15/15 perfect → 3; one item → 1.5; 95% health
	// → 1.425; penalty 2 early + 2 missed → 0.3 → 7.125 → 7.1.
	if res.extras[extraComposite] != 7.1 {
		t.Errorf("composite = %v, want 7.1", res.extras[extraComposite])
	}
	if res.raw != 225 {
		t.Errorf("raw = %v, want 225", res.raw)
	}
	if res.authoritative != nil {
		t.Errorf("authoritative = %v, want nil (curve pending)", *res.authoritative)
	}
}

func TestEconomicSimWideGolden(t *testing.T) {
	hits := make([]*esportspb.EconomicSimLastHit, 0, 20)
	for i := 0; i < 20; i++ {
		hits = append(hits, esHit(int32(i*2000), "melee", 3, true, 21))
	}
	req := esRequest(hits, []*esportspb.EconomicSimPurchase{esBuy(39000, 400)}, 0, 1500)
	res, err := economicSimScorer{}.recompute(req, esParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 20/30 cs rate → 2; all perfect → 3; one item → 1.5; full health →
	// 1.5 → 8.0.
	if res.extras[extraComposite] != 8.0 {
		t.Errorf("composite = %v, want 8.0", res.extras[extraComposite])
	}
	if res.raw != 300 {
		t.Errorf("raw = %v, want 300", res.raw)
	}
}

func TestEconomicSimIdleFloor(t *testing.T) {
	res, err := economicSimScorer{}.recompute(esRequest(nil, nil, 0, 1500), esParams())
	if err != nil {
		t.Fatalf("idle session error: %v", err)
	}
	// Never tapping still banks the health term — the legacy idle floor of
	// 1.5 rides for parity, a disclosed carry.
	if res.raw != 0 || res.extras[extraComposite] != 1.5 {
		t.Errorf("idle session: raw = %v composite = %v, want 0/1.5", res.raw, res.extras[extraComposite])
	}
}

func TestEconomicSimUnitScoreAndDifficultyOverride(t *testing.T) {
	req := esStrongSession()
	p := esParams()
	p.difficultyFactor = 1.5
	p.engineParams = map[string]string{unitScoreKey: "20"}
	res, err := economicSimScorer{}.recompute(req, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 15 confirmed last hits × overridden unit 20 × difficulty 1.5 = 450.
	if res.raw != 450 {
		t.Errorf("raw = %v, want 450", res.raw)
	}
}

func TestEconomicSimRejections(t *testing.T) {
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
			name:   "unknown minion type",
			req:    esRequest([]*esportspb.EconomicSimLastHit{esHit(0, "siege", 3, true, 21)}, nil, 0, 1500),
			reason: reasonDetailMismatch,
		},
		{
			name:   "verdict contradicting the health window",
			req:    esRequest([]*esportspb.EconomicSimLastHit{esHit(0, "melee", 3, false, 6)}, nil, 0, 1500),
			reason: reasonDetailMismatch,
		},
		{
			name:   "payout contradicting the window",
			req:    esRequest([]*esportspb.EconomicSimLastHit{esHit(0, "melee", 3, true, 999)}, nil, 0, 1500),
			reason: reasonDetailMismatch,
		},
		{
			name: "purchase before the gold exists",
			req: esRequest([]*esportspb.EconomicSimLastHit{esHit(200, "melee", 3, true, 21)},
				[]*esportspb.EconomicSimPurchase{esBuy(100, 400)}, 0, 1500),
			reason: reasonDetailMismatch,
		},
		{
			name: "cost off the equipment ladder",
			req: esRequest([]*esportspb.EconomicSimLastHit{esHit(0, "melee", 3, true, 21)},
				[]*esportspb.EconomicSimPurchase{esBuy(1000, 500)}, 0, 1500),
			reason: reasonDetailMismatch,
		},
		{
			name: "more purchases than the ladder allows",
			req: esRequest(nil,
				[]*esportspb.EconomicSimPurchase{esBuy(1000, 400), esBuy(2000, 400), esBuy(3000, 400), esBuy(4000, 400)}, 0, 1500),
			reason: reasonDetailMismatch,
		},
		{
			name:   "final health off the wire step",
			req:    esRequest(nil, nil, 0, 1420),
			reason: reasonDetailMismatch,
		},
		{
			name: "hits out of order",
			req: esRequest([]*esportspb.EconomicSimLastHit{
				esHit(5000, "melee", 3, true, 21), esHit(4000, "melee", 3, true, 21),
			}, nil, 0, 1500),
			reason: reasonDetailMismatch,
		},
		{
			name:   "span past the clock bonus ceiling",
			req:    esRequest([]*esportspb.EconomicSimLastHit{esHit(48000, "melee", 60, false, 0)}, nil, 0, 1500),
			reason: reasonAbsurdTimeSpan,
		},
	}
	for _, tc := range cases {
		_, err := economicSimScorer{}.recompute(tc.req, esParams())
		var invalid *invalidAttemptError
		if !errors.As(err, &invalid) || invalid.reason != tc.reason {
			t.Errorf("%s: err = %v, want %s", tc.name, err, tc.reason)
		}
	}
}
