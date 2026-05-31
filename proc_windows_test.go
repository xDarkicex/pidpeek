//go:build windows

package pidpeek

import (
	"testing"
	"unsafe"

	"github.com/xDarkicex/pidpeek/internal/windows"
)

func TestPROCESSMEMORYCOUNTERSEXSize(t *testing.T) {
	if got := unsafe.Sizeof(windows.PROCESS_MEMORY_COUNTERS_EX{}); got != 80 {
		t.Fatalf("PROCESS_MEMORY_COUNTERS_EX size=%d want 80", got)
	}
}

func TestPROCESSMEMORYCOUNTERSEXOffsets(t *testing.T) {
	var mc windows.PROCESS_MEMORY_COUNTERS_EX
	tests := []struct {
		field string
		got   uintptr
		want  uintptr
	}{
		{"Cb", unsafe.Offsetof(mc.Cb), 0},
		{"PageFaultCount", unsafe.Offsetof(mc.PageFaultCount), 4},
		{"PeakWorkingSetSize", unsafe.Offsetof(mc.PeakWorkingSetSize), 8},
		{"WorkingSetSize", unsafe.Offsetof(mc.WorkingSetSize), 16},
		{"QuotaPeakPagedPoolUsage", unsafe.Offsetof(mc.QuotaPeakPagedPoolUsage), 24},
		{"QuotaPagedPoolUsage", unsafe.Offsetof(mc.QuotaPagedPoolUsage), 32},
		{"QuotaPeakNonPagedPoolUsage", unsafe.Offsetof(mc.QuotaPeakNonPagedPoolUsage), 40},
		{"QuotaNonPagedPoolUsage", unsafe.Offsetof(mc.QuotaNonPagedPoolUsage), 48},
		{"PagefileUsage", unsafe.Offsetof(mc.PagefileUsage), 56},
		{"PeakPagefileUsage", unsafe.Offsetof(mc.PeakPagefileUsage), 64},
		{"PrivateUsage", unsafe.Offsetof(mc.PrivateUsage), 72},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s offset=%d want %d", tt.field, tt.got, tt.want)
		}
	}
}
