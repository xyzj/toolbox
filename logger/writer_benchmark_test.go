package logger

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/xyzj/toolbox/json"
)

func oldFormatData(data []byte, timeformat string) []byte {
	xp := make([]byte, 0, len(data)+len(timeformat)+1)
	if timeformat != "" {
		xp = append(xp, json.Bytes(time.Now().Format(timeformat))...)
	}
	xp = append(xp, data...)
	if !bytes.HasSuffix(xp, lineEnd) {
		xp = append(xp, lineEnd...)
	}
	return xp
}

func newSplitWrite(w *bufio.Writer, data []byte, timeformat string) error {
	if timeformat != "" {
		var tbuf [96]byte
		timePrefix := time.Now().AppendFormat(tbuf[:0], timeformat)
		if _, err := w.Write(timePrefix); err != nil {
			return err
		}
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if data[len(data)-1] != lineEnd[0] {
		if err := w.WriteByte(lineEnd[0]); err != nil {
			return err
		}
	}
	return nil
}

func BenchmarkWriterFormatStrategy(b *testing.B) {
	payload := []byte("this is a benchmark log line payload")
	timefmt := "2006-01-02 15:04:05 "

	b.Run("old-pack-and-write", func(b *testing.B) {
		w := bufio.NewWriterSize(io.Discard, 32*1024)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := w.Write(oldFormatData(payload, timefmt)); err != nil {
				b.Fatal(err)
			}
			if err := w.Flush(); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("new-split-write", func(b *testing.B) {
		w := bufio.NewWriterSize(io.Discard, 32*1024)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := newSplitWrite(w, payload, timefmt); err != nil {
				b.Fatal(err)
			}
			if err := w.Flush(); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func newBenchmarkWriter(timeFormat string) *Writer {
	return &Writer{
		cnf: &writerOpt{
			timeformat: timeFormat,
			maxsize:    0,
		},
		buff:                    bufio.NewWriterSize(io.Discard, 32*1024),
		fno:                     os.Stdout,
		closed:                  false,
		cacheTimePrefixBySecond: canCacheTimePrefixBySecond(timeFormat),
	}
}

func BenchmarkWriterWriteDirect(b *testing.B) {
	payload := []byte("this is a benchmark log line payload")

	b.Run("timeformat-on", func(b *testing.B) {
		w := newBenchmarkWriter("2006-01-02 15:04:05 ")
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := w.Write(payload); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("timeformat-off", func(b *testing.B) {
		w := newBenchmarkWriter("")
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := w.Write(payload); err != nil {
				b.Fatal(err)
			}
		}
	})
}
