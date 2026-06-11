package bench

import (
	"crypto/sha1"
	"encoding/hex"
	"hash/crc32"
	"testing"
)

var benchPayload = []byte("the quick brown fox jumps over the lazy dog 1234567890")

func BenchmarkCRC32Hex(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(benchPayload)))
	var sumBuf [4]byte
	for i := 0; i < b.N; i++ {
		v := crc32.ChecksumIEEE(benchPayload)
		sumBuf[0] = byte(v >> 24)
		sumBuf[1] = byte(v >> 16)
		sumBuf[2] = byte(v >> 8)
		sumBuf[3] = byte(v)
		_ = hex.EncodeToString(sumBuf[:])
	}
}

func BenchmarkSHA1Hex(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(benchPayload)))
	for i := 0; i < b.N; i++ {
		sum := sha1.Sum(benchPayload)
		_ = hex.EncodeToString(sum[:])
	}
}
