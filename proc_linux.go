//go:build linux

// Package pidpeek retrieves process metrics (RSS, CPU, thread count) on Linux.
package pidpeek

import (
	"fmt"
	"os"
	"runtime"

	"github.com/xDarkicex/pidpeek/internal/linux"
)

// Ensure defaultInspector satisfies Inspector interface.
var _ Inspector = (*defaultInspector)(nil)

func Get(pid int) (Metrics, error) {
	if pid == 0 {
		return Metrics{}, ErrProcessNotFound
	}
	metrics, err := linux.ProcessMetrics(pid)
	if err != nil {
		return Metrics{}, wrapErr("Get", pid, err)
	}
	return metrics, nil
}

func GetIdentity(pid int) (Identity, error) {
	if pid == 0 {
		return Identity{}, ErrProcessNotFound
	}
	identity, err := linux.ProcessIdentity(pid)
	if err != nil {
		return Identity{}, wrapErr("GetIdentity", pid, err)
	}
	return identity, nil
}

func GetAll(pid int) (ProcessInfo, error) {
	if pid == 0 {
		return ProcessInfo{}, ErrProcessNotFound
	}
	metrics, err := linux.ProcessMetrics(pid)
	if err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, err)
	}
	identity, err := linux.ProcessIdentity(pid)
	if err != nil {
		return ProcessInfo{}, wrapErr("GetAll", pid, err)
	}
	return ProcessInfo{Metrics: metrics, Identity: identity}, nil
}

func Self() (Metrics, error) {
	return Get(os.Getpid())
}

func SelfIdentity() (Identity, error) {
	return GetIdentity(os.Getpid())
}

var _ = runtime.Pinner // ensure pinner is linked