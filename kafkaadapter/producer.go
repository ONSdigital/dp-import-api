package kafkaadapter

import kafka "github.com/ONSdigital/dp-kafka"

func NewProducerAdapter(producer *kafka.Producer) *Producer {
	return &Producer{kafkaProducer: producer}
}

// exposes an output function, to satisfy the interface used by go-ns libraries
type Producer struct {
	kafkaProducer *kafka.Producer
}

func (p Producer) Output() chan []byte {
	return p.kafkaProducer.Channels().Output
}
