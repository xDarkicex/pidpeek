//go:build darwin

// Package darwin provides Darwin-specific process metrics via libproc.
package darwin

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Flavor constants for proc_pidinfo.
const (
	ProcPidTbsdInfo = 3  // proc_bsdinfo — 136 bytes
	ProcPidTaskInfo = 4  // proc_taskinfo — 96 bytes
	ProcPidPathInfo = 11 // path buffer
)

// ProcPidPathInfoMaxSize is the maximum path buffer size for proc_pidpath.
const ProcPidPathInfoMaxSize = 4096

// MachTimebaseInfoData holds mach timebase info numerator and denominator.
type MachTimebaseInfoData struct {
	Numer uint32
	Denom uint32
}

// TimebaseRatioVal is the cached timebase ratio (numer/denom as float64).
var TimebaseRatioVal float64

// proc_pidinfo and proc_pidpath are in libproc.dylib.
var ProcPidinfo func(pid int32, flavor int32, arg uint64, buffer unsafe.Pointer, buffersize int32) int32

// ProcPidpath calls proc_pidpath via purego.
var ProcPidpath func(pid int32, buffer unsafe.Pointer, buffersize uint32) int32

// MachTimebaseInfo calls mach_timebase_info via purego.
var MachTimebaseInfo func(info *MachTimebaseInfoData) int32

var (
	initOnce sync.Once
	initErr  error
)

// EnsureInit initializes the darwin package by loading libSystem and libproc.
// Safe to call concurrently; uses sync.Once internally.
func EnsureInit() error {
	initOnce.Do(func() {
		initErr = loadLibraries()
	})
	return initErr
}

// ensureInit is the unexported entry point for internal use within the darwin package.
func ensureInit() error {
	return EnsureInit()
}

// loadLibraries dlopens libSystem.B.dylib and libproc.dylib, registers
// mach_timebase_info, proc_pidinfo, proc_pidpath, and caches the timebase ratio.
func loadLibraries() error {
	libSystem, err := purego.Dlopen("/usr/lib/libSystem.B.dylib", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("pidpeek: dlopen libSystem.B.dylib: %w", err)
	}
	libProc, err := purego.Dlopen("/usr/lib/libproc.dylib", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("pidpeek: dlopen libproc.dylib: %w", err)
	}
	purego.RegisterLibFunc(&MachTimebaseInfo, libSystem, "mach_timebase_info")
	purego.RegisterLibFunc(&ProcPidinfo, libProc, "proc_pidinfo")
	purego.RegisterLibFunc(&ProcPidpath, libProc, "proc_pidpath")

	var info MachTimebaseInfoData
	MachTimebaseInfo(&info)
	TimebaseRatioVal = float64(info.Numer) / float64(info.Denom)
	return nil
}
