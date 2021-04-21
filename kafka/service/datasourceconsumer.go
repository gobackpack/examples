package service

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

// DataSourceConsumer is responsible to consume and marshal data from DataSource (third-party service)
// This is because we usually do not send data from third-party services directly to Kafka
// We usually need to marshal it in some way, and possibly do something else with the data from DataSource (third-party service)
type DataSourceConsumer struct{}

// MarshalData from DataSource (third-party service)
// Do what you please before sending it to Kafka
func (dsc *DataSourceConsumer) MarshalData(done chan bool, svcId string, dataSource <-chan *Payload) chan []KafkaMessage {
	result := make(chan []KafkaMessage)
	kafkaMessages := make([]KafkaMessage, 0)

	go func() {
		defer func() {
			// send all data at once
			result <- kafkaMessages

			close(result)
		}()

		for {
			select {
			case p, ok := <-dataSource:
				if !ok {
					continue
				}

				if p.Error != nil {
					logrus.Error("data source consumer received error: ", p.Error)
					continue
				}

				b, err := json.Marshal(p)
				if err != nil {
					logrus.Errorf("data source consumer failed to marshal payload: %v, error: %v", p, err)
					continue
				}

				kafkaMessages = append(kafkaMessages, KafkaMessage{
					Key:   []byte(svcId),
					Value: b,
				})
			case <-done:
				return
			}
		}
	}()

	return result
}
