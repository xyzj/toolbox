package crypto

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
)

// CompressType 压缩编码类型
type CompressType byte

const (
	CompressZlib CompressType = iota
	CompressGZip
	CompressZstd
)

type zstdEnc struct {
	buf   *bytes.Buffer
	in    *bytes.Reader
	coder *zstd.Encoder
}

func (e *zstdEnc) Encode(src []byte) ([]byte, error) {
	e.buf.Reset()
	e.in.Reset(src)
	e.coder.Reset(e.buf)
	_, err := io.Copy(e.coder, e.in)
	if err != nil {
		e.coder.Close()
		return nil, err
	}
	e.coder.Close()
	return e.buf.Bytes(), nil
}

type zstdDec struct {
	buf   *bytes.Buffer
	in    *bytes.Reader
	coder *zstd.Decoder
}

func (e *zstdDec) Decode(src []byte) ([]byte, error) {
	e.buf.Reset()
	e.in.Reset(src)
	e.coder.Reset(e.in)
	// _, err := io.Copy(e.buf, e.coder)
	// if err != nil {
	// 	// e.coder.Close()
	// 	return nil, err
	// }
	// e.coder.Close()
	e.coder.WriteTo(e.buf)
	return e.buf.Bytes(), nil
}

type gzipEnc struct {
	buf   *bytes.Buffer
	in    *bytes.Reader
	coder *gzip.Writer
}

func (e *gzipEnc) Encode(src []byte) ([]byte, error) {
	e.buf.Reset()
	e.in.Reset(src)
	e.coder.Reset(e.buf)
	_, err := io.Copy(e.coder, e.in)
	if err != nil {
		e.coder.Close()
		return nil, err
	}
	e.coder.Close()
	return e.buf.Bytes(), nil
}

type gzipDec struct {
	buf   *bytes.Buffer
	in    *bytes.Reader
	coder *gzip.Reader
}

func (e *gzipDec) Decode(src []byte) ([]byte, error) {
	e.buf.Reset()
	e.in.Reset(src)
	e.coder.Reset(e.in)
	_, err := io.Copy(e.buf, e.coder)
	if err != nil {
		e.coder.Close()
		return nil, err
	}
	e.coder.Close()
	return e.buf.Bytes(), nil
}

type zlibEnc struct {
	buf   *bytes.Buffer
	in    *bytes.Reader
	coder *zlib.Writer
}

func (e *zlibEnc) Encode(src []byte) ([]byte, error) {
	e.buf.Reset()
	e.in.Reset(src)
	e.coder.Reset(e.buf)
	_, err := io.Copy(e.coder, e.in)
	if err != nil {
		e.coder.Close()
		return nil, err
	}
	e.coder.Close()
	return e.buf.Bytes(), nil
}

type zlibDec struct {
	buf   *bytes.Buffer
	in    *bytes.Reader
	coder io.ReadCloser
}

func (e *zlibDec) Decode(src []byte) ([]byte, error) {
	e.buf.Reset()
	e.in.Reset(src)
	var err error
	e.coder, err = zlib.NewReader(e.in)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(e.buf, e.coder)
	if err != nil {
		e.coder.Close()
		return nil, err
	}
	e.coder.Close()
	return e.buf.Bytes(), nil
}

type Compressor struct {
	t       CompressType
	encpool sync.Pool
	decpool sync.Pool
}

func (z *Compressor) Encode(src []byte) ([]byte, error) {
	tool := z.encpool.Get()
	defer z.encpool.Put(tool)
	switch z.t {
	case CompressGZip:
		return tool.(*gzipEnc).Encode(src)
	case CompressZlib:
		return tool.(*zlibEnc).Encode(src)
	default:
		return tool.(*zstdEnc).Encode(src)
	}
}

func (z *Compressor) Decode(src []byte) ([]byte, error) {
	tool := z.decpool.Get()
	defer z.decpool.Put(tool)
	switch z.t {
	case CompressGZip:
		return tool.(*gzipDec).Decode(src)
	case CompressZlib:
		return tool.(*zlibDec).Decode(src)
	default:
		return tool.(*zstdDec).Decode(src)
	}
}

func NewCompressor(t CompressType) *Compressor {
	var encnew func() any
	var decnew func() any
	switch t {
	case CompressGZip:
		encnew = func() any {
			return &gzipEnc{
				buf:   &bytes.Buffer{},
				in:    bytes.NewReader([]byte{}),
				coder: gzip.NewWriter(nil),
			}
		}
		decnew = func() any {
			return &gzipDec{
				buf:   &bytes.Buffer{},
				in:    bytes.NewReader([]byte{}),
				coder: new(gzip.Reader),
			}
		}
	case CompressZlib:
		encnew = func() any {
			return &zlibEnc{
				buf:   &bytes.Buffer{},
				in:    bytes.NewReader([]byte{}),
				coder: zlib.NewWriter(nil),
			}
		}
		decnew = func() any {
			dec, _ := zlib.NewReader(bytes.NewReader([]byte{}))
			return &zlibDec{
				buf:   &bytes.Buffer{},
				in:    bytes.NewReader([]byte{}),
				coder: dec,
			}
		}
	case CompressZstd: // zstd
		encnew = func() any {
			enc, _ := zstd.NewWriter(nil)
			return &zstdEnc{
				buf:   &bytes.Buffer{},
				in:    bytes.NewReader([]byte{}),
				coder: enc,
			}
		}
		decnew = func() any {
			dec, _ := zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
			return &zstdDec{
				buf:   &bytes.Buffer{},
				in:    bytes.NewReader([]byte{}),
				coder: dec,
			}
		}
	}
	return &Compressor{
		t: t,
		encpool: sync.Pool{
			New: encnew,
		},
		decpool: sync.Pool{
			New: decnew,
		},
	}
}
