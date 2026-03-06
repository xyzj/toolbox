package storage

import (
	"strconv"
	"testing"
)

func benchmarkPayload(n int) []string {
	payload := make([]string, n)
	for i := 0; i < n; i++ {
		payload[i] = "item-" + strconv.Itoa(i)
	}
	return payload
}

func BenchmarkStorageStore(b *testing.B) {
	payload := benchmarkPayload(64)

	b.Run("memory", func(b *testing.B) {
		s := NewMemory(128)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = s.Store(payload...)
		}
	})

	b.Run("ring", func(b *testing.B) {
		s := NewRing(128)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = s.Store(payload...)
		}
	})
}

func BenchmarkStorageLoad(b *testing.B) {
	payload := benchmarkPayload(128)

	b.Run("memory", func(b *testing.B) {
		s := NewMemory(128)
		_ = s.Store(payload...)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = s.Load()
		}
	})

	b.Run("ring", func(b *testing.B) {
		s := NewRing(128)
		_ = s.Store(payload...)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = s.Load()
		}
	})
}
