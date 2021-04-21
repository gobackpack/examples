package service

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

// KafkaProducer is responsible to produce data to Kafka
type KafkaProducer struct {
	KafkaProducerLib *kafka.Producer
}

// NewKafkaProducer will create new KafkaProducer with underlying library
func NewKafkaProducer(brokers, clientId string) *KafkaProducer {
	kafkaProducerLib, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"client.id":         clientId,
		"acks":              "all"})
	if err != nil {
		logrus.Fatal(err)
	}

	return &KafkaProducer{KafkaProducerLib: kafkaProducerLib}
}

// Send data to Kafka
func (producer *KafkaProducer) Send(topic string, kafkaMessages []KafkaMessage) {
	done := make(chan bool)

	go producer.listenEvents(done)

	for _, m := range kafkaMessages {
		producer.KafkaProducerLib.ProduceChannel() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            m.Key,
			Value:          m.Value,
		}
	}

	<-done
}

// Close KafkaProducer
// Why is it wrapped in a function? For the same reason we wrapped kafka producer library in KafkaProducer struct
// The less we expose underlying library, more flexible we are
func (producer *KafkaProducer) Close() {
	producer.KafkaProducerLib.Close()
}

func (producer *KafkaProducer) listenEvents(done chan bool) {
	defer close(done)

	for e := range producer.KafkaProducerLib.Events() {
		switch e.(type) {
		case *kafka.Message:
			return
		}
	}
}