// Package tcpfactory 高可用性的tcp服务
package tcpfactory

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/loopfunc"
	"github.com/xyzj/toolbox/queue"
)

type tcpCore struct {
	conn               *net.TCPConn // 连接实例
	sendQueue          *queue.HighLowQueue[*SendMessage]
	closeOnce          *sync.Once
	readCache          *bytes.Buffer // 数据读取临时缓存
	tcpClient          Client        // 设备功能模块
	logg               logger.Logger
	timeLastWrite      time.Time          // 上次发送时间
	timeLastRead       time.Time          // 上次数据读取时间
	readTimeout        time.Duration      // 读取超时
	writeTimeout       time.Duration      // 发送超时
	writeIntervalTimer *time.Timer        // 发送间隔计时
	closeCtx           context.Context    // 关闭上下文
	closeFunc          context.CancelFunc // 关闭事件
	readBuffer         []byte             // 读取缓存
	sockID             uint64             // 实例id
	remoteAddr         string             // 远端地址
	closed             atomic.Bool        // 是否已关闭
}

func (t *tcpCore) formatLog(s string) string {
	return fmt.Sprintf("[%s] %s", t.remoteAddr, s)
}

func (t *tcpCore) connect(conn *net.TCPConn, msgs ...*SendMessage) {
	t.conn = conn
	t.closed.Store(false)
	t.remoteAddr = conn.RemoteAddr().String()
	t.closeOnce = new(sync.Once)
	t.timeLastRead = time.Now()
	t.timeLastWrite = time.Now()
	t.sendQueue.Open()
	t.logg.Info(t.formatLog("new connection established"))
	t.tcpClient.OnConnect(conn)
	loopfunc.GoFunc(func(params ...any) {
		t1 := time.NewTicker(time.Second)
		t1.Stop()
		for _, msg := range msgs {
			if t.closed.Load() {
				return
			}
			if len(msg.Data) == 0 && msg.Interval > 0 {
				t1.Reset(msg.Interval)
				select {
				case <-t1.C:
				case <-t.closeCtx.Done():
					return
				}
			} else {
				t.sendQueue.Put(msg)
			}
		}
	}, "say hello", t.logg.DefaultWriter())
}

func (t *tcpCore) disconnect(s string) {
	t.closeOnce.Do(func() {
		t.closed.Store(true)
		t.conn.Close()
		t.closeFunc()
		t.sendQueue.Close()
		t.readCache.Reset()
		t.writeIntervalTimer.Stop()
		t.logg.Debug(t.formatLog("close:" + s))
		t.tcpClient.OnDisconnect(s)
	})
}

func (t *tcpCore) recv() {
	var err error
	var n int
	var d, unfinish []byte
	var echo []*SendMessage
	for !t.closed.Load() {
		if err = t.conn.SetReadDeadline(time.Now().Add(t.readTimeout)); err != nil { // time.Duration(tcpReadTimeout)
			t.disconnect("set read timeout error: " + err.Error())
			return
		}
		n, err = t.conn.Read(t.readBuffer)
		if err != nil {
			if err == io.EOF {
				t.disconnect("remote close: " + err.Error())
			} else {
				t.disconnect("read error: " + err.Error())
			}
			return
		}
		if n == 0 {
			continue
		}
		t.timeLastRead = time.Now()
		d = t.readBuffer[:n]
		t.logg.Debug(t.formatLog("read:" + hex.EncodeToString(d)))
		// 检查缓存
		if t.readCache.Len() > 0 {
			t.readCache.Write(d)
			d = t.readCache.Bytes()
		}
		// 清理缓存
		t.readCache.Reset()
		// 数据解析
		unfinish, echo = t.tcpClient.OnRecive(d)
		if len(unfinish) > 0 {
			t.readCache.Write(unfinish)
			t.logg.Debug(t.formatLog("read unfinish:" + hex.EncodeToString(unfinish)))
		}
		if len(echo) > 0 {
			for _, s := range echo {
				t.sendQueue.PutFront(s)
			}
		}
	}
}

func (t *tcpCore) send() {
	var msg *SendMessage
	var ok bool
	var err error
	for !t.closed.Load() {
		if msg, ok = t.sendQueue.Get(); ok {
			if t.writeTimeout > 0 {
				err = t.conn.SetWriteDeadline(time.Now().Add(t.writeTimeout))
				if err != nil {
					t.disconnect("set send timeout error: " + err.Error())
					return
				}
			}
			if len(msg.Data) > 0 {
				_, err = t.conn.Write(msg.Data)
				if err != nil {
					t.disconnect("send error: " + err.Error())
					return
				}
				t.timeLastWrite = time.Now()
				t.logg.Debug(t.formatLog("send:" + hex.EncodeToString(msg.Data)))
				t.tcpClient.OnSend(msg.Data)
			}
			if msg.Interval > 0 {
				t.writeIntervalTimer.Reset(msg.Interval)
				select {
				case <-t.writeIntervalTimer.C:
					continue
				case <-t.closeCtx.Done():
					return
				}
			}
		} else {
			t.disconnect("load send message failed")
			return
		}
	}
}

func (t *tcpCore) writeTo(target string, msgs ...*SendMessage) bool {
	if t.closed.Load() {
		return false
	}
	if t.tcpClient.MatchTarget(target, false) {
		for _, msg := range msgs {
			t.sendQueue.Put(msg)
		}
		return true
	}
	return false
}

func (t *tcpCore) healthReport() (any, bool) {
	if t.closed.Load() {
		return "", false
	}
	return t.tcpClient.Report()
}
