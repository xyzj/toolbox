package logger

import (
	"path/filepath"
	"strings"

	"github.com/xyzj/toolbox/pathtool"
)

const (
	fileTimeFormat  = "0102-1504"           // 日志文件命名格式
	maxFileSize     = 1024 * 1024 * 1000    // 1000mb
	ShortTimeFormat = "15:04:05.000 "       // ShortTimeFormat 日志事件戳格式
	LongTimeFormat  = "Jan02 15:04:05.000 " // 2006-01-02 15:04:05.000 "
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
	compress     CompressMethod // compress 压缩方式， none-不压缩, snappy-snappy压缩, zstd-zstd压缩, gzip-gzip压缩
	filename     string         // filename 日志文件名，不需要扩展名，会自动追加时间戳以及.log扩展名，为空时其他参数无效，仅输出到控制台
	timeformat   string         // timeformat 日志时间戳格式
	maxsize      int64          // maxsize 日志文件最大大小，超过后会自动切割
	maxdays      int            // maxdays 日志文件最大保存天数
	maxbackups   int            // maxbackups 最大备份数量
	backupformat string         // backupformat 备份文件时间戳格式
	dir          string         // dir 日志文件存放目录
	file         string         // file 日志文件名，不包含路径
}

type writerOpts func(opt *writerOpt)

func WithMaxDays(days int) writerOpts {
	return func(o *writerOpt) {
		o.maxdays = days
	}
}

func WithCompressMethod(method CompressMethod) writerOpts {
	return func(o *writerOpt) {
		o.compress = method
	}
}

func WithLogTimeFormater(format string) writerOpts {
	return func(o *writerOpt) {
		o.timeformat = format
	}
}

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

func WithMaxSize(size int64) writerOpts {
	return func(o *writerOpt) {
		o.maxsize = size
	}
}

func WithMaxBackups(backups int) writerOpts {
	return func(o *writerOpt) {
		o.maxbackups = backups
	}
}

func WithBackupFormat(format string) writerOpts {
	return func(o *writerOpt) {
		o.backupformat = format
	}
}

func defaultWriterOpts() *writerOpt {
	return &writerOpt{
		compress:     CompressNone,
		filename:     "",
		timeformat:   LongTimeFormat,
		maxsize:      maxFileSize,
		maxdays:      0,
		maxbackups:   10,
		dir:          pathtool.GetExecDir(),
		backupformat: fileTimeFormat,
	}
}
