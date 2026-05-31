//go:build darwin

package pidpeek

import (
	"errors"

	"github.com/xDarkicex/pidpeek/internal/darwin"
)

var _ Inspector = (*defaultInspector)(nil)

// Get retrieves process metrics for the given PID on Darwin.
func (d *defaultInspector) Get(pid int) (Metrics, error) {
	if pid == 0 {
		return Metrics{}, ErrProcessNotFound
	}
	if err := darwin.EnsureInit(); err != nil {
		// coverage:skip-reason init errors only occur when libSystem/libproc are missing from the system
		return Metrics{}, wrapErr("Get", pid, err)
	}
	m, err := darwin.ProcessMetrics(pid)
	if err != nil {
		return Metrics{}, wrapErr("Get", pid, mapDarwinError(err))
	}
	return Metrics{RSS: m.RSS, VMSSize: m.VMSSize, CPUTotalSec: m.CPUTotalSec, ThreadNum: m.ThreadNum}, nil
}

// GetIdentity retrieves process identity for the given PID on Darwin.
func (d *defaultInspector) GetIdentity(pid int) (Identity, error) {
	if pid == 0 {
		return Identity{}, ErrProcessNotFound
	}
	if err := darwin.EnsureInit(); err != nil {
		// coverage:skip-reason init errors only occur when libSystem/libproc are missing from the system
		return Identity{}, wrapErr("GetIdentity", pid, err)
	}
	id, err := darwin.ProcessIdentity(pid)
	if err != nil {
		return Identity{}, wrapErr("GetIdentity", pid, mapDarwinError(err))
	}
	return Identity{Name: id.Name, Ppid: id.Ppid, CreateTime: id.CreateTime, ExePath: id.ExePath}, nil
}

// GetAll retrieves both metrics and identity for the given PID on Darwin.
func (d *defaultInspector) GetAll(pid int) (ProcessInfo, error) {
	if pid == 0 {
		return ProcessInfo{}, ErrProcessNotFound
	}
	if err := darwin.EnsureInit(); err != nil {
		// coverage:skip-reason init errors only occur when libSystem/libproc are missing from the system
		return ProcessInfo{}, wrapErr("GetAll", pid, err)
	}
	m, err := darwin.ProcessMetrics(pid)
	if err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, mapDarwinError(err))
	}
	id, err := darwin.ProcessIdentity(pid)
	if err != nil {
		// coverage:skip-reason requires process to exit between ProcessMetrics and ProcessIdentity calls
		return ProcessInfo{}, wrapErr("GetAll", pid, mapDarwinError(err))
	}
	return ProcessInfo{
		Metrics:  Metrics{RSS: m.RSS, VMSSize: m.VMSSize, CPUTotalSec: m.CPUTotalSec, ThreadNum: m.ThreadNum},
		Identity: Identity{Name: id.Name, Ppid: id.Ppid, CreateTime: id.CreateTime, ExePath: id.ExePath},
	}, nil
}

// Self retrieves metrics for the current process on Darwin.
func (d *defaultInspector) Self() (Metrics, error) {
	if err := darwin.EnsureInit(); err != nil {
		// coverage:skip-reason init errors only occur when libSystem/libproc are missing from the system
		return Metrics{}, wrapErr("Self", 0, err)
	}
	m, err := darwin.ProcessMetrics(-1)
	if err != nil {
		// coverage:skip-reason ProcessMetrics(-1) always succeeds for the calling process
		return Metrics{}, wrapErr("Self", 0, mapDarwinError(err))
	}
	return Metrics{RSS: m.RSS, VMSSize: m.VMSSize, CPUTotalSec: m.CPUTotalSec, ThreadNum: m.ThreadNum}, nil
}

// SelfIdentity retrieves identity for the current process on Darwin.
func (d *defaultInspector) SelfIdentity() (Identity, error) {
	if err := darwin.EnsureInit(); err != nil {
		// coverage:skip-reason init errors only occur when libSystem/libproc are missing from the system
		return Identity{}, wrapErr("SelfIdentity", 0, err)
	}
	id, err := darwin.ProcessIdentity(-1)
	if err != nil {
		// coverage:skip-reason ProcessIdentity(-1) always succeeds for the calling process
		return Identity{}, wrapErr("SelfIdentity", 0, mapDarwinError(err))
	}
	return Identity{Name: id.Name, Ppid: id.Ppid, CreateTime: id.CreateTime, ExePath: id.ExePath}, nil
}

// mapDarwinError maps internal darwin sentinel errors to the unified taxonomy
// so that callers can use errors.Is against pidpeek.ErrProcessNotFound etc.
func mapDarwinError(err error) error {
	switch {
	case errors.Is(err, darwin.ErrProcessNotFound):
		return ErrProcessNotFound
	case errors.Is(err, darwin.ErrAccessDenied):
		return ErrAccessDenied
	case errors.Is(err, darwin.ErrResourceExhausted):
		return ErrResourceExhausted
	default:
		return err
	}
}
