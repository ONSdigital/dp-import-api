package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Configuration structure which hold information for configuring the import API
type Configuration struct {
	BindAddr                string   `envconfig:"BIND_ADDR"`
	Host                    string   `envconfig:"HOST"`
	Brokers                 []string `envconfig:"KAFKA_ADDR"`
	DatabakerImportTopic    string   `envconfig:"DATABAKER_IMPORT_TOPIC"`
	InputFileAvailableTopic string   `envconfig:"INPUT_FILE_AVAILABLE_TOPIC"`
	KafkaMaxBytes           int      `envconfig:"KAFKA_MAX_BYTES"`
	SecretKey               string   `envconfig:"SECRET_KEY"`
	MongoDBURL              string   `envconfig:"MONGODB_IMPORTS_ADDR"`
	MongoDBCollection       string   `envconfig:"MONGODB_IMPORTS_DATABASE"`
	MongoDBDatabase         string   `envconfig:"MONGODB_IMPORTS_COLLECTION"`
	DatasetAPIURL           string   `envconfig:"DATASET_API_URL"`
	DatasetAPIAuthToken     string   `envconfig:"DATASET_AUTH_TOKEN"`
	RecipeAPIURL            string   `envconfig:"RECIPE_API_URL"`
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
		SecretKey:               "FD0108EA-825D-411C-9B1D-41EF7727F465",
		DatasetAPIURL:           "http://localhost:22000",
		DatasetAPIAuthToken:     "FD0108EA-825D-411C-9B1D-41EF7727F465",
		RecipeAPIURL:            "http://localhost:22300",
	}

	return cfg, envconfig.Process("", cfg)
}
