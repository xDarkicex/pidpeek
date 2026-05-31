//go:build darwin

// Package darwin provides Darwin-specific process metrics via libproc.
package darwin

import "syscall"

// getpid returns the current process ID as int32, the native Darwin pid type.
func getpid() int32 {
	return int32(syscall.Getpid())
}