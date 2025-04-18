// Package mq mqtt 和 rmq 相关功能模块
package mq

import (
	"crypto/tls"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/loopfunc"
)

// MqttClient mqtt客户端
type MqttClient struct {
	client mqtt.Client
}

// Close close the mqtt client
func (m *MqttClient) Close() error {
	if m.client == nil {
		return fmt.Errorf("not connect to the server")
	}
	m.client.Disconnect(3000)
	return nil
}

// Client return mqtt.Client
func (m *MqttClient) Client() mqtt.Client {
	return m.client
}

// IsConnectionOpen 返回在线状态
func (m *MqttClient) IsConnectionOpen() bool {
	if m.client == nil {
		return false
	}
	return m.client.IsConnectionOpen()
}

// Write 以qos0发送消息
func (m *MqttClient) Write(topic string, body []byte) error {
	return m.WriteWithQos(topic, body, 0)
}

// WriteWithQos 发送消息，可自定义qos
func (m *MqttClient) WriteWithQos(topic string, body []byte, qos byte) error {
	if m.client == nil {
		return fmt.Errorf("not connect to the server")
	}
	t := m.client.Publish(topic, qos, false, body)
	t.Wait()
	return t.Error()
}

// NewMQTTClient 创建一个mqtt客户端 3.11
func NewMQTTClient(opt *MqttOpt, recvCallback func(topic string, body []byte)) (*MqttClient, error) {
	if opt == nil {
		return nil, fmt.Errorf("mqtt opt error")
	}
	if opt.SendTimeo == 0 {
		opt.SendTimeo = time.Second * 5
	}
	if opt.LogHeader == "" {
		opt.LogHeader = "[MQTT3]"
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

	if opt.ClientID == "" {
		opt.ClientID += "_" + toolbox.GetRandomString(20, true)
	}
	if len(opt.ClientID) > 22 {
		opt.ClientID = opt.ClientID[:22]
	}
	needSub := len(opt.Subscribe) > 0
	doneSub := false
	xopt := mqtt.NewClientOptions()
	xopt.AddBroker("tcp://" + opt.Addr)
	xopt.SetClientID(opt.ClientID)
	xopt.SetUsername(opt.Username)
	xopt.SetPassword(opt.Passwd)
	xopt.SetTLSConfig(opt.TLSConf)
	xopt.SetWriteTimeout(opt.SendTimeo) // 发送3秒超时
	xopt.SetConnectTimeout(time.Second * 10)
	xopt.SetConnectRetry(true)
	xopt.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		opt.Logg.Error(opt.LogHeader + " connection lost, " + err.Error())
		doneSub = false
	})
	xopt.SetOnConnectHandler(func(client mqtt.Client) {
		opt.Logg.System(opt.LogHeader + " Success connect to " + opt.Addr)
	})
	client := mqtt.NewClient(xopt)
	go loopfunc.LoopFunc(func(params ...interface{}) {
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			opt.Logg.Error(opt.LogHeader + " " + token.Error().Error())
			panic(token.Error())
		}
		for {
			if needSub && !doneSub && client.IsConnectionOpen() {
				client.SubscribeMultiple(opt.Subscribe, func(client mqtt.Client, msg mqtt.Message) {
					defer func() {
						if err := recover(); err != nil {
							opt.Logg.Error(opt.LogHeader + fmt.Sprintf(" %+v", errors.WithStack(err.(error))))
						}
					}()
					opt.Logg.Debug(opt.LogHeader + " DR:" + msg.Topic() + "; " + json.String(msg.Payload()))
					recvCallback(msg.Topic(), msg.Payload())
				})
				doneSub = true
			}
			time.Sleep(time.Second * 20)
		}
	}, opt.LogHeader, opt.Logg.DefaultWriter())
	return &MqttClient{client: client}, nil
}
