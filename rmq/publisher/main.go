package main

import (
	"github.com/gobackpack/rmq"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func main() {
	credentials := rmq.NewCredentials()

	// config
	config := rmq.NewConfig()
	config.Exchange = "test_exchange_a"
	config.Queue = "test_queue_a"
	config.RoutingKey = "test_queue_a"

	// setup connection
	publisher := &rmq.Connection{
		Credentials: credentials,
		Config:      config,
		ResetSignal: make(chan int),
	}

	// pass true if there is only one publisher config
	// else manually call publisher.ApplyConfig(*Config) for each configuration (declare queues) and
	// call publisher.PublishWithConfig(*Config) if publisher.Config was not set!
	if err := publisher.Connect(true); err != nil {
		logrus.Fatal(err)
	}

	// optionally ListenNotifyClose and HandleResetSignalPublisher
	done := make(chan bool)

	go publisher.ListenNotifyClose(done)

	go publisher.HandleResetSignalPublisher(done)

	configB := rmq.NewConfig()
	configB.Exchange = "test_exchange_b"
	configB.Queue = "test_queue_b"
	configB.RoutingKey = "test_queue_b"

	if err := publisher.ApplyConfig(configB); err != nil {
		logrus.Error(err)
		return
	}

	configC := rmq.NewConfig()
	configC.Exchange = "test_exchange_c"
	configC.Queue = "test_queue_c"
	configC.RoutingKey = "test_queue_c"

	if err := publisher.ApplyConfig(configC); err != nil {
		logrus.Error(err)
		return
	}

	for i := 0; i < 10; i++ {
		// publish to default config exchange/queue
		if err := publisher.Publish([]byte(uuid.New().String())); err != nil {
			logrus.Error(err)
		}

		// publish to configB exchange/queue
		if err := publisher.PublishWithConfig(configB, []byte(uuid.New().String())); err != nil {
			logrus.Error(err)
		}

		// publish to configC exchange/queue
		if err := publisher.PublishWithConfig(configC, []byte(uuid.New().String())); err != nil {
			logrus.Error(err)
		}
	}

	close(done)

	<-done

	logrus.Info("rmq publisher sent messages")
}
