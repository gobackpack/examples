package main

import (
	"github.com/gobackpack/examples/kafka"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("examples")

	kafka.RunExample()
}