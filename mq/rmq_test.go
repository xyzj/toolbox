package mq

import (
	"os"
	"testing"
	"time"
)

func TestRMQ(t *testing.T) {
	var t1 time.Time
	sender := NewRMQProducer(&RabbitMQOpt{
		Addr:         "192.168.50.83:5672",
		Username:     "arx7",
		Passwd:       "arbalest",
		VHost:        "/",
		ExchangeName: "luwak_topic",
	}, nil)
	NewRMQConsumer(&RabbitMQOpt{
		Addr:         "192.168.50.83:5672",
		Username:     "arx7",
		Passwd:       "arbalest",
		VHost:        "/",
		ExchangeName: "luwak_topic",
		QueueName:    "test-queue",
		Subscribe:    []string{"aaa.#"},
	}, nil, func(topic string, body []byte) {
		println("--- recv len ---", len(body), time.Since(t1).String())
	})
	b, _ := os.ReadFile("a.json")
	for {
		time.Sleep(time.Second * 7)
		println("--- send")
		t1 = time.Now()
		sender.Send("aaa.longdata", b, time.Second)
	}
}
