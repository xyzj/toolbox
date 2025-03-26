package tcpfactory

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/xyzj/deepcopy"
	"github.com/xyzj/toolbox/loopfunc"
	"github.com/xyzj/toolbox/mapfx"
	"github.com/xyzj/toolbox/queue"
	"golang.org/x/net/context"
)

type TCPManager struct {
	members  *mapfx.StructMap[uint64, tcpCore]
	opt      *Opt
	listener *net.TCPListener
	recycle  sync.Pool
	shutdown atomic.Bool
	port     int
}

func (t *TCPManager) HealthReport() map[uint64]any {
	dis := make(map[uint64]string)
	a := make(map[uint64]any)
	var ok bool
	t.members.ForEachWithRLocker(func(key uint64, value *tcpCore) bool {
		if value.closed.Load() {
			dis[key] = ""
			return true
		}
		if time.Since(value.timeLastRead) > t.opt.readTimeout+time.Second*20 { // 读取超时，但却没有被关闭，通常为虚连接
			value.disconnect("socket anomaly")
			return true
		}
		if t.opt.registTimeout > 0 && time.Since(value.timeLastWrite) > t.opt.registTimeout && value.sendQueue.Len() == 0 {
			if _, ok = value.healthReport(); !ok {
				value.disconnect("unregistered connection")
				return true
			}
		}
		if z, ok := value.healthReport(); ok {
			a[key] = z
		}
		return true
	})
	return a
}

func (t *TCPManager) WriteTo(target string, msgs ...*SendMessage) {
	t.members.ForEachWithRLocker(func(key uint64, value *tcpCore) bool {
		if value.writeTo(target, msgs...) {
			return t.opt.multiTargets
		}
		return true
	})
}

func (t *TCPManager) Listen() error {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(t.opt.bind), Port: t.port, Zone: ""})
	if err != nil {
		t.opt.logg.Error(err.Error())
		return err
	}
	t.opt.logg.System(fmt.Sprintf("TCP listening to: %s", listener.Addr().String()))
	t.listener = listener
	loopfunc.LoopFunc(func(params ...any) {
		for !t.shutdown.Load() {
			conn, err := t.listener.AcceptTCP()
			if err != nil {
				t.opt.logg.Error(err.Error())
				continue
			}
			go func(conn *net.TCPConn) {
				cli := t.recycle.Get().(*tcpCore)
				defer func() {
					if err := recover(); err != nil {
						cli.disconnect(err.(error).Error())
					} else {
						cli.disconnect("socket closed")
					}
					if !t.shutdown.Load() {
						t.members.Delete(cli.sockID)
						t.recycle.Put(cli)
					}
				}()
				if t.opt.keepAlive > 0 {
					conn.SetKeepAliveConfig(net.KeepAliveConfig{
						Enable:   true,
						Idle:     t.opt.keepAlive,
						Interval: t.opt.keepAlive,
					})
				} else {
					conn.SetKeepAlive(false)
				}
				conn.SetLinger(0)
				cli.connect(conn, t.opt.helloMsg...)
				t.members.Store(cli.sockID, cli)
				// recv
				go func() {
					defer func() {
						if err := recover(); err != nil {
							cli.disconnect(fmt.Sprintf("tcp reciver crash: %+v", errors.WithStack(err.(error))))
						}
					}()
					cli.recv()
				}()
				// send
				cli.send()
			}(conn)
		}
	}, "tcplistener", t.opt.logg.DefaultWriter())
	t.opt.logg.System("Shutting down")
	t.members.ForEachWithRLocker(func(key uint64, value *tcpCore) bool {
		value.disconnect("server shutdown")
		return true
	})
	return nil
}

func (t *TCPManager) Shutdown() {
	t.shutdown.Store(true)
	t.listener.Close()
}

func NewTcpFactory(port int, opts ...Opts) (*TCPManager, error) {
	if port <= 0 || port > 65535 {
		return nil, ErrPortNotValid
	}
	opt := defaultOpt
	for _, o := range opts {
		o(&opt)
	}
	if net.ParseIP(opt.bind) == nil {
		return nil, ErrHostNotValid
	}
	sid := atomic.Uint64{}
	return &TCPManager{
		port: port,
		opt:  &opt,
		recycle: sync.Pool{
			New: func() any {
				ctx, cancel := context.WithCancel(context.Background())
				return &tcpCore{
					sockID:        sid.Add(1),
					sendQueue:     queue.NewHighLowQueue[*SendMessage](opt.maxQueue),
					closeOnce:     new(sync.Once),
					closed:        atomic.Bool{},
					readBuffer:    make([]byte, 8192),
					readCache:     &bytes.Buffer{},
					readTimeout:   opt.readTimeout,
					writeTimeout:  opt.writeTimeout,
					writeTimetick: time.NewTicker(time.Second),
					tcpMod:        deepcopy.CopyAny(opt.mod),
					closeCtx:      ctx,
					closeFunc:     cancel,
					logg:          opt.logg,
				}
			},
		},
		shutdown: atomic.Bool{},
		members:  mapfx.NewStructMap[uint64, tcpCore](),
	}, nil
}
