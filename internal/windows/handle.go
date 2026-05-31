//go:build windows

// Package windows provides Windows-specific process metrics via Windows API.
package windows

import (
	"unsafe"

	syswindows "golang.org/x/sys/windows"
)

// PROCESS_MEMORY_COUNTERS_EX is not exported by x/sys/windows.
// Defined here per Windows SDK psapi.h — 80 bytes on amd64.
type PROCESS_MEMORY_COUNTERS_EX struct {
	Cb                         uint32 // offset 0
	PageFaultCount             uint32 // offset 4
	PeakWorkingSetSize         uint64 // offset 8
	WorkingSetSize             uint64 // offset 16
	QuotaPeakPagedPoolUsage    uint64 // offset 24
	QuotaPagedPoolUsage        uint64 // offset 32
	QuotaPeakNonPagedPoolUsage uint64 // offset 40
	QuotaNonPagedPoolUsage     uint64 // offset 48
	PagefileUsage              uint64 // offset 56
	PeakPagefileUsage          uint64 // offset 64
	PrivateUsage               uint64 // offset 72
}

var (
	modKernel32                      = syswindows.NewLazySystemDLL("kernel32.dll")
	modPsapi                         = syswindows.NewLazySystemDLL("psapi.dll")
	procGetProcessMemoryInfo         = modPsapi.NewProc("GetProcessMemoryInfo")
	procQueryFullProcessImageNameW   = modKernel32.NewProc("QueryFullProcessImageNameW")
	procCreateToolhelp32Snapshot     = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procThread32First                = modKernel32.NewProc("Thread32First")
	procThread32Next                 = modKernel32.NewProc("Thread32Next")
)

// THREADENTRY32 is defined per Windows SDK tlhelp32.h — 28 bytes.
type THREADENTRY32 struct {
	DwSize             uint32 // offset 0
	CntUsage           uint32 // offset 4
	Th32ThreadID       uint32 // offset 8
	Th32OwnerProcessID uint32 // offset 12
	TpBasePri          int32  // offset 16
	TpDeltaPri         int32  // offset 20
	DwFlags            uint32 // offset 24
}

const (
	processQueryLimitedInformation = 0x1000
	th32csSnapthread               = 0x00000004
)

const fileTimeEpochOffset = 116444736000000000 // 100ns ticks from 1601-01-01 to 1970-01-01

// filetimeToUnix converts a Windows FILETIME to Unix seconds.
func filetimeToUnix(ft *syswindows.Filetime) int64 {
	ticks := uint64(ft.HighDateTime)<<32 | uint64(ft.LowDateTime)
	return int64((ticks - fileTimeEpochOffset) / 10_000_000)
}

// openProcess opens a handle to the given PID with limited query access.
func openProcess(pid uint32) (syswindows.Handle, error) {
	return syswindows.OpenProcess(processQueryLimitedInformation, false, pid)
}

// closeHandle wraps syswindows.CloseHandle. Skips pseudo-handles.
func closeHandle(handle syswindows.Handle) {
	if handle == syswindows.CurrentProcess() {
		return
	}
	syswindows.CloseHandle(handle)
}

// getProcessMemoryInfo retrieves memory counters via psapi.GetProcessMemoryInfo.
func getProcessMemoryInfo(handle syswindows.Handle) (*PROCESS_MEMORY_COUNTERS_EX, error) {
	var mc PROCESS_MEMORY_COUNTERS_EX
	mc.Cb = uint32(unsafe.Sizeof(mc))
	r1, _, err := procGetProcessMemoryInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&mc)),
		uintptr(mc.Cb),
	)
	if r1 == 0 {
		return nil, err
	}
	return &mc, nil
}

// getProcessTimes retrieves CPU time and creation time via kernel32.GetProcessTimes.
func getProcessTimes(handle syswindows.Handle) (cpuSec float64, createTime int64, err error) {
	var creation, exit, kernel, user syswindows.Filetime
	err = syswindows.GetProcessTimes(handle, &creation, &exit, &kernel, &user)
	if err != nil {
		return 0, 0, err
	}

	kernNanos := (uint64(kernel.HighDateTime)<<32 | uint64(kernel.LowDateTime)) % 10_000_000
	userNanos := (uint64(user.HighDateTime)<<32 | uint64(user.LowDateTime)) % 10_000_000

	cpuSec = float64(filetimeToUnix(&kernel)) + float64(filetimeToUnix(&user)) +
		float64(kernNanos+userNanos)/1e7
	createTime = filetimeToUnix(&creation)
	return cpuSec, createTime, nil
}

// queryFullProcessImageName retrieves the executable path via QueryFullProcessImageNameW.
func queryFullProcessImageName(handle syswindows.Handle) string {
	const nameWin32 = 0
	var size uint32 = syswindows.MAX_PATH
	buf := make([]uint16, size)
	r1, _, _ := procQueryFullProcessImageNameW.Call(
		uintptr(handle),
		uintptr(nameWin32),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if r1 == 0 {
		return ""
	}
	return syswindows.UTF16ToString(buf[:size])
}

// countProcessThreads counts threads for the given PID via ToolHelp snapshot.
// Returns 0 if the snapshot fails — thread count is best-effort.
func countProcessThreads(pid uint32) int32 {
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(
		uintptr(th32csSnapthread),
		0, // all processes
	)
	if snapshot == ^uintptr(0) { // INVALID_HANDLE_VALUE
		return 0
	}
	defer syswindows.CloseHandle(syswindows.Handle(snapshot))

	var te THREADENTRY32
	te.DwSize = uint32(unsafe.Sizeof(te))

	r1, _, _ := procThread32First.Call(
		snapshot,
		uintptr(unsafe.Pointer(&te)),
	)
	if r1 == 0 {
		return 0
	}

	var count int32
	for {
		if te.Th32OwnerProcessID == pid {
			count++
		}
		r1, _, _ := procThread32Next.Call(
			snapshot,
			uintptr(unsafe.Pointer(&te)),
		)
		if r1 == 0 {
			break
		}
	}
	return count
}
