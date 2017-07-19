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
		job, jobError := CreateJob(strings.NewReader("{ }"))
		So(jobError, ShouldBeNil)
		So(job.Validate(), ShouldNotBeNil)
	})
}

func TestCreateJobWithDataset(t *testing.T) {
	Convey("When a job has a valid json body, a message is returned", t, func() {
		reader := strings.NewReader("{ \"recipe\": \"test123\", \"number_of_instances\":1}")
		job, jobError := CreateJob(reader)
		So(jobError, ShouldBeNil)
		So(job.Validate(), ShouldBeNil)
		So(job.Recipe, ShouldEqual, "test123")
		So(job.NumberOfInstances, ShouldEqual, 1)
	})
}

func TestCreateJobWithInvalidJson(t *testing.T) {
	Convey("When a job message has an invalid json, an error is returned", t, func() {
		reader := strings.NewReader("{ ")
		_, jobError := CreateJob(reader)
		So(jobError, ShouldNotBeNil)
	})
}

func TestCreateS3FilehNoBody(t *testing.T) {
	Convey("When a uploaded file message has no body, an error is returned", t, func() {
		_, uploadedFileError := CreateUploadedFile(mocks.Reader{})
		So(uploadedFileError, ShouldNotBeNil)
	})
}

func TestCreateS3FileWithEmptyJson(t *testing.T) {
	Convey("When a uploaded file message has an empty json, an error is returned", t, func() {
		_, uploadedFileError := CreateUploadedFile(strings.NewReader("{ }"))
		So(uploadedFileError, ShouldNotBeNil)
	})
}

func TestCreateS3FileWithInvalidJson(t *testing.T) {
	Convey("When an uploaded file message has an empty json, an error is returned", t, func() {
		_, uploadedFileError := CreateUploadedFile(strings.NewReader("{}}}"))
		So(uploadedFileError, ShouldNotBeNil)
	})
}

func TestCreateUploadedFileWithValidJson(t *testing.T) {
	Convey("When an uploaded file message has valid json, an uploaded file struct is returned", t, func() {
		file, uploadedFileError := CreateUploadedFile(strings.NewReader("{ \"alias_name\":\"n1\",\"url\":\"https://aws.s3/ons/myfile.exel\"}"))
		So(uploadedFileError, ShouldBeNil)
		So(file.AliasName, ShouldEqual, "n1")
		So(file.URL, ShouldEqual, "https://aws.s3/ons/myfile.exel")
	})
}

func TestCreateEventWithValidJson(t *testing.T) {
	Convey("When an event message has valid json, an event struct is returned", t, func() {
		reader := strings.NewReader("{ \"type\":\"info\",\"message\":\"123 123\",\"time\":\"7789789\",\"messageOffset\":\"321\"}")
		event, eventError := CreateEvent(reader)
		So(eventError, ShouldBeNil)
		So(event.Validate(), ShouldBeNil)
		So("info", ShouldEqual, event.Type)
	})
}
