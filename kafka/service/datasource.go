package service

import (
	"math/rand"
	"time"
)

// DataSource is responsible to produce some random data for kafka producer
// This data is usually received from some third-party service
type DataSource struct{}

// Produce data until timeEnd for each timeTick
func (ds *DataSource) Produce(done chan bool, svcType string, svcId string, timeEnd time.Duration, timeTick time.Duration) chan *Payload {
	result := make(chan *Payload)

	go func() {
		select {
		case <-time.After(timeEnd):
			close(done)
		}
	}()

	go func() {
		defer close(result)

		for {
			select {
			case <-time.After(timeTick):
				in := rand.Intn(1000)

				result <- &Payload{
					Error: nil,
					Content: &Content{
						Type:      svcType,
						ServiceId: svcId,
						TimeStamp: time.Now(),
						Attr: &Attribute{
							Status:      true,
							Consumption: in,
						},
					},
				}
			case <-done:
				return
			}
		}
	}()

	return result
}

// ProduceN will produce data of total times
func (ds *DataSource) ProduceN(done chan bool, data *Content, total int) chan *Payload {
	result := make(chan *Payload)

	go func() {
		defer func() {
			// NOTE: This is wrong way to stop producer!
			// It's here just for educational purposes!
			close(done)

			close(result)
		}()

		for i := 0; i < total; i++ {
			result <- &Payload{
				Error: nil,
				Content: &Content{
					ServiceId: data.ServiceId,
					Type: data.Type,
					TimeStamp: time.Now(),
					Attr: &Attribute{
						Status: data.Attr.Status,
						Consumption: data.Attr.Consumption,
					},
				},
			}
		}
	}()

	return result
}
