package compressor

import (
	"bytes"
	"compress/gzip"
	"compress/zlib" // 引入 zlib
	"fmt"
	"io"
	"sync"

	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"
)

type Algorithm string

const (
	Gzip   Algorithm = "gzip"
	Zlib   Algorithm = "zlib" // 新增 Zlib
	Snappy Algorithm = "snappy"
	Zstd   Algorithm = "zstd"
)

var (
	// Gzip 池
	gzipWriterPool = sync.Pool{New: func() any { return gzip.NewWriter(nil) }}
	gzipReaderPool = sync.Pool{New: func() any { return new(gzip.Reader) }}

	// Zlib 池
	zlibWriterPool = sync.Pool{New: func() any { return zlib.NewWriter(nil) }}
	// zlib.Reader 比较特殊，它的 Reset 需要指定字典，这里我们通过接口复用
	zlibReaderPool = sync.Pool{New: func() any { return nil }}

	// Zstd 单例 (内部自带池)
	zstdEncoder *zstd.Encoder
	zstdDecoder *zstd.Decoder
)

func init() {
	zstdEncoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedFastest))
	zstdDecoder, _ = zstd.NewReader(nil)
}

// Compress 压缩方法
func Compress(alg Algorithm, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var buf bytes.Buffer
	switch alg {
	case Gzip:
		gw := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gw)
		gw.Reset(&buf)
		gw.Write(data)
		gw.Close()
		return buf.Bytes(), nil

	case Zlib:
		zw := zlibWriterPool.Get().(*zlib.Writer)
		defer zlibWriterPool.Put(zw)
		zw.Reset(&buf) // 复用现有的 writer 并指向新的 buffer
		zw.Write(data)
		zw.Close()
		return buf.Bytes(), nil

	case Snappy:
		return snappy.Encode(nil, data), nil

	case Zstd:
		return zstdEncoder.EncodeAll(data, make([]byte, 0, len(data)/2)), nil

	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", alg)
	}
}

// Decompress 解压方法
func Decompress(alg Algorithm, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	switch alg {
	case Gzip:
		gr := gzipReaderPool.Get().(*gzip.Reader)
		defer gzipReaderPool.Put(gr)
		if err := gr.Reset(bytes.NewReader(data)); err != nil {
			return nil, err
		}
		defer gr.Close()
		return io.ReadAll(gr)

	case Zlib:
		// Zlib Reader 的复用逻辑
		var zr io.ReadCloser
		if iface := zlibReaderPool.Get(); iface != nil {
			zr = iface.(io.ReadCloser)
			// Reset 接口在 zlib 中需要类型断言，因为标准库返回的是私有结构
			zr.(interface {
				Reset(r io.Reader, dict []byte) error
			}).Reset(bytes.NewReader(data), nil)
		} else {
			var err error
			zr, err = zlib.NewReader(bytes.NewReader(data))
			if err != nil {
				return nil, err
			}
		}
		defer zlibReaderPool.Put(zr)
		defer zr.Close()
		return io.ReadAll(zr)

	case Snappy:
		return snappy.Decode(nil, data)

	case Zstd:
		return zstdDecoder.DecodeAll(data, nil)

	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", alg)
	}
}
