# pidpeek

[![Go Reference](https://pkg.go.dev/badge/github.com/xDarkicex/pidpeek.svg)](https://pkg.go.dev/github.com/xDarkicex/pidpeek)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8.svg?style=flat-square)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)

Cross-platform process metrics for Go — RSS, VMS, CPU time, thread count, and
identity (name, parent PID, executable path, creation time). **No CGO, no
shell execution, no stateful handles.** One function per need.

- **Darwin** — `libproc.dylib` via [purego](https://github.com/ebitengine/purego)
- **Linux** — `/proc` filesystem, standard library only
- **Windows** — `GetProcessMemoryInfo`, `GetProcessTimes`, `CreateToolhelp32Snapshot`

## Why pidpeek exists

On macOS, the traditional approach — `sysctl` with `kinfo_proc` — zeros out memory
and CPU fields. The `task_for_pid` Mach trap is blocked by SIP/AMFI entitlements,
and raw syscalls are volatile across releases. **No existing pure-Go library
correctly reads process metrics on all three platforms without CGO or shell
execution.**

pidpeek uses `proc_pidinfo` (a stable, public libproc API) via purego on Darwin,
procfs on Linux, and the Win32 process API on Windows. The result is a single
unified API that works everywhere with `CGO_ENABLED=0`.

## Install

```
go get github.com/xDarkicex/pidpeek
```

## Quickstart

```go
package main

import (
    "fmt"
    "os"

    "github.com/xDarkicex/pidpeek"
)

func main() {
    // Current process
    m, err := pidpeek.Self()
    if err != nil {
        panic(err)
    }
    fmt.Printf("RSS: %s, Threads: %d, CPU: %.3fs\n",
        pidpeek.FormatBytes(m.RSS), m.ThreadNum, m.CPUTotalSec)

    // Current process identity
    id, err := pidpeek.SelfIdentity()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Name: %s, PPID: %d, Exe: %s\n", id.Name, id.Ppid, id.ExePath)

    // Arbitrary PID — combine metrics + identity in one call
    info, err := pidpeek.GetAll(os.Getpid())
    if err != nil {
        panic(err)
    }
    fmt.Printf("VMS: %s, Created: %d\n",
        pidpeek.FormatBytes(info.Metrics.VMSSize), info.Identity.CreateTime)

    // Go runtime heap
    fmt.Printf("Go heap: %s\n", pidpeek.FormatBytes(pidpeek.GoHeapAlloc()))
}
```

Run the full example:

```
go run ./examples/process-info/
go run ./examples/process-info/ 1   # inspect PID 1
```

## API

### Functions

| Function | Returns | Description |
|----------|---------|-------------|
| `Get(pid int)` | `(Metrics, error)` | Resource usage for a PID |
| `GetIdentity(pid int)` | `(Identity, error)` | Identity for a PID |
| `GetAll(pid int)` | `(ProcessInfo, error)` | Metrics + identity in one platform call |
| `Self()` | `(Metrics, error)` | Resource usage for the current process |
| `SelfIdentity()` | `(Identity, error)` | Identity for the current process |
| `FormatBytes(uint64)` | `string` | IEC binary formatting: KiB, MiB, GiB |
| `GoHeapAlloc()` | `uint64` | Live heap bytes from `runtime/metrics` |

`Get(0)` returns `ErrProcessNotFound` on all platforms.

### Types

```go
type Metrics struct {
    RSS         uint64  // resident set size in bytes
    VMSSize     uint64  // virtual memory size in bytes (platform-mapped)
    CPUTotalSec float64 // user + system CPU seconds
    ThreadNum   int32   // number of threads
}

type Identity struct {
    Name       string // process name
    Ppid       int    // parent PID
    CreateTime int64  // Unix timestamp (seconds since epoch)
    ExePath    string // absolute path (empty if inaccessible)
}

type ProcessInfo struct {
    Metrics
    Identity
}
```

### Inspector interface

For callers that need a testability seam — inject a mock rather than hitting
real OS calls:

```go
type Inspector interface {
    Get(pid int) (Metrics, error)
    GetIdentity(pid int) (Identity, error)
    GetAll(pid int) (ProcessInfo, error)
    Self() (Metrics, error)
    SelfIdentity() (Identity, error)
}
```

Reference the global `pidpeek.DefaultInspector` or pass your own. The exported
package-level functions delegate to `DefaultInspector`.

### Errors

```go
var (
    ErrProcessNotFound   = errors.New("process not found")
    ErrAccessDenied      = errors.New("access denied")
    ErrResourceExhausted = errors.New("resource exhausted")
)
```

Errors are wrapped with operation and PID context — `errors.Is` still works on
the sentinel:

```
pidpeek.Get 999999: process not found
```

| Platform Error | Unified Sentinel |
|----------------|-----------------|
| ESRCH / ENOENT / `ERROR_FILE_NOT_FOUND` | `ErrProcessNotFound` |
| EACCES / EPERM / `ERROR_ACCESS_DENIED` | `ErrAccessDenied` |
| EMFILE / ENFILE | `ErrResourceExhausted` |

## Platform details

### macOS (Darwin)

Calls `proc_pidinfo` (flavors 3, 4, 11) and `proc_pidpath` from `libproc.dylib`,
and `mach_timebase_info` from `libSystem.B.dylib` — all dynamically linked via
[purego](https://github.com/ebitengine/purego). No CGO, no `sysctl`, no
`task_for_pid`.

- `ResidentSize` from `proc_taskinfo` is already in bytes — **not** multiplied by page size
- Name resolution tries `pbi_name` first, falls back to `pbi_comm`
- Zombie processes return `ErrProcessNotFound` (proc_pidinfo returns 0)
- CPU time uses `mach_timebase_info` for nanosecond precision

### Linux

Reads `/proc/<pid>/stat`, `/proc/<pid>/statm`, `/proc/<pid>/exe`, `/proc/stat`,
and `/proc/self/auxv`. Standard library only.

- `CLK_TCK` resolved from `/proc/self/auxv` at first use; falls back to 100 if
  the entry is missing (Alpine/musl containers)
- RSS from `statm` field 1 × `pageSize`
- `vsize` from `stat` field 23 is already in bytes
- Zombie processes (`state == 'Z'`) return zeroed `Metrics` with `nil` error
- Auxv reader stops at `AT_NULL` sentinel; does not assume a fixed buffer size

### Windows

Uses `GetProcessMemoryInfo` (psapi), `GetProcessTimes` (kernel32),
`QueryFullProcessImageNameW` (kernel32), and `CreateToolhelp32Snapshot` /
`Thread32First` / `Thread32Next` (kernel32) for thread enumeration.

- `PROCESS_QUERY_LIMITED_INFORMATION` only — minimal privilege
- `Self()` uses `GetCurrentProcess()` pseudo-handle (never closed)
- VMS is `PagefileUsage` (Commit Charge), not raw `VirtualSize`
- Thread count via ToolHelp snapshot enumeration (best-effort; returns 0 on failure)
- FILETIME to Unix epoch: offset 116444736000000000 (100ns ticks)

### RSS semantic differences

| Platform | RSS meaning |
|----------|------------|
| Linux | Includes shared memory pages; NUMA reports total pages across all nodes; huge pages inflate by 2MB per page |
| Darwin | Physical RAM pages attributed to the task |
| Windows | Working set (private + shared pages) |

### VMS semantic differences

| Platform | VMS meaning |
|----------|------------|
| Linux | `vsize` — total virtual address space |
| Darwin | `VirtualSize` — total virtual address space |
| Windows | `PagefileUsage` — Commit Charge (private committed memory) |

## Memory allocation strategy

pidpeek uses **[xDarkicex/memory](https://github.com/xDarkicex/memory)** — an
off-heap, lock-free, mmap-backed allocator — to eliminate GC pressure from
per-call struct and buffer allocations on Darwin.

The initial implementation allocated a 96-byte `ProcTaskInfo`, a 136-byte
`ProcBsdInfo`, and a 4KB path buffer on the Go heap for every `Get` /
`GetIdentity` call. Under `memprofile`, `readExePath` alone accounted for
**70% of total allocation volume**.

Moving these to `memory.FreeList` (off-heap mmap, invisible to GC) reduced
allocation volume by **84% on `GetIdentity`** and **78% on `GetAll`**:

| Benchmark | Before (B/op) | After (B/op) | Reduction |
|-----------|:------------:|:-----------:|:---------:|
| `Get` | 480 | 384 | -20% |
| `GetIdentity` | 5,120 | 816 | **-84%** |
| `GetAll` | 5,608 | 1,208 | **-78%** |

This matters for latency-sensitive observability tools: fewer heap allocations
means fewer GC cycles and more predictable p99 tail latency. The remaining
allocations are in purego's C-call trampoline and string conversions for
returned fields — not fixable from the pidpeek side.

Full benchmark data: [BENCHMARK.md](BENCHMARK.md).

## Concurrency

All exported functions are goroutine-safe. Global constants (CLK_TCK, boot time,
timebase ratio) are initialized once via `sync.Once`. No mutable state is shared
between calls. No mutexes, no atomics in the hot path.

## Constraints

- **CGO_ENABLED=0** — enforced at build time; no C toolchain required
- **No `os/exec`** — no shelling out to `ps`, `top`, or `tasklist`
- **No `runtime.ReadMemStats`** — uses `runtime/metrics` exclusively (no STW pauses)
- **Go 1.25+** — required for `runtime.Pinner` hygiene (Darwin purego path)

## License

MIT — see [LICENSE](LICENSE).

Purego is used under the Apache License 2.0. Attribution in [NOTICE](NOTICE),
full text at `licenses/purego/LICENSE`.
