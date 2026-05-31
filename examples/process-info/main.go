// Process info: demonstrates the pidpeek API for inspecting process metrics
// and identity — both for the current process and for arbitrary PIDs.
//
//	go run ./examples/process-info/ [pid]
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/xDarkicex/pidpeek"
)

func main() {
	if len(os.Args) > 1 {
		pid, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid pid: %s\n", os.Args[1])
			os.Exit(1)
		}
		printProcess(pid)
		return
	}

	printSelf()
	fmt.Println()
	printGoHeap()
}

func printSelf() {
	info, err := pidpeek.GetAll(os.Getpid())
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetAll self: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Self ===")
	fmt.Printf("  PID:       %d\n", os.Getpid())
	fmt.Printf("  Name:      %s\n", info.Identity.Name)
	fmt.Printf("  PPID:      %d\n", info.Identity.Ppid)
	fmt.Printf("  Exe:       %s\n", info.Identity.ExePath)
	fmt.Printf("  RSS:       %s\n", pidpeek.FormatBytes(info.Metrics.RSS))
	fmt.Printf("  VMS:       %s\n", pidpeek.FormatBytes(info.Metrics.VMSSize))
	fmt.Printf("  CPU:       %.3fs\n", info.Metrics.CPUTotalSec)
	fmt.Printf("  Threads:   %d\n", info.Metrics.ThreadNum)
}

func printProcess(pid int) {
	info, err := pidpeek.GetAll(pid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetAll %d: %v\n", pid, err)
		os.Exit(1)
	}

	fmt.Printf("=== PID %d ===\n", pid)
	fmt.Printf("  Name:      %s\n", info.Identity.Name)
	fmt.Printf("  PPID:      %d\n", info.Identity.Ppid)
	fmt.Printf("  Exe:       %s\n", info.Identity.ExePath)
	fmt.Printf("  Created:   %d\n", info.Identity.CreateTime)
	fmt.Printf("  RSS:       %s\n", pidpeek.FormatBytes(info.Metrics.RSS))
	fmt.Printf("  VMS:       %s\n", pidpeek.FormatBytes(info.Metrics.VMSSize))
	fmt.Printf("  CPU:       %.3fs\n", info.Metrics.CPUTotalSec)
	fmt.Printf("  Threads:   %d\n", info.Metrics.ThreadNum)
}

func printGoHeap() {
	heapBytes := pidpeek.GoHeapAlloc()
	fmt.Printf("Go heap live: %s\n", pidpeek.FormatBytes(heapBytes))
}
