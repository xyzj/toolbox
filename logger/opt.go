package logger

import "github.com/xyzj/toolbox/pathtool"

// opt OptLog
type opt struct {
	// filename 日志文件名，不需要扩展名，会自动追加时间戳以及.log扩展名，为空时其他参数无效，仅输出到控制台
	filename string
	// filedir 日志存放目录
	filedir string
	// filesize 单个日志文件最大大小，AutoRoll==true时有效
	filesize int64
	// filedays 日志最大保留天数，AutoRoll==true时有效
	filedays int
	// autoroll 是否自动滚动日志文件，true-依据FileDays和FileSize自动切分日志文件，日志文件名会额外追加日期时间戳‘yymmdd’和序号
	autoroll bool
	// compressfile 是否压缩旧日志文件，AutoRoll==true时有效
	compressfile bool
	// DelayWrite 延迟写入，每秒检查写入缓存，并写入文件，非实时写入
	// DelayWrite bool
	// LogLevel 日志等级
	// LogLevel LogLevel
}

type Options func(opt *opt)

func WithFilename(name string) Options {
	return func(o *opt) {
		o.filename = name
	}
}

func WithFileDir(name string) Options {
	if name == "" || name == "." {
		name = pathtool.GetExecDir()
	}
	return func(o *opt) {
		o.filedir = name
	}
}

func WithFileDays(n int) Options {
	if n > 0 {
		return func(o *opt) {
			o.filedays = n
			o.autoroll = true
		}
	}
	return func(o *opt) {
		o.filedays = n
	}
}

func WithFileSize(n int64) Options {
	return func(o *opt) {
		o.filesize = max(n, maxFileSize)
		o.autoroll = true
	}
}

func WithCompressFile(b bool) Options {
	return func(o *opt) {
		o.compressfile = b
	}
}
