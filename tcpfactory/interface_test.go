package tcpfactory

import (
	"bytes"
	"math/rand"
	"net"
	"strconv"
	"testing"
)

type EmptyTCP1 struct {
	s string
}

func (t *EmptyTCP1) OnConnect(*net.TCPConn) {}
func (t *EmptyTCP1) OnDisconnect(string) {
	t.s = strconv.Itoa(rand.Intn(100))
}
func (t *EmptyTCP1) OnRecive([]byte) ([]byte, []*SendMessage) { return nil, nil }
func (t *EmptyTCP1) FormatDataToLog(b []byte) string          { return string(b) }
func (t *EmptyTCP1) MatchTarget(string) bool                  { return false }
func (t *EmptyTCP1) Report() any                              { return t.s }

func TestFac(t *testing.T) {
	a := bytes.Buffer{}
	a.WriteString("*&^%%$^&*(GHJDKSDHJGFSDDF)")
	b := a.Bytes()
	c := a.Bytes()
	println(len("*&^%%$^&*(GHJDKSDHJGFSDDF)"), string(b), a.Available(), a.Len(), string(c))
}
