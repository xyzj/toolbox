package logger

import (
	"io"

	"github.com/xyzj/toolbox/json"
)

type LogLevel byte

const (
	LogDebug            LogLevel = 10
	LogInfo             LogLevel = 20
	LogWarning          LogLevel = 30
	LogError            LogLevel = 40
	LogSystem           LogLevel = 90
	logformater                  = "%s %s"
	logformaterWithName          = "%s [%s] %s"
)

// Logger 日志接口
type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warning(msg string)
	Error(msg string)
	System(msg string)
	DefaultWriter() io.Writer
	SetLevel(l LogLevel)
}

// NilLogger 空日志
type NilLogger struct{}

// Debug Debug
func (l *NilLogger) Debug(msg string) {}

// Info Info
func (l *NilLogger) Info(msg string) {}

// Warning Warning
func (l *NilLogger) Warning(msg string) {}

// Error Error
func (l *NilLogger) Error(msg string) {}

// System System
func (l *NilLogger) System(msg string) {}

// DefaultWriter 返回日志Writer
func (l *NilLogger) DefaultWriter() io.Writer { return io.Discard }
func (l *NilLogger) SetLevel(LogLevel)        {}

// StdLogger mx log
type StdLogger struct {
	out      io.Writer
	logLevel LogLevel
}

// Debug Debug
func (l *StdLogger) Debug(msg string) {
	if LogDebug >= l.logLevel {
		l.out.Write(json.Bytes(msg))
	}
}

// Info Info
func (l *StdLogger) Info(msg string) {
	if LogInfo >= l.logLevel {
		l.out.Write(json.Bytes(msg))
	}
}

// Warning Warning
func (l *StdLogger) Warning(msg string) {
	if LogWarning >= l.logLevel {
		l.out.Write(json.Bytes(msg))
	}
}

// Error Error
func (l *StdLogger) Error(msg string) {
	if LogError >= l.logLevel {
		l.out.Write(json.Bytes(msg))
	}
}

// System System
func (l *StdLogger) System(msg string) {
	if LogSystem >= l.logLevel {
		l.out.Write(json.Bytes(msg))
	}
}

// DefaultWriter 返回日志Writer
func (l *StdLogger) DefaultWriter() io.Writer {
	return l.out
}

func (l *StdLogger) SetLevel(ll LogLevel) {
	l.logLevel = ll
}

// NewLogger init logger
//
// d: 日志保存路径
//
// f:日志文件名，为空时仅输出到控制台
//
// logLevel 日志等级，1-输出到控制台，10-debug（输出到控制台和文件）,20-info（输出到文件）,30-warning（输出到文件）,40-error（输出到控制台和文件）,90-system（输出到控制台和文件）
//
// logDays 日志文件保留天数
//
// delayWrite 是否延迟写入，在日志密集时，可减少磁盘io，但可能导致日志丢失
func NewLogger(l LogLevel, opts ...Options) Logger {
	return &StdLogger{
		out:      NewWriter(opts...),
		logLevel: l,
	}
	// return &MultiLogger{
	// 	outs: []*StdLogger{
	// 		{
	// 			logLevel: opt.LogLevel,
	// 			out:      NewWriter(opt),
	// 		},
	// 	},
	// }
}

// NewConsoleLogger 返回一个纯控制台日志输出器
func NewConsoleLogger() Logger {
	return &StdLogger{
		out:      NewConsoleWriter(),
		logLevel: LogDebug,
	}
	// return &MultiLogger{
	// 	outs: []*StdLogger{
	// 		{
	// 			logLevel: LogDebug,
	// 			out:      NewConsoleWriter(),
	// 		},
	// 	},
	// }
}

func NewConsoleLoggerWithLevel(l LogLevel) Logger {
	return &StdLogger{
		out:      NewConsoleWriter(),
		logLevel: l,
	}
	// return &MultiLogger{
	// 	outs: []*StdLogger{
	// 		{
	// 			logLevel: LogDebug,
	// 			out:      NewConsoleWriter(),
	// 		},
	// 	},
	// }
}

func NewNilLogger() Logger {
	return &NilLogger{}
}
