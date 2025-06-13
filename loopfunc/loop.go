/*
Package loopfunc ： 用于控制需要持续运行的循环方法，当方法漰溃时会自动重启
*/
package loopfunc

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/xyzj/toolbox/json"
)

const (
	longTimeFormat = "2006-01-02 15:04:05.000"
)

// CrashLogger 主进程崩溃用日志
type CrashLogger struct {
	FilePath string
	fn       *os.File
}

func (m *CrashLogger) Write(p []byte) (n int, err error) {
	m.fn, err = os.OpenFile(m.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o664)
	if err != nil {
		return 0, err
	}
	defer m.fn.Close()
	b := []byte(time.Now().Format(longTimeFormat) + " ")
	b = append(b, p...)
	return m.fn.Write(b)
}

// LoopFunc 执行循环工作，并提供panic恢复
//
// f: 要执行的循环方法，可控制传入参数
//
// name：这个方法的名称，用于错误标识
//
// logWriter：方法崩溃时的日志记录器，默认os.stdout
//
// params： 需要传给f的参数，f内需要进行类型转换
func LoopFunc(f func(params ...any), name string, logWriter io.Writer, params ...any) {
	loopAndRetry(f, name, logWriter, time.Second*20, 0, params...)
}

// LoopWithWait 执行循环工作，并在指定的等待时间后提供panic恢复
//
// f: 要执行的循环方法，可控制传入参数
//
// name：这个方法的名称，用于错误标识
//
// logWriter：方法崩溃时的日志记录器，默认os.stdout
//
// params： 需要传给f的参数，f内需要进行类型转换
func LoopWithWait(f func(params ...any), name string, logWriter io.Writer, timewait time.Duration, params ...any) {
	loopAndRetry(f, name, logWriter, timewait, 0, params...)
}

// LoopWithRetry 执行循环工作，并在指定的等待时间后提供panic恢复，panic次数可设置
//
// f: 要执行的循环方法，可控制传入参数
//
// name：这个方法的名称，用于错误标识
//
// logWriter：方法崩溃时的日志记录器，默认os.stdout
//
// retry：panic最大次数
//
// params： 需要传给f的参数，f内需要进行类型转换
func LoopWithRetry(f func(params ...any), name string, logWriter io.Writer, timewait time.Duration, retry int, params ...any) {
	loopAndRetry(f, name, logWriter, timewait, retry, params...)
}

func loopAndRetry(f func(params ...any), name string, logWriter io.Writer, timewait time.Duration, retry int, params ...any) {
	locker := &sync.WaitGroup{}
	end := false
	if logWriter == nil {
		logWriter = os.Stdout
	}
	errCount := 0
RUN:
	locker.Add(1)
	func() {
		defer func() {
			if err := recover(); err == nil {
				// 非panic,不需要恢复
				end = true
			} else {
				logWriter.Write(json.Bytes(fmt.Sprintf(name+" [LOOP] crash: %+v\n", errors.WithStack(err.(error)))))
				errCount++
				if retry > 0 && errCount >= retry {
					logWriter.Write([]byte(name + " [LOOP] the maximum number of retries has been reached, the end.\n"))
					end = true
				}
			}
			locker.Done()
		}()
		f(params...)
	}()
	locker.Wait()
	if end {
		return
	}
	time.Sleep(timewait)
	goto RUN
}

// GoFunc 执行安全的子线程工作，包含panic捕获
//
// f: 要执行的循环方法，可控制传入参数
//
// name：这个方法的名称，用于错误标识
//
// logWriter：方法崩溃时的日志记录器，默认os.stdout
//
// params： 需要传给f的参数，f内需要进行类型转换
func GoFunc(f func(params ...any), name string, logWriter io.Writer, params ...any) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logWriter.Write(json.Bytes(fmt.Sprintf(name+" [LOOP] crash: %+v\n", err)))
			}
		}()
		f(params...)
	}()
}
