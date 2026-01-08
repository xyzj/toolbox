package logger

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/loopfunc"
	"github.com/xyzj/toolbox/pathtool"
)

func writePanic(f, s string) {
	ff, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o664)
	if err != nil {
		os.WriteFile(filepath.Join(os.TempDir(), filepath.Base(f)), []byte("open panic log file error:"+err.Error()), 0o664)
		println("open panic log file error:", err.Error())
		return
	}
	ff.WriteString(s)
	ff.Close()
}

type Writer struct {
	cnf         *writerOpt
	buff        *bufio.Writer
	fno         *os.File
	ctxClose    context.Context
	ctxCancel   context.CancelFunc
	currentSize atomic.Int64
	locker      sync.Mutex
	closed      atomic.Bool
	n           int
	err         error
}

func (w *Writer) formatdata(data []byte) *[]byte {
	xp := make([]byte, 0, len(data)+len(w.cnf.timeformat)+1)
	if w.cnf.timeformat != "" {
		xp = append(xp, json.Bytes(time.Now().Format(w.cnf.timeformat))...)
	}
	xp = append(xp, data...)
	if !bytes.HasSuffix(xp, lineEnd) {
		xp = append(xp, lineEnd...)
	}
	return &xp
}
func (w *Writer) compressAndClean(oldfile string) {
	defer func() {
		if r := recover(); r != nil {
			writePanic(pathtool.JoinPathFromHere(w.cnf.file+".err"),
				"compress old log file panic:"+r.(error).Error())
		}
	}()
	var err error
	_, err = pathtool.CompressFile(oldfile, w.cnf.compress)
	if err != nil {
		writePanic(pathtool.JoinPathFromHere(w.cnf.file+".err"),
			"compress old log file error:"+err.Error())
	} else {
		os.Remove(oldfile)
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
}
func (w *Writer) openfile(trunc bool) error {
	if w.cnf.filename == "" {
		return errors.New("filename not specified")
	}
	w.currentSize.Store(0)
	var err error
	flags := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	if trunc {
		flags = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	}
	w.fno, err = os.OpenFile(w.cnf.filename, flags, 0o664)
	if err != nil {
		w.buff.Reset(os.Stdout)
		writePanic(pathtool.JoinPathFromHere(w.cnf.file+".err"),
			"open log file error, write to console:"+err.Error())
		return err
	}
	info, err := w.fno.Stat()
	if err == nil {
		w.currentSize.Store(info.Size())
	}
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
	if w.fno == nil {
		return 0, errors.New("log file not opened")
	}
	if w.cnf.filename == "" { // console writer
		return w.fno.Write(*w.formatdata(b))
	}
	w.n, w.err = w.buff.Write(*w.formatdata(b))
	if w.n > 0 && w.cnf.maxsize > 0 {
		w.currentSize.Add(int64(w.n))
	}
	return w.n, w.err
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
	var err error
	if err = w.openfile(false); err != nil {
		panic(err)
	}
	ctxClose, ctxCancel := context.WithCancel(context.Background())
	w.ctxClose = ctxClose
	w.ctxCancel = ctxCancel
	go loopfunc.LoopFunc(func(params ...any) {
		t := time.NewTicker(time.Minute)
		for {
			select {
			case <-w.ctxClose.Done():
				return
			case <-t.C:
				w.locker.Lock()
				if n := w.buff.Buffered(); n > 0 {
					w.buff.Flush()
				}
				if w.cnf.maxsize > 0 && w.currentSize.Load() >= w.cnf.maxsize {
					if w.fno != nil {
						w.fno.Close()
					}
					oldfile := filepath.Join(w.cnf.dir, strings.TrimSuffix(w.cnf.file, ".log")+"."+time.Now().Format(w.cnf.backupformat)+".log")
					err = os.Rename(w.cnf.filename, oldfile)
					if err != nil {
						writePanic(pathtool.JoinPathFromHere(w.cnf.file+".err"),
							"rename log file error:"+err.Error())
						// Should I rename the file by copying it?
					} else {
						// 压缩旧的，并清理
						go w.compressAndClean(oldfile)
					}
					w.openfile(false)
					if w.cnf.rollmsg != "" {
						w.buff.WriteString(w.cnf.rollmsg + "\n")
					}
				}
				w.locker.Unlock()
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
