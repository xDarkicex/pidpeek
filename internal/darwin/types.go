// Package darwin provides Darwin-specific process metrics via libproc.
package darwin

import (
	"bytes"
	"runtime"
	"unsafe"
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
	_             [4]byte // explicit padding to 96 bytes
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
	_               [8]byte   // 136 trailing padding to struct size 136
}

var _ = unsafe.Offsetof(ProcBsdInfo{}.Pbi_ppid) // assert offset at compile time

func ProcessMetrics(pid int) (Metrics, error) {
	if err := ensureInit(); err != nil {
		return Metrics{}, err
	}
	// Special case for self
	if pid == -1 {
		pid = int(getpid())
	}
	return processMetricsForPid(pid)
}

func processMetricsForPid(pid int) (Metrics, error) {
	var pinner runtime.Pinner
	buf := make([]byte, 96)
	pinner.Pin(&buf[0])
	defer pinner.Unpin()

	nb := ProcPidinfo(int32(pid), ProcPidTaskInfo, 0, unsafe.Pointer(&buf[0]), int32(len(buf)))
	if nb <= 0 {
		return Metrics{}, ErrProcessNotFound
	}
	if nb < 96 {
		return Metrics{}, ErrProcessNotFound
	}

	pti := (*ProcTaskInfo)(unsafe.Pointer(&buf[0]))
	cpuSec := float64(pti.TotalUser+pti.TotalSystem) * TimebaseRatioVal / 1e9

	return Metrics{
		RSS:         pti.ResidentSize,
		VMSSize:     pti.VirtualSize,
		CPUTotalSec: cpuSec,
		ThreadNum:   pti.Threadnum,
	}, nil
}

func ProcessIdentity(pid int) (Identity, error) {
	if err := ensureInit(); err != nil {
		return Identity{}, err
	}
	// Special case for self
	if pid == -1 {
		pid = int(getpid())
	}
	return processIdentityForPid(pid)
}

func processIdentityForPid(pid int) (Identity, error) {
	var pinner runtime.Pinner
	buf := make([]byte, 136)
	pinner.Pin(&buf[0])
	defer pinner.Unpin()

	nb := ProcPidinfo(int32(pid), ProcPidTbsdInfo, 0, unsafe.Pointer(&buf[0]), int32(len(buf)))
	if nb <= 0 {
		return Identity{}, ErrProcessNotFound
	}
	if nb < 136 {
		return Identity{}, ErrProcessNotFound
	}

	pbi := (*ProcBsdInfo)(unsafe.Pointer(&buf[0]))

	// Name: try Pbi_name first, fall back to Pbi_comm if empty
	name := readNullTerminated(pbi.Pbi_name[:])
	if len(name) == 0 {
		name = readNullTerminated(pbi.Pbi_comm[:])
	}

	// CreateTime from start_tvsec and start_tvusec
	createTime := int64(pbi.Pbi_start_tvsec) + int64(pbi.Pbi_start_tvusec)/1_000_000

	// ExePath via proc_pidpath
	exePath := readExePath(pid)

	return Identity{
		Name:       name,
		Ppid:       int(pbi.Pbi_ppid),
		CreateTime: createTime,
		ExePath:    exePath,
	}, nil
}

func readNullTerminated(buf []byte) string {
	n := bytes.IndexByte(buf, 0)
	if n < 0 {
		return string(buf)
	}
	return string(buf[:n])
}

func readExePath(pid int) string {
	var pinner runtime.Pinner
	pathBuf := make([]byte, ProcPidPathInfoMaxSize)
	pinner.Pin(&pathBuf[0])
	defer pinner.Unpin()

	nb := ProcPidpath(int32(pid), unsafe.Pointer(&pathBuf[0]), uint32(len(pathBuf)))
	if nb <= 0 {
		return ""
	}
	n := bytes.IndexByte(pathBuf[:nb], 0)
	if n < 0 {
		return string(pathBuf[:nb])
	}
	return string(pathBuf[:n])
}