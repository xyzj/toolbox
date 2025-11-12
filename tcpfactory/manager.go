package tcpfactory

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/xyzj/deepcopy"
	gopool "github.com/xyzj/go-pool"
	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/cache"
	"github.com/xyzj/toolbox/loopfunc"
	"github.com/xyzj/toolbox/mapfx"
	"github.com/xyzj/toolbox/queue"
)

type TCPManager struct {
	members     *mapfx.StructMap[uint64, tcpCore]
	opt         *opt
	listener    *net.TCPListener
	addr        *net.TCPAddr
	recycle     *gopool.GoPool[*tcpCore]
	targetCache *mapfx.BaseMap[uint64]
	reportCache *cache.AnyCache[*reportItem]
	shutdown    atomic.Bool
}

func (t *TCPManager) HealthReport() map[uint64]any {
	// dis := make([]uint64, 0, t.members.Len())
	a := make(map[uint64]any, t.members.Len())
	t.reportCache.ForEach(func(key string, value *reportItem) bool {
		if time.Since(value.lastRead) > t.opt.readTimeout+time.Second*20 {
			xc, ok := t.members.LoadForUpdate(value.id)
			if ok {
				xc.disconnect("socket anomaly")
			}
			return true
		}
		if !value.status && t.opt.registTimeout > 0 && time.Since(value.connTime) > t.opt.registTimeout {
			xc, ok := t.members.LoadForUpdate(value.id)
			if ok {
				xc.disconnect("unregistered connection")
			}
			return true
		}
		a[value.id] = value.msg
		return true
	})
	// t.members.ForEachWithRLocker(func(key uint64, value *tcpCore) bool {
	// 	if value.closed.Load() {
	// 		dis = append(dis, key)
	// 		return true
	// 	}
	// 	if time.Since(value.timeLastRead) > t.opt.readTimeout+time.Second*20 { // 读取超时，但却没有被关闭，通常为虚连接
	// 		value.disconnect("socket anomaly")
	// 		return true
	// 	}

	// 	if t.opt.registTimeout > 0 && time.Since(value.timeConnection) > t.opt.registTimeout { //  && value.sendQueue.Len() == 0
	// 		if z, ok := value.healthReport(); !ok {
	// 			value.disconnect("unregistered connection")
	// 		} else {
	// 			a[key] = z
	// 		}
	// 		return true
	// 	}
	// 	if z, ok := value.healthReport(); ok {
	// 		a[key] = z
	// 	}
	// 	return true
	// })
	// t.members.DeleteMore(dis...)
	return a
}

func (t *TCPManager) writeToSocket(target string, front bool, msgs ...*SendMessage) bool {
	sid, ok := t.targetCache.Load(target)
	if ok {
		v, ok := t.members.LoadForUpdate(sid)
		if ok {
			return v.writeTo(target, front, msgs...)
		}
	}
	return false
}

// WriteTo sends the given messages to the specified target connections.
// If the multiTargets option is enabled, it continues to send messages to other connections even after
// sending to the specified target.
//
// Parameters:
// - targets: A string list representing the target connection identifier.
// - msgs: Variadic parameter of type *SendMessage, representing the messages to be sent.
//
// Return:
// - None
func (t *TCPManager) WriteTo(target string, msgs ...*SendMessage) {
	if len(msgs) == 0 || strings.TrimSpace(target) == "" {
		return
	}
	if t.writeToSocket(target, false, msgs...) {
		return
	}
	t.members.ForEachWithRLocker(func(key uint64, value *tcpCore) bool {
		if value.writeTo(target, false, msgs...) {
			t.targetCache.Store(target, value.sockID)
			return t.opt.multiTargets
		}
		return true
	})
}

// WriteToFront sends one or more messages to the specified target connection(s) in the front-end.
// If the target string is empty or no messages are provided, the function returns immediately.
// The method iterates over all managed TCP connections and attempts to write the messages to the
// connection matching the target identifier. If the write is successful and the multiTargets option
// is enabled, the iteration continues to allow sending to multiple targets; otherwise, it stops after
// the first successful write.
func (t *TCPManager) WriteToFront(target string, msgs ...*SendMessage) {
	if len(msgs) == 0 || strings.TrimSpace(target) == "" {
		return
	}
	if t.writeToSocket(target, true, msgs...) {
		return
	}
	t.members.ForEachWithRLocker(func(key uint64, value *tcpCore) bool {
		if value.writeTo(target, true, msgs...) {
			t.targetCache.Store(target, value.sockID)
			return t.opt.multiTargets
		}
		return true
	})
}

// WriteTo sends the given messages to the specified target connections.
// If the multiTargets option is enabled, it continues to send messages to other connections even after
// sending to the specified target.
//
// Parameters:
// - targets: A string list representing the target connection identifier.
// - msgs: Variadic parameter of type *SendMessage, representing the messages to be sent.
//
// Return:
// - None
func (t *TCPManager) WriteToMultiTargets(msg *SendMessage, targets ...string) {
	if msg == nil || len(targets) == 0 {
		return
	}
	nt := make([]string, 0, len(targets))
	for _, a := range targets {
		if t.writeToSocket(a, false, msg) {
			continue
		}
		nt = append(nt, a)
	}
	t.members.ForEachWithRLocker(func(key uint64, value *tcpCore) bool {
		for _, a := range nt {
			if strings.TrimSpace(a) == "" {
				continue
			}
			if value.writeTo(a, false, msg) {
				t.targetCache.Store(a, value.sockID)
				return t.opt.multiTargets
			}
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
				cli := t.recycle.Get()
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
				defer func() {
					if err := recover(); err != nil {
						cli.disconnect(fmt.Sprintf("%+v", err))
					} else {
						cli.disconnect("socket closed")
					}
					if !t.shutdown.Load() {
						t.members.Delete(cli.sockID)
						t.recycle.Put(cli)
						t.reportCache.Delete(fmt.Sprintf("%d", cli.sockID))
					}
				}()
				cli.connect(conn, t.opt.helloMsg...) // conn
				// checkhealth
				go func() {
					freport := func() {
						x, ok := cli.healthReport()
						t.reportCache.Store(fmt.Sprintf("%d", cli.sockID), &reportItem{
							id:        cli.sockID,
							connTime:  cli.timeConnection,
							lastRead:  cli.timeLastRead,
							lastWrite: cli.timeLastWrite,
							msg:       x,
							status:    ok,
						})
					}
					time.Sleep(time.Second * 10)
					freport()
					t1 := time.NewTicker(time.Second * time.Duration(rand.Int31n(25)+25))
					for !cli.closed.Load() {
						select {
						case <-cli.closeCtx.Done():
							return
						case <-t1.C:
							freport()
						}
					}
				}()
				go func() {
					defer func() {
						if err := recover(); err != nil {
							cli.disconnect(fmt.Sprintf("send, %+v", err))
						}
					}()
					// send
					cli.send()
				}()
				// recv
				cli.recv()
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

func (t *TCPManager) Len() int {
	return t.members.Len()
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
func NewTcpFactory(opts ...Options) (*TCPManager, error) {
	opt := defaultOpt
	for _, o := range opts {
		o(&opt)
	}
	b, ok := toolbox.ValidateIPPort(opt.bind)
	if !ok {
		return nil, fmt.Errorf("invalid bind address: %s", opt.bind)
	}
	rep := cache.NewAnyCache[*reportItem](time.Second * 50)
	rep.SetCleanUp(time.Minute * 3)
	sid := atomic.Uint64{}
	return &TCPManager{
		addr:        b,
		opt:         &opt,
		targetCache: mapfx.NewBaseMap[uint64](),
		recycle: gopool.New(
			func() *tcpCore {
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
					logg:               opt.logg,
				}
			},
			gopool.WithMaxIdleSize(uint32(opt.poolSize)),
		),
		// recycle: sync.Pool{
		// 	New: func() any {
		// 		ctx, cancel := context.WithCancel(context.Background())
		// 		t1 := time.NewTimer(time.Minute)
		// 		t1.Stop()
		// 		return &tcpCore{
		// 			sockID:             sid.Add(1),
		// 			sendQueue:          queue.NewHighLowQueue[*SendMessage](opt.maxQueue),
		// 			closed:             atomic.Bool{},
		// 			readBuffer:         make([]byte, 8192),
		// 			readCache:          &bytes.Buffer{},
		// 			readTimeout:        opt.readTimeout,
		// 			writeTimeout:       opt.writeTimeout,
		// 			writeIntervalTimer: t1,
		// 			tcpClient:          deepcopy.CopyAny(opt.client),
		// 			closeCtx:           ctx,
		// 			closeFunc:          cancel,
		// 			logg:               opt.logg,
		// 		}
		// 	},
		// },
		shutdown:    atomic.Bool{},
		members:     mapfx.NewStructMap[uint64, tcpCore](),
		reportCache: rep,
	}, nil
}
