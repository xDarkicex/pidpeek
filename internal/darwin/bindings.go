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
	ProcPidTbsdInfo  = 3  // proc_bsdinfo — 136 bytes
	ProcPidTaskInfo  = 4  // proc_taskinfo — 96 bytes
	ProcPidPathInfo  = 11 // path buffer — PROC_PIDPATHINFO_MAXSIZE = 4096
)

const (
	ProcPidPathInfoMaxSize = 4096
)

// MachTimebaseInfoData holds mach timebase info numerator and denominator.
type MachTimebaseInfoData struct {
	Numer uint32
	Denom uint32
}

// TimebaseRatioVal is the cached timebase ratio (numer/denom as float64).
var TimebaseRatioVal float64

// ProcPidinfo calls proc_pidinfo via purego.
var ProcPidinfo func(pid int32, flavor int32, arg uint64, buffer unsafe.Pointer, buffersize int32) int32

// ProcPidpath calls proc_pidpath via purego.
var ProcPidpath func(pid int32, buffer unsafe.Pointer, buffersize uint32) int32

// MachTimebaseInfo calls mach_timebase_info via purego.
var MachTimebaseInfo func(info *MachTimebaseInfoData) int32

var (
	initOnce sync.Once
	initErr  error
	initFn   func() error
)

// EnsureInit initializes the Darwin package via the registered init function.
func EnsureInit() error {
	initOnce.Do(func() {
		if initFn != nil {
			initErr = initFn()
		}
	})
	return initErr
}

// SetInitFunction registers the init function from the platform bridge.
func SetInitFunction(fn func() error) {
	initFn = fn
}

var _ = fmt.Sprintf // ensure fmt is linked