/*
Package logger 日志专用写入器，可设置是否自动依据日期以及文件大小滚动日志文件
*/
package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xyzj/toolbox/crypto"
	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/loopfunc"
	"github.com/xyzj/toolbox/pathtool"
)

const (
	fileTimeFormat = "060102"    // 日志文件命名格式
	maxFileSize    = 1024 * 1024 // 100mb
	// ShortTimeFormat 日志事件戳格式
	ShortTimeFormat = "15:04:05.000 "
	LongTimeFormat  = "Jan 02 15:04:05.000 " // 2006-01-02 15:04:05.000 "
)

var (
	lineEnd = []byte{10}
	comp    = crypto.NewCompressor(crypto.CompressZstd)
)

// OptLog OptLog
type OptLog struct {
	// Filename 日志文件名，不需要扩展名，会自动追加时间戳以及.log扩展名，为空时其他参数无效，仅输出到控制台
	Filename string
	// FileDir 日志存放目录
	FileDir string
	// FileSize 单个日志文件最大大小，AutoRoll==true时有效
	FileSize int64
	// FileDays 日志最大保留天数，AutoRoll==true时有效
	FileDays int
	// AutoRoll 是否自动滚动日志文件，true-依据FileDays和FileSize自动切分日志文件，日志文件名会额外追加日期时间戳‘yymmdd’和序号
	AutoRoll bool
	// CompressFile 是否压缩旧日志文件，AutoRoll==true时有效
	CompressFile bool
	// DelayWrite 延迟写入，每秒检查写入缓存，并写入文件，非实时写入
	DelayWrite bool
	// LogLevel 日志等级
	LogLevel LogLevel
}

func (o *OptLog) ensureDefaults() {
	if o.FileDays == 0 && o.FileSize == 0 {
		o.AutoRoll = false
	} else {
		o.AutoRoll = true
	}
	if o.FileSize < maxFileSize {
		o.FileSize = maxFileSize
	}
}

func NewConsoleWriter() io.Writer {
	return &Writer{
		timeFormat: LongTimeFormat,
		out:        os.Stdout,
	}
}

// NewWriter 一个新的log写入器
//
// opt: 日志写入器配置
func NewWriter(opt *OptLog) io.Writer {
	if opt == nil || opt.Filename == "" {
		return NewConsoleWriter()
	}
	opt.ensureDefaults()
	t := time.Now()
	mylog := &Writer{
		out:          os.Stdout,
		expired:      int64(opt.FileDays)*24*60*60 - 10,
		fileFileSize: opt.FileSize,
		fname:        opt.Filename,
		rollfile:     opt.AutoRoll,
		fileDay:      t.Day(),
		fileHour:     t.Hour(),
		logDir:       opt.FileDir,
		chanGoWrite:  make(chan []byte, 2000),
		enablegz:     opt.CompressFile,
		withFile:     opt.Filename != "",
		delayWrite:   opt.DelayWrite,
		timeFormat:   LongTimeFormat,
	}
	if opt.AutoRoll {
		mylog.timeFormat = ShortTimeFormat
	}
	if opt.Filename != "" && opt.AutoRoll {
		ymd := t.Format(fileTimeFormat)
		for i := byte(255); i > 0; i-- {
			if pathtool.IsExist(filepath.Join(mylog.logDir, fmt.Sprintf("%s.%s.%d.log", mylog.fname, ymd, i))) {
				mylog.fileIndex = i
				break
			}
		}
		// for i := 1; i < 255; i++ {
		// 	if !pathtool.IsExist(filepath.Join(mylog.logDir, fmt.Sprintf("%s.%s.%d.log", mylog.fname, ymd, i))) {
		// 		mylog.fileIndex = byte(i) - 1
		// 		break
		// 	}
		// }
	}
	mylog.newFile()
	mylog.startWrite()
	return mylog
}

// Writer 自定义Writer
type Writer struct {
	chanGoWrite  chan []byte
	out          io.Writer
	fno          *os.File
	pathNow      string
	fname        string
	nameNow      string
	nameOld      string
	logDir       string
	timeFormat   string
	expired      int64
	fileFileSize int64
	fileDay      int
	fileHour     int
	fileIndex    byte
	enablegz     bool
	rollfile     bool
	withFile     bool
	delayWrite   bool
}

// Write 异步写入日志，返回固定为 0, nil
func (w *Writer) Write(p []byte) (n int, err error) {
	xp := json.Bytes(time.Now().Format(w.timeFormat))
	xp = append(xp, p...)
	if !bytes.HasSuffix(xp, lineEnd) {
		xp = append(xp, lineEnd...)
	}
	if w.withFile {
		if w.delayWrite {
			w.chanGoWrite <- xp
		} else {
			w.fno.Write(xp)
		}
	} else {
		w.out.Write(xp)
	}
	return 0, nil
}

func (w *Writer) startWrite() {
	if !w.withFile {
		return
	}
	go loopfunc.LoopFunc(func(params ...interface{}) {
		tc := time.NewTicker(time.Minute * 10)
		tw := time.NewTicker(time.Second)
		buftick := &bytes.Buffer{}
		if !w.delayWrite {
			tw.Stop()
		}
		for {
			select {
			case p := <-w.chanGoWrite:
				// if w.withFile {
				// 	if w.delayWrite {
				buftick.Write(p)
				// } else {
				// 	w.fno.Write(p)
				// }
				// }
				// w.out.Write(buf.Bytes())
			case <-tc.C:
				if w.rollfile {
					w.rollingFileNoLock()
				}
			case <-tw.C:
				if w.withFile {
					if buftick.Len() > 0 {
						w.fno.Write(buftick.Bytes())
						buftick.Reset()
					}
				}
			}
		}
	}, "log writer", nil)
}

// 创建新日志文件
func (w *Writer) newFile() {
	if !w.withFile {
		return
	}
	if w.rollfile {
		t := time.Now()
		if w.fileDay != t.Day() {
			w.fileDay = t.Day()
			w.fileIndex = 0
		}
		w.nameNow = fmt.Sprintf("%s.%s.%d.log", w.fname, t.Format(fileTimeFormat), w.fileIndex)
	} else {
		w.nameNow = fmt.Sprintf("%s.log", w.fname)
	}
	if w.nameOld == w.nameNow {
		return
	}
	// 关闭旧fno
	if w.fno != nil {
		w.fno.Close()
	}
	if !pathtool.IsExist(w.logDir) {
		os.MkdirAll(w.logDir, 0o755)
	}
	w.pathNow = filepath.Join(w.logDir, w.nameNow)
	// 直接写入当日日志
	var err error
	w.fno, err = os.OpenFile(w.pathNow, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o664)
	if err != nil {
		os.WriteFile("logerr.log", []byte("log file open error: "+err.Error()), 0o664)
		w.withFile = false
		return
	}
	w.withFile = true
	// 判断是否压缩旧日志
	if w.enablegz {
		w.zipFile(w.nameOld)
	}
	// w.fno.Write(lineEnd)
}

// 检查文件大小,返回是否需要切分文件
func (w *Writer) rolledWithFileSize() bool {
	// if w.fileHour == time.Now().Hour() {
	// 	return false
	// }
	w.nameOld = w.nameNow
	w.fileHour = time.Now().Hour()
	fs, ex := os.Stat(w.pathNow)
	if ex == nil {
		if fs.Size() > w.fileFileSize {
			if w.fileIndex == 255 {
				w.fileIndex = 0
			} else {
				w.fileIndex++
			}
			return true
		}
	}
	return false
}

func (w *Writer) rollingFileNoLock() bool {
	if time.Now().Day() == w.fileDay && !w.rolledWithFileSize() {
		return false
	}
	// t := time.Now()
	// w.nameNow = fmt.Sprintf("%s.%v.%d.log", w.fname, t.Format(fileTimeFormat), w.fileIndex)
	// // 比对文件名，若不同则重新设置io
	// if w.nameNow == w.nameOld {
	// 	return false
	// }
	// 创建新日志
	w.newFile()
	// 清理旧日志
	w.clearFile()

	return true
}

// 压缩旧日志
func (w *Writer) zipFile(s string) {
	if !w.enablegz || len(s) == 0 {
		return
	}
	if xs := filepath.Join(w.logDir, s); pathtool.IsExist(xs) {
		go func(s string) {
			b, err := os.ReadFile(s)
			if err != nil {
				w.Write([]byte("read log file error: " + s + " " + err.Error()))
				return
			}
			bb, err := comp.Encode(b)
			if err != nil {
				w.Write([]byte("compress log file error: " + s + " " + err.Error()))
				return
			}
			os.WriteFile(s+".zst", bb, 0o664)
			time.Sleep(time.Second * 5)
			os.Remove(s)
		}(xs)
	}
}

// 清理旧日志
func (w *Writer) clearFile() {
	// 若未设置超时，则不清理
	if !w.rollfile || w.expired <= 0 {
		return
	}
	go func() {
		defer func() { recover() }()
		// 遍历文件夹
		lstfno, ex := os.ReadDir(w.logDir)
		if ex != nil {
			w.Write([]byte("clear log files error: " + ex.Error()))
			return
		}
		t := time.Now().Unix()
		for _, d := range lstfno {
			if d.IsDir() { // 忽略目录，不含日志名的文件，以及当前文件
				continue
			}
			fno, err := d.Info()
			if err != nil {
				continue
			}
			if !strings.Contains(fno.Name(), w.fname) {
				continue
			}
			// 比对文件生存期
			if t-fno.ModTime().Unix() >= w.expired {
				os.Remove(filepath.Join(w.logDir, fno.Name()))
			}
		}
	}()
}
