package config

import (
	"time"

	"encoding/json"

	"github.com/kelseyhightower/envconfig"
)

// Configuration structure which hold information for configuring the import API
type Configuration struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	Host                       string        `envconfig:"HOST"`
	Brokers                    []string      `envconfig:"KAFKA_ADDR"`
	DatabakerImportTopic       string        `envconfig:"DATABAKER_IMPORT_TOPIC"`
	InputFileAvailableTopic    string        `envconfig:"INPUT_FILE_AVAILABLE_TOPIC"`
	KafkaMaxBytes              int           `envconfig:"KAFKA_MAX_BYTES"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"          json:"-"`
	MongoDBURL                 string        `envconfig:"MONGODB_IMPORTS_ADDR"        json:"-"`
	MongoDBCollection          string        `envconfig:"MONGODB_IMPORTS_COLLECTION"`
	MongoDBDatabase            string        `envconfig:"MONGODB_IMPORTS_DATABASE"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	RecipeAPIURL               string        `envconfig:"RECIPE_API_URL"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	AuditEventsTopic           string        `envconfig:"AUDIT_EVENTS_TOPIC"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

var cfg *Configuration

// Get the application and returns the configuration structure
func Get() (*Configuration, error) {
	if cfg != nil {
		return cfg, nil
	}

	brokers := []string{"localhost:9092"}

	cfg = &Configuration{
		BindAddr:                   ":21800",
		Host:                       "http://localhost:21800",
		Brokers:                    brokers,
		DatabakerImportTopic:       "data-bake-job-available",
		InputFileAvailableTopic:    "input-file-available",
		KafkaMaxBytes:              2000000,
		MongoDBURL:                 "localhost:27017",
		MongoDBDatabase:            "imports",
		MongoDBCollection:          "imports",
		ServiceAuthToken:           "0C30662F-6CF6-43B0-A96A-954772267FF5",
		DatasetAPIURL:              "http://localhost:22000",
		RecipeAPIURL:               "http://localhost:22300",
		GracefulShutdownTimeout:    time.Second * 5,
		ZebedeeURL:                 "http://localhost:8082",
		AuditEventsTopic:           "audit-events",
		HealthCheckInterval:        10 * time.Second,
		HealthCheckCriticalTimeout: 1 * time.Minute,
	}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Configuration) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
