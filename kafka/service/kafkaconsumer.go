package service

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

// KafkaConsumer is responsible to consume data from Kafka
type KafkaConsumer struct {
	*kafka.Consumer
}

func NewKafkaConsumer(brokers, groupId string) *KafkaConsumer {
	kafkaConsumerLib, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               brokers,
		"group.id":                        groupId,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.partition.eof":            true,
		"auto.offset.reset":               "earliest"})
	if err != nil {
		logrus.Fatal("failed to create kafka consumer: ", err)
	}

	return &KafkaConsumer{Consumer: kafkaConsumerLib}
}

// Consume data from Kafka
func (consumer *KafkaConsumer) Consume(done chan bool, topics []string) chan *KafkaMessage {
	if err := consumer.SubscribeTopics(topics, nil); err != nil {
		logrus.Fatal("subscription to topics failed: ", err)
	}

	response := make(chan *KafkaMessage)

	go func() {
		defer func() {
			close(response)
			logrus.Info("consumer stopped reading")
		}()

		for {
			select {
			case ev := <-consumer.Events():
				switch e := ev.(type) {
				case kafka.AssignedPartitions:
					logrus.Info("partitions assignment: ", e)

					if err := consumer.Assign(e.Partitions); err != nil {
						logrus.Error("failed to assign partitions: ", err)
					}
				case kafka.RevokedPartitions:
					logrus.Info("partitions revoke: ", e)

					if err := consumer.Unassign(); err != nil {
						logrus.Error("failed to revoke partitions: ", err)
					}
				case *kafka.Message:
					response <- &KafkaMessage{
						Key: e.Key,
						Value: e.Value,
					}
				case kafka.PartitionEOF:
					logrus.Info("eof for partition: ", e.Partition)
				case kafka.Error:
					logrus.Error("received error: ", e.Error())
				}
			case <-done:
				return
			}
		}
	}()

	return response
}

// Close KafkaConsumer
// Why is it wrapped in a function? For the same reason we wrapped kafka consumer library in KafkaConsumer struct
// The less we expose underlying library, more flexible we are
func (consumer *KafkaConsumer) Close() {
	if err := consumer.Consumer.Close(); err != nil {
		logrus.Error("failed to close kafka consumer: ", err)
	}
}