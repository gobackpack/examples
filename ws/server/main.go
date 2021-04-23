package main

import (
	"github.com/gobackpack/websocket"
	"github.com/sirupsen/logrus"
)

func main() {
	server := &websocket.Server{
		Host:           "localhost",
		Port:           "8080",
		Endpoint:       "/",
		MessageHandler: &mHandler{},
	}

	done := make(chan bool)

	go server.Run(done)

	logrus.Info("listening for messages...")

	<-done
}

// mHandler example impl
type mHandler struct{
	counter int
}

func (h *mHandler) OnMessage(in []byte, reply func(int, []byte) error) {
	h.counter++
}

func (h *mHandler) OnError(err error) {
	// handle error from ws channel
	logrus.Error("error from ws connection: ", err)
	logrus.Info("received total: ", h.counter)
}