package ginmiddleware

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	json "github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/logger"
)

const logStr = "%s |%3d |%-13s|%-15s|%-4s %s|%s"

// LogToWriter LogToWriter
func LogToWriter(out io.Writer, skippath ...string) gin.HandlerFunc {
	// 设置io
	gin.DefaultWriter = out
	gin.DefaultErrorWriter = out
	skip := map[string]struct{}{}
	for _, v := range skippath {
		skip[v] = struct{}{}
	}
	return func(c *gin.Context) {
		path := c.Request.RequestURI
		for _, v := range skippath {
			if strings.HasPrefix(path, v) {
				return
			}
		}
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		jsn, _ := json.MarshalToString(c.Keys)
		var s = make([]byte, 0, len(jsn)+len(path)+50)

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)
		s = fmt.Appendf(s, logStr,
			end.Format(logger.LongTimeFormat),
			c.Writer.Status(),
			latency,
			GetClientIP(c),
			c.Request.Method,
			path,
			jsn,
		)
		out.Write(s)
		if gin.IsDebugging() {
			println(json.String(s))
		}
	}
}
