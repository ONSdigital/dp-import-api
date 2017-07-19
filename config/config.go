package config

import "github.com/ian-kent/gofigure"

type appConfiguration struct {
	BindAddr             string   `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"The port to bind to"`
	Brokers              []string `env:"KAFKA_ADDR" flag:"kafka-addr" flagDesc:"The kafka broker addresses"`
	DatabakerImportTopic string   `env:"DATABAKER_IMPORT_TOPIC" flag:"databaker-import-topic" flagDesc:"The Kafka topic to import job via databaker"`
	DirectImportTopic    string   `env:"DIRECT_IMPORT_TOPIC" flag:"direct-import-topic" flagDesc:"The Kafka topic to import job directly"`
	KafkaMaxBytes        int      `env:"KAFKA_MAX_BYTES" flag:"kafka-max-bytes" flagDesc:"The maximum permitted size of a message. Should be set equal to or smaller than the broker's 'message.max.bytes'"`
	PostgresURL          string   `env:"POSTGRES_URL" flag:"postgres-url" flagDesc:"The URL address to connect to a postgres instance'"`
}

var configuration *appConfiguration

// Get - configures the application and returns the configuration
func Get() (*appConfiguration, error) {
	if configuration != nil {
		return configuration, nil
	}

	var brokers []string

	brokers = append(brokers, "localhost:9092")

	configuration = &appConfiguration{
		BindAddr:             ":21800",
		Brokers:              brokers,
		DatabakerImportTopic: "data-bake",
		DirectImportTopic:    "dimensions-extracted",
		KafkaMaxBytes:        2000000,
		PostgresURL:          "user=dp dbname=ImportJobs sslmode=disable",
	}

	if err := gofigure.Gofigure(configuration); err != nil {
		return configuration, err
	}

	return configuration, nil
}