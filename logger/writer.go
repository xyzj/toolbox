package logger

import (
	"bufio"
	"bytes"
	"compress/flate"
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/loopfunc"
	"github.com/xyzj/toolbox/pathtool"
)

// type logData struct {
// 	t time.Time
// 	d []byte
// }

// func (l *logData) Bytes(timeformat string) []byte {
// 	xp := json.Bytes(l.t.Format(timeformat))
// 	xp = append(xp, l.d...)
// 	if !bytes.HasSuffix(xp, lineEnd) {
// 		xp = append(xp, lineEnd...)
// 	}
// 	return xp
// }

type Writer struct {
	cnf         *writerOpt
	buff        *bufio.Writer
	fno         *os.File
	ctxClose    context.Context
	ctxCancel   context.CancelFunc
	currentSize atomic.Int64
	locker      sync.Mutex
	chanWorker  chan *[]byte
	closed      atomic.Bool
}

func (w *Writer) formatdata(data []byte) *[]byte {
	xp := make([]byte, 0, len(data)+len(w.cnf.timeformat)+1)
	xp = append(xp, json.Bytes(time.Now().Format(w.cnf.timeformat))...)
	xp = append(xp, data...)
	if !bytes.HasSuffix(xp, lineEnd) {
		xp = append(xp, lineEnd...)
	}
	return &xp
}

func (w *Writer) openfile() error {
	if w.cnf.filename == "" {
		return errors.New("filename not specified")
	}
	var err error
	w.fno, err = os.OpenFile(w.cnf.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o664)
	if err != nil {
		return err
	}
	info, err := w.fno.Stat()
	if err != nil {
		return err
	}
	w.currentSize.Store(int64(info.Size()))
	w.buff.Reset(w.fno)
	return nil
}

// Close closes the log writer and releases any associated resources.
func (w *Writer) Close() error {
	w.locker.Lock()
	defer w.locker.Unlock()
	w.closed.Store(true)
	if w.cnf.filename == "" {
		return nil
	}
	w.ctxCancel()
	close(w.chanWorker)
	if w.buff.Buffered() > 0 {
		w.buff.Flush()
	}
	if w.fno != nil {
		return w.fno.Close()
	}
	return nil
}

// Write writes the given byte slice to the log writer.
func (w *Writer) Write(b []byte) (n int, err error) {
	w.locker.Lock()
	defer w.locker.Unlock()
	if w.closed.Load() {
		return 0, nil
	}
	if len(b) == 0 {
		return 0, nil
	}
	if w.cnf.filename == "" {
		return w.fno.Write(*w.formatdata(b))
	}
	w.chanWorker <- w.formatdata(b)
	return len(b), nil
}

// NewWriter creates a new log writer with the given options.
func NewWriter(opts ...writerOpts) io.Writer {
	opt := defaultWriterOpts()
	for _, o := range opts {
		o(opt)
	}
	if opt.filename == "" {
		return NewConsoleWriter()
	}
	w := &Writer{
		cnf:         opt,
		currentSize: atomic.Int64{},
		locker:      sync.Mutex{},
		closed:      atomic.Bool{},
		buff:        bufio.NewWriterSize(io.Discard, opt.bufferSize),
	}
	if err := w.openfile(); err != nil {
		return NewConsoleWriter()
	}
	ctxClose, ctxCancel := context.WithCancel(context.Background())
	w.ctxClose = ctxClose
	w.ctxCancel = ctxCancel
	w.chanWorker = make(chan *[]byte, 100)
	go loopfunc.LoopFunc(func(params ...any) {
		n := 0
		t := time.NewTicker(time.Minute)
		if opt.bufferDisabled {
			t.Stop()
		}
		for {
			select {
			case <-w.ctxClose.Done():
				return
			case <-t.C:
				if w.buff.Buffered() > 0 {
					w.buff.Flush()
				}
			case ld := <-w.chanWorker:
				if w.cnf.bufferDisabled {
					n, _ = w.fno.Write(*ld)
				} else {
					n, _ = w.buff.Write(*ld)
				}
				// 如果是文件输出，执行文件相关操作
				if w.cnf.maxsize == 0 {
					continue
				}
				w.currentSize.Add(int64(n))
				if w.currentSize.Load() >= w.cnf.maxsize {
					if w.buff.Buffered() > 0 {
						w.buff.Flush()
					}
					w.fno.Close()
					oldfile := w.cnf.filename + "." + time.Now().Format(w.cnf.backupformat)
					os.Rename(w.cnf.filename, oldfile)
					if err := w.openfile(); err != nil {
						panic(err)
					}
					go func(oldfile string) {
						switch opt.compress {
						case CompressSnappy:
							compressFileSnappy(oldfile)
						case CompressZstd:
							compressFileZstd(oldfile, zstd.SpeedFastest)
						case CompressGzip:
							compressFileGzip(oldfile, flate.BestSpeed)
						}
						if opt.maxbackups == 0 && opt.maxdays == 0 {
							return
						}
						// 定期清理过期日志
						olderthen := time.Time{}
						files, _ := pathtool.SearchFilesByTime(opt.dir, opt.file)
						if len(files) < 2 {
							return
						}
						delfiles := make([]string, 0, len(files))
						keepfiles := make([]string, 0, len(files))
						if opt.maxdays > 0 {
							olderthen = time.Now().AddDate(0, 0, -opt.maxdays)
						}
						for _, f := range files {
							if f.Path == opt.filename {
								continue
							}
							if !olderthen.IsZero() && f.ModTime.Before(olderthen) {
								delfiles = append(delfiles, f.Path)
								continue
							}
							keepfiles = append(keepfiles, f.Path)
						}
						if l := len(keepfiles); l > opt.maxbackups {
							delfiles = append(delfiles, keepfiles[:l-opt.maxbackups]...)
						}
						if len(delfiles) > 0 {
							for _, f := range delfiles {
								os.Remove(f)
							}
						}
					}(oldfile)
				}
			}
		}
	}, "log writer", os.Stdout)
	return w
}

func NewConsoleWriter() io.Writer {
	o := defaultWriterOpts()
	return &Writer{
		fno:         os.Stdout,
		cnf:         o,
		currentSize: atomic.Int64{},
		locker:      sync.Mutex{},
		closed:      atomic.Bool{},
	}
}
