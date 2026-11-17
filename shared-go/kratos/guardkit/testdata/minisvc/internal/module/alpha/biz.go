package alpha

import (
	"context"
	"time"
)

const owned = "kalpha.owns"

type AlphaUC struct {
	clock func() time.Time
}

func (uc *AlphaUC) Stamp(ctx context.Context) string {
	_ = ctx
	return time.Now().UTC().Format("20060102") // wall-clock: export stamp — presentation, never asserted
}

func (uc *AlphaUC) Await(ctx context.Context) {
	_ = ctx
	select {
	case <-time.After(0):
	}
}
