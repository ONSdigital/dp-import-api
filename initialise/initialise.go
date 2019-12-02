package initialise

import (
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

const (
	// AuditProducer represents a name for
	// the producer that writes to an audit topic
	AuditProducer = "audit-producer"
	// DataBakerProducer represents a name for the producer
	// that writes to a databaker import topic
	DataBakerProducer = "databaker-producer"
	// DirectProducer represents a name for the producer
	// that writes to an input file available topic
	DirectProducer = "direct-producer"
)

// GetMongoDataStore returns an initialised connection to import store (mongo database)
func (e *ExternalServiceList) GetMongoDataStore(cfg *config.Configuration) (dataStore *mongo.Mongo, err error) {
	dataStore, err = mongo.NewDatastore(cfg.MongoDBURL, cfg.MongoDBDatabase, cfg.MongoDBCollection)
	if err == nil {
		e.MongoDataStore = true
	}

	return
}

// GetProducer returns a kafka producer
func (e *ExternalServiceList) GetProducer(kafkaBrokers []string, topic, name string, envMax int) (kafkaProducer kafka.Producer, err error) {
	kafkaProducer, err = kafka.NewProducer(kafkaBrokers, topic, envMax)
	if err == nil {
		switch {
		case name == AuditProducer:
			e.AuditProducer = true
		case name == DataBakerProducer:
			e.DataBakerProducer = true
		case name == DirectProducer:
			e.DirectProducer = true
		}
	}

	return
}
