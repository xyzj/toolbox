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
	"github.com/xyzj/toolbox/loopfunc"
	"github.com/xyzj/toolbox/queue"
)

type TCPManager struct {
	members     *members //*mapfx.StructMap[uint64, tcpCore]
	opt         *opt
	listener    *net.TCPListener
	addr        *net.TCPAddr
	recycle     *gopool.GoPool[*tcpCore]
	reportCache *reportData
	shutdown    atomic.Bool
}

// HealthReport generates a health report for all members.
func (t *TCPManager) HealthReport() map[uint64]any {
	a := make(map[uint64]any, t.members.Len())
	removekey := make([]uint64, 0)
	t.reportCache.ForEach(func(key uint64, value *reportItem) bool {
		if time.Since(value.dtReport).Minutes() > 2 {
			removekey = append(removekey, key)
			return true
		}
		a[value.id] = value.msg
		return true
	})
	t.reportCache.Delete(removekey...)
	return a
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
	t.WriteWithPriority(queue.PriorityNormal, target, msgs...)
}

func (t *TCPManager) WriteWithPriority(priority queue.Priority, target string, msgs ...*SendMessage) {
	if len(msgs) == 0 || strings.TrimSpace(target) == "" {
		return
	}
	t.members.SendTo(priority, target, msgs...)
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
	for _, a := range targets {
		t.members.SendTo(queue.PriorityNormal, a, msg)
	}
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
						t.reportCache.Delete(cli.sockID)
						t.recycle.Put(cli)
					}
				}()
				cli.connect(conn, t.opt.helloMsg...) // conn
				// checkhealth
				go func() {
					freport := func() {
						x, ok, shutdown := cli.healthReport()
						if shutdown {
							cli.disconnect("client said shutdown")
							return
						}
						if time.Since(cli.timeLastRead) > t.opt.readTimeout+time.Second*20 {
							cli.disconnect("socket anomaly")
							return
						}
						if !ok && t.opt.registTimeout > 0 && time.Since(cli.timeConnection) > t.opt.registTimeout {
							cli.disconnect("unregistered connection")
							return
						}
						if ok {
							t.reportCache.Store(cli.sockID, &reportItem{
								id:       cli.sockID,
								msg:      x,
								status:   ok,
								dtReport: time.Now(),
							})
						}
					}
					defer func() {
						if err := recover(); err != nil {
							cli.disconnect(fmt.Sprintf("send panic, %+v", err))
						}
					}()
					t1 := time.NewTicker(time.Second * time.Duration(rand.Int31n(10)+20))
					for !cli.closed.Load() {
						select {
						case <-cli.closeCtx.Done():
							return
						case <-t1.C:
							freport()
						default:
							cli.send()
						}
					}
				}()
				// recv
				cli.recv()
			}(conn)
		}
	}, "tcplistener", t.opt.logg.DefaultWriter())
	t.opt.logg.System("Shutting down")
	t.members.ShutdownAll()
	return nil
}

// Shutdown gracefully shuts down the TCPManager.
func (t *TCPManager) Shutdown() {
	t.shutdown.Store(true)
	t.listener.Close()
}

// Len returns the number of active members.
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
	sid := atomic.Uint64{}
	t := &TCPManager{
		addr:        b,
		opt:         &opt,
		shutdown:    atomic.Bool{},
		members:     newMembers(int(opt.predictedClients), opt.multiTargets), // mapfx.NewStructMap[uint64, tcpCore](),
		reportCache: newReportData(int(opt.predictedClients)),
		recycle: gopool.New(
			func() *tcpCore {
				t1 := time.NewTimer(time.Minute)
				t1.Stop()
				socketid := sid.Add(1)
				return &tcpCore{
					sockID:             socketid,
					sendQueue:          queue.NewPriorityQueue[*SendMessage](int(opt.maxQueue)),
					closed:             atomic.Bool{},
					readBuffer:         make([]byte, opt.readBufferSize),
					readCache:          &bytes.Buffer{},
					readTimeout:        opt.readTimeout,
					writeTimeout:       opt.writeTimeout,
					sendQueueTimeout:   time.Second * 30,
					writeIntervalTimer: t1,
					tcpClient:          deepcopy.CopyAny(opt.client),
					logg:               opt.logg,
				}
			},
			gopool.WithMaxIdleSize(uint32(opt.poolSize)),
		),
	}
	return t, nil
}
