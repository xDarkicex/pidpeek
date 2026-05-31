// Package pidpeek provides process metrics retrieval across platforms.
package pidpeek

// Metrics holds resource usage for a process.
// RSS semantics differ across platforms: Linux includes shared memory pages,
// Darwin reflects physical RAM attributed to the task, Windows includes private+shared working set.
// VMS semantics differ: Linux/Darwin = total virtual address space; Windows = Commit Charge (PagefileUsage).
// Huge pages (MAP_HUGETLB on Linux) inflate RSS by 2MB per huge page; the byte value is still accurate.
type Metrics struct {
	RSS         uint64  // resident set size in bytes
	VMSSize     uint64  // virtual memory size in bytes (platform-mapped semantics)
	CPUTotalSec float64 // user + system CPU seconds; Linux resolution is CLK_TCK (~10ms), Darwin/Windows sub-ms
	ThreadNum   int32   // number of threads
}

// Identity holds process identity information.
type Identity struct {
	Name       string // process name (untruncated where available)
	Ppid       int    // parent PID
	CreateTime int64  // Unix timestamp (seconds since epoch)
	ExePath    string // absolute path (may be empty if inaccessible)
}

// ProcessInfo bundles Metrics and Identity for a single process.
type ProcessInfo struct {
	Metrics
	Identity
}

// Inspector is the testability seam. Callers building observability tools can inject a mock
// in tests rather than hitting real OS calls.
type Inspector interface {
	Get(pid int) (Metrics, error)
	GetIdentity(pid int) (Identity, error)
	GetAll(pid int) (ProcessInfo, error)
	Self() (Metrics, error)
	SelfIdentity() (Identity, error)
}

// defaultInspector implements Inspector using real OS calls.
type defaultInspector struct{}

// DefaultInspector is the global Inspector instance using real OS calls.
var DefaultInspector Inspector = &defaultInspector{}
