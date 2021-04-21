package service

import "time"

// Payload is produced from DataSource (third-party service data)
// Then it is marshaled into KafkaMessage
type Payload struct {
	Error   error
	Content *Content
}

// Content from Payload
type Content struct {
	Type      string
	ServiceId string
	TimeStamp time.Time
	Attr      *Attribute
}

// Attributed from Content
type Attribute struct {
	Status      bool
	Consumption int
}

// KafkaMessage is passed to KafkaProducer and received from KafkaConsumer
// This data should be compatible with kafka data (usually key-value pairs is sent/received)
type KafkaMessage struct {
	Key   []byte
	Value []byte
}