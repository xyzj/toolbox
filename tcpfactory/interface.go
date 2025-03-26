package tcpfactory

import (
	"net"
	"time"
)

type SendMessage struct {
	Data     []byte
	Interval time.Duration
}

type TCPFactory interface {
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

type EmptyTCP struct{}

func (t *EmptyTCP) OnConnect(*net.TCPConn)                   {}
func (t *EmptyTCP) OnDisconnect(string)                      {}
func (t *EmptyTCP) OnRecive([]byte) ([]byte, []*SendMessage) { return nil, nil }
func (t *EmptyTCP) FormatDataToLog(b []byte) string          { return string(b) }
func (t *EmptyTCP) MatchTarget(string) bool                  { return false }
func (t *EmptyTCP) Report() (any, bool)                      { return "", false }
