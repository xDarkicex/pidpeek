# Benchmarks

Darwin (arm64), Apple M2, Go 1.25. `go test -bench=. -benchtime=1s -count=5`.

## Current

| Benchmark | ops/sec | ns/op | B/op | allocs/op |
|-----------|--------|-------|------|-----------|
| `Get` | 868,810 | 1,151 | 384 | 8 |
| `GetIdentity` | 274,725 | 3,640 | 816 | 16 |
| `GetAll` | 202,306 | 4,943 | 1,208 | 24 |
| `Self` | 831,946 | 1,202 | 384 | 8 |
| `SelfIdentity` | 271,960 | 3,677 | 816 | 16 |
| `FormatBytes` | 4,405,286 | 227 | 16 | 2 |
| `GoHeapAlloc` | 2,481,390 | 403 | 48 | 1 |

## Before (heap-allocated structs + path buffer)

| Benchmark | ops/sec | ns/op | B/op | allocs/op |
|-----------|--------|-------|------|-----------|
| `Get` | 945,869 | 1,232 | 480 | 9 |
| `GetIdentity` | 267,580 | 4,478 | 5,120 | 19 |
| `GetAll` | 206,369 | 5,850 | 5,608 | 28 |
| `Self` | 949,304 | 1,273 | 480 | 9 |
| `SelfIdentity` | 261,348 | 4,485 | 5,120 | 19 |
| `FormatBytes` | 5,282,907 | 227 | 16 | 2 |
| `GoHeapAlloc` | 2,983,646 | 402 | 48 | 1 |

## Delta

| Benchmark | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| `Get` | -6.6% | -20% | -1 |
| `GetIdentity` | -18.7% | **-84%** | -3 |
| `GetAll` | -15.5% | **-78%** | -4 |
| `Self` | -5.6% | -20% | -1 |
| `SelfIdentity` | -18.0% | **-84%** | -3 |

Struct and path buffer allocations moved off-heap via `github.com/xDarkicex/memory.FreeList`.
The 4KB `make([]byte, 4096)` in `readExePath` — previously 70% of total allocation volume — is eliminated.

Remaining 8 allocs/op on `Get` are dominated by `purego` internal C-call marshaling (~41%),
`FreeList` slotGen metadata, and `string()` conversions for returned identity fields.
None of these are fixable from the pidpeek side.

Linux and Windows benchmarks pending — require native execution.
