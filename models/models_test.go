package models

import (
	"github.com/ONSdigital/dp-import-api/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func TestCreateImportJobWithNoBody(t *testing.T) {
	Convey("When a import message has no body, an error is returned", t, func() {
		_, errorMessage := CreateImportJob(mocks.Reader{})
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateImportJobWithEmptyJson(t *testing.T) {
	Convey("When a import message has an empty json body, an error is returned", t, func() {
		_, errorMessage := CreateImportJob(strings.NewReader("{ }"))
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateImportJobWithDataset(t *testing.T) {
	Convey("When a import message has a valid json body, a message is returned", t, func() {
		reader := strings.NewReader("{ \"dataset\": \"test123\"}")
		message, errorMessage := CreateImportJob(reader)
		So(errorMessage, ShouldBeNil)
		So("test123", ShouldEqual, message.Dataset)
	})
}

func TestCreateImportJobWithInvalidJson(t *testing.T) {
	Convey("When a import message has an invalid json, an error is returned", t, func() {
		reader := strings.NewReader("{ ")
		_, errorMessage := CreateImportJob(reader)
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateS3FilehNoBody(t *testing.T) {
	Convey("When a S3 file message has no body, an error is returned", t, func() {
		_, errorMessage := CreateS3File(mocks.Reader{})
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateS3FileWithEmptyJson(t *testing.T) {
	Convey("When a S3 file message has an empty json, an error is returned", t, func() {
		_, errorMessage := CreateS3File(strings.NewReader("{ }"))
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateS3FileWithInvalidJson(t *testing.T) {
	Convey("When a S3 file message has an empty json, an error is returned", t, func() {
		_, errorMessage := CreateS3File(strings.NewReader("{}}}"))
		So(errorMessage, ShouldNotBeNil)
	})
}

func TestCreateS3FileWithValidJson(t *testing.T) {
	Convey("When a S3 file message has valid json, a s3 file struct is returned", t, func() {
		message, errorMessage := CreateS3File(strings.NewReader("{ \"aliasName\":\"n1\",\"s3Url\":\"https://aws.s3/ons/myfile.exel\"}"))
		So(errorMessage, ShouldBeNil)
		So(message.AliasName, ShouldEqual, "n1")
		So(message.S3Url, ShouldEqual, "https://aws.s3/ons/myfile.exel")
	})
}

func TestCreateEventWithValidJson(t *testing.T) {
	Convey("When an event message has valid json, a event struct is returned", t, func() {
		reader := strings.NewReader("{ \"type\":\"info\",\"message\":\"123 123\",\"time\":\"7789789\",\"messageOffset\":\"321\"}")
		message, errorMessage := CreateEvent(reader)
		So(errorMessage, ShouldBeNil)
		So("info", ShouldEqual, message.Type)
	})
}

func TestCreateDimensionWithValidJson(t *testing.T) {
	Convey("hen a dimension message has valid json, a dimension struct is returned", t, func() {
		reader := strings.NewReader("{ \"nodeName\":\"321\",\"value\":\"123 123\"}")
		message, errorMessage := CreateDimension(reader)
		So(errorMessage, ShouldBeNil)
		So("321", ShouldEqual, message.NodeName)
	})
}
