package tcpfactory

import (
	"net"
	"strings"
	"time"
)

type SendMessage struct {
	Data     []byte
	Interval time.Duration
}

// Client defines the interface for handling TCP connection lifecycle and message processing.
// It provides methods for connection events, data handling, logging, and client status reporting.
type Client interface {
	// Format formats data to log, like hex string, json string
	Format(data []byte) string
	// MatchTarget is used to match if the target matches the client
	MatchTarget(target string, prefix bool) bool
	// Report is used to report client status, return status data and if the client is registered
	Report() (any, bool)
	// OnConnect is called when the connection is established
	OnConnect(conn *net.TCPConn)
	// OnDisconnect is called when the connection is closed
	OnDisconnect(reson string)
	// OnRecive is called when received data, return unfinished data, message need send to client
	OnRecive(data []byte) ([]byte, []*SendMessage)
}

// EmptyClient is a no-op implementation of the Client interface that provides default empty method implementations.
type EmptyClient struct{}

func (t *EmptyClient) Format(data []byte) string                { return string(data) }
func (t *EmptyClient) MatchTarget(string, bool) bool            { return false }
func (t *EmptyClient) Report() (any, bool)                      { return "", false }
func (t *EmptyClient) OnConnect(*net.TCPConn)                   {}
func (t *EmptyClient) OnDisconnect(string)                      {}
func (t *EmptyClient) OnRecive([]byte) ([]byte, []*SendMessage) { return nil, nil }

// EchoClient is a simple implementation of the Client interface that echoes received data
// and stores the remote connection's address as its name.
type EchoClient struct {
	name string
}

func (t *EchoClient) Format(data []byte) string { return string(data) }
func (t *EchoClient) MatchTarget(n string, prefix bool) bool {
	if prefix {
		return strings.HasPrefix(t.name, n)
	} else {
		return t.name == string(n)
	}
}
func (t *EchoClient) Report() (any, bool) { return "", true }

// OnConnect is called when the connection is established, save the remote address as the client name.
func (t *EchoClient) OnConnect(n *net.TCPConn) { t.name = n.RemoteAddr().String() }
func (t *EchoClient) OnDisconnect(string)      {}

// OnRecive is called when data is received. It returns nil for unfinished data
// and a slice containing a single SendMessage with the received data.
func (t *EchoClient) OnRecive(data []byte) ([]byte, []*SendMessage) {
	return nil, []*SendMessage{{Data: data}}
}
