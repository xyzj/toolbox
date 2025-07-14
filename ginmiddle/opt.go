package ginmiddleware

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xyzj/toolbox/crypto"
	"github.com/xyzj/toolbox/logger"
)

var defaultOpt = Opt{
	engine:       gin.New(),
	readTimeout:  time.Second * 120,
	writeTimeout: time.Second * 120,
	idleTimeout:  time.Second * 60,
	logg:         logger.NewConsoleLogger(),
	hosts:        make([]string, 0),
	http:         ":6880",
	debug:        false,
	tlsc:         nil,
}

// Opt 通用化http框架
type Opt struct {
	engineFunc   func() *gin.Engine
	hosts        []string
	engine       *gin.Engine
	tlsc         *tls.Config
	logg         logger.Logger
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	http         string
	https        string
	debug        bool
}
type Opts func(opt *Opt)

func OptEngine(r *gin.Engine) Opts {
	return func(opt *Opt) {
		opt.engine = r
	}
}
func OptEngineFunc(f func() *gin.Engine) Opts {
	return func(opt *Opt) {
		opt.engineFunc = f
	}
}

func OptHosts(hosts ...string) Opts {
	return func(opt *Opt) {
		opt.hosts = hosts
	}
}

func OptLogger(l logger.Logger) Opts {
	return func(opt *Opt) {
		opt.logg = l
	}
}

func OptReadTimeout(t time.Duration) Opts {
	return func(opt *Opt) {
		opt.readTimeout = t
	}
}

func OptWriteTimeout(t time.Duration) Opts {
	return func(opt *Opt) {
		opt.writeTimeout = t
	}
}

func OptIdleTimeout(t time.Duration) Opts {
	return func(opt *Opt) {
		opt.idleTimeout = t
	}
}

func OptHTTP(s string) Opts {
	return func(opt *Opt) {
		if _, err := net.ResolveTCPAddr("tcp", s); err != nil {
			opt.http = ""
			return
		}
		opt.http = s
	}
}

func OptHTTPS(s string, t *tls.Config) Opts {
	return func(opt *Opt) {
		if _, err := net.ResolveTCPAddr("tcp", s); err != nil {
			opt.https = ""
			return
		}
		if t == nil || t.Certificates == nil {
			opt.https = ""
			return
		}
		opt.https = s
		opt.tlsc = t
	}
}

func OptHTTPSFromFile(s, certFile, keyFile string) Opts {
	return func(opt *Opt) {
		t, err := crypto.TLSConfigFromFile(certFile, keyFile, "")
		if err != nil {
			opt.tlsc = nil
			return
		}
		OptHTTPS(s, t)
	}
}

func OptDebug(d bool) Opts {
	return func(opt *Opt) {
		opt.debug = d
	}
}
