package ginmiddleware

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xyzj/toolbox"
	json "github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/loopfunc"
)

const logStr = "|%3d |%-13s|%-15s|%-4s %s |%s"

type logParam struct {
	timer      time.Duration
	keys       map[string]any
	jsn        []byte
	clientIP   string
	method     string
	path       string
	token      string
	body       string
	username   string
	statusCode int
}

// LogToWriter LogToWriter
func LogToWriter(w io.Writer, skippath ...string) gin.HandlerFunc {
	// 设置io
	gin.DefaultWriter = w
	gin.DefaultErrorWriter = w
	chanlog := make(chan *logParam, 200)
	if len(skippath) == 0 {
		skippath = []string{"/favicon.ico", "/showroutes", "/static"}
	}
	go loopfunc.LoopFunc(func(params ...any) {
		for a := range chanlog {
			if len(a.keys) > 0 {
				a.jsn, _ = json.Marshal(a.keys)
			}
			if a.token != "" {
				if a.username != "" {
					a.path = "(" + a.username + "-" + toolbox.CalcCRC32String(json.Bytes(a.token)) + ")" + a.path
				} else {
					a.path = "(" + toolbox.CalcCRC32String(json.Bytes(a.token)) + ")" + a.path
				}
			}
			if a.body != "" {
				a.path += " |" + a.body
			}
			s := fmt.Appendf([]byte{}, logStr, a.statusCode, a.timer, a.clientIP, a.method, a.path, a.jsn)
			w.Write(s)
			if gin.IsDebugging() {
				println(time.Now().Format(logger.ShortTimeFormat) + json.String(s))
			}
		}
	}, "http log", w)
	return func(c *gin.Context) {
		// |,(,) 124,40,41,32
		for _, v := range skippath {
			if strings.HasPrefix(c.Request.URL.Path, v) {
				return
			}
		}
		start := time.Now()
		c.Next()
		// Stop timer
		chanlog <- &logParam{
			timer:      time.Since(start),
			path:       c.Request.URL.Path,
			token:      c.GetHeader("User-Token"),
			body:       c.Param("_body"),
			clientIP:   c.ClientIP(),
			method:     c.Request.Method,
			statusCode: c.Writer.Status(),
			username:   c.Param("_userTokenName"),
			keys:       c.Keys,
		}
	}
}

// LoggerWithRolling 滚动日志
// logdir: 日志存放目录。
// filename：日志文件名。
// maxdays：日志文件最大保存天数。
func LoggerWithRolling(logdir, filename string, maxdays int, skippath ...string) gin.HandlerFunc {
	lo := logger.NewWriter(logger.WithCompressFile(true),
		logger.WithFileDays(maxdays),
		logger.WithFileDir(logdir),
		logger.WithFilename(filename))
	return LogToWriter(lo, skippath...)
}
