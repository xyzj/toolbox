package logger

import "github.com/xyzj/toolbox/pathtool"

// Opt OptLog
type Opt struct {
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
	// DelayWrite bool
	// LogLevel 日志等级
	// LogLevel LogLevel
}

type Opts func(opt *Opt)

func OptFilename(name string) Opts {
	return func(o *Opt) {
		o.Filename = name
	}
}

func OptFileDir(name string) Opts {
	if name == "" || name == "." {
		name = pathtool.GetExecDir()
	}
	return func(o *Opt) {
		o.FileDir = name
	}
}

func OptFileDays(n int) Opts {
	if n > 0 {
		return func(o *Opt) {
			o.FileDays = n
			o.AutoRoll = true
		}
	}
	return func(o *Opt) {
		o.FileDays = n
	}
}

func OptFileSize(n int64) Opts {
	return func(o *Opt) {
		o.FileSize = max(n, maxFileSize)
		o.AutoRoll = true
	}
}

func OptCompressFile(b bool) Opts {
	return func(o *Opt) {
		o.CompressFile = b
	}
}
