package pidpeek

import (
	"os"
	"testing"
)

func BenchmarkGet(b *testing.B) {
	pid := os.Getpid()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := Get(pid)
		if err != nil {
			b.Fatalf("Get: %v", err)
		}
	}
}

func BenchmarkGetIdentity(b *testing.B) {
	pid := os.Getpid()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := GetIdentity(pid)
		if err != nil {
			b.Fatalf("GetIdentity: %v", err)
		}
	}
}

func BenchmarkGetAll(b *testing.B) {
	pid := os.Getpid()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := GetAll(pid)
		if err != nil {
			b.Fatalf("GetAll: %v", err)
		}
	}
}

func BenchmarkSelf(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := Self()
		if err != nil {
			b.Fatalf("Self: %v", err)
		}
	}
}

func BenchmarkSelfIdentity(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := SelfIdentity()
		if err != nil {
			b.Fatalf("SelfIdentity: %v", err)
		}
	}
}

func BenchmarkFormatBytes(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = FormatBytes(1073741824)
	}
}

func BenchmarkGoHeapAlloc(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = GoHeapAlloc()
	}
}
