package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var expectedConfig = &Configuration{
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
		Brokers:                               []string{"localhost:9092"},
		DatabakerImportTopic:                  "data-bake-job-available",
		InputFileAvailableTopic:               "input-file-available",
		CantabularDatasetInstanceStartedTopic: "cantabular-dataset-instance-started",
		Version:                               "1.0.2",
		MaxBytes:                              2000000,
		SecProtocol:                           "",
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

func TestGetReturnsDefaultValues(t *testing.T) {
	Convey("Given a clean environment", t, func() {
		os.Clearenv()
		cfg = nil

		Convey("When default configuration is obtained", func() {
			configuration, err := Get()

			Convey("Then expected configuration is returned", func() {
				So(err, ShouldBeNil)
				So(configuration, ShouldResemble, expectedConfig)
			})
		})

		Convey("When configuration is called with an invalid security protocol", func() {
			_ = os.Setenv("KAFKA_SEC_PROTO", "ssl")
			configuration, err := Get()

			Convey("Then an error is returned", func() {
				So(configuration, ShouldBeNil)
				So(err.Error(), ShouldEqual, "validation of config failed: KAFKA_SEC_PROTO has invalid value")
			})
		})

		Convey("When configuration is called with an invalid cert setting", func() {
			_ = os.Setenv("KAFKA_SEC_CLIENT_KEY", "open sesame")
			configuration, err := Get()

			Convey("Then an error is returned", func() {
				So(configuration, ShouldBeNil)
				So(err.Error(), ShouldEqual, "validation of config failed: got a KAFKA_SEC_CLIENT_KEY value, so require KAFKA_SEC_CLIENT_CERT to have a value")
			})
		})

		Convey("When configuration is called with a valid cert and key", func() {
			secExpectedConfig := *expectedConfig
			secExpectedConfig.KafkaConfig.SecClientKey = "open sesame"
			secExpectedConfig.KafkaConfig.SecClientCert = "please"

			_ = os.Setenv("KAFKA_SEC_CLIENT_KEY", "open sesame")
			_ = os.Setenv("KAFKA_SEC_CLIENT_CERT", "please")

			configuration, err := Get()

			Convey("Then expected configuration is returned", func() {
				So(err, ShouldBeNil)
				So(configuration, ShouldResemble, &secExpectedConfig)
			})
		})
	})
}
