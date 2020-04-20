package config

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetReturnsDefaultValues(t *testing.T) {
	Convey("When a loading a configuration, default values are return", t, func() {
		configuration, error := Get()
		So(error, ShouldBeNil)
		So(configuration, ShouldResemble, &Configuration{
			BindAddr:                   ":21800",
			Host:                       "http://localhost:21800",
			Brokers:                    []string{"localhost:9092"},
			DatabakerImportTopic:       "data-bake-job-available",
			InputFileAvailableTopic:    "input-file-available",
			KafkaMaxBytes:              2000000,
			MongoDBURL:                 "localhost:27017",
			MongoDBDatabase:            "imports",
			MongoDBCollection:          "imports",
			ServiceAuthToken:           "0C30662F-6CF6-43B0-A96A-954772267FF5",
			DatasetAPIURL:              "http://localhost:22000",
			RecipeAPIURL:               "http://localhost:22300",
			GracefulShutdownTimeout:    time.Second * 5,
			ZebedeeURL:                 "http://localhost:8082",
			AuditEventsTopic:           "audit-events",
			HealthCheckInterval:        30 * time.Second,
			HealthCheckCriticalTimeout: 90 * time.Second,
		})
	})
}
