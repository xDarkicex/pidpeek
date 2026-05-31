// Package pidpeek provides process metrics retrieval across platforms.
package pidpeek

import "runtime/metrics"

var _ = metrics.Alloc

// GoHeapAlloc returns the current live heap bytes via runtime/metrics.
// This uses "/gc/heap/live:bytes" which lags and is updated at each GC cycle.
func GoHeapAlloc() uint64 {
	reader := metrics.MakeReader(nil)
	reader.Read(metrics.Snapshot())
	return reader.ReadUint64("/gc/heap/live:bytes")
}