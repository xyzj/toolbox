package mq

import (
	"net/url"
	"testing"
	"time"

	"github.com/xyzj/toolbox/logger"
)

func TestURL(t *testing.T) {
	a := "mqtts://192.168.50.97:6821"
	u, _ := url.Parse(a)
	println(u.Host, u.Path, u.Port(), u.Scheme)
}

var (
	v3 *MqttClientV5
	v5 *MqttClientV5
)

func TestMQ5(t *testing.T) {
	opt := &MqttOpt{
		Addr: "mqtt://10.3.8.34:1883",
		// Addr:     "tls://192.168.50.83:1881",
		LogHeader: "one",
		Username:  "arx7",
		Passwd:    "arbalest",
		// Username: "SH50_DEV",
		// Passwd:   "dasfs@8124545",
		ClientID: "123122334e234d",
		Subscribe: map[string]byte{
			"22/#":        1,
			"123df/3/#":   1,
			"13323/#":     1,
			"123d32f/3/#": 1,
		},
		Logg: logger.NewConsoleLogger(),
	}
	v3, _ = NewMQTTClientV5(opt, func(topic string, body []byte) {
		println("v3 recv:", topic)
	})
	opt5 := &MqttOpt{
		Addr: "mqtt://10.3.8.34:1883",
		// Addr: "tls://192.168.50.83:1881",
		LogHeader: "two",
		Subscribe: map[string]byte{
			"22/#":      1,
			"133/#":     1,
			"123df/3/#": 1,
		},
		Username: "mqtest",
		Passwd:   "test0307",
		// Username: "SH50_DEV",
		// Passwd:   "dasfs@8124545",
		ClientID: "123122334e234d",
		Logg:     logger.NewConsoleLogger(),
	}
	v5, _ = NewMQTTClientV5(opt5, func(topic string, body []byte) {
		println("v5 recv:", topic)
	})

	for {
		err := v5.Write("123/12321", []byte("123123"))
		if err != nil {
			println(err.Error())
		}
		time.Sleep(time.Second * 10)
		// err = v5.Write("23842/2382", []byte("189273gksdhfksf"))
		// if err != nil {
		// 	t.Fatal(err)
		// 	return
		// }
		// time.Sleep(time.Second * 2)
	}
	// time.Sleep(time.Minute * 2)
}

func TestCli(t *testing.T) {
	// mqttdStart()
	opt := &MqttOpt{
		Addr:     "mqtt://127.0.0.1:1883",
		Username: "arx7",
		Passwd:   "arbalest",
		ClientID: "123122334",
		Subscribe: map[string]byte{
			"#": 1,
		},
		Logg: logger.NewConsoleLogger(),
	}
	v3, _ := NewMQTTClientV5(opt, func(topic string, body []byte) {
		time.Sleep(time.Second * 10)
		println("v3 recv:", topic)
	})

	for {
		time.Sleep(time.Second * 6)
		err := v3.Write("yiyang/asdfsdf", []byte("123123"))
		if err != nil {
			println(err.Error())
			continue
		}
		// err = v5.Write("23842/2382", []byte("189273gksdhfksf"))
		// if err != nil {
		// 	t.Fatal(err)
		// 	return
		// }
		// time.Sleep(time.Second * 2)
	}
	// time.Sleep(time.Minute * 2)
}
