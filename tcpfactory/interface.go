package tcpfactory

import (
	"net"
	"strings"
	"time"
)

var shutmedown = &SendMessage{
	Data:     []byte{0x73, 0x68, 0x75, 0x74, 0x20, 0x6d, 0x65, 0x20, 0x64, 0x6f, 0x77, 0x6e},
	Interval: 0,
}

type SendMessage struct {
	Data     []byte
	Interval time.Duration
}

func TakeANap(t time.Duration) *SendMessage {
	return &SendMessage{
		Data:     nil,
		Interval: t,
	}
}
func ShutMeDown() *SendMessage {
	return shutmedown
}

// Client defines the interface for handling TCP connection lifecycle and message processing.
// It provides methods for connection events, data handling, logging, and client status reporting.
type Client interface {
	// MatchTarget is used to match if the target matches the client
	MatchTarget(target string, prefix bool) bool
	// Report is used to report client status, return status data and if the client is registered, and if the client is shutting down
	Report() (data any, legal bool, shutdown bool)
	// OnConnect is called when the connection is established
	OnConnect(conn *net.TCPConn)
	// OnDisconnect is called when the connection is closed
	OnDisconnect(reason string)
	// OnRecive is called when received data, return unfinished data, message need send to client
	OnRecive(data []byte) ([]byte, []*SendMessage)
	// OnSend is called when data is about to be sent, allowing the client to format or modify the outgoing data.
	OnSend(data []byte)
}

// EmptyClient is a no-op implementation of the Client interface that provides default empty method implementations.
type EmptyClient struct{}

func (t *EmptyClient) OnConnect(n *net.TCPConn)                   {}
func (t *EmptyClient) OnDisconnect(s string)                      {}
func (t *EmptyClient) MatchTarget(s string, prefix bool) bool     { return false }
func (t *EmptyClient) Report() (any, bool, bool)                  { return "", false, false }
func (t *EmptyClient) OnSend(b []byte)                            {}
func (t *EmptyClient) OnRecive(b []byte) ([]byte, []*SendMessage) { return nil, nil }

// EchoClient is a simple implementation of the Client interface that echoes received data
// and stores the remote connection's address as its name.
type EchoClient struct {
	name string
}

// OnConnect is called when the connection is established, save the remote address as the client name.
func (t *EchoClient) OnConnect(n *net.TCPConn) { t.name = n.RemoteAddr().String() }
func (t *EchoClient) OnDisconnect(string)      {}
func (t *EchoClient) MatchTarget(s string, prefix bool) bool {
	if prefix {
		return strings.HasPrefix(t.name, s)
	} else {
		return t.name == string(s)
	}
}
func (t *EchoClient) Report() (any, bool, bool) { return "", true, false }
func (t *EchoClient) OnSend(b []byte)           { println("send:" + string(b)) }

// OnRecive is called when data is received. It returns nil for unfinished data
// and a slice containing a single SendMessage with the received data.
func (t *EchoClient) OnRecive(b []byte) ([]byte, []*SendMessage) {
	return nil, []*SendMessage{{Data: b}}
}
