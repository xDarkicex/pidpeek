//go:build linux

package pidpeek

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xDarkicex/pidpeek/internal/linux"
)

const benchTestdataDir = "testdata/linux"

func readBenchTestdata(b *testing.B, name string) []byte {
	b.Helper()
	path := filepath.Join(benchTestdataDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("read testdata %s: %v", name, err)
	}
	return data
}

func BenchmarkParseComm(b *testing.B) {
	data := readBenchTestdata(b, "self_stat.ubuntu")
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := linux.ParseComm(data)
		if err != nil {
			b.Fatalf("ParseComm: %v", err)
		}
	}
}

func BenchmarkParseStatFields(b *testing.B) {
	data := readBenchTestdata(b, "self_stat.ubuntu")
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := linux.ParseStatFields(data)
		if err != nil {
			b.Fatalf("ParseStatFields: %v", err)
		}
	}
}

func BenchmarkParseStatmPages(b *testing.B) {
	data := readBenchTestdata(b, "self_statm.ubuntu")
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := linux.ParseStatmPages(data)
		if err != nil {
			b.Fatalf("ParseStatmPages: %v", err)
		}
	}
}

func BenchmarkReadClkTck(b *testing.B) {
	data := readBenchTestdata(b, "self_auxv.ubuntu")
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = linux.ReadClkTck(data)
	}
}

func BenchmarkParseProcStatBtime(b *testing.B) {
	data := readBenchTestdata(b, "proc_stat.ubuntu")
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := linux.ParseProcStatBtime(data)
		if err != nil {
			b.Fatalf("ParseProcStatBtime: %v", err)
		}
	}
}
