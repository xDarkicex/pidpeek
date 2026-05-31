//go:build windows

package pidpeek

import (
	"errors"

	"github.com/xDarkicex/pidpeek/internal/windows"
)

var _ Inspector = (*defaultInspector)(nil)

// Get retrieves process metrics for the given PID on Windows.
func (d *defaultInspector) Get(pid int) (Metrics, error) {
	if pid == 0 {
		return Metrics{}, ErrProcessNotFound
	}
	m, err := windows.ProcessMetrics(pid)
	if err != nil {
		return Metrics{}, wrapErr("Get", pid, mapWindowsError(err))
	}
	return Metrics{RSS: m.RSS, VMSSize: m.VMSSize, CPUTotalSec: m.CPUTotalSec, ThreadNum: m.ThreadNum}, nil
}

// GetIdentity retrieves process identity for the given PID on Windows.
func (d *defaultInspector) GetIdentity(pid int) (Identity, error) {
	if pid == 0 {
		return Identity{}, ErrProcessNotFound
	}
	id, err := windows.ProcessIdentity(pid)
	if err != nil {
		return Identity{}, wrapErr("GetIdentity", pid, mapWindowsError(err))
	}
	return Identity{Name: id.Name, Ppid: id.Ppid, CreateTime: id.CreateTime, ExePath: id.ExePath}, nil
}

// GetAll retrieves both metrics and identity for the given PID on Windows.
func (d *defaultInspector) GetAll(pid int) (ProcessInfo, error) {
	if pid == 0 {
		return ProcessInfo{}, ErrProcessNotFound
	}
	m, err := windows.ProcessMetrics(pid)
	if err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, mapWindowsError(err))
	}
	id, err := windows.ProcessIdentity(pid)
	if err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, mapWindowsError(err))
	}
	return ProcessInfo{
		Metrics:  Metrics{RSS: m.RSS, VMSSize: m.VMSSize, CPUTotalSec: m.CPUTotalSec, ThreadNum: m.ThreadNum},
		Identity: Identity{Name: id.Name, Ppid: id.Ppid, CreateTime: id.CreateTime, ExePath: id.ExePath},
	}, nil
}

// Self retrieves metrics for the current process on Windows.
func (d *defaultInspector) Self() (Metrics, error) {
	m, err := windows.ProcessMetrics(-1)
	if err != nil {
		return Metrics{}, wrapErr("Self", 0, mapWindowsError(err))
	}
	return Metrics{RSS: m.RSS, VMSSize: m.VMSSize, CPUTotalSec: m.CPUTotalSec, ThreadNum: m.ThreadNum}, nil
}

// SelfIdentity retrieves identity for the current process on Windows.
func (d *defaultInspector) SelfIdentity() (Identity, error) {
	id, err := windows.ProcessIdentity(-1)
	if err != nil {
		return Identity{}, wrapErr("SelfIdentity", 0, mapWindowsError(err))
	}
	return Identity{Name: id.Name, Ppid: id.Ppid, CreateTime: id.CreateTime, ExePath: id.ExePath}, nil
}

// mapWindowsError maps internal windows sentinel errors to the unified taxonomy
// so that callers can use errors.Is against pidpeek.ErrProcessNotFound etc.
func mapWindowsError(err error) error {
	switch {
	case errors.Is(err, windows.ErrProcessNotFound):
		return ErrProcessNotFound
	case errors.Is(err, windows.ErrAccessDenied):
		return ErrAccessDenied
	case errors.Is(err, windows.ErrResourceExhausted):
		return ErrResourceExhausted
	default:
		return err
	}
}
