//go:build windows

// Package windows provides Windows-specific process metrics via Windows API.
package windows

import (
	"errors"

	syswindows "golang.org/x/sys/windows"
)

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

// ProcessMetrics retrieves resource usage for the given PID.
// Pass pid=-1 for the current process.
func ProcessMetrics(pid int) (Metrics, error) {
	handle, err := getHandle(pid)
	if err != nil {
		return Metrics{}, mapProcessError(err)
	}
	defer closeHandle(handle)

	mc, err := getProcessMemoryInfo(handle)
	if err != nil {
		return Metrics{}, mapProcessError(err)
	}

	cpuSec, _, err := getProcessTimes(handle)
	if err != nil {
		return Metrics{}, mapProcessError(err)
	}

	return Metrics{
		RSS:         mc.WorkingSetSize,
		VMSSize:     mc.PagefileUsage,
		CPUTotalSec: cpuSec,
		ThreadNum:   0,
	}, nil
}

// ProcessIdentity retrieves process identity for the given PID.
// Pass pid=-1 for the current process.
func ProcessIdentity(pid int) (Identity, error) {
	handle, err := getHandle(pid)
	if err != nil {
		return Identity{}, mapProcessError(err)
	}
	defer closeHandle(handle)

	_, createTime, err := getProcessTimes(handle)
	if err != nil {
		return Identity{}, mapProcessError(err)
	}

	exePath := queryFullProcessImageName(handle)

	return Identity{
		Name:       "",
		Ppid:       0,
		CreateTime: createTime,
		ExePath:    exePath,
	}, nil
}

// getHandle returns a process handle for the given PID.
// pid=-1 returns the current process pseudo-handle.
func getHandle(pid int) (syswindows.Handle, error) {
	if pid == -1 {
		return syswindows.CurrentProcess(), nil
	}
	return openProcess(uint32(pid))
}

// mapProcessError maps Windows API errors to sentinel errors.
func mapProcessError(err error) error {
	if errors.Is(err, syswindows.ERROR_ACCESS_DENIED) {
		return ErrAccessDenied
	}
	return err
}
