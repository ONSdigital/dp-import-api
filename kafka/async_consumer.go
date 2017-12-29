package kafka

import (
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"context"
	"errors"
)

// AsyncConsumer provides a higher level of abstraction over a kafka consumer.
// Instead of dealing with a channel of messages, the AsyncConsumer allows
// a function to be provided that is called for each message consumed.
// It maintains its own go routine to consume messages, and provides a simple
// graceful shutdown with a blocking close method.
type AsyncConsumer struct {
	closing chan bool
	closed  chan bool
}

// MessageConsumer provides a generic interface for consuming []byte messages
type MessageConsumer interface {
	Incoming() chan kafka.Message
}

// NewAsyncConsumer returns a new consumer instance.
func NewAsyncConsumer() *AsyncConsumer {
	return &AsyncConsumer{
		closing: make(chan bool),
		closed:  make(chan bool),
	}
}

// Consume converts messages to event instances, and pass the event to the provided handler.
func (consumer *AsyncConsumer) Consume(messageConsumer MessageConsumer, handlerFunc func(message kafka.Message)) {

	go func() {
		defer close(consumer.closed)

		for {
			select {

			case message := <-messageConsumer.Incoming():
				handlerFunc(message)

			case <-consumer.closing:
				log.Info("closing event consumer loop", nil)
				return
			}
		}
	}()
}

// Close safely closes the consumer and releases all resources
func (consumer *AsyncConsumer) Close(ctx context.Context) (err error) {

	if ctx == nil {
		ctx = context.Background()
	}

	close(consumer.closing)

	select {
	case <-consumer.closed:
		log.Info("successfully closed event consumer", nil)
		return nil
	case <-ctx.Done():
		log.Info("shutdown context time exceeded, skipping graceful shutdown of kafka consumer", nil)
		return errors.New("shutdown context timed out")
	}

}