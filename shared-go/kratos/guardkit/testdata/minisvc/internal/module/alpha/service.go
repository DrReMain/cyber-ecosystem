package alpha

import "time"

func latency() time.Duration {
	start := time.Now()
	return time.Since(start)
}
