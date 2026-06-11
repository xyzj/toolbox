package main

import (
	"strings"
	"testing"
)

var (
	benchSinkBytes  []byte
	benchSinkString string
)

func benchmarkSourceString(size int) string {
	if size <= 0 {
		return ""
	}
	return strings.Repeat("abcdef0123456789", size/16+1)[:size]
}

func BenchmarkStringToBytes(b *testing.B) {
	cases := []struct {
		name  string
		size  int
		sconv func(string) []byte
	}{
		{name: "raw/small", size: 32, sconv: toRawBytes},
		{name: "reflect/small", size: 32, sconv: toReflectBytes},
		{name: "unsafe-slice/small", size: 32, sconv: toBytes},
		{name: "k8s/small", size: 32, sconv: toK8sBytes},
		{name: "raw/large", size: 4096, sconv: toRawBytes},
		{name: "reflect/large", size: 4096, sconv: toReflectBytes},
		{name: "unsafe-slice/large", size: 4096, sconv: toBytes},
		{name: "k8s/large", size: 4096, sconv: toK8sBytes},
	}

	for _, tc := range cases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			src := benchmarkSourceString(tc.size)
			b.SetBytes(int64(len(src)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchSinkBytes = tc.sconv(src)
			}
		})
	}
}

func BenchmarkBytesToString(b *testing.B) {
	cases := []struct {
		name  string
		size  int
		bconv func([]byte) string
	}{
		{name: "raw/small", size: 32, bconv: toRawString},
		{name: "reflect/small", size: 32, bconv: toReflectString},
		{name: "unsafe-string/small", size: 32, bconv: toString},
		{name: "k8s/small", size: 32, bconv: toK8sString},
		{name: "raw/large", size: 4096, bconv: toRawString},
		{name: "reflect/large", size: 4096, bconv: toReflectString},
		{name: "unsafe-string/large", size: 4096, bconv: toString},
		{name: "k8s/large", size: 4096, bconv: toK8sString},
	}

	for _, tc := range cases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			src := []byte(benchmarkSourceString(tc.size))
			b.SetBytes(int64(len(src)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchSinkString = tc.bconv(src)
			}
		})
	}
}
