package kafka

import (
	"github.com/gin-gonic/gin"
	"github.com/gobackpack/examples/kafka/service"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

// configuration values
var (
	ServiceId   = "b789ab42-6f26-4868-94da-03224b94f3d9"
	GroupId     = "kafka-go-consumer-1"
	ServiceType = "kafka-go-service-type"
	Topic       = "kafka-go-topic"
	Brokers     = "123.123.123.123:4567,124.124.124.124:4568"
)

func RunExample() {
	logrus.Info("bravo")

	// 1.) kafka producer usage example

	dataSource := &service.DataSource{}
	dataSourceConsumer := &service.DataSourceConsumer{}
	producer := service.NewKafkaProducer(Brokers, ServiceId)

	r := gin.Default()

	r.POST("/publish", func(c *gin.Context) {
		total, _ := strconv.Atoi(c.Query("total"))

		data := &service.Content{}

		if err := c.ShouldBind(&data); err != nil {
			logrus.Fatal(err)
		}

		done := make(chan bool)

		dscStartTime := time.Now()

		payload := dataSource.ProduceN(done, data, total)

		// do some other work, dataSource.Produce is non-blocking

		kafkaMessages := <-dataSourceConsumer.MarshalData(done, ServiceId, payload)

		// do some other work, dataSourceConsumer.MarshalData is non-blocking

		dscEndTime := time.Now()

		logrus.Infof("data source consumer transformed/marshaled %v messages in %v", len(kafkaMessages), dscEndTime.Sub(dscStartTime))

		kafkaSendStartTime := time.Now()

		producer.Send(Topic, kafkaMessages)

		kafkaSendEndTime := time.Now()

		logrus.Infof("sent %v messages to kafka in %v", len(kafkaMessages), kafkaSendEndTime.Sub(kafkaSendStartTime))
	})

	// 2.) kafka consumer usage example

	consumer := service.NewKafkaConsumer(Brokers, GroupId)

	done := make(chan bool)

	message := consumer.Consume(done, []string{Topic})

	// do some other work, consumer.Consume is non-blocking

	go func() {
		for {
			select {
			case m := <-message:
				logrus.Infof("consumer received message: %s, %s", m.Key, m.Value)
			}
		}
	}()

	if err := r.Run(); err != nil {
		logrus.Fatal(err)
	}

	// 3.) make sure we close our producer and consumer
	// This part is unreachable, this is for demonstration only

	producer.Close()
	close(done)
	consumer.Close()
}
