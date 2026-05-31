//go:build darwin

package pidpeek

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/xDarkicex/pidpeek/internal/darwin"
)

func TestProcTaskInfoSize(t *testing.T) {
	if got := unsafe.Sizeof(darwin.ProcTaskInfo{}); got != 96 {
		t.Fatalf("ProcTaskInfo size=%d want 96", got)
	}
}

func TestProcBsdInfoSize(t *testing.T) {
	if got := unsafe.Sizeof(darwin.ProcBsdInfo{}); got != 136 {
		t.Fatalf("ProcBsdInfo size=%d want 136", got)
	}
}

func TestProcTaskInfoOffsets(t *testing.T) {
	var s darwin.ProcTaskInfo
	tests := []struct {
		field  string
		got    uintptr
		want   uintptr
	}{
		{"VirtualSize", unsafe.Offsetof(s.VirtualSize), 0},
		{"ResidentSize", unsafe.Offsetof(s.ResidentSize), 8},
		{"TotalUser", unsafe.Offsetof(s.TotalUser), 16},
		{"TotalSystem", unsafe.Offsetof(s.TotalSystem), 24},
		{"Threadnum", unsafe.Offsetof(s.Threadnum), 84},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s offset=%d want %d", tt.field, tt.got, tt.want)
		}
	}
}

func TestProcBsdInfoOffsets(t *testing.T) {
	var s darwin.ProcBsdInfo
	tests := []struct {
		field string
		got   uintptr
		want  uintptr
	}{
		{"Pbi_flags", unsafe.Offsetof(s.Pbi_flags), 0},
		{"Pbi_pid", unsafe.Offsetof(s.Pbi_pid), 12},
		{"Pbi_ppid", unsafe.Offsetof(s.Pbi_ppid), 16},
		{"Pbi_comm", unsafe.Offsetof(s.Pbi_comm), 48},
		{"Pbi_name", unsafe.Offsetof(s.Pbi_name), 64},
		{"Pbi_start_tvsec", unsafe.Offsetof(s.Pbi_start_tvsec), 120},
		{"Pbi_start_tvusec", unsafe.Offsetof(s.Pbi_start_tvusec), 128},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s offset=%d want %d", tt.field, tt.got, tt.want)
		}
	}
}

func TestMapDarwinError(t *testing.T) {
	unknownErr := errors.New("unknown error")
	tests := []struct {
		name   string
		err    error
		wantIs error
	}{
		{
			name:   "process_not_found",
			err:    darwin.ErrProcessNotFound,
			wantIs: ErrProcessNotFound,
		},
		{
			name:   "access_denied",
			err:    darwin.ErrAccessDenied,
			wantIs: ErrAccessDenied,
		},
		{
			name:   "resource_exhausted",
			err:    darwin.ErrResourceExhausted,
			wantIs: ErrResourceExhausted,
		},
		{
			name:   "unknown_passthrough",
			err:    unknownErr,
			wantIs: unknownErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapDarwinError(tt.err)
			if !errors.Is(got, tt.wantIs) {
				t.Errorf("errors.Is(%v, %v) = false, want true", got, tt.wantIs)
			}
		})
	}
}
