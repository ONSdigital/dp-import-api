package models

import (
	"github.com/ONSdigital/dp-import-api/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func TestCreateJobWithNoBody(t *testing.T) {
	Convey("When a job message has no body, an error is returned", t, func() {
		_, errorMessage := CreateJob(mocks.Reader{})
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateJobWithEmptyJson(t *testing.T) {
	Convey("When a job message has an empty json body, an error is returned", t, func() {
		_, errorMessage := CreateJob(strings.NewReader("{ }"))
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateJobWithDataset(t *testing.T) {
	Convey("When a job has a valid json body, a message is returned", t, func() {
		reader := strings.NewReader("{ \"recipe\": \"test123\", \"datasets\":[\"RPI\"]}")
		message, errorMessage := CreateJob(reader)
		So(errorMessage, ShouldBeNil)
		So(message.Recipe, ShouldEqual, "test123")
		So(message.Datasets, ShouldContain, "RPI")
	})
}

func TestCreateJobWithInvalidJson(t *testing.T) {
	Convey("When a job message has an invalid json, an error is returned", t, func() {
		reader := strings.NewReader("{ ")
		_, errorMessage := CreateJob(reader)
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateS3FilehNoBody(t *testing.T) {
	Convey("When a uploaded file message has no body, an error is returned", t, func() {
		_, errorMessage := CreateUploadedFile(mocks.Reader{})
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateS3FileWithEmptyJson(t *testing.T) {
	Convey("When a uploaded file message has an empty json, an error is returned", t, func() {
		_, errorMessage := CreateUploadedFile(strings.NewReader("{ }"))
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateS3FileWithInvalidJson(t *testing.T) {
	Convey("When an uploaded file message has an empty json, an error is returned", t, func() {
		_, errorMessage := CreateUploadedFile(strings.NewReader("{}}}"))
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateUploadedFileWithValidJson(t *testing.T) {
	Convey("When an uploaded file message has valid json, an uploaded file struct is returned", t, func() {
		message, errorMessage := CreateUploadedFile(strings.NewReader("{ \"aliasName\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}"))
		So(errorMessage, ShouldBeNil)
		So(message.AliasName, ShouldEqual, "n1")
		So(message.URL, ShouldEqual, "https://aws.s3/ons/myfile.exel")
	})
}

func TestCreateEventWithValidJson(t *testing.T) {
	Convey("When an event message has valid json, an event struct is returned", t, func() {
		reader := strings.NewReader("{ \"type\":\"info\",\"message\":\"123 123\",\"time\":\"7789789\",\"messageOffset\":\"321\"}")
		message, errorMessage := CreateEvent(reader)
		So(errorMessage, ShouldBeNil)
		So("info", ShouldEqual, message.Type)
	})
}

func TestCreateDimensionWithValidJson(t *testing.T) {
	Convey("When a dimension message has valid json, a dimension struct is returned", t, func() {
		reader := strings.NewReader("{ \"nodeName\":\"321\",\"value\":\"123 123\"}")
		message, errorMessage := CreateDimension(reader)
		So(errorMessage, ShouldBeNil)
		So("321", ShouldEqual, message.NodeName)
	})
}

func TestCreateJobState(t *testing.T) {
	Convey("When a JobState has valid json, a jobstate struct is returned", t, func() {
		reader := strings.NewReader("{ \"state\":\"start\"}")
		jobState, errorMessage := CreateJobState(reader)
		So(errorMessage, ShouldBeNil)
		So("start", ShouldEqual, jobState.State)
	})
}
