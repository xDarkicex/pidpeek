//go:build windows

// Package windows provides Windows-specific process metrics via Windows API.
package windows

import "errors"

// Sentinels for internal windows package.
var (
	ErrProcessNotFound   = errors.New("process not found")
	ErrAccessDenied      = errors.New("access denied")
	ErrResourceExhausted = errors.New("resource exhausted")
)

// Metrics holds resource usage for a Windows process.
type Metrics struct {
	RSS         uint64
	VMSSize     uint64
	CPUTotalSec float64
	ThreadNum   int32
}

// Identity holds process identity information for Windows.
type Identity struct {
	Name       string
	Ppid       int
	CreateTime int64
	ExePath    string
}

// ProcessMetrics retrieves process metrics for the given PID.
func ProcessMetrics(pid int) (Metrics, error) {
	// TODO: implement Windows API calls
	return Metrics{}, ErrProcessNotFound
}

// ProcessIdentity retrieves process identity for the given PID.
func ProcessIdentity(pid int) (Identity, error) {
	// TODO: implement Windows API calls
	return Identity{}, ErrProcessNotFound
}
