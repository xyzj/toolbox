package mq

// broker: https://github.com/nanomq/nanomq/releases/tag/0.21.6

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/cache"
	"github.com/xyzj/toolbox/logger"
)

var (
	payloadFormat byte   = 1
	messageExpiry uint32 = 600
)

var (
	ErrorResendCache  = fmt.Errorf("not connect to the server, cache sent")
	ErrorOptions      = fmt.Errorf("mqtt opt error")
	ErrorNotConnected = fmt.Errorf("not connect to the server")
)

var EmptyMQTTClientV5 = &MqttClientV5{
	empty: true,
	st:    &atomic.Bool{},
}

type mqttMessage struct {
	qos   byte
	body  []byte
	topic string
}

// MqttOpt mqtt 配置
type MqttOpt struct {
	// TLSConf 日志
	Logg logger.Logger
	// tls配置，默认为 InsecureSkipVerify: true
	TLSConf *tls.Config
	// 订阅消息，map[topic]qos
	Subscribe map[string]byte
	// 发送超时
	SendTimeo time.Duration
	// ClientID 客户端标示，会添加随机字符串尾巴，最大22个字符
	ClientID string
	// 服务端ip:port
	Addr string
	// 登录用户名
	Username string
	// 登录密码
	Passwd string
	// 日志前缀，默认 [MQTT]
	LogHeader string
	// 是否启用断连消息暂存
	EnableFailureCache bool
	// 最大缓存消息数量，默认10000
	FailureCacheMax int
	// 消息缓存时间，默认一小时
	FailureCacheExpire time.Duration
	// 消息失效的处置方法
	FailureCacheExpireFunc func(topic string, body []byte)
}

// MqttClientV5 mqtt客户端 5.0
type MqttClientV5 struct {
	cnf         *MqttOpt
	client      *autopaho.ConnectionManager
	failedCache *cache.AnyCache[*mqttMessage]
	st          *atomic.Bool
	ctxCancel   context.CancelFunc
	empty       bool
}

// Close close the mqtt client
func (m *MqttClientV5) Close() error {
	if m.empty {
		return nil
	}
	if m.client == nil {
		return nil
	}
	m.failedCache.Close()
	m.st.Store(false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := m.client.Disconnect(ctx)
	if err != nil {
		return err
	}
	m.ctxCancel()
	return nil
}

// Client return autopaho.ConnectionManager
func (m *MqttClientV5) Client() *autopaho.ConnectionManager {
	if m.empty {
		return nil
	}
	return m.client
}

// IsConnectionOpen 返回在线状态
func (m *MqttClientV5) IsConnectionOpen() bool {
	if m.empty {
		return false
	}
	return m.st.Load()
}

// Write 以qos0发送消息
func (m *MqttClientV5) Write(topic string, body []byte) error {
	return m.WriteWithQos(topic, body, 0)
}

// WriteWithQos 发送消息，可自定义qos
func (m *MqttClientV5) WriteWithQos(topic string, body []byte, qos byte) error {
	if m.empty {
		return nil
	}
	if !m.st.Load() || m.client == nil { // 未连接状态
		if m.cnf.EnableFailureCache {
			if m.failedCache.Len() < m.cnf.FailureCacheMax {
				m.failedCache.StoreWithExpire(
					time.Now().Format(time.RFC3339Nano),
					&mqttMessage{
						topic: topic,
						body:  body,
						qos:   qos,
					},
					m.cnf.FailureCacheExpire)
				return ErrorResendCache
			}
		}
		return ErrorNotConnected
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()
	err := m.client.PublishViaQueue(ctx, &autopaho.QueuePublish{
		Publish: &paho.Publish{
			QoS:     qos,
			Topic:   topic,
			Payload: body,
			Retain:  false,
			Properties: &paho.PublishProperties{
				PayloadFormat: &payloadFormat,
				MessageExpiry: &messageExpiry,
				ContentType:   "text/plain",
			},
		},
	})
	if err != nil {
		m.cnf.Logg.Debug(m.cnf.LogHeader + "serr:" + topic + "|" + err.Error())
		return err
	}
	m.cnf.Logg.Debug(m.cnf.LogHeader + " s:" + topic)
	return nil
}

// NewMQTTClientV5 创建一个5.0的mqtt client
func NewMQTTClientV5(opt *MqttOpt, recvCallback func(topic string, body []byte)) (*MqttClientV5, error) {
	if opt == nil {
		return EmptyMQTTClientV5, ErrorOptions
	}
	if opt.SendTimeo == 0 {
		opt.SendTimeo = time.Second * 5
	}
	if opt.ClientID == "" {
		opt.ClientID = toolbox.GetRandomString(19, true)
	}
	if opt.LogHeader == "" {
		opt.LogHeader = "[mqtt]"
	}
	if opt.TLSConf == nil {
		opt.TLSConf = &tls.Config{InsecureSkipVerify: true}
	}
	if recvCallback == nil {
		recvCallback = func(topic string, body []byte) {}
	}
	if opt.Logg == nil {
		opt.Logg = &logger.NilLogger{}
	}
	if opt.FailureCacheMax == 0 {
		opt.FailureCacheMax = 10000
	}
	if opt.FailureCacheExpire == 0 {
		opt.FailureCacheExpire = time.Hour
	}
	if !strings.Contains(opt.Addr, "://") {
		switch {
		case strings.Contains(opt.Addr, ":1881"):
			opt.Addr = "tls://" + opt.Addr
		default: // case strings.Contains(opt.Addr,":1883"):
			opt.Addr = "mqtt://" + opt.Addr
		}
	}
	u, err := url.Parse(opt.Addr)
	if err != nil {
		return EmptyMQTTClientV5, err
	}
	connSt := &atomic.Bool{}
	code142 := &atomic.Bool{}
	code133 := &atomic.Bool{}
	failedCache := cache.NewAnyCacheWithExpireFunc(opt.FailureCacheExpire, func(m map[string]*mqttMessage) {
		for _, v := range m {
			opt.FailureCacheExpireFunc(v.topic, v.body)
		}
	})
	conf := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{u},
		KeepAlive:                     55,
		SessionExpiryInterval:         0,
		CleanStartOnInitialConnection: true,
		ConnectUsername:               opt.Username,
		ConnectPassword:               []byte(opt.Passwd),
		TlsCfg:                        opt.TLSConf,
		ConnectTimeout:                time.Second * 5,
		ReconnectBackoff: func(i int) time.Duration {
			if i <= 0 {
				return 0
			}
			return time.Second * time.Duration(rand.Int31n(30)+30)
		},
		ConnectPacketBuilder: func(c *paho.Connect, u *url.URL) (*paho.Connect, error) {
			c.CleanStart = true
			if code142.Load() || code133.Load() {
				c.ClientID = opt.ClientID + "_" + toolbox.GetRandomString(9, true)
			}
			return c, nil
		},
		OnConnectionUp: func(cm *autopaho.ConnectionManager, c *paho.Connack) {
			connSt.Store(true)
			code133.Store(false)
			code142.Store(false)
			if len(opt.Subscribe) > 0 {
				x := make([]paho.SubscribeOptions, 0, len(opt.Subscribe))
				for k, v := range opt.Subscribe {
					x = append(x, paho.SubscribeOptions{
						Topic: k,
						QoS:   v,
					})
				}
				cm.Subscribe(context.Background(), &paho.Subscribe{
					Subscriptions: x,
				})
			}
			opt.Logg.System(opt.LogHeader + " success connect to " + opt.Addr)
			// 对失败消息进行补发
			if opt.EnableFailureCache {
				var err error
				failedCache.ForEach(func(key string, value *mqttMessage) bool {
					ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
					defer cancel()
					err = cm.PublishViaQueue(ctx, &autopaho.QueuePublish{
						Publish: &paho.Publish{
							QoS:     value.qos,
							Topic:   value.topic,
							Payload: value.body,
							Retain:  false,
							Properties: &paho.PublishProperties{
								PayloadFormat: &payloadFormat,
								MessageExpiry: &messageExpiry,
								ContentType:   "text/plain",
							},
						},
					})
					if err != nil {
						opt.Logg.Error(opt.LogHeader + " reerr:" + value.topic + "|" + err.Error())
					} else {
						opt.Logg.Info(fmt.Sprintf("%s re:%s|%v", opt.LogHeader, value.topic, value.body))
					}
					return true
				})
			}
		},
		OnConnectError: func(err error) {
			connSt.Store(false)
			if strings.Contains(err.Error(), "reason: 133") { // exmq, need rename client id
				code133.Store(true)
			}
			opt.Logg.Error(opt.LogHeader + " connect error: " + err.Error())
		},
		ClientConfig: paho.ClientConfig{
			ClientID: opt.ClientID,
			OnServerDisconnect: func(d *paho.Disconnect) {
				connSt.Store(false)
				if d.ReasonCode == 142 { // need rename client id
					code142.Store(true)
				}
				if d.Properties != nil {
					opt.Logg.Error(opt.LogHeader + " server requested disconnect, reason code: " + strconv.Itoa(int(d.ReasonCode)) + " " + d.Properties.ReasonString)
				} else {
					opt.Logg.Error(opt.LogHeader + " server requested disconnect, reason code: " + strconv.Itoa(int(d.ReasonCode)))
				}
			},
			OnClientError: func(err error) {
				connSt.Store(false)
				if err == io.EOF {
					code142.Store(true)
				}
				opt.Logg.Error(opt.LogHeader + " client error: " + err.Error())
			},
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					opt.Logg.Debug(opt.LogHeader + " r:" + pr.Packet.Topic)
					recvCallback(pr.Packet.Topic, pr.Packet.Payload)
					return true, nil
				},
			},
		},
	}
	ctxClose, funClose := context.WithCancel(context.TODO())
	cm, err := autopaho.NewConnection(ctxClose, conf)
	if err != nil {
		opt.Logg.Error(opt.LogHeader + " new connection error: " + err.Error())
		funClose()
		return EmptyMQTTClientV5, err
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()
	cm.AwaitConnection(ctx)

	return &MqttClientV5{
		client:      cm,
		st:          connSt,
		cnf:         opt,
		ctxCancel:   funClose,
		failedCache: failedCache,
	}, nil
}
