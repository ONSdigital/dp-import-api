package config

import "github.com/ian-kent/gofigure"

// Configuration structure which hold information for configuring the import API
type Configuration struct {
	BindAddr                string   `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"The port to bind to"`
	Host                    string   `env:"HOST" flag:"host" flagDesc:"The host name used to build URLs"`
	Brokers                 []string `env:"KAFKA_ADDR" flag:"kafka-addr" flagDesc:"The kafka broker addresses"`
	DatabakerImportTopic    string   `env:"DATABAKER_IMPORT_TOPIC" flag:"databaker-import-topic" flagDesc:"The Kafka topic to import job via databaker"`
	InputFileAvailableTopic string   `env:"INPUT_FILE_AVAILABLE_TOPIC" flag:"input-file-available-topic" flagDesc:"The Kafka topic to import job directly"`
	KafkaMaxBytes           int      `env:"KAFKA_MAX_BYTES" flag:"kafka-max-bytes" flagDesc:"The maximum permitted size of a message. Should be set equal to or smaller than the broker's 'message.max.bytes'"`
	MongoDBURL              string   `env:"MONGODB_IMPORTS_ADDR" flag:"mongodb-bind-addr" flagDesc:"MongoDB bind address"`
	MongoDBCollection       string   `env:"MONGODB_IMPORTS_DATABASE" flag:"mongodb-database" flagDesc:"MongoDB import database"`
	MongoDBDatabase         string   `env:"MONGODB_IMPORTS_COLLECTION" flag:"mongodb-collection" flagDesc:"MongoDB import collection"`
	SecretKey               string   `env:"SECRET_KEY" flag:"secret-key" flagDesc:"A secret key used in client authentication"`
	DatasetAPIURL           string   `env:"DATASET_API_URL" flag:"dataset-api" flagDesc:"The URL of the Dataset API"`
	DatasetAPIAuthToken     string   `env:"DATASET_AUTH_TOKEN" flag:"dataset-auth-token" flagDesc:"Authentication token to access the Dataset API"`
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
	}

	return cfg, gofigure.Gofigure(cfg)

}
