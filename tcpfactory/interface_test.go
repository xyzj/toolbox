package tcpfactory

import (
	"bytes"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/logger"
)

type EmptyTCP1 struct {
	s string
}

func (t *EmptyTCP1) OnConnect(*net.TCPConn) {}
func (t *EmptyTCP1) OnDisconnect(string) {
	t.s = strconv.Itoa(rand.Intn(100))
}

func (t *EmptyTCP1) OnRecive(b []byte) ([]byte, []*SendMessage) {
	println("recv", string(b))
	time.Sleep(time.Second * 4)
	panic("test close")
}
func (t *EmptyTCP1) OnSend(b []byte)               {}
func (t *EmptyTCP1) MatchTarget(string, bool) bool { return true }
func (t *EmptyTCP1) Report() (any, bool, bool)     { return t.s, true, false }

func TestFac(t *testing.T) {
	tm, _ := NewTcpFactory(WithBindAddr(":6819"),
		WithLogger(logger.NewConsoleLogger()),
		WithTcpClient(&EmptyTCP1{}))
	go tm.Listen()
	time.Sleep(time.Second * 1)
	cli, _ := net.DialTCP("tcp", nil, tm.addr)
	name := cli.LocalAddr().String()
	println(name)
	t1 := time.NewTicker(time.Second * 10)
	t2 := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-t1.C:
			cli.Write([]byte("hello"))
		case <-t2.C:
			b := bytes.Buffer{}
			for range 1 {
				b.Reset()
				x := toolbox.GetRandomString(10, true)
				println("-----x", x)
				b.WriteString(x)
				z := &SendMessage{
					Data:     json.Bytes(b.String()),
					Interval: time.Second * 2,
				}
				println("-----z", z)
				tm.WriteTo(name, z)
			}
		}
	}
}

func TestAAA(t *testing.T) {
	var chanaaa = make(chan *SendMessage, 10)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for m := range chanaaa {
			if m == shutmedown {
				println("shut me down")
				return
			} else {
				println("else")
			}
		}
	}()
	chanaaa <- &SendMessage{
		Data:     []byte{0x73, 0x68, 0x75, 0x74, 0x20, 0x6d, 0x65, 0x20, 0x64, 0x6f, 0x77, 0x6e},
		Interval: 0,
	}
	chanaaa <- ShutMeDown()
	wg.Wait()
}
