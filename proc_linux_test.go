//go:build linux

package pidpeek

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xDarkicex/pidpeek/internal/linux"
)

const testdataDir = "testdata/linux"

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join(testdataDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read testdata %s: %v", name, err)
	}
	return data
}

func TestParseComm(t *testing.T) {
	data := readTestdata(t, "self_stat.ubuntu")

	name, err := linux.ParseComm(data)
	if err != nil {
		t.Fatalf("ParseComm: %v", err)
	}
	if name != "cat" {
		t.Errorf("ParseComm = %q, want %q", name, "cat")
	}
}

func TestParseStatFields(t *testing.T) {
	data := readTestdata(t, "self_stat.ubuntu")

	fields, err := linux.ParseStatFields(data)
	if err != nil {
		t.Fatalf("ParseStatFields: %v", err)
	}

	if fields.State != 'R' {
		t.Errorf("State = %c, want R", fields.State)
	}
	if fields.Ppid != 1 {
		t.Errorf("Ppid = %d, want 1", fields.Ppid)
	}
	if fields.NumThreads != 1 {
		t.Errorf("NumThreads = %d, want 1", fields.NumThreads)
	}
	if fields.VSize != 2420736 {
		t.Errorf("VSize = %d, want 2420736", fields.VSize)
	}
	if fields.Starttime != 8246 {
		t.Errorf("Starttime = %d, want 8246", fields.Starttime)
	}
}

func TestParseStatmPages(t *testing.T) {
	data := readTestdata(t, "self_statm.ubuntu")

	pages, err := linux.ParseStatmPages(data)
	if err != nil {
		t.Fatalf("ParseStatmPages: %v", err)
	}
	if pages != 178 {
		t.Errorf("ParseStatmPages = %d, want 178", pages)
	}
}

func TestParseProcStatBtime(t *testing.T) {
	data := readTestdata(t, "proc_stat.ubuntu")

	btime, err := linux.ParseProcStatBtime(data)
	if err != nil {
		t.Fatalf("ParseProcStatBtime: %v", err)
	}
	if btime != 1780201036 {
		t.Errorf("ParseProcStatBtime = %d, want 1780201036", btime)
	}
}

func TestReadClkTck(t *testing.T) {
	t.Run("ubuntu_glibc", func(t *testing.T) {
		data := readTestdata(t, "self_auxv.ubuntu")
		clkTck := linux.ReadClkTck(data)
		if clkTck != 100.0 {
			t.Errorf("ReadClkTck = %f, want 100.0", clkTck)
		}
	})

	t.Run("alpine_musl", func(t *testing.T) {
		data := readTestdata(t, "self_auxv.alpine")
		clkTck := linux.ReadClkTck(data)
		if clkTck != 100.0 {
			t.Errorf("ReadClkTck alpine = %f, want 100.0", clkTck)
		}
	})

	t.Run("missing_at_clktck", func(t *testing.T) {
		data := readTestdata(t, "self_auxv.no_clktck")
		clkTck := linux.ReadClkTck(data)
		if clkTck != 100.0 {
			t.Errorf("ReadClkTck = %f, want fallback 100.0", clkTck)
		}
	})
}

func TestParseStatFieldsAlpine(t *testing.T) {
	data := readTestdata(t, "self_stat.alpine")

	fields, err := linux.ParseStatFields(data)
	if err != nil {
		t.Fatalf("ParseStatFields alpine: %v", err)
	}

	if fields.State != 'R' {
		t.Errorf("State = %c, want R", fields.State)
	}
	if fields.Ppid != 1 {
		t.Errorf("Ppid = %d, want 1", fields.Ppid)
	}
}

func TestParseStatmPagesAlpine(t *testing.T) {
	data := readTestdata(t, "self_statm.alpine")

	pages, err := linux.ParseStatmPages(data)
	if err != nil {
		t.Fatalf("ParseStatmPages alpine: %v", err)
	}
	if pages != 220 {
		t.Errorf("ParseStatmPages = %d, want 220", pages)
	}
}

func TestParseProcStatBtimeAlpine(t *testing.T) {
	data := readTestdata(t, "proc_stat.alpine")

	btime, err := linux.ParseProcStatBtime(data)
	if err != nil {
		t.Fatalf("ParseProcStatBtime alpine: %v", err)
	}
	if btime != 1780201036 {
		t.Errorf("ParseProcStatBtime = %d, want 1780201036", btime)
	}
}

func TestZombieDetection(t *testing.T) {
	data := readTestdata(t, "self_stat.ubuntu")

	// Replace 'R' with 'Z' to simulate a zombie
	// Find the state field after the last ')'
	lastParen := 0
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] == ')' {
			lastParen = i
			break
		}
	}
	zombieData := make([]byte, len(data))
	copy(zombieData, data)
	// State is at lastParen+2 (after ") " prefix)
	zombieData[lastParen+2] = 'Z'

	fields, err := linux.ParseStatFields(zombieData)
	if err != nil {
		t.Fatalf("ParseStatFields zombie: %v", err)
	}
	if fields.State != 'Z' {
		t.Errorf("State = %c, want Z", fields.State)
	}
}

// --- Integration tests (require real /proc) ---

func TestProcessMetricsSelf(t *testing.T) {
	pid := os.Getpid()
	m, err := linux.ProcessMetrics(pid)
	if err != nil {
		t.Fatalf("ProcessMetrics(%d): %v", pid, err)
	}
	if m.RSS == 0 {
		t.Error("ProcessMetrics RSS is zero")
	}
	if m.VMSSize == 0 {
		t.Error("ProcessMetrics VMSSize is zero")
	}
	if m.ThreadNum < 1 {
		t.Errorf("ProcessMetrics ThreadNum = %d, want >= 1", m.ThreadNum)
	}
	if m.CPUTotalSec < 0 {
		t.Errorf("ProcessMetrics CPUTotalSec is negative: %f", m.CPUTotalSec)
	}
}

func TestProcessIdentitySelf(t *testing.T) {
	pid := os.Getpid()
	id, err := linux.ProcessIdentity(pid)
	if err != nil {
		t.Fatalf("ProcessIdentity(%d): %v", pid, err)
	}
	if id.Name == "" {
		t.Error("ProcessIdentity Name is empty")
	}
	if id.Ppid == 0 {
		t.Error("ProcessIdentity Ppid is 0")
	}
	if id.ExePath == "" {
		t.Error("ProcessIdentity ExePath is empty")
	}
}

func TestProcessMetricsPID1(t *testing.T) {
	m, err := linux.ProcessMetrics(1)
	if err != nil {
		t.Fatalf("ProcessMetrics(1): %v", err)
	}
	if m.ThreadNum < 1 {
		t.Errorf("ProcessMetrics PID 1 ThreadNum = %d, want >= 1", m.ThreadNum)
	}
}

func TestProcessIdentityPID1(t *testing.T) {
	id, err := linux.ProcessIdentity(1)
	if err != nil {
		t.Fatalf("ProcessIdentity(1): %v", err)
	}
	if id.Name == "" {
		t.Error("ProcessIdentity PID 1 Name is empty")
	}
}

func TestProcessMetricsNotFound(t *testing.T) {
	_, err := linux.ProcessMetrics(999999)
	if err != linux.ErrProcessNotFound {
		t.Errorf("ProcessMetrics(999999) error = %v, want ErrProcessNotFound", err)
	}
}

func TestProcessIdentityNotFound(t *testing.T) {
	_, err := linux.ProcessIdentity(999999)
	if err != linux.ErrProcessNotFound {
		t.Errorf("ProcessIdentity(999999) error = %v, want ErrProcessNotFound", err)
	}
}
