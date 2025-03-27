package tcpfactory

import (
	"net"
	"time"
)

type SendMessage struct {
	Data     []byte
	Interval time.Duration
}

type Client interface {
	// OnConnect is called when the connection is established
	OnConnect(*net.TCPConn)
	// OnDisconnect is called when the connection is closed
	OnDisconnect(string)
	// OnRecive is called when received data, return unfinished data, message need send to client
	OnRecive([]byte) ([]byte, []*SendMessage)
	// FormatDataToLog formats data to log, like hex string, json string
	FormatDataToLog(b []byte) string
	// MatchTarget is used to match if the target matches the client
	MatchTarget(string) bool
	// Report is used to report client status, return status data and if the client is registered
	Report() (any, bool)
}

type EmptyClient struct{}

func (t *EmptyClient) OnConnect(*net.TCPConn)                   {}
func (t *EmptyClient) OnDisconnect(string)                      {}
func (t *EmptyClient) OnRecive([]byte) ([]byte, []*SendMessage) { return nil, nil }
func (t *EmptyClient) FormatDataToLog(b []byte) string          { return string(b) }
func (t *EmptyClient) MatchTarget(string) bool                  { return false }
func (t *EmptyClient) Report() (any, bool)                      { return "", false }

type EchoClient struct {
	name string
}

// OnConnect is called when the connection is established, save the remote address as the client name.
func (t *EchoClient) OnConnect(n *net.TCPConn) {
	t.name = n.RemoteAddr().String()
}
func (t *EchoClient) OnDisconnect(string) {}

// OnRecive is called when data is received. It returns nil for unfinished data
// and a slice containing a single SendMessage with the received data.
func (t *EchoClient) OnRecive(b []byte) ([]byte, []*SendMessage) {
	return nil, []*SendMessage{
		{
			Data: b,
		},
	}
}
func (t *EchoClient) FormatDataToLog(b []byte) string { return string(b) }
func (t *EchoClient) MatchTarget(n string) bool       { return string(n) == t.name }
func (t *EchoClient) Report() (any, bool)             { return "", true }
