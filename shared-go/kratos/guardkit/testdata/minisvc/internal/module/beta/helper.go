package beta

import (
	_ "example.com/minisvc/internal/module/alpha"
	tm "time"
)

func frozen() tm.Time { return tm.Now() }
