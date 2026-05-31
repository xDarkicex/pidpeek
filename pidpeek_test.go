package pidpeek

import (
	"os"
	"runtime"
	"testing"
)

func TestSelf(t *testing.T) {
	m, err := Self()
	if err != nil {
		t.Fatalf("Self: %v", err)
	}
	if m.ThreadNum < 1 {
		t.Errorf("Self ThreadNum = %d, want >= 1", m.ThreadNum)
	}
	if m.RSS == 0 {
		t.Errorf("Self RSS is zero")
	}
	if m.VMSSize == 0 {
		t.Errorf("Self VMSSize is zero")
	}
	if m.CPUTotalSec < 0 {
		t.Errorf("Self CPUTotalSec is negative: %f", m.CPUTotalSec)
	}
}

func TestSelfIdentity(t *testing.T) {
	id, err := SelfIdentity()
	if err != nil {
		t.Fatalf("SelfIdentity: %v", err)
	}
	if id.Name == "" {
		t.Errorf("SelfIdentity Name is empty")
	}
	// Ppid should be non-zero for any process except init
	if id.Ppid == 0 {
		t.Errorf("SelfIdentity Ppid is 0")
	}
	if id.ExePath == "" {
		t.Errorf("SelfIdentity ExePath is empty")
	}
}

func TestGetSelf(t *testing.T) {
	pid := os.Getpid()
	m, err := Get(pid)
	if err != nil {
		t.Fatalf("Get(%d): %v", pid, err)
	}
	if m.ThreadNum < 1 {
		t.Errorf("Get ThreadNum = %d, want >= 1", m.ThreadNum)
	}
}

func TestGetIdentitySelf(t *testing.T) {
	pid := os.Getpid()
	id, err := GetIdentity(pid)
	if err != nil {
		t.Fatalf("GetIdentity(%d): %v", pid, err)
	}
	if id.Name == "" {
		t.Errorf("GetIdentity Name is empty")
	}
}

func TestGetPID0(t *testing.T) {
	_, err := Get(0)
	if err != ErrProcessNotFound {
		t.Errorf("Get(0) error = %v, want ErrProcessNotFound", err)
	}
}

func TestGetIdentityPID0(t *testing.T) {
	_, err := GetIdentity(0)
	if err != ErrProcessNotFound {
		t.Errorf("GetIdentity(0) error = %v, want ErrProcessNotFound", err)
	}
}

func TestGetAll(t *testing.T) {
	info, err := GetAll(os.Getpid())
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if info.Metrics.ThreadNum < 1 {
		t.Errorf("GetAll ThreadNum = %d, want >= 1", info.Metrics.ThreadNum)
	}
	if info.Identity.Name == "" {
		t.Errorf("GetAll Name is empty")
	}
}

func TestGoHeapAlloc(t *testing.T) {
	// Allocate and force GC so /gc/heap/live:bytes reflects live objects.
	_ = make([]byte, 1<<20)
	runtime.GC()
	alloc := GoHeapAlloc()
	if alloc == 0 {
		t.Error("GoHeapAlloc is zero after GC with live heap")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		n    uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1500, "1.5 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
	}
	for _, tt := range tests {
		got := FormatBytes(tt.n)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
