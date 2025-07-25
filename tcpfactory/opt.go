package tcpfactory

import (
	"time"

	"github.com/xyzj/toolbox/logger"
)

type Opt struct {
	logg          logger.Logger
	client        Client
	readTimeout   time.Duration
	writeTimeout  time.Duration
	registTimeout time.Duration
	keepAlive     time.Duration
	helloMsg      []*SendMessage
	bind          string
	maxQueue      int32
	poolSize      int32
	multiTargets  bool
}
type Opts func(opt *Opt)

var defaultOpt = Opt{
	logg:          logger.NewConsoleLogger(),
	client:        &EmptyClient{},
	bind:          ":6880",
	readTimeout:   time.Second * 100,
	writeTimeout:  0,
	registTimeout: 0,
	keepAlive:     time.Second * 30,
	helloMsg:      make([]*SendMessage, 0),
	maxQueue:      1000,
	poolSize:      3000,
	multiTargets:  false,
}

func OptBindAddr(s string) Opts {
	return func(o *Opt) {
		o.bind = s
	}
}

// OptReadTimeout is an option function for the TCPFactory that sets the read timeout for client connections.
// The provided duration is clamped to a minimum of 1 second and a maximum of 100 minutes.
//
// Parameters:
//
//	t: A time.Duration representing the read timeout duration.
//
// Returns:
//
//	An Opts function that can be used to configure the TCPFactory with the provided read timeout.
func OptReadTimeout(t time.Duration) Opts {
	return func(o *Opt) {
		o.readTimeout = min(max(t, 1000000000), 6000000000000) // 1s～100m
	}
}

// OptWriteTimeout is an option function for the TCPFactory that sets the write timeout for client connections.
// The provided duration is clamped to a minimum of 0 and a maximum of 1 minute.
//
// Parameters:
//
//	t: A time.Duration representing the write timeout duration.
//
// Returns:
//
//	An Opts function that can be used to configure the TCPFactory with the provided write timeout.
func OptWriteTimeout(t time.Duration) Opts {
	return func(o *Opt) {
		o.writeTimeout = min(max(t, 0), 60000000000) // 0~1m
	}
}

// OptRegistTimeout is an option function for the TCPFactory that sets the registration timeout for client connections.
// The provided duration is clamped to a minimum of 10 seconds.
//
// Parameters:
//
//	t: A time.Duration representing the registration timeout duration.
//
// Returns:
//
//	An Opts function that can be used to configure the TCPFactory with the provided registration timeout.
func OptRegistTimeout(t time.Duration) Opts {
	return func(o *Opt) {
		o.registTimeout = min(max(t, 10000000000), 100000000000)
	}
}

// OptKeepAlive is an option function for the TCPFactory that sets the keep-alive
// period for client connections. The provided duration is clamped to a minimum of
// 10 seconds.
//
// Parameters:
//
//	t: A time.Duration representing the keep-alive period duration.
//
// Returns:
//
//	An Opts function that can be used to configure the TCPFactory with the provided
//	keep-alive period.
func OptKeepAlive(t time.Duration) Opts {
	return func(o *Opt) {
		o.keepAlive = min(max(t, 10000000000), 100000000000)
	}
}

// OptLogger is an option function for the TCPFactory that allows setting a custom logger.
// If no logger is provided, it defaults to a console logger.
//
// The function accepts a single parameter:
// - l: A logger.Logger interface that defines the behavior of the custom logger.
//
// The function returns an Opts function, which can be used to configure the TCPFactory
// with the provided logger implementation.
//
// Example usage:
//
//	factory := NewTCPFactory(
//		OptLogger(&CustomLogger{}),
//	)
func OptLogger(l logger.Logger) Opts {
	return func(o *Opt) {
		o.logg = l
	}
}

// OptMaxSendQueue is an option function for the TCPFactory that allows setting
// the maximum size of the send queue for each client connection.
//
// The function accepts a single parameter:
//   - l: An integer value representing the maximum size of the send queue.
//     If the provided value is less than 10000, it will be set to 10000.
//
// The function returns an Opts function, which can be used to configure the TCPFactory
// with the provided maximum send queue size option.
//
// Example usage:
//
//	factory := NewTCPFactory(
//		OptMaxSendQueue(20000),
//	)
func OptMaxSendQueue(l int32) Opts {
	return func(o *Opt) {
		o.maxQueue = max(l, 10000)
	}
}

// OptMatchMultiTargets is an option function for the TCPFactory that allows setting
// whether the factory should match messages to multiple targets or not.
//
// The function accepts a single parameter:
//   - l: A boolean value indicating whether to match messages to multiple targets or not.
//     If true, messages will be sent to all matching targets.
//     If false, messages will only be sent to the first matching target.
//
// The function returns an Opts function, which can be used to configure the TCPFactory
// with the provided match multi targets option.
//
// Example usage:
//
//	factory := NewTCPFactory(
//		OptMatchMultiTargets(true),
//	)
func OptMatchMultiTargets(l bool) Opts {
	return func(o *Opt) {
		o.multiTargets = l
	}
}

// OptTcpClient is an option function for the TCPFactory that allows setting a custom
// TCPFactory implementation. If no TCPFactory is provided, it defaults to an EmptyTCP.
//
// The function accepts a single parameter:
// - t: A TCPFactory interface that defines the behavior of the custom TCPFactory.
//
// The function returns an Opts function, which can be used to configure the TCPFactory
// with the provided TCPFactory implementation.
//
// Example usage:
//
//	factory := NewTCPFactory(
//		OptTcpClient(&CustomTCPFactory{}),
//	)
func OptTcpClient(t Client) Opts {
	return func(o *Opt) {
		if t == nil {
			o.client = &EmptyClient{}
		} else {
			o.client = t
		}
	}
}

// OptHelloMessages is an option function for the TCPFactory that allows setting
// custom hello messages to be sent to the connected clients.
//
// The function accepts a variadic number of pointers to SendMessage structs.
// These messages will be appended to the existing hello messages list in the Opt struct.
//
// This option function returns an Opts function, which can be used to configure
// the TCPFactory with the provided hello messages.
//
// Example usage:
//
//	factory := NewTCPFactory(
//		OptHelloMessages(
//			&SendMessage{Data: []byte("Hello, client 1")},
//			&SendMessage{Data: []byte("Hello, client 2")},
//		),
//	)
func OptHelloMessages(t ...*SendMessage) Opts {
	return func(o *Opt) {
		o.helloMsg = append(o.helloMsg, t...)
	}
}

// OptMaxClientPoolSize returns an option function that sets the maximum client pool size.
// The pool size will be set to the greater of 2 or the provided value t.
func OptMaxClientPoolSize(t int32) Opts {
	return func(o *Opt) {
		o.poolSize = max(2, t)
	}
}
