//go:build linux

// Package linux provides Linux-specific process metrics via procfs.
package linux

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

// Sentinels for internal linux package.
var (
	ErrProcessNotFound   = errors.New("process not found")
	ErrAccessDenied      = errors.New("access denied")
	ErrResourceExhausted = errors.New("resource exhausted")
)

// Metrics holds resource usage for a Linux process.
type Metrics struct {
	RSS         uint64
	VMSSize     uint64
	CPUTotalSec float64
	ThreadNum   int32
}

// Identity holds process identity information for Linux.
type Identity struct {
	Name       string
	Ppid       int
	CreateTime int64
	ExePath    string
}

var (
	clkTckOnce sync.Once
	clkTckVal  float64

	bootTimeOnce sync.Once
	bootTimeVal  int64
)

// pageSize is resolved once at startup — os.Getpagesize cannot fail.
var pageSize = uint64(os.Getpagesize())

// getClkTck returns the system clock tick rate, cached on first call.
// Resolved from /proc/self/auxv; falls back to 100.0 if unavailable.
func getClkTck() float64 {
	clkTckOnce.Do(func() {
		data, err := os.ReadFile("/proc/self/auxv")
		if err != nil {
			clkTckVal = 100.0
			return
		}
		clkTckVal = ReadClkTck(data)
	})
	return clkTckVal
}

// getBootTime returns the system boot time in Unix seconds, cached on first call.
func getBootTime() int64 {
	bootTimeOnce.Do(func() {
		data, err := os.ReadFile("/proc/stat")
		if err != nil {
			return
		}
		btime, err := ParseProcStatBtime(data)
		if err != nil {
			return
		}
		bootTimeVal = btime
	})
	return bootTimeVal
}

// ProcessMetrics retrieves resource usage for the given PID from procfs.
// Returns zeroed Metrics without error for zombie processes (state Z).
func ProcessMetrics(pid int) (Metrics, error) {
	statData, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return Metrics{}, ErrProcessNotFound
	}

	fields, err := ParseStatFields(statData)
	if err != nil {
		return Metrics{}, ErrProcessNotFound
	}

	if fields.State == 'Z' {
		return Metrics{RSS: 0, VMSSize: 0, CPUTotalSec: 0, ThreadNum: 0}, nil
	}

	statmData, err := os.ReadFile(fmt.Sprintf("/proc/%d/statm", pid))
	if err != nil {
		return Metrics{}, ErrProcessNotFound
	}

	rssPages, err := ParseStatmPages(statmData)
	if err != nil {
		return Metrics{}, ErrProcessNotFound
	}

	cpuSec := float64(fields.Utime+fields.Stime) / getClkTck()

	return Metrics{
		RSS:         rssPages * pageSize,
		VMSSize:     fields.VSize,
		CPUTotalSec: cpuSec,
		ThreadNum:   fields.NumThreads,
	}, nil
}

// ProcessIdentity retrieves process identity for the given PID from procfs.
func ProcessIdentity(pid int) (Identity, error) {
	statData, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return Identity{}, ErrProcessNotFound
	}

	name, err := ParseComm(statData)
	if err != nil {
		return Identity{}, ErrProcessNotFound
	}

	fields, err := ParseStatFields(statData)
	if err != nil {
		return Identity{}, ErrProcessNotFound
	}

	createTime := int64(0)
	btime := getBootTime()
	if btime > 0 && fields.Starttime > 0 {
		createTime = btime + int64(float64(fields.Starttime)/getClkTck())
	}

	exePath, _ := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))

	return Identity{
		Name:       name,
		Ppid:       int(fields.Ppid),
		CreateTime: createTime,
		ExePath:    exePath,
	}, nil
}
