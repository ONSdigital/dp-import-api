package config

import (
	"errors"
	"time"

	"encoding/json"

	"github.com/kelseyhightower/envconfig"
)

const KafkaSecProtocolTLS = "TLS"

// Configuration structure which hold information for configuring the import API
type Configuration struct {
	BindAddr                              string        `envconfig:"BIND_ADDR"`
	Host                                  string        `envconfig:"HOST"`
	DatabakerImportTopic                  string        `envconfig:"DATABAKER_IMPORT_TOPIC"`
	InputFileAvailableTopic               string        `envconfig:"INPUT_FILE_AVAILABLE_TOPIC"`
	CantabularDatasetInstanceStartedTopic string        `envconfig:"CANTABULAR_DATASET_INSTANCE_STARTED_TOPIC"`
	KafkaAddr                             []string      `envconfig:"KAFKA_ADDR"`
	KafkaVersion                          string        `envconfig:"KAFKA_VERSION"`
	KafkaLegacyAddr                       []string      `envconfig:"KAFKA_LEGACY_ADDR"`
	KafkaLegacyVersion                    string        `envconfig:"KAFKA_LEGACY_VERSION"`
	KafkaMaxBytes                         int           `envconfig:"KAFKA_MAX_BYTES"`
	KafkaSecProtocol                      string        `envconfig:"KAFKA_SEC_PROTO"`
	KafkaSecCACerts                       string        `envconfig:"KAFKA_SEC_CA_CERTS"`
	KafkaSecClientCert                    string        `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	KafkaSecClientKey                     string        `envconfig:"KAFKA_SEC_CLIENT_KEY"        json:"-"`
	KafkaSecSkipVerify                    bool          `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	ServiceAuthToken                      string        `envconfig:"SERVICE_AUTH_TOKEN"          json:"-"`
	MongoDBURL                            string        `envconfig:"MONGODB_IMPORTS_ADDR"        json:"-"`
	MongoDBCollection                     string        `envconfig:"MONGODB_IMPORTS_COLLECTION"`
	MongoDBDatabase                       string        `envconfig:"MONGODB_IMPORTS_DATABASE"`
	DatasetAPIURL                         string        `envconfig:"DATASET_API_URL"`
	RecipeAPIURL                          string        `envconfig:"RECIPE_API_URL"`
	GracefulShutdownTimeout               time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	ZebedeeURL                            string        `envconfig:"ZEBEDEE_URL"`
	HealthCheckInterval                   time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout            time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	DefaultLimit                          int           `envconfig:"DEFAULT_LIMIT"`
	DefaultMaxLimit                       int           `envconfig:"DEFAULT_MAXIMUM_LIMIT"`
	DefaultOffset                         int           `envconfig:"DEFAULT_OFFSET"`
}

var cfg *Configuration

// Get the application and returns the configuration structure
func Get() (*Configuration, error) {
	if cfg != nil {
		return cfg, nil
	}

	brokers := []string{"localhost:9092"}

	cfg = &Configuration{
		BindAddr:                              ":21800",
		Host:                                  "http://localhost:21800",
		DatabakerImportTopic:                  "data-bake-job-available",
		InputFileAvailableTopic:               "input-file-available",
		CantabularDatasetInstanceStartedTopic: "cantabular-dataset-instance-started",
		KafkaAddr:                             brokers,
		KafkaVersion:                          "1.0.2",
		KafkaSecProtocol:                      "",
		KafkaLegacyAddr:                       brokers,
		KafkaLegacyVersion:                    "1.0.2",
		KafkaMaxBytes:                         2000000,
		MongoDBURL:                            "localhost:27017",
		MongoDBDatabase:                       "imports",
		MongoDBCollection:                     "imports",
		ServiceAuthToken:                      "0C30662F-6CF6-43B0-A96A-954772267FF5",
		DatasetAPIURL:                         "http://localhost:22000",
		RecipeAPIURL:                          "http://localhost:22300",
		GracefulShutdownTimeout:               time.Second * 5,
		ZebedeeURL:                            "http://localhost:8082",
		HealthCheckInterval:                   30 * time.Second,
		HealthCheckCriticalTimeout:            90 * time.Second,
		DefaultLimit:                          20,
		DefaultMaxLimit:                       1000,
		DefaultOffset:                         0,
	}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	if cfg.KafkaSecProtocol != "" && cfg.KafkaSecProtocol != KafkaSecProtocolTLS {
		return nil, errors.New("KAFKA_SEC_PROTO has invalid value")
	}

	return cfg, nil
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Configuration) String() string {
	jsonConfig, _ := json.Marshal(config)
	return string(jsonConfig)
}
