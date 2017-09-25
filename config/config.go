package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Configuration structure which hold information for configuring the import API
type Configuration struct {
	BindAddr                string        `envconfig:"BIND_ADDR"`
	Host                    string        `envconfig:"HOST"`
	Brokers                 []string      `envconfig:"KAFKA_ADDR"`
	DatabakerImportTopic    string        `envconfig:"DATABAKER_IMPORT_TOPIC"`
	InputFileAvailableTopic string        `envconfig:"INPUT_FILE_AVAILABLE_TOPIC"`
	KafkaMaxBytes           int           `envconfig:"KAFKA_MAX_BYTES"`
	SecretKey               string        `envconfig:"SECRET_KEY"`
	MongoDBURL              string        `envconfig:"MONGODB_IMPORTS_ADDR"`
	MongoDBCollection       string        `envconfig:"MONGODB_IMPORTS_DATABASE"`
	MongoDBDatabase         string        `envconfig:"MONGODB_IMPORTS_COLLECTION"`
	DatasetAPIURL           string        `envconfig:"DATASET_API_URL"`
	DatasetAPIAuthToken     string        `envconfig:"DATASET_API_AUTH_TOKEN"`
	GracefulShutdownTimeout time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
}

var cfg *Configuration

// Get the application and returns the configuration structure
func Get() (*Configuration, error) {
	if cfg != nil {
		return cfg, nil
	}

	brokers := []string{"localhost:9092"}

	cfg = &Configuration{
		BindAddr:                ":21800",
		Host:                    "http://localhost:21800",
		Brokers:                 brokers,
		DatabakerImportTopic:    "data-bake-job-available",
		InputFileAvailableTopic: "input-file-available",
		KafkaMaxBytes:           2000000,
		MongoDBURL:              "localhost:27017",
		MongoDBDatabase:         "imports",
		MongoDBCollection:       "imports",
		DatasetAPIURL:           "http://localhost:22000",
		GracefulShutdownTimeout: time.Second * 5,
	}

	return cfg, envconfig.Process("", cfg)
}
