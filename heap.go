// Package pidpeek provides process metrics retrieval across platforms.
package pidpeek

import (
	"runtime/metrics"
	"sync"
)

// GoRuntimeMetrics is a fixed-cost snapshot of metrics exposed by the Go
// runtime for this process only. It intentionally is not part of Metrics:
// Metrics supports arbitrary PIDs, while these counters exist only for the
// current Go runtime.
//
// CumulativeAllocBytes and CumulativeAllocObjects are monotonic counters and
// are suitable for before/after allocation deltas. HeapLiveBytes is updated at
// GC boundaries and must not be interpreted as an instantaneous allocation
// counter.
type GoRuntimeMetrics struct {
	HeapLiveBytes          uint64
	CumulativeAllocBytes   uint64
	CumulativeAllocObjects uint64
}

const (
	goHeapLiveMetric    = "/gc/heap/live:bytes"
	goHeapAllocsBytes   = "/gc/heap/allocs:bytes"
	goHeapAllocsObjects = "/gc/heap/allocs:objects"
)

type goRuntimeSampleSet struct {
	samples [3]metrics.Sample
}

type goHeapSample struct {
	sample [1]metrics.Sample
}

// metrics.Read retains the sample slice for the duration of the call, which
// makes a stack-local sample array escape. Reusing an initialized set through
// sync.Pool removes that allocation while remaining safe for concurrent
// callers. The common per-P pool path does not contend; callers needing a
// hard real-time guarantee should use their own measurement schedule because
// Go runtime metrics are not a real-time API.
var goRuntimeSamplePool = sync.Pool{
	New: func() any {
		return &goRuntimeSampleSet{samples: [3]metrics.Sample{
			{Name: goHeapLiveMetric},
			{Name: goHeapAllocsBytes},
			{Name: goHeapAllocsObjects},
		}}
	},
}

var goHeapSamplePool = sync.Pool{
	New: func() any {
		return &goHeapSample{sample: [1]metrics.Sample{{Name: goHeapLiveMetric}}}
	},
}

// GoRuntime returns a fixed-size self-runtime snapshot. It has no per-call
// heap allocation after pool warmup and no maps or reflection. Its latency is
// a measured property of the Go runtime version and platform; callers should
// use the included benchmark rather than assume an impossible sub-nanosecond
// guarantee.
func GoRuntime() GoRuntimeMetrics {
	set := goRuntimeSamplePool.Get().(*goRuntimeSampleSet)
	metrics.Read(set.samples[:])
	result := GoRuntimeMetrics{
		HeapLiveBytes:          set.samples[0].Value.Uint64(),
		CumulativeAllocBytes:   set.samples[1].Value.Uint64(),
		CumulativeAllocObjects: set.samples[2].Value.Uint64(),
	}
	goRuntimeSamplePool.Put(set)
	return result
}

// GoHeapAlloc returns the current live heap bytes via runtime/metrics.
// Uses "/gc/heap/live:bytes" which lags and is updated at each GC cycle.
func GoHeapAlloc() uint64 {
	set := goHeapSamplePool.Get().(*goHeapSample)
	metrics.Read(set.sample[:])
	result := set.sample[0].Value.Uint64()
	goHeapSamplePool.Put(set)
	return result
}
