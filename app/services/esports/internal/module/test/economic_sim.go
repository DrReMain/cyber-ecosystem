package test

import (
	"math"
	"slices"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

const (
	economicSimPerfectWindowMeleeDefault  = 5.0
	economicSimPerfectWindowCannonDefault = 8.0
	economicSimGoodWindowMeleeDefault     = 10.0
	economicSimGoodWindowCannonDefault    = 15.0
	economicSimGoldMeleeDefault           = 21.0
	economicSimGoldCannonDefault          = 90.0
	economicSimGoodShare                  = 0.6
	economicSimEarlyShare                 = 0.3
	economicSimEarlyHealthCap             = 30.0
	economicSimPerfectTimeBonusSec        = 1.0
	economicSimHealthWireScale            = 1500.0
	economicSimHealthStep                 = 15.0
	economicSimItemsMaxDefault            = 3.0
	economicSimUnitScoreDefault           = 15.0
)

func economicSimWindow(health, perfectPx, goodPx float64) string {
	// The four tap windows the minion's remaining health lands in; the
	// first two count as confirmed last hits.
	switch {
	case health <= perfectPx:
		return "perfect"
	case health <= goodPx:
		return "good"
	case health <= economicSimEarlyHealthCap:
		return "early"
	default:
		return "too_early"
	}
}

func economicSimGold(window string, base float64) float64 {
	switch window {
	case "perfect":
		return base
	case "good":
		return math.Floor(base * economicSimGoodShare)
	case "early":
		return math.Floor(base * economicSimEarlyShare)
	default:
		return 0
	}
}

type economicSimScorer struct{}

func init() {
	scorers = append(scorers, economicSimScorer{})
}

func (economicSimScorer) testCode() string {
	return "economic_sim"
}

func (economicSimScorer) recompute(req *esportspb.SubmitTestAttemptRequest, params scoringParams) (scoreResult, error) {
	detail := req.GetEconomicSim()
	if detail == nil || detail.FinalHealthHp == nil {
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	durationSec := params.effectiveDurationSec()
	perfectMelee := params.num("perfect_window_melee", economicSimPerfectWindowMeleeDefault)
	perfectCannon := params.num("perfect_window_cannon", economicSimPerfectWindowCannonDefault)
	goodMelee := params.num("good_window_melee", economicSimGoodWindowMeleeDefault)
	goodCannon := params.num("good_window_cannon", economicSimGoodWindowCannonDefault)
	goldMelee := params.num("gold_melee", economicSimGoldMeleeDefault)
	goldCannon := params.num("gold_cannon", economicSimGoldCannonDefault)
	tiers := [3]float64{
		params.num("item_cost_1", 400),
		params.num("item_cost_2", 800),
		params.num("item_cost_3", 1200),
	}
	itemsMax := params.num("max_items", economicSimItemsMaxDefault)
	hits := detail.GetLastHits()
	purchases := detail.GetPurchases()
	if float64(len(purchases)) > itemsMax {
		// The shop ladder carries at most three items in a session.
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	cs, perfectCS, earlyHits := 0, 0, 0
	prevHitMs := float64(-1)
	for _, h := range hits {
		minionType := h.GetMinionType()
		cannon := minionType == "cannon"
		if minionType != "melee" && minionType != "ranged" && !cannon {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		perfectPx, goodPx, base := perfectMelee, goodMelee, goldMelee
		if cannon {
			perfectPx, goodPx, base = perfectCannon, goodCannon, goldCannon
		}
		health := float64(h.GetMinionHealth())
		window := economicSimWindow(health, perfectPx, goodPx)
		// The verdict and the payout are both recomputed from the reported
		// health — the client's flags are cross-checked, never trusted.
		if confirmed := window == "perfect" || window == "good"; confirmed != h.GetIsLastHit() {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		if gold := economicSimGold(window, base); gold != float64(h.GetGoldEarned()) {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		elapsed := float64(h.GetElapsedMs())
		if prevHitMs >= 0 && elapsed <= prevHitMs {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		prevHitMs = elapsed
		if h.GetIsLastHit() {
			cs++
			if window == "perfect" {
				perfectCS++
			}
		} else {
			earlyHits++
		}
	}
	prevPurchaseMs := float64(-1)
	for _, p := range purchases {
		cost := float64(p.GetCost())
		if !slices.Contains(tiers[:], cost) {
			// Costs ride the three-rung equipment ladder (the legacy gate's
			// 400/800/1200 set).
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		elapsed := float64(p.GetElapsedMs())
		if prevPurchaseMs >= 0 && elapsed <= prevPurchaseMs {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		prevPurchaseMs = elapsed
	}
	type event struct {
		ms    float64
		delta float64
	}
	// Gold conservation walks both streams on one timeline — a purchase may
	// only land once the taps in front of it have paid for it.
	timeline := make([]event, 0, len(hits)+len(purchases))
	for _, h := range hits {
		timeline = append(timeline, event{ms: float64(h.GetElapsedMs()), delta: float64(h.GetGoldEarned())})
	}
	for _, p := range purchases {
		timeline = append(timeline, event{ms: float64(p.GetElapsedMs()), delta: -float64(p.GetCost())})
	}
	// Stable with taps ahead of purchases at the same millisecond — the
	// player-friendly reading of a same-instant spend (an unstable sort
	// could dip the balance negative purchase-first).
	slices.SortStableFunc(timeline, func(a, b event) int {
		if a.ms != b.ms {
			switch {
			case a.ms < b.ms:
				return -1
			default:
				return 1
			}
		}
		switch {
		case a.delta > b.delta:
			return -1
		case a.delta < b.delta:
			return 1
		default:
			return 0
		}
	})
	balance := 0.0
	spanMs := 0.0
	for _, e := range timeline {
		balance += e.delta
		if balance < 0 {
			return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
		}
		spanMs = e.ms
	}
	// Perfect taps extend the clock a second each — the span ceiling rides
	// the config mirror plus that bonus (a fabrication closure family).
	if spanMs > (durationSec+float64(perfectCS)*economicSimPerfectTimeBonusSec)*1000+spanGraceMs {
		return scoreResult{}, &invalidAttemptError{reason: reasonAbsurdTimeSpan}
	}
	finalHp := float64(detail.GetFinalHealthHp())
	if finalHp != math.Floor(finalHp/economicSimHealthStep)*economicSimHealthStep {
		// The wire scale steps in fifteens (health% × 15) — anything else
		// never left the legacy formula.
		return scoreResult{}, &invalidAttemptError{reason: reasonDetailMismatch}
	}
	csEfficiency := float64(cs) / durationSec * 3
	accuracy := 0.0
	if cs > 0 {
		accuracy = float64(perfectCS) / float64(cs) * 3
	}
	items := float64(len(purchases)) * 1.5
	health := finalHp / economicSimHealthWireScale * 1.5
	penalty := float64(earlyHits)*0.1 + float64(detail.GetMissedMinions())*0.05
	composite := clamp(csEfficiency+accuracy+items+health-penalty, 0, 10)
	// The linear track counts confirmed last hits.
	return curvePendingResult(cs, params, economicSimUnitScoreDefault, composite), nil
}
