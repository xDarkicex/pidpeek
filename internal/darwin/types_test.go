//go:build darwin

package darwin

import (
	"os"
	"testing"
)

func TestReadNullTerminated(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		buf := [32]byte{}
		copy(buf[:], "bash")
		got := readNullTerminated(buf[:])
		if got != "bash" {
			t.Errorf("readNullTerminated = %q, want %q", got, "bash")
		}
	})

	t.Run("no_null_byte", func(t *testing.T) {
		buf := []byte("hello")
		got := readNullTerminated(buf)
		if got != "hello" {
			t.Errorf("readNullTerminated = %q, want %q", got, "hello")
		}
	})

	t.Run("null_at_start", func(t *testing.T) {
		buf := [32]byte{}
		buf[0] = 0
		copy(buf[1:], "world")
		got := readNullTerminated(buf[:])
		if got != "" {
			t.Errorf("readNullTerminated = %q, want empty", got)
		}
	})

	t.Run("pbi_name_buffer", func(t *testing.T) {
		var buf [32]byte
		copy(buf[:], "com.docker.sbom")
		got := readNullTerminated(buf[:])
		if got != "com.docker.sbom" {
			t.Errorf("readNullTerminated = %q, want %q", got, "com.docker.sbom")
		}
	})
}

func TestGetpid(t *testing.T) {
	pid := getpid()
	if pid <= 0 {
		t.Errorf("getpid = %d, want > 0", pid)
	}
}

func TestProcessMetricsSelf(t *testing.T) {
	pid := os.Getpid()
	m, err := ProcessMetrics(pid)
	if err != nil {
		t.Fatalf("ProcessMetrics(%d): %v", pid, err)
	}
	if m.RSS == 0 {
		t.Error("RSS is zero")
	}
	if m.VMSSize == 0 {
		t.Error("VMSSize is zero")
	}
	if m.ThreadNum < 1 {
		t.Errorf("ThreadNum = %d, want >= 1", m.ThreadNum)
	}
	if m.CPUTotalSec < 0 {
		t.Errorf("CPUTotalSec is negative: %f", m.CPUTotalSec)
	}
}

func TestProcessMetricsSelfMinusOne(t *testing.T) {
	m, err := ProcessMetrics(-1)
	if err != nil {
		t.Fatalf("ProcessMetrics(-1): %v", err)
	}
	if m.RSS == 0 {
		t.Error("RSS is zero")
	}
	if m.ThreadNum < 1 {
		t.Errorf("ThreadNum = %d, want >= 1", m.ThreadNum)
	}
}

func TestProcessMetricsNotFound(t *testing.T) {
	_, err := ProcessMetrics(999999)
	if err != ErrProcessNotFound {
		t.Errorf("ProcessMetrics(999999) error = %v, want ErrProcessNotFound", err)
	}
}

func TestProcessMetricsPIDZero(t *testing.T) {
	_, err := ProcessMetrics(0)
	if err != ErrProcessNotFound {
		t.Errorf("ProcessMetrics(0) error = %v, want ErrProcessNotFound", err)
	}
}

func TestProcessIdentitySelf(t *testing.T) {
	pid := os.Getpid()
	id, err := ProcessIdentity(pid)
	if err != nil {
		t.Fatalf("ProcessIdentity(%d): %v", pid, err)
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
	if id.CreateTime <= 0 {
		t.Errorf("CreateTime = %d, want > 0", id.CreateTime)
	}
}

func TestProcessIdentitySelfMinusOne(t *testing.T) {
	id, err := ProcessIdentity(-1)
	if err != nil {
		t.Fatalf("ProcessIdentity(-1): %v", err)
	}
	if id.Name == "" {
		t.Error("Name is empty")
	}
	if id.ExePath == "" {
		t.Error("ExePath is empty")
	}
}

func TestProcessIdentityNotFound(t *testing.T) {
	_, err := ProcessIdentity(999999)
	if err != ErrProcessNotFound {
		t.Errorf("ProcessIdentity(999999) error = %v, want ErrProcessNotFound", err)
	}
}

func TestProcessIdentityPIDZero(t *testing.T) {
	_, err := ProcessIdentity(0)
	if err != ErrProcessNotFound {
		t.Errorf("ProcessIdentity(0) error = %v, want ErrProcessNotFound", err)
	}
}

func TestReadExePathSelf(t *testing.T) {
	path := readExePath(os.Getpid())
	if path == "" {
		t.Error("readExePath(self) is empty")
	}
}

func TestReadExePathPIDZero(t *testing.T) {
	path := readExePath(0)
	if path != "" {
		t.Errorf("readExePath(0) = %q, want empty", path)
	}
}

func TestEnsureInit(t *testing.T) {
	if err := EnsureInit(); err != nil {
		t.Fatalf("EnsureInit: %v", err)
	}
}
