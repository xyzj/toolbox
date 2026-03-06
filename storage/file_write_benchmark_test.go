package storage

import (
	"bufio"
	"io"
	"testing"
	"unsafe"
)

func writeLineString(w *bufio.Writer, line string) error {
	if _, err := w.WriteString(line); err != nil {
		return err
	}
	return w.WriteByte('\n')
}

func writeLineAppendCopy(w *bufio.Writer, line string) error {
	_, err := w.Write(append([]byte(line), '\n'))
	return err
}

func unsafeBytesFromString(line string) []byte {
	return unsafe.Slice(unsafe.StringData(line), len(line))
}

func writeLineUnsafe(w *bufio.Writer, line string) error {
	if _, err := w.Write(unsafeBytesFromString(line)); err != nil {
		return err
	}
	return w.WriteByte('\n')
}

func BenchmarkFileLineWrite(b *testing.B) {
	lines := benchmarkPayload(256)

	b.Run("WriteString+WriteByte", func(b *testing.B) {
		w := bufio.NewWriterSize(io.Discard, 64*1024)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, line := range lines {
				if err := writeLineString(w, line); err != nil {
					b.Fatal(err)
				}
			}
			if err := w.Flush(); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("append-copy", func(b *testing.B) {
		w := bufio.NewWriterSize(io.Discard, 64*1024)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, line := range lines {
				if err := writeLineAppendCopy(w, line); err != nil {
					b.Fatal(err)
				}
			}
			if err := w.Flush(); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("unsafe-zero-copy", func(b *testing.B) {
		w := bufio.NewWriterSize(io.Discard, 64*1024)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, line := range lines {
				if err := writeLineUnsafe(w, line); err != nil {
					b.Fatal(err)
				}
			}
			if err := w.Flush(); err != nil {
				b.Fatal(err)
			}
		}
	})
}
