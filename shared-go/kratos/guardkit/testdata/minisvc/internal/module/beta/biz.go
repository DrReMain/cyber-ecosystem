package beta

import "time"

var legacy = "legacy.code"

type BetaUC struct{}

func (uc *BetaUC) Slide(expires time.Time) time.Duration {
	return time.Until(expires)
}
