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
