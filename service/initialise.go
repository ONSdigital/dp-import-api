package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-api/config"
	"github.com/ONSdigital/dp-import-api/datastore"
	"github.com/ONSdigital/dp-import-api/mongo"
	kafka "github.com/ONSdigital/dp-kafka"
	dphttp "github.com/ONSdigital/dp-net/http"
)

// ExternalServiceList represents a list of services
type ExternalServiceList struct {
	MongoDataStore    bool
	DataBakerProducer bool
	DirectProducer    bool
	HealthCheck       bool
	Init              Initialiser
}

// KafkaProducerName represents a type for kafka producer name used by iota constants
type KafkaProducerName int

// Possible names of Kafka Producers
const (
	DataBaker = iota
	Direct
)

var kafkaProducerNames = []string{"DataBaker", "Direct"}

// Values of the kafka producers names
func (k KafkaProducerName) String() string {
	return kafkaProducerNames[k]
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		Init: initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetHealthCheck creates a healthcheck with versionInfo and sets the HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Configuration, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// GetMongoDataStore returns an initialised connection to import store (mongo database)
func (e *ExternalServiceList) GetMongoDataStore(cfg *config.Configuration) (dataStore datastore.DataStorer, err error) {
	dataStore, err = e.Init.DoGetMongoDataStore(cfg)
	if err != nil {
		return
	}
	e.MongoDataStore = true
	return
}

// GetProducer returns a kafka producer
func (e *ExternalServiceList) GetProducer(ctx context.Context, kafkaBrokers []string, topic string, name KafkaProducerName, envMax int) (kafkaProducer kafka.IProducer, err error) {

	kafkaProducer, err = e.Init.DoGetKafkaProducer(ctx, kafkaBrokers, topic, envMax)
	if err != nil {
		return
	}

	switch {
	case name == DataBaker:
		e.DataBakerProducer = true
	case name == Direct:
		e.DirectProducer = true
	default:
		err = fmt.Errorf("kafka producer name not recognised: '%s'. Valid names: %v", name.String(), kafkaProducerNames)
	}

	return
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Configuration, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

// DoGetMongoDataStore creates a mongoDB connection
func (e *Init) DoGetMongoDataStore(cfg *config.Configuration) (datastore.DataStorer, error) {
	return mongo.NewDatastore(cfg.MongoDBURL, cfg.MongoDBDatabase, cfg.MongoDBCollection)
}

// DoGetKafkaProducer cretes a new Kafka Producer
func (e *Init) DoGetKafkaProducer(ctx context.Context, kafkaBrokers []string, topic string, envMax int) (kafka.IProducer, error) {
	producerChannels := kafka.CreateProducerChannels()
	return kafka.NewProducer(ctx, kafkaBrokers, topic, envMax, producerChannels)
}
