package initialise

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/mongo"
	kafka "github.com/ONSdigital/dp-kafka"
)

// ExternalServiceList represents a list of services
type ExternalServiceList struct {
	AuditProducer     bool
	DataBakerProducer bool
	DirectProducer    bool
	MongoDataStore    bool
}

// KafkaProducerName represents a type for kafka producer name used by iota constants
type KafkaProducerName int

// Possible names of Kafka Producers
const (
	Audit = iota
	DataBaker
	Direct
)

var kafkaProducerNames = []string{"Audit", "DataBaker", "Direct"}

// Values of the kafka producers names
func (k KafkaProducerName) String() string {
	return kafkaProducerNames[k]
}

// GetMongoDataStore returns an initialised connection to import store (mongo database)
func (e *ExternalServiceList) GetMongoDataStore(cfg *config.Configuration) (dataStore *mongo.Mongo, err error) {
	dataStore, err = mongo.NewDatastore(cfg.MongoDBURL, cfg.MongoDBDatabase, cfg.MongoDBCollection)
	if err != nil {
		return
	}
	e.MongoDataStore = true

	return
}

// GetProducer returns a kafka producer
func (e *ExternalServiceList) GetProducer(ctx context.Context, kafkaBrokers []string, topic string, name KafkaProducerName, envMax int) (kafkaProducer *kafka.Producer, err error) {

	producerChannels := kafka.CreateProducerChannels()
	kafkaProducer, err = kafka.NewProducer(ctx, kafkaBrokers, topic, envMax, producerChannels)

	if err != nil {
		return
	}

	switch {
	case name == Audit:
		e.AuditProducer = true
	case name == DataBaker:
		e.DataBakerProducer = true
	case name == Direct:
		e.DirectProducer = true
	default:
		err = fmt.Errorf("kafka producer name not recognised: '%s'. Valid names: %v", name.String(), kafkaProducerNames)
	}

	return
}
