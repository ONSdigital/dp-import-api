package models

import (
	"strings"
	"testing"

	"github.com/ONSdigital/dp-import-api/mocks"
	. "github.com/smartystreets/goconvey/convey"
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
		reader := strings.NewReader(`{ "recipe": "1234-sdfsdf"}`)
		job, jobError := CreateJob(reader)
		So(jobError, ShouldBeNil)
		So(job.Validate(), ShouldBeNil)
		So(job.RecipeID, ShouldEqual, "1234-sdfsdf")
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
		file, uploadedFileError := CreateUploadedFile(strings.NewReader(`{ "alias_name":"n1","url":"https://aws.s3/ons/myfile.exel"}`))
		So(uploadedFileError, ShouldBeNil)
		So(file.AliasName, ShouldEqual, "n1")
		So(file.URL, ShouldEqual, "https://aws.s3/ons/myfile.exel")
	})
}

func TestCreateInstance(t *testing.T) {

	Convey("Given a dummy job and slice of codelists", t, func() {

		job := &Job{}
		datasetID := "123"
		datasetURL := "/wut"
		codelists := []CodeList{
			{
				Name:        "codelist1",
				ID:          "1",
				IsHierarchy: false,
			},
			{
				Name:        "codelist2",
				ID:          "2",
				IsHierarchy: true,
			},
		}

		Convey("When CreateInstance is called", func() {

			instance := CreateInstance(job, datasetID, datasetURL, codelists)

			Convey("Then there should be a single build hierarchy task for the codelist that is marked as a hierarchy", func() {

				So(instance.ImportTasks.BuildHierarchyTasks[0].State, ShouldEqual, CreatedState)
				So(instance.ImportTasks.BuildHierarchyTasks[0].DimensionName, ShouldEqual, "codelist2")
				So(instance.ImportTasks.BuildHierarchyTasks[0].CodeListID, ShouldEqual, "2")
			})
		})
	})
}
