package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetReturnsDefaultValues(t *testing.T) {
	Convey("Given a clean environment", t, func() {
		os.Clearenv()
		cfg = nil

		Convey("When default configuration is obtained", func() {
			configuration, err := Get()

			Convey("Then expected configuration is returned", func() {
				So(err, ShouldBeNil)
				So(configuration, ShouldResemble, &Configuration{
					BindAddr:                              ":21800",
					Host:                                  "http://localhost:21800",
					DatabakerImportTopic:                  "data-bake-job-available",
					InputFileAvailableTopic:               "input-file-available",
					CantabularDatasetInstanceStartedTopic: "cantabular-dataset-instance-started",
					KafkaAddr:                             []string{"localhost:9092"},
					KafkaVersion:                          "1.0.2",
					KafkaLegacyAddr:                       []string{"localhost:9092"},
					KafkaLegacyVersion:                    "1.0.2",
					KafkaMaxBytes:                         2000000,
					KafkaSecProtocol:                      "",
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
				})
			})
		})

		Convey("When configuration is called with an invalid security setting", func() {
			os.Setenv("KAFKA_SEC_PROTO", "ssl")
			configuration, err := Get()

			Convey("Then an error is returned", func() {
				So(configuration, ShouldBeNil)
				So(err.Error(), ShouldEqual, "KAFKA_SEC_PROTO has invalid value")
			})
		})

	})
}
