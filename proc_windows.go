//go:build windows

// Package pidpeek retrieves process metrics (RSS, CPU, thread count) on Windows.
package pidpeek

import (
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/xDarkicex/pidpeek/internal/windows"
)

// Ensure defaultInspector satisfies Inspector interface.
var _ Inspector = (*defaultInspector)(nil)

func Get(pid int) (Metrics, error) {
	if pid == 0 {
		return Metrics{}, ErrProcessNotFound
	}
	metrics, err := windows.ProcessMetrics(pid)
	if err != nil {
		return Metrics{}, wrapErr("Get", pid, err)
	}
	return metrics, nil
}

func GetIdentity(pid int) (Identity, error) {
	if pid == 0 {
		return Identity{}, ErrProcessNotFound
	}
	identity, err := windows.ProcessIdentity(pid)
	if err != nil {
		return Identity{}, wrapErr("GetIdentity", pid, err)
	}
	return identity, nil
}

func GetAll(pid int) (ProcessInfo, error) {
	if pid == 0 {
		return ProcessInfo{}, ErrProcessNotFound
	}
	metrics, err := windows.ProcessMetrics(pid)
	if err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, err)
	}
	identity, err := windows.ProcessIdentity(pid)
	if err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, err)
	}
	return ProcessInfo{Metrics: metrics, Identity: identity}, nil
}

func Self() (Metrics, error) {
	metrics, err := windows.ProcessMetrics(-1)
	if err != nil {
		return Metrics{}, wrapErr("Self", 0, err)
	}
	return metrics, nil
}

func SelfIdentity() (Identity, error) {
	identity, err := windows.ProcessIdentity(-1)
	if err != nil {
		return Identity{}, wrapErr("SelfIdentity", 0, err)
	}
	return identity, nil
}

var _ = runtime.Pinner // ensure pinner is linked
var _ unsafe.Pointer   // ensure unsafe is linked