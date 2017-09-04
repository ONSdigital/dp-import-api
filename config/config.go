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
	PostgresURL             string   `envconfig:"POSTGRES_URL"`
	SecretKey               string   `envconfig:"SECRET_KEY"`
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
		PostgresURL:             "user=dp dbname=ImportJobs sslmode=disable",
		SecretKey:               "FD0108EA-825D-411C-9B1D-41EF7727F465",
	}

	return cfg, envconfig.Process("", cfg)
}
