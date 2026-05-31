//go:build darwin

// Package darwin provides Darwin-specific process metrics via libproc.
package darwin

import (
	"bytes"
	"unsafe"

	"github.com/xDarkicex/memory"
)

// Metrics holds resource usage for a Darwin process.
type Metrics struct {
	RSS         uint64
	VMSSize     uint64
	CPUTotalSec float64
	ThreadNum   int32
}

// Identity holds process identity information for Darwin.
type Identity struct {
	Name       string
	Ppid       int
	CreateTime int64
	ExePath    string
}

// ProcTaskInfo is flavor 4 of proc_pidinfo — 96 bytes.
// Authoritative source: XNU bsd/sys/proc_info.h ProcTaskInfo flavor 4.
type ProcTaskInfo struct {
	VirtualSize   uint64 // offset 0
	ResidentSize  uint64 // offset 8  ← RSS (already in bytes, NOT pages)
	TotalUser     uint64 // offset 16
	TotalSystem   uint64 // offset 24
	ThreadsUser   uint64 // offset 32
	ThreadsSystem uint64 // offset 40
	Policy        int32  // offset 48
	Faults        int32  // offset 52
	Pageins       int32  // offset 56
	CowFaults     int32  // offset 60
	MessagesSent  int32  // offset 64
	MessagesRcvd  int32  // offset 68
	SyscallsMach  int32  // offset 72
	SyscallsUnix  int32  // offset 76
	Csw           int32  // offset 80
	Threadnum     int32  // offset 84  ← ThreadNum
	Numrunning    int32  // offset 88
	Priority      int32  // offset 92
}

// ProcBsdInfo is flavor 3 of proc_pidinfo — 136 bytes.
// Authoritative source: XNU bsd/sys/proc_info.h ProcBsdInfo flavor 3.
type ProcBsdInfo struct {
	Pbi_flags       uint32    // 0
	Pbi_status      uint32    // 4
	Pbi_xstatus     uint32    // 8
	Pbi_pid         uint32    // 12
	Pbi_ppid        uint32    // 16
	Pbi_uid         uint32    // 20
	Pbi_gid         uint32    // 24
	Pbi_ruid        uint32    // 28
	Pbi_rgid        uint32    // 32
	Pbi_svuid       uint32    // 36
	Pbi_svgid       uint32    // 40
	_               uint32    // 44 reserved
	Pbi_comm        [16]byte // 48 MAXCOMLEN=16
	Pbi_name        [32]byte // 64 2*MAXCOMLEN
	Pbi_nfiles      uint32    // 96
	Pbi_pgid        uint32    // 100
	Pbi_pjobc       uint32    // 104
	E_tdev          uint32    // 108
	E_tpgid         uint32    // 112
	Pbi_nice        int32     // 116
	Pbi_start_tvsec  uint64   // 120
	Pbi_start_tvusec uint64   // 128
}

// ProcessMetrics retrieves resource usage for the given PID.
// Pass pid=-1 for the current process.
func ProcessMetrics(pid int) (Metrics, error) {
	if err := ensureInit(); err != nil {
		// coverage:skip-reason init errors only occur when libSystem/libproc are missing from the system
		return Metrics{}, err
	}
	if pid == -1 {
		pid = int(getpid())
	}
	return processMetricsForPid(pid)
}

// processMetricsForPid calls proc_pidinfo with PROC_PIDTASKINFO flavor.
// Allocates ProcTaskInfo from off-heap FreeList — zero GC pressure.
func processMetricsForPid(pid int) (Metrics, error) {
	pti, err := memory.FreeListAlloc[ProcTaskInfo](structFL)
	if err != nil {
		// coverage:skip-reason FreeList exhaustion requires 1600+ concurrent calls; not triggerable in unit tests
		return Metrics{}, err
	}
	defer memory.FreeListDealloc(structFL, pti)

	size := int32(unsafe.Sizeof(*pti))
	nb := ProcPidinfo(int32(pid), ProcPidTaskInfo, 0, unsafe.Pointer(pti), size)
	if nb <= 0 {
		return Metrics{}, ErrProcessNotFound
	}
	// coverage:skip-reason partial read requires process to exit between proc_pidinfo start and return
	if nb < size {
		return Metrics{}, ErrProcessNotFound
	}

	cpuSec := float64(pti.TotalUser+pti.TotalSystem) * TimebaseRatioVal / 1e9

	return Metrics{
		RSS:         pti.ResidentSize,
		VMSSize:     pti.VirtualSize,
		CPUTotalSec: cpuSec,
		ThreadNum:   pti.Threadnum,
	}, nil
}

// ProcessIdentity retrieves identity information for the given PID.
// Pass pid=-1 for the current process.
func ProcessIdentity(pid int) (Identity, error) {
	if err := ensureInit(); err != nil {
		// coverage:skip-reason init errors only occur when libSystem/libproc are missing from the system
		return Identity{}, err
	}
	if pid == -1 {
		pid = int(getpid())
	}
	return processIdentityForPid(pid)
}

// processIdentityForPid calls proc_pidinfo with PROC_PIDTBSDINFO flavor.
// Allocates from off-heap FreeLists — zero GC pressure.
func processIdentityForPid(pid int) (Identity, error) {
	pbi, err := memory.FreeListAlloc[ProcBsdInfo](structFL)
	if err != nil {
		// coverage:skip-reason FreeList exhaustion requires 1600+ concurrent calls; not triggerable in unit tests
		return Identity{}, err
	}
	defer memory.FreeListDealloc(structFL, pbi)

	size := int32(unsafe.Sizeof(*pbi))
	nb := ProcPidinfo(int32(pid), ProcPidTbsdInfo, 0, unsafe.Pointer(pbi), size)
	if nb <= 0 {
		return Identity{}, ErrProcessNotFound
	}
	// coverage:skip-reason partial read requires process to exit between proc_pidinfo start and return
	if nb < size {
		return Identity{}, ErrProcessNotFound
	}

	// Name: try Pbi_name first, fall back to Pbi_comm if empty
	name := readNullTerminated(pbi.Pbi_name[:])
	// coverage:skip-reason all Darwin processes have Pbi_name set; fallback for kernel threads only
	if len(name) == 0 {
		name = readNullTerminated(pbi.Pbi_comm[:])
	}

	createTime := int64(pbi.Pbi_start_tvsec) + int64(pbi.Pbi_start_tvusec)/1_000_000

	exePath := readExePath(pid)

	return Identity{
		Name:       name,
		Ppid:       int(pbi.Pbi_ppid),
		CreateTime: createTime,
		ExePath:    exePath,
	}, nil
}

// readNullTerminated converts a null-terminated byte array to a string.
func readNullTerminated(buf []byte) string {
	n := bytes.IndexByte(buf, 0)
	if n < 0 {
		return string(buf)
	}
	return string(buf[:n])
}

// readExePath reads the executable path for the given PID via proc_pidpath.
// Allocates the path buffer from the off-heap path FreeList.
func readExePath(pid int) string {
	buf, err := memory.FreeListAlloc[byte](pathFL)
	if err != nil {
		// coverage:skip-reason FreeList exhaustion requires 127+ concurrent path reads; not triggerable in unit tests
		return ""
	}
	defer memory.FreeListDealloc(pathFL, buf)
	// buf is a *byte pointing to a 4096-byte usable region.
	// Construct a slice over the full usable buffer.
	pathSlice := unsafe.Slice(buf, ProcPidPathInfoMaxSize)

	nb := ProcPidpath(int32(pid), unsafe.Pointer(unsafe.SliceData(pathSlice)), uint32(len(pathSlice)))
	if nb <= 0 {
		return ""
	}
	n := bytes.IndexByte(pathSlice[:nb], 0)
	// coverage:skip-reason kernel always null-terminates path buffers; branch exists as defensive guard
	if n < 0 {
		return string(pathSlice[:nb])
	}
	return string(pathSlice[:n])
}
