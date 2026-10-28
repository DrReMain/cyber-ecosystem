package test

import (
	"errors"
	"testing"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

func mtParams() scoringParams {
	return scoringParams{durationSec: 30}
}

// mtStream: a full 30s session at 20Hz — pattern blocks rotate every 100
// samples (five seconds each) exactly as the legacy rotation does.
func mtStream(deviations []float64) []*esportspb.MovingTargetShot {
	shots := make([]*esportspb.MovingTargetShot, 0, len(deviations))
	for i, dev := range deviations {
		shots = append(shots, &esportspb.MovingTargetShot{
			Pattern:   movingTargetPatterns[i/100],
			ElapsedMs: int32(i * 50),
			Deviation: dev,
		})
	}
	return shots
}

func mtRequest(shots []*esportspb.MovingTargetShot) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_MovingTarget{
			MovingTarget: &esportspb.MovingTargetDetail{Shots: shots},
		},
	}
}

func TestMovingTargetPerfectGolden(t *testing.T) {
	// Six hundred zero deviations: the zero-value slice plays a perfect run.
	res, err := movingTargetScorer{}.recompute(mtRequest(mtStream(make([]float64, 600))), mtParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 30s perfect at zero deviation: coverage 3 + perfect 4 + stability 5 +
	// adaptation 10 = 22 clamped to 10.
	if res.extras[extraComposite] != 10.0 {
		t.Errorf("composite = %v, want 10.0", res.extras[extraComposite])
	}
	// 30 perfect seconds × unit 10 = 300.
	if res.raw != 300 {
		t.Errorf("raw = %v, want 300", res.raw)
	}
	if res.authoritative != nil {
		t.Errorf("authoritative = %v, want nil (curve pending)", *res.authoritative)
	}
}

func TestMovingTargetMixedGolden(t *testing.T) {
	deviations := make([]float64, 0, 600)
	deviations = append(deviations, repeatDev(100, 10)...)  // perfect block
	deviations = append(deviations, repeatDev(100, 25)...)  // good-only block
	deviations = append(deviations, repeatDev(400, 120)...) // lost block
	res, err := movingTargetScorer{}.recompute(mtRequest(mtStream(deviations)), mtParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 5s perfect, 10s good, stddev 48.51: coverage 1 + perfect 0.667 +
	// stability 0.149 + adaptation 3.333 − one loss transition 0.3 = 4.8.
	if res.extras[extraComposite] != 4.8 {
		t.Errorf("composite = %v, want 4.8", res.extras[extraComposite])
	}
	if res.raw != 50 {
		t.Errorf("raw = %v, want 50", res.raw)
	}
}

func TestMovingTargetUnitScoreAndDifficultyOverride(t *testing.T) {
	deviations := append(append([]float64{}, repeatDev(100, 10)...), repeatDev(500, 120)...)
	p := mtParams()
	p.difficultyFactor = 1.5
	p.engineParams = map[string]string{unitScoreKey: "15"}
	res, err := movingTargetScorer{}.recompute(mtRequest(mtStream(deviations)), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 5 perfect seconds × overridden unit 15 × difficulty 1.5 = 112.5.
	if res.raw != 112.5 {
		t.Errorf("raw = %v, want 112.5", res.raw)
	}
}

func TestMovingTargetHonestZero(t *testing.T) {
	res, err := movingTargetScorer{}.recompute(mtRequest(nil), mtParams())
	if err != nil {
		t.Fatalf("empty stream error: %v", err)
	}
	if res.raw != 0 || res.extras[extraComposite] != 0 {
		t.Errorf("empty stream: raw = %v composite = %v, want 0/0", res.raw, res.extras[extraComposite])
	}
}

func TestMovingTargetRejections(t *testing.T) {
	dense := make([]*esportspb.MovingTargetShot, 0, 600)
	for i := 0; i < 600; i++ {
		dense = append(dense, &esportspb.MovingTargetShot{
			Pattern: "straight", ElapsedMs: int32(i * 5), Deviation: 10,
		})
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
			name: "unknown pattern",
			req: mtRequest([]*esportspb.MovingTargetShot{
				{Pattern: "spiral", ElapsedMs: 0, Deviation: 10},
			}),
			reason: reasonDetailMismatch,
		},
		{
			name: "negative deviation",
			req: mtRequest([]*esportspb.MovingTargetShot{
				{Pattern: "straight", ElapsedMs: 0, Deviation: -1},
			}),
			reason: reasonDetailMismatch,
		},
		{
			name: "timestamps out of order",
			req: mtRequest([]*esportspb.MovingTargetShot{
				{Pattern: "straight", ElapsedMs: 5000, Deviation: 10},
				{Pattern: "straight", ElapsedMs: 4000, Deviation: 10},
			}),
			reason: reasonDetailMismatch,
		},
		{
			name: "span past the config duration",
			req: mtRequest([]*esportspb.MovingTargetShot{
				{Pattern: "straight", ElapsedMs: 33000, Deviation: 10},
			}),
			reason: reasonAbsurdTimeSpan,
		},
		{
			name:   "counts minted past the span",
			req:    mtRequest(dense),
			reason: reasonAbsurdTimeSpan,
		},
		{
			name: "late-starting dense burst",
			req: mtRequest(func() []*esportspb.MovingTargetShot {
				late := make([]*esportspb.MovingTargetShot, 0, 100)
				for i := 0; i < 100; i++ {
					late = append(late, &esportspb.MovingTargetShot{
						Pattern: "straight", ElapsedMs: int32(28000 + i*20), Deviation: 10,
					})
				}
				return late
			}()),
			reason: reasonAbsurdTimeSpan,
		},
	}
	for _, tc := range cases {
		_, err := movingTargetScorer{}.recompute(tc.req, mtParams())
		var invalid *invalidAttemptError
		if !errors.As(err, &invalid) || invalid.reason != tc.reason {
			t.Errorf("%s: err = %v, want %s", tc.name, err, tc.reason)
		}
	}
}

func repeatDev(n int, dev float64) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = dev
	}
	return out
}
