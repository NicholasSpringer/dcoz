package util

import "time"

func AdjustLatency(latency time.Duration, pauseDuration time.Duration, pausePeriod time.Duration) time.Duration {
	adjusted := float64(latency) * (1 - float64(pauseDuration)/float64(pausePeriod))
	return time.Duration(adjusted)
}
