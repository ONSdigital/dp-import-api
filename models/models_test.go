package models

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
	errs "github.com/ONSdigital/dp-import-api/apierrors"
	"github.com/ONSdigital/dp-import-api/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

// Regression test for new official golang driver which conforms to golang json spec on marshalling zero value structs.
func TestBSONMarshalJob_ZeroValues(t *testing.T) {
	goldenJob := Job{State: SubmittedState}
	// This is the golden BSON value for the above job, where all zero value attributes that have the 'omitempty' BSON tag are indeed
	// omitted. If other, zero value attributes that are tagged as 'omitempty' are marshalled, this is an error
	goldenBSON := []byte{26, 0, 0, 0, 2, 115, 116, 97, 116, 101, 0, 10, 0, 0, 0, 115, 117, 98, 109, 105, 116, 116, 101, 100, 0, 0}

	bsn, e := bson.Marshal(goldenJob)
	if e != nil {
		t.Fatalf("failed to marshal goldenJob to bson: %v", e)
	}
	if !bytes.Equal(bsn, goldenBSON) {
		t.Errorf("Job incorrectly marshalled to BSON")
	}
}

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

		hierarchy := true
		notHierarchy := false

		job := &Job{Links: &LinksMap{}}
		datasetID := "123"
		datasetURL := "/wut"
		codelists := []recipe.CodeList{
			{
				Name:        "codelist1",
				ID:          "1",
				IsHierarchy: &notHierarchy,
			},
			{
				Name:        "codelist2",
				ID:          "2",
				IsHierarchy: &hierarchy,
			},
		}

		Convey("When CreateInstance is called", func() {

			instance := CreateInstance(job, datasetID, datasetURL, codelists)

			Convey("Then there should be a single build hierarchy task for the codelist that is marked as a hierarchy", func() {

				So(len(instance.ImportTasks.BuildHierarchyTasks), ShouldEqual, 1)

				So(instance.ImportTasks.BuildHierarchyTasks[0].State, ShouldEqual, CreatedState)
				So(instance.ImportTasks.BuildHierarchyTasks[0].DimensionName, ShouldEqual, "codelist2")
				So(instance.ImportTasks.BuildHierarchyTasks[0].CodeListID, ShouldEqual, "2")
			})
		})
	})
}

func TestValidateState(t *testing.T) {
	t.Parallel()
	Convey("Given job has a valid state", t, func() {
		listOfValidStates := []string{
			CompletedState,
			CreatedState,
			SubmittedState,
		}

		for _, state := range listOfValidStates {
			Convey("When validating job state of "+state, func() {
				job := &Job{
					State: state,
				}
				Convey("Then error should be nil", func() {
					err := job.ValidateState()
					So(err, ShouldBeNil)
				})
			})
		}
	})

	Convey("Given job contains no state field", t, func() {
		Convey("When validating job state ", func() {
			job := &Job{}
			Convey("Then error should be nil", func() {
				err := job.ValidateState()
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given job contains an invalid state", t, func() {
		Convey("When validating job state ", func() {
			job := &Job{
				State: "start",
			}
			Convey("Then error should be returned", func() {
				err := job.ValidateState()
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrInvalidState)
			})
		})
	})
}
