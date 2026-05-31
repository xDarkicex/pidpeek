// Package pidpeek provides process metrics retrieval across platforms.
package pidpeek

import "runtime/metrics"

// GoHeapAlloc returns the current live heap bytes via runtime/metrics.
// Uses "/gc/heap/live:bytes" which lags and is updated at each GC cycle.
func GoHeapAlloc() uint64 {
	const name = "/gc/heap/live:bytes"
	sample := []metrics.Sample{{Name: name}}
	metrics.Read(sample)
	return sample[0].Value.Uint64()
}