package url_test

import (
	"testing"

	"github.com/ONSdigital/dp-import-api/url"
	. "github.com/smartystreets/goconvey/convey"
)

var importAPIHost string = "http://import-api"
var datasetAPIHost string = "http://dataset-api"
var builder = url.NewBuilder(importAPIHost, datasetAPIHost)

func TestBuilder_GetInstanceURL(t *testing.T) {

	Convey("Given an instance ID", t, func() {

		instanceID := "1234"

		Convey("When the instance URL is requested", func() {

			builtURL := builder.GetInstanceURL(instanceID)

			Convey("Then the expected URL is returned", func() {
				expectedURL := datasetAPIHost + "/instances/" + instanceID
				So(builtURL, ShouldEqual, expectedURL)
			})
		})
	})
}

func TestBuilder_GetJobURL(t *testing.T) {

	Convey("Given a job ID", t, func() {

		jobID := "3456"

		Convey("When the instance URL is requested", func() {

			builtURL := builder.GetJobURL(jobID)

			Convey("Then the expected URL is returned", func() {
				expectedURL := importAPIHost + "/jobs/" + jobID
				So(builtURL, ShouldEqual, expectedURL)
			})
		})
	})
}
