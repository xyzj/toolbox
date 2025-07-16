// Package ginmiddleware 基于gin的web框架封装
package ginmiddleware

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/loopfunc"
)

/*
// ServiceProtocol http协议类型
type ServiceProtocol int

const (
	// ProtocolHTTP http协议
	ProtocolHTTP ServiceProtocol = iota
	// ProtocolHTTPS https协议
	ProtocolHTTPS
	// PtorocolBoth 2种协议
	PtorocolBoth
)
*/

// ServiceOption 通用化http框架
type ServiceOption struct {
	EngineFunc   func() *gin.Engine
	Engine       *gin.Engine
	Hosts        []string
	CertFile     string
	KeyFile      string
	LogFile      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	HTTPPort     string
	HTTPSPort    string
	LogDays      int
	Debug        bool
}

// ListenAndServe 启用监听
// port：端口号
// h： http.hander, like gin.New()
func ListenAndServe(port int, h *gin.Engine) error {
	ListenAndServeWithOption(
		OptHTTP(fmt.Sprintf(":%d", port)),
		OptEngineFunc(func() *gin.Engine { return h }))
	return nil
}

// ListenAndServeTLS 启用TLS监听
// port：端口号
// h： http.hander, like gin.New()
// certfile： cert file path
// keyfile： key file path
// clientca: 客户端根证书用于验证客户端合法性
func ListenAndServeTLS(port int, h *gin.Engine, certfile, keyfile string) error {
	ListenAndServeWithOption(
		OptHTTPSFromFile(fmt.Sprintf(":%d", port), certfile, keyfile),
		OptEngineFunc(func() *gin.Engine { return h }),
	)
	return nil
}

// ListenAndServeWithOption 启动服务
func ListenAndServeWithOption(opts ...Opts) {
	opt := defaultOpt
	for _, o := range opts {
		o(&opt)
	}
	if opt.http+opt.https == "" {
		opt.logg.Error("no service port is valid")
		os.Exit(1)
	}
	// 路由处理
	findRoot := false
	findIcon := false
	h := opt.engine
	if opt.engineFunc != nil {
		h = opt.engineFunc()
	}
	if h == nil {
		h = gin.New()
	}
	for _, v := range h.Routes() {
		if v.Path == "/" {
			findRoot = true
			continue
		}
		if v.Path == "/favicon.ico" {
			findIcon = true
		}
		if findRoot && findIcon {
			break
		}
	}
	if !findRoot {
		h.GET("/", PageDefault)
	}
	if !findIcon {
		h.GET("/favicon.ico", func(c *gin.Context) {
			c.Writer.Write(favicon)
		})
	}
	wg := sync.WaitGroup{}
	// 启动https服务
	if opt.https != "" {
		wg.Add(1)
		loopfunc.GoFunc(func(params ...interface{}) {
			defer wg.Done()
			s := &http.Server{
				Addr:         opt.https,
				ReadTimeout:  opt.readTimeout,
				WriteTimeout: opt.writeTimeout,
				IdleTimeout:  opt.idleTimeout,
				Handler:      h,
				TLSConfig:    opt.tlsc,
			}
			fmt.Fprintf(os.Stdout, "%s [%s] %s\n", time.Now().Format(toolbox.ShortTimeFormat), "HTTP", "Start HTTPS server at "+opt.https)
			if err := s.ListenAndServeTLS("", ""); err != nil {
				fmt.Fprintf(os.Stdout, "%s [%s] %s\n", time.Now().Format(toolbox.ShortTimeFormat), "HTTP", "Start HTTPS server error: "+err.Error())
			}
		}, "https", os.Stdout)
	}
	if opt.http != "" {
		wg.Add(1)
		loopfunc.GoFunc(func(params ...interface{}) {
			defer wg.Done()
			s := &http.Server{
				Addr:         opt.http,
				ReadTimeout:  opt.readTimeout,
				WriteTimeout: opt.writeTimeout,
				IdleTimeout:  opt.idleTimeout,
				Handler:      h,
			}
			fmt.Fprintf(os.Stdout, "%s [%s] %s\n", time.Now().Format(toolbox.ShortTimeFormat), "HTTP", "Start HTTP server at "+opt.http)
			if err := s.ListenAndServe(); err != nil {
				fmt.Fprintf(os.Stdout, "%s [%s] %s\n", time.Now().Format(toolbox.ShortTimeFormat), "HTTP", "Start HTTP server error: "+err.Error())
			}
		}, "http", os.Stdout)
	}
	wg.Wait()
}

// LiteEngine 轻量化基础引擎
func LiteEngine(w io.Writer, hosts ...string) *gin.Engine {
	r := gin.New()
	// 特殊路由处理
	r.HandleMethodNotAllowed = true
	r.NoMethod(Page405)
	r.NoRoute(Page404Big)
	// 允许跨域
	r.Use(cors.New(cors.Config{
		MaxAge:           time.Hour * 24,
		AllowWebSockets:  true,
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
	}))
	// 处理转发ip
	r.Use(XForwardedIP())
	// 配置日志
	r.Use(LogToWriter(w))
	// 故障恢复
	r.Use(Recovery())
	// 绑定域名
	r.Use(bindHosts(hosts...))
	// 数据压缩
	// r.Use(gingzip.Gzip(6))
	return r
}

func bindHosts(hosts ...string) gin.HandlerFunc {
	if len(hosts) == 0 {
		return func(c *gin.Context) {}
	}
	return func(c *gin.Context) {
		host, _, _ := net.SplitHostPort(c.Request.Host)
		nohost := true
		for _, v := range hosts {
			if v == host {
				nohost = false
				break
			}
		}
		if nohost {
			c.Set("status", 0)
			c.Set("detail", "forbidden")
			c.AbortWithStatusJSON(http.StatusForbidden, c.Keys)
		}
	}
}
