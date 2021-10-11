package config

import (
	"fmt"
	"strings"
	"time"

	"encoding/json"

	"github.com/kelseyhightower/envconfig"
)

const KafkaSecProtocolTLS = "TLS"

// KafkaConfig contains all configuration relating to kafka
type KafkaConfig struct {
	Addr                                  []string `envconfig:"KAFKA_ADDR"`
	Version                               string   `envconfig:"KAFKA_VERSION"`
	LegacyAddr                            []string `envconfig:"KAFKA_LEGACY_ADDR"`
	LegacyVersion                         string   `envconfig:"KAFKA_LEGACY_VERSION"`
	MaxBytes                              int      `envconfig:"KAFKA_MAX_BYTES"`
	SecProtocol                           string   `envconfig:"KAFKA_SEC_PROTO"`
	SecCACerts                            string   `envconfig:"KAFKA_SEC_CA_CERTS"`
	SecClientCert                         string   `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	SecClientKey                          string   `envconfig:"KAFKA_SEC_CLIENT_KEY"                         json:"-"`
	SecSkipVerify                         bool     `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	DatabakerImportTopic                  string   `envconfig:"DATABAKER_IMPORT_TOPIC"`
	InputFileAvailableTopic               string   `envconfig:"INPUT_FILE_AVAILABLE_TOPIC"`
	CantabularDatasetInstanceStartedTopic string   `envconfig:"CANTABULAR_DATASET_INSTANCE_STARTED_TOPIC"`
}

// MongoConfig contains the config required to connect to MongoDB.
type MongoConfig struct {
	URI                string        `envconfig:"MONGODB_BIND_ADDR"   json:"-"`
	Collection         string        `envconfig:"MONGODB_COLLECTION"`
	Database           string        `envconfig:"MONGODB_DATABASE"`
	Username           string        `envconfig:"MONGODB_USERNAME"    json:"-"`
	Password           string        `envconfig:"MONGODB_PASSWORD"    json:"-"`
	IsSSL              bool          `envconfig:"MONGODB_IS_SSL"`
	EnableReadConcern  bool          `envconfig:"MONGODB_ENABLE_READ_CONCERN"`
	EnableWriteConcern bool          `envconfig:"MONGODB_ENABLE_WRITE_CONCERN"`
	QueryTimeout       time.Duration `envconfig:"MONGODB_QUERY_TIMEOUT"`
	ConnectionTimeout  time.Duration `envconfig:"MONGODB_CONNECT_TIMEOUT"`
}

// Configuration structure which hold information for configuring the import API
type Configuration struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	Host                       string        `envconfig:"HOST"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"            json:"-"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	RecipeAPIURL               string        `envconfig:"RECIPE_API_URL"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	DefaultLimit               int           `envconfig:"DEFAULT_LIMIT"`
	DefaultMaxLimit            int           `envconfig:"DEFAULT_MAXIMUM_LIMIT"`
	DefaultOffset              int           `envconfig:"DEFAULT_OFFSET"`
	KafkaConfig
	MongoConfig
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
		ServiceAuthToken:           "0C30662F-6CF6-43B0-A96A-954772267FF5",
		DatasetAPIURL:              "http://localhost:22000",
		RecipeAPIURL:               "http://localhost:22300",
		GracefulShutdownTimeout:    time.Second * 5,
		ZebedeeURL:                 "http://localhost:8082",
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		DefaultLimit:               20,
		DefaultMaxLimit:            1000,
		DefaultOffset:              0,
		KafkaConfig: KafkaConfig{
			DatabakerImportTopic:                  "data-bake-job-available",
			InputFileAvailableTopic:               "input-file-available",
			CantabularDatasetInstanceStartedTopic: "cantabular-dataset-instance-started",
			Addr:                                  brokers,
			Version:                               "1.0.2",
			SecProtocol:                           "",
			LegacyAddr:                            brokers,
			LegacyVersion:                         "1.0.2",
			MaxBytes:                              2000000,
		},
		MongoConfig: MongoConfig{
			URI:                "localhost:27017",
			Database:           "imports",
			Collection:         "imports",
			Username:           "",
			Password:           "",
			IsSSL:              false,
			QueryTimeout:       15 * time.Second,
			ConnectionTimeout:  5 * time.Second,
			EnableReadConcern:  false,
			EnableWriteConcern: true,
		},
	}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	if validationIssues := cfg.validate(); len(validationIssues) > 0 {
		return nil, fmt.Errorf("validation of config failed: %s", strings.Join(validationIssues, "; "))
	}

	return cfg, nil
}

func (config Configuration) validate() (issues []string) {

	if config.KafkaConfig.SecProtocol != "" && config.KafkaConfig.SecProtocol != KafkaSecProtocolTLS {
		issues = append(issues, "KAFKA_SEC_PROTO has invalid value")
	}

	keyEmpty := len(config.KafkaConfig.SecClientKey) == 0
	certEmpty := len(config.KafkaConfig.SecClientCert) == 0
	if keyEmpty && !certEmpty {
		issues = append(issues, "got a KAFKA_SEC_CLIENT_CERT value, so require KAFKA_SEC_CLIENT_KEY to have a value")
	} else if certEmpty && !keyEmpty {
		issues = append(issues, "got a KAFKA_SEC_CLIENT_KEY value, so require KAFKA_SEC_CLIENT_CERT to have a value")
	}

	return
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Configuration) String() string {
	jsonConfig, _ := json.Marshal(config)
	return string(jsonConfig)
}
