// Package darwin provides Darwin-specific process metrics via libproc.
package darwin

import "syscall"

func getpid() int32 {
	return int32(syscall.Getpid())
}