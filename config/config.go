package config

import "github.com/ian-kent/gofigure"

type appConfiguration struct {
	BindAddr                string   `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"The port to bind to"`
	Host                    string   `env:"HOST" flag:"host" flagDesc:"The host name used to build URLs"`
	Brokers                 []string `env:"KAFKA_ADDR" flag:"kafka-addr" flagDesc:"The kafka broker addresses"`
	DatabakerImportTopic    string   `env:"DATABAKER_IMPORT_TOPIC" flag:"databaker-import-topic" flagDesc:"The Kafka topic to import job via databaker"`
	InputFileAvailableTopic string   `env:"INPUT_FILE_AVAILABLE_TOPIC" flag:"input-file-available-topic" flagDesc:"The Kafka topic to import job directly"`
	KafkaMaxBytes           int      `env:"KAFKA_MAX_BYTES" flag:"kafka-max-bytes" flagDesc:"The maximum permitted size of a message. Should be set equal to or smaller than the broker's 'message.max.bytes'"`
	PostgresURL             string   `env:"POSTGRES_URL" flag:"postgres-url" flagDesc:"The URL address to connect to a postgres instance'"`
}

var cfg *appConfiguration

// Get - configures the application and returns the configuration
func Get() (*appConfiguration, error) {
	if cfg != nil {
		return cfg, nil
	}

	var brokers []string

	brokers = append(brokers, "localhost:9092")

	cfg = &appConfiguration{
		BindAddr:                ":21800",
		Host:                    "http://localhost:21800",
		Brokers:                 brokers,
		DatabakerImportTopic:    "data-bake-job-available",
		InputFileAvailableTopic: "input-file-available",
		KafkaMaxBytes:           2000000,
		PostgresURL:             "user=dp dbname=ImportJobs sslmode=disable",
	}

	if err := gofigure.Gofigure(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
