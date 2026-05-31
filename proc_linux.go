//go:build linux

package pidpeek

import (
	"errors"
	"os"

	"github.com/xDarkicex/pidpeek/internal/linux"
)

var _ Inspector = (*defaultInspector)(nil)

// Get retrieves process metrics for the given PID on Linux.
func (d *defaultInspector) Get(pid int) (Metrics, error) {
	if pid == 0 {
		return Metrics{}, ErrProcessNotFound
	}
	m, err := linux.ProcessMetrics(pid)
	if err != nil {
		return Metrics{}, wrapErr("Get", pid, mapLinuxError(err))
	}
	return Metrics{RSS: m.RSS, VMSSize: m.VMSSize, CPUTotalSec: m.CPUTotalSec, ThreadNum: m.ThreadNum}, nil
}

// GetIdentity retrieves process identity for the given PID on Linux.
func (d *defaultInspector) GetIdentity(pid int) (Identity, error) {
	if pid == 0 {
		return Identity{}, ErrProcessNotFound
	}
	id, err := linux.ProcessIdentity(pid)
	if err != nil {
		return Identity{}, wrapErr("GetIdentity", pid, mapLinuxError(err))
	}
	return Identity{Name: id.Name, Ppid: id.Ppid, CreateTime: id.CreateTime, ExePath: id.ExePath}, nil
}

// GetAll retrieves both metrics and identity for the given PID on Linux.
func (d *defaultInspector) GetAll(pid int) (ProcessInfo, error) {
	if pid == 0 {
		return ProcessInfo{}, ErrProcessNotFound
	}
	m, err := linux.ProcessMetrics(pid)
	if err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, mapLinuxError(err))
	}
	id, err := linux.ProcessIdentity(pid)
	if err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, mapLinuxError(err))
	}
	return ProcessInfo{
		Metrics:  Metrics{RSS: m.RSS, VMSSize: m.VMSSize, CPUTotalSec: m.CPUTotalSec, ThreadNum: m.ThreadNum},
		Identity: Identity{Name: id.Name, Ppid: id.Ppid, CreateTime: id.CreateTime, ExePath: id.ExePath},
	}, nil
}

// Self retrieves metrics for the current process on Linux.
func (d *defaultInspector) Self() (Metrics, error) {
	return d.Get(os.Getpid())
}

// SelfIdentity retrieves identity for the current process on Linux.
func (d *defaultInspector) SelfIdentity() (Identity, error) {
	return d.GetIdentity(os.Getpid())
}

// mapLinuxError maps internal linux sentinel errors to the unified taxonomy
// so that callers can use errors.Is against pidpeek.ErrProcessNotFound etc.
func mapLinuxError(err error) error {
	switch {
	case errors.Is(err, linux.ErrProcessNotFound):
		return ErrProcessNotFound
	case errors.Is(err, linux.ErrAccessDenied):
		return ErrAccessDenied
	case errors.Is(err, linux.ErrResourceExhausted):
		return ErrResourceExhausted
	default:
		return err
	}
}
