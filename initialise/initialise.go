package initialise

import (
	"fmt"

	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/mongo"
	"github.com/ONSdigital/go-ns/kafka"
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

// Possible names of Kafa Producsers
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
func (e *ExternalServiceList) GetProducer(kafkaBrokers []string, topic string, name KafkaProducerName, envMax int) (kafkaProducer kafka.Producer, err error) {
	kafkaProducer, err = kafka.NewProducer(kafkaBrokers, topic, envMax)
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
		err = fmt.Errorf("Kafka producer name not recognised: '%s'. Valid names: %v", name.String(), kafkaProducerNames)
	}

	return
}
