package config

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetRetrunsDefaultValues(t *testing.T) {
	t.Parallel()
	Convey("When a loading a configuration, default values are return", t, func() {
		configuration, error := Get()
		So(error, ShouldBeNil)
		So(configuration.BindAddr, ShouldEqual, ":21800")
		So(configuration.PublishDatasetTopic, ShouldEqual, "publish-dataset")
		So(configuration.KafkaMaxBytes, ShouldEqual, 2000000)
	})
}
