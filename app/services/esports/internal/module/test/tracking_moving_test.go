package test

import (
	"errors"
	"testing"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

func tmParams() scoringParams {
	return scoringParams{durationSec: 30}
}

func tmPoint(x, y float64) *esportspb.Point {
	return &esportspb.Point{X: x, Y: y}
}

func tmMove(ms int32, x, y float64) *esportspb.TrackingMovingFrame {
	return &esportspb.TrackingMovingFrame{
		ElapsedMs: ms,
		Deviation: 0,
		Target:    tmPoint(0, 0),
		Crosshair: tmPoint(x, y),
		Player:    tmPoint(x, y),
	}
}

func tmShot(ms int32, dev float64, hit bool, life int32) *esportspb.TrackingMovingFrame {
	return &esportspb.TrackingMovingFrame{
		ElapsedMs:     ms,
		Deviation:     dev,
		Target:        tmPoint(100, 100),
		Crosshair:     tmPoint(100, 100),
		Player:        tmPoint(100, 100),
		Hit:           &hit,
		DotLifetimeMs: &life,
	}
}

func tmRequest(frames ...*esportspb.TrackingMovingFrame) *esportspb.SubmitTestAttemptRequest {
	return &esportspb.SubmitTestAttemptRequest{
		RawDetail: &esportspb.SubmitTestAttemptRequest_TrackingMoving{
			TrackingMoving: &esportspb.TrackingMovingDetail{Frames: frames},
		},
	}
}

// tmMidSession: ten shots (eight hits at 30px, two misses at 60px, kills at
// 1200ms) over a 800px movement track — the mid golden.
func tmMidSession() []*esportspb.TrackingMovingFrame {
	frames := []*esportspb.TrackingMovingFrame{
		tmMove(0, 0, 0), tmMove(500, 400, 0), tmMove(1000, 800, 0),
	}
	for i := 0; i < 10; i++ {
		if i == 3 || i == 7 {
			frames = append(frames, tmShot(int32(2000+i*1000), 60, false, 1500))
		} else {
			frames = append(frames, tmShot(int32(2000+i*1000), 30, true, 1200))
		}
	}
	return frames
}

func TestTrackingMovingMidGolden(t *testing.T) {
	res, err := trackingMovingScorer{}.recompute(tmRequest(tmMidSession()...), tmParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Accuracy 0.8 → 3.2, stability (40−36)/8 = 0.5, movement 1.6,
	// multitasking (0.32+0.3+0.3)×2 = 1.84 → 7.14 → 7.1.
	if res.extras[extraComposite] != 7.1 {
		t.Errorf("composite = %v, want 7.1", res.extras[extraComposite])
	}
	if res.raw != 80 {
		t.Errorf("raw = %v, want 80", res.raw)
	}
	if res.authoritative != nil {
		t.Errorf("authoritative = %v, want nil (curve pending)", *res.authoritative)
	}
}

func TestTrackingMovingSaturatedGolden(t *testing.T) {
	frames := []*esportspb.TrackingMovingFrame{
		tmMove(0, 0, 0), tmMove(500, 600, 0), tmMove(1000, 1200, 0),
	}
	for i := 0; i < 10; i++ {
		frames = append(frames, tmShot(int32(2000+i*1000), 20, true, 1000))
	}
	res, err := trackingMovingScorer{}.recompute(tmRequest(frames...), tmParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 4 + 2.5 + 2 + 2 = 10.5 clamped to 10.
	if res.extras[extraComposite] != 10.0 {
		t.Errorf("composite = %v, want 10.0", res.extras[extraComposite])
	}
	if res.raw != 100 {
		t.Errorf("raw = %v, want 100", res.raw)
	}
}

func TestTrackingMovingUnitScoreAndDifficultyOverride(t *testing.T) {
	p := tmParams()
	p.difficultyFactor = 1.5
	p.engineParams = map[string]string{unitScoreKey: "15"}
	res, err := trackingMovingScorer{}.recompute(tmRequest(tmMidSession()...), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 8 confirmed hits × overridden unit 15 × difficulty 1.5 = 180.
	if res.raw != 180 {
		t.Errorf("raw = %v, want 180", res.raw)
	}
}

func TestTrackingMovingHonestZero(t *testing.T) {
	res, err := trackingMovingScorer{}.recompute(tmRequest(tmMove(0, 0, 0), tmMove(500, 300, 0)), tmParams())
	if err != nil {
		t.Fatalf("shotless session error: %v", err)
	}
	if res.raw != 0 || res.extras[extraComposite] != 0 {
		t.Errorf("shotless session: raw = %v composite = %v, want 0/0", res.raw, res.extras[extraComposite])
	}
}

func TestTrackingMovingRejections(t *testing.T) {
	life := int32(500)
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
			name: "movement frame carrying a lifetime",
			req: tmRequest(&esportspb.TrackingMovingFrame{
				ElapsedMs: 0, Player: tmPoint(0, 0), DotLifetimeMs: &life,
			}),
			reason: reasonDetailMismatch,
		},
		{
			name: "shot frame missing its lifetime",
			req: tmRequest(&esportspb.TrackingMovingFrame{
				ElapsedMs: 1000, Deviation: 30, Hit: func() *bool { v := true; return &v }(),
			}),
			reason: reasonDetailMismatch,
		},
		{
			name:   "hit verdict contradicting the deviation",
			req:    tmRequest(tmShot(1000, 60, true, 1200)),
			reason: reasonDetailMismatch,
		},
		{
			name:   "frames out of order",
			req:    tmRequest(tmShot(5000, 30, true, 1200), tmShot(4000, 30, true, 1200)),
			reason: reasonDetailMismatch,
		},
		{
			name:   "span past the config duration",
			req:    tmRequest(tmMove(33000, 0, 0)),
			reason: reasonAbsurdTimeSpan,
		},
		{
			name:   "kill faster than reaction floor",
			req:    tmRequest(tmShot(1000, 30, true, 50)),
			reason: reasonAbsurdReactionTime,
		},
		{
			name: "negative deviation",
			req: tmRequest(&esportspb.TrackingMovingFrame{
				ElapsedMs: 1000, Deviation: -1, Player: tmPoint(0, 0),
			}),
			reason: reasonDetailMismatch,
		},
	}
	for _, tc := range cases {
		_, err := trackingMovingScorer{}.recompute(tc.req, tmParams())
		var invalid *invalidAttemptError
		if !errors.As(err, &invalid) || invalid.reason != tc.reason {
			t.Errorf("%s: err = %v, want %s", tc.name, err, tc.reason)
		}
	}
}
