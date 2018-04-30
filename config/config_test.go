package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetReturnsDefaultValues(t *testing.T) {
	Convey("When a loading a configuration, default values are return", t, func() {
		configuration, error := Get()
		So(error, ShouldBeNil)
		So(configuration.BindAddr, ShouldEqual, ":21800")
		So(configuration.DatabakerImportTopic, ShouldEqual, "data-bake-job-available")
		So(configuration.SecretKey, ShouldEqual, "0C30662F-6CF6-43B0-A96A-954772267FF5")
		So(configuration.ServiceAuthToken, ShouldEqual, "0C30662F-6CF6-43B0-A96A-954772267FF5")
		So(configuration.ZebedeeURL, ShouldEqual, "http://localhost:8082")
		So(configuration.KafkaMaxBytes, ShouldEqual, 2000000)
	})
}
