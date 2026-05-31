// Package pidpeek provides cross-platform process metrics retrieval.
//
// pidpeek retrieves resource usage (RSS, VMS, CPU time, thread count) and
// identity (name, parent PID, create time, executable path) for arbitrary
// processes. It supports Darwin, Linux, and Windows with a unified API.
//
// All exported functions are safe for concurrent use. Initialization is lazy
// via sync.Once — no init() panics.
//
// Usage:
//
//	m, err := pidpeek.Get(pid)
//	id, err := pidpeek.GetIdentity(pid)
//	info, err := pidpeek.GetAll(pid)
//
//	// Current process
//	m, err := pidpeek.Self()
//
//	// Go runtime heap
//	heapBytes := pidpeek.GoHeapAlloc()
//	formatted := pidpeek.FormatBytes(heapBytes)
//
// Build constraints: CGO_ENABLED=0 is required. The package uses purego on
// Darwin, procfs on Linux, and Windows API via golang.org/x/sys/windows.
package pidpeek
