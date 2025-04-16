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
	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/loopfunc"
	"github.com/xyzj/toolbox/mapfx"
	"github.com/xyzj/toolbox/queue"
	"golang.org/x/net/context"
)

type TCPManager struct {
	members  *mapfx.StructMap[uint64, tcpCore]
	opt      *Opt
	listener *net.TCPListener
	addr     *net.TCPAddr
	recycle  sync.Pool
	shutdown atomic.Bool
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

// WriteTo sends the given messages to the specified target connections.
// If the multiTargets option is enabled, it continues to send messages to other connections even after
// sending to the specified target.
//
// Parameters:
// - target: A string representing the target connection identifier.
// - msgs: Variadic parameter of type *SendMessage, representing the messages to be sent.
//
// Return:
// - None
func (t *TCPManager) WriteTo(target string, msgs ...*SendMessage) {
	t.members.ForEachWithRLocker(func(key uint64, value *tcpCore) bool {
		if value.writeTo(target, msgs...) {
			return t.opt.multiTargets
		}
		return true
	})
}

// Listen starts listening for incoming TCP connections on the specified address.
// It creates a TCP listener, logs the listening address, and handles incoming connections.
// For each accepted connection, it creates a new tcpCore instance, sets up keep-alive, linger options,
// and starts separate goroutines for receiving and sending data.
// If any error occurs during the listening process, it logs the error and returns the error.
//
// Parameters:
// - None
//
// Return:
// - An error if any, otherwise nil.
func (t *TCPManager) Listen() error {
	listener, err := net.ListenTCP("tcp", t.addr)
	if err != nil {
		t.opt.logg.Error(err.Error())
		return err
	}
	t.opt.logg.System(fmt.Sprintf("[tcp] listening to: %s", listener.Addr().String()))
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
				t.members.Store(cli.sockID, cli)
				cli.connect(conn, t.opt.helloMsg...)
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

// NewTcpFactory creates a new TCPManager instance with the specified bind address and options.
// It resolves the bind address and initializes the TCPManager with default or provided options.
// The function returns a pointer to the created TCPManager instance and an error if any.
//
// Parameters:
// - bind: A string representing the bind address in the format "host:port".
// - opts: Variadic parameter of type Opts, which are optional configuration functions for the TCPManager.
//
// Return:
// - A pointer to the created TCPManager instance.
// - An error if any, otherwise nil.
func NewTcpFactory(opts ...Opts) (*TCPManager, error) {
	opt := defaultOpt
	for _, o := range opts {
		o(&opt)
	}
	b, ok := toolbox.CheckTCPAddr(opt.bind)
	if !ok {
		return nil, fmt.Errorf("invalid bind address: %s", opt.bind)
	}
	sid := atomic.Uint64{}
	return &TCPManager{
		addr: b,
		opt:  &opt,
		recycle: sync.Pool{
			New: func() any {
				ctx, cancel := context.WithCancel(context.Background())
				t1 := time.NewTimer(time.Minute)
				t1.Stop()
				return &tcpCore{
					sockID:             sid.Add(1),
					sendQueue:          queue.NewHighLowQueue[*SendMessage](opt.maxQueue),
					closed:             atomic.Bool{},
					readBuffer:         make([]byte, 8192),
					readCache:          &bytes.Buffer{},
					readTimeout:        opt.readTimeout,
					writeTimeout:       opt.writeTimeout,
					writeIntervalTimer: t1,
					tcpClient:          deepcopy.CopyAny(opt.client),
					closeCtx:           ctx,
					closeFunc:          cancel,
					logg:               opt.logg,
				}
			},
		},
		shutdown: atomic.Bool{},
		members:  mapfx.NewStructMap[uint64, tcpCore](),
	}, nil
}
