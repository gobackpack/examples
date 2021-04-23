package main

import (
	"fmt"
	"github.com/gobackpack/websocket"
	"github.com/sirupsen/logrus"
)

func main() {
	client := &websocket.Client{
		MessageHandler: &mHandler{},
	}

	done := make(chan bool)
	ready := make(chan bool)

	go client.Connect(done, ready, "ws://localhost:8080")

	<-ready

	// ready to send messages to websocket channel
	for i := 0; i < 50000; i++ {
		if err := client.SendText([]byte(fmt.Sprint("message: ", i))); err != nil {
			logrus.Fatal("failed to send message: ", err)
		}
	}

	close(done)

	<-done
}

// mHandler example impl
type mHandler struct{}

func (h *mHandler) OnMessage(in []byte, reply func(int, []byte) error) {}

func (h *mHandler) OnError(err error) {
	logrus.Error("error from ws connection: ", err)
}