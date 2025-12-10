package logger

import (
	"path/filepath"
	"strings"

	"github.com/xyzj/toolbox/pathtool"
)

const (
	defaultFileSize   = 1024 * 1024 * 1000    // 1000mb
	defaultBufferSize = 1024 * 4              // 4kb
	fileTimeFormat    = "Jan021504"           // 日志文件命名格式
	ShortTimeFormat   = "15:04:05.000 "       // ShortTimeFormat 日志事件戳格式
	LongTimeFormat    = "Jan02 15:04:05.000 " // 2006-01-02 15:04:05.000 "
)

var lineEnd = []byte{10}

type CompressMethod string

const (
	CompressNone   CompressMethod = "none"
	CompressSnappy CompressMethod = "snappy"
	CompressZstd   CompressMethod = "zstd"
	CompressGzip   CompressMethod = "gzip"
)

type writerOpt struct {
	compress       CompressMethod // compress 压缩方式， none-不压缩, snappy-snappy压缩, zstd-zstd压缩, gzip-gzip压缩
	filename       string         // filename 日志文件名，不需要扩展名，会自动追加时间戳以及.log扩展名，为空时其他参数无效，仅输出到控制台
	timeformat     string         // timeformat 日志时间戳格式
	backupformat   string         // backupformat 备份文件时间戳格式
	dir            string         // dir 日志文件存放目录
	file           string         // file 日志文件名，不包含路径
	maxsize        int64          // maxsize 日志文件最大大小，超过后会自动切割
	bufferSize     int            // bufferSize 日志缓冲区大小
	maxdays        int            // maxdays 日志文件最大保存天数
	maxbackups     int            // maxbackups 最大备份数量
	bufferDisabled bool           // bufferDisabled 是否禁用缓冲区，禁用后每次写入都会直接写入文件，影响性能但不易丢失日志
}

type writerOpts func(opt *writerOpt)

// WithBufferSize sets the buffer size for the logger.
// If the size is less than 0, it will be set to 0.
// If the size is greater than 1024*1024, it will be set to 1024*1024.
// If the size is 0, it will not use a buffer.
// The default buffer size is 4KB.
func WithBufferSize(size int) writerOpts {
	return func(o *writerOpt) {
		if size == 0 {
			o.bufferDisabled = true
		} else {
			o.bufferSize = min(max(size, 0), 1024*1024)
		}
	}
}

// WithMaxDays sets the maximum number of days to keep log files.
// If the value is less than 0, it will be set to 0.
// If the value is greater than 365, it will be set to 365.
// The default value is 0, which means no limit.
func WithMaxDays(days int) writerOpts {
	return func(o *writerOpt) {
		o.maxdays = min(max(days, 0), 365)
	}
}

// WithCompressMethod sets the compression method for the logger.
func WithCompressMethod(method CompressMethod) writerOpts {
	return func(o *writerOpt) {
		o.compress = method
	}
}

// WithLogTimeFormat sets the time format for the logger.
func WithLogTimeFormat(format string) writerOpts {
	return func(o *writerOpt) {
		o.timeformat = format
	}
}

// WithFilename sets the filename for the logger.
func WithFilename(name string) writerOpts {
	return func(o *writerOpt) {
		o.dir, o.file, _ = pathtool.EnsureDirAndSplit(name)
		if o.file == "" {
			o.filename = ""
			return
		}
		if !strings.HasSuffix(o.file, ".log") {
			o.file += ".log"
		}
		o.filename = filepath.Join(o.dir, o.file)
	}
}

// WithMaxSize sets the maximum size of the log file.
// If the size is less than or equal to 0, it will be set to 0, which means no limit.
func WithMaxSize(size int64) writerOpts {
	return func(o *writerOpt) {
		o.maxsize = max(size, 0)
	}
}

// WithMaxBackups sets the maximum number of backup log files to keep.
func WithMaxBackups(backups int) writerOpts {
	return func(o *writerOpt) {
		o.maxbackups = max(backups, 0)
	}
}

// WithBackupFormat sets the backup file name format.
func WithBackupFormat(format string) writerOpts {
	return func(o *writerOpt) {
		o.backupformat = format
	}
}

func defaultWriterOpts() *writerOpt {
	return &writerOpt{
		filename:     "",
		compress:     CompressNone,
		backupformat: fileTimeFormat,
		timeformat:   LongTimeFormat,
		bufferSize:   defaultBufferSize,
		maxsize:      defaultFileSize,
		maxdays:      0,
		maxbackups:   10,
		dir:          pathtool.GetExecDir(),
		file:         "",
	}
}
