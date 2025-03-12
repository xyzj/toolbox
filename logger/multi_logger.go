package logger

import (
	"io"
)

type MultiLogger struct {
	outs []Logger
}

func NewMultiLogger(writers ...Logger) Logger {
	return &MultiLogger{
		outs: writers,
	}
}

// DefaultWriter out
func (l *MultiLogger) DefaultWriter() io.Writer {
	if len(l.outs) > 0 {
		return l.outs[0].DefaultWriter()
	}
	return nil
}

// Debug writelog with level 10
func (l *MultiLogger) Debug(msg string) {
	for _, o := range l.outs {
		o.Debug(msg)
	}
}

// Info writelog with level 20
func (l *MultiLogger) Info(msg string) {
	for _, o := range l.outs {
		o.Info(msg)
	}
}

// Warning writelog with level 30
func (l *MultiLogger) Warning(msg string) {
	for _, o := range l.outs {
		o.Warning(msg)
	}
}

// Error writelog with level 40
func (l *MultiLogger) Error(msg string) {
	for _, o := range l.outs {
		o.Error(msg)
	}
}

// System writelog with level 90
func (l *MultiLogger) System(msg string) {
	for _, o := range l.outs {
		o.System(msg)
	}
}
