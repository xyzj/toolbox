package logger

import (
	"bufio"
	"bytes"
	"compress/flate"
	"context"
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

type logData struct {
	t time.Time
	d []byte
}

func (l *logData) Bytes(timeformat string) []byte {
	xp := json.Bytes(l.t.Format(timeformat))
	xp = append(xp, l.d...)
	if !bytes.HasSuffix(xp, lineEnd) {
		xp = append(xp, lineEnd...)
	}
	return xp
}

type Writer struct {
	cnf         *writerOpt
	buff        *bufio.Writer
	fno         *os.File
	ctxClose    context.Context
	ctxCancel   context.CancelFunc
	currentSize atomic.Int64
	closeOnce   sync.Once
	locker      sync.Mutex
	chanWorker  chan *logData
	closed      atomic.Bool
}

func (w *Writer) openfile() error {
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

func (w *Writer) Close() error {
	w.locker.Lock()
	defer w.locker.Unlock()
	w.closeOnce.Do(func() {
		w.closed.Store(true)
		if w.cnf.filename == "" {
			return
		}
		w.ctxCancel()
		close(w.chanWorker)
		if w.buff.Available() > 0 {
			w.buff.Flush()
		}
		if w.fno != nil {
			w.fno.Close()
			w.fno = nil
			return
		}
	})
	return nil
}

func (w *Writer) Write(b []byte) (n int, err error) {
	w.locker.Lock()
	defer w.locker.Unlock()
	if w.closed.Load() {
		return 0, nil
	}
	l := &logData{
		t: time.Now(),
		d: b,
	}
	if w.cnf.filename == "" {
		return w.fno.Write(l.Bytes(w.cnf.timeformat))
	}
	w.chanWorker <- l
	return len(b), nil
}

func NewWriter(opts ...writerOpts) io.Writer {
	opt := defaultWriterOpts()
	for _, o := range opts {
		o(opt)
	}
	ctxClose, ctxCancel := context.WithCancel(context.Background())
	w := &Writer{
		cnf:         opt,
		chanWorker:  make(chan *logData, 200),
		currentSize: atomic.Int64{},
		buff:        bufio.NewWriterSize(os.Stdout, 8192*4),
		ctxClose:    ctxClose,
		ctxCancel:   ctxCancel,
		locker:      sync.Mutex{},
		closeOnce:   sync.Once{},
		closed:      atomic.Bool{},
	}
	if opt.filename != "" {
		go loopfunc.LoopFunc(func(params ...any) {
			if err := w.openfile(); err != nil {
				panic(err)
			}
			for {
				select {
				case <-w.ctxClose.Done():
					return
				case ld := <-w.chanWorker:
					bs := ld.Bytes(w.cnf.timeformat)
					n, _ := w.buff.Write(bs)
					// 如果是文件输出，执行文件相关操作
					if w.cnf.maxsize > 0 {
						w.currentSize.Add(int64(n))
						if w.currentSize.Load() >= w.cnf.maxsize {
							w.buff.Flush()
							w.fno.Close()
							oldfile := w.cnf.filename + "." + time.Now().Format(w.cnf.backupformat)
							os.Rename(w.cnf.filename, oldfile)
							if err := w.openfile(); err != nil {
								panic(err)
							}
							go func() {
								switch w.cnf.compress {
								case CompressSnappy:
									compressFileSnappy(oldfile)
								case CompressZstd:
									compressFileZstd(oldfile, zstd.SpeedFastest)
								case CompressGzip:
									compressFileGzip(oldfile, flate.BestSpeed)
								}
								if w.cnf.maxbackups == 0 && w.cnf.maxdays == 0 {
									return
								}
								// 定期清理过期日志
								olderthen := time.Time{}
								files, _ := pathtool.SearchFilesByTime(w.cnf.dir, w.cnf.file)
								if len(files) < 2 {
									return
								}
								delfiles := make([]string, 0, len(files))
								keepfiles := make([]string, 0, len(files))
								if w.cnf.maxdays > 0 {
									olderthen = time.Now().AddDate(0, 0, -w.cnf.maxdays)
								}
								for _, f := range files {
									if f.Path == w.cnf.filename {
										continue
									}
									if !olderthen.IsZero() && f.ModTime.Before(olderthen) {
										delfiles = append(delfiles, f.Path)
										continue
									}
									keepfiles = append(keepfiles, f.Path)
								}
								if l := len(keepfiles); l > w.cnf.maxbackups {
									delfiles = append(delfiles, keepfiles[:l-w.cnf.maxbackups]...)
								}
								if len(delfiles) > 0 {
									for _, f := range delfiles {
										os.Remove(f)
									}
								}
							}()
						}
					}
				}
			}
		}, "log writer", os.Stdout)
	}
	return w
}

func NewConsoleWriter() io.Writer {
	o := defaultWriterOpts()
	o.maxsize = 0
	return &Writer{
		fno: os.Stdout,
		cnf: o,
	}
}
