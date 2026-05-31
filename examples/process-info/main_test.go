package main

import (
	"os"
	"runtime"
	"testing"

	"github.com/xDarkicex/pidpeek"
)

func TestSelfMetrics(t *testing.T) {
	m, err := pidpeek.Self()
	if err != nil {
		t.Fatalf("Self: %v", err)
	}
	if m.ThreadNum < 1 {
		t.Errorf("ThreadNum = %d, want >= 1", m.ThreadNum)
	}
	if m.RSS == 0 {
		t.Error("RSS is zero")
	}
	if m.VMSSize == 0 {
		t.Error("VMSSize is zero")
	}
	if m.CPUTotalSec < 0 {
		t.Errorf("CPUTotalSec is negative: %f", m.CPUTotalSec)
	}
}

func TestSelfIdentity(t *testing.T) {
	id, err := pidpeek.SelfIdentity()
	if err != nil {
		t.Fatalf("SelfIdentity: %v", err)
	}
	if id.Name == "" {
		t.Error("Name is empty")
	}
	if id.Ppid == 0 {
		t.Error("Ppid is 0")
	}
	if id.ExePath == "" {
		t.Error("ExePath is empty")
	}
}

func TestGetSelf(t *testing.T) {
	pid := os.Getpid()
	info, err := pidpeek.GetAll(pid)
	if err != nil {
		t.Fatalf("GetAll(%d): %v", pid, err)
	}
	if info.Metrics.ThreadNum < 1 {
		t.Errorf("ThreadNum = %d, want >= 1", info.Metrics.ThreadNum)
	}
	if info.Identity.Name == "" {
		t.Error("Name is empty")
	}
	if info.Metrics.RSS == 0 {
		t.Error("RSS is zero")
	}
}

func TestGetPID0(t *testing.T) {
	_, err := pidpeek.Get(0)
	if err == nil {
		t.Error("Get(0) should return an error")
	}
}

func TestGetNotFound(t *testing.T) {
	_, err := pidpeek.Get(999999)
	if err == nil {
		t.Error("Get(999999) should return an error")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		n    uint64
		want string
	}{
		{0, "0 B"},
		{1024, "1.0 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
	}
	for _, tt := range tests {
		got := pidpeek.FormatBytes(tt.n)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestGoHeapAlloc(t *testing.T) {
	_ = make([]byte, 1<<20)
	runtime.GC()
	alloc := pidpeek.GoHeapAlloc()
	if alloc == 0 {
		t.Error("GoHeapAlloc is zero after GC with live heap")
	}
}

func BenchmarkSelf(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, err := pidpeek.Self()
		if err != nil {
			b.Fatalf("Self: %v", err)
		}
	}
}

func BenchmarkGetAllSelf(b *testing.B) {
	pid := os.Getpid()
	b.ReportAllocs()
	for b.Loop() {
		_, err := pidpeek.GetAll(pid)
		if err != nil {
			b.Fatalf("GetAll: %v", err)
		}
	}
}

func BenchmarkFormatBytes(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = pidpeek.FormatBytes(1073741824)
	}
}
