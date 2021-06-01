package importqueue

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import/events"
	. "github.com/smartystreets/goconvey/convey"
)

func TestQueueV4File(t *testing.T) {
	ctx := context.Background()

	job := models.ImportData{
		InstanceIDs:   []string{"1"},
		Recipe:        "b944be78-f56d-409b-9ebd-ab2b77ffe187",
		Format:        "v4",
		UploadedFiles: &[]models.UploadedFile{{AliasName: "aliasV4", URL: "s3//aws/000/v4.csv"}}}

	Convey("Given a mocked importQueue without a v4 queue", t, func() {
		importer := CreateImportQueue(nil, nil, nil)

		Convey("Then importing a 'v4' recipe results in the expected error being returned", func() {
			err := importer.Queue(ctx, &job)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "v4 queue (kafka producer) is not available")
		})
	})

	Convey("Given a mocked importQueue with a valid v4 queue", t, func() {
		v4Queue := make(chan []byte, 1)
		importer := CreateImportQueue(nil, v4Queue, nil)

		Convey("Then importing an invalid 'v4' recipe fails with the expected error", func() {
			err := importer.Queue(ctx, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "job not available")
		})

		Convey("Then importing a 'v4' recipe sends the expected import event to the v4 queue", func() {
			err := importer.Queue(ctx, &job)
			So(err, ShouldBeNil)

			bytes := <-v4Queue

			var file events.InputFileAvailable
			events.InputFileAvailableSchema.Unmarshal(bytes, &file)

			So(file.URL, ShouldEqual, (*job.UploadedFiles)[0].URL)
			So(file.InstanceID, ShouldEqual, job.InstanceIDs[0])
		})
	})
}

func TestQueueCantabularFile(t *testing.T) {
	ctx := context.Background()

	job := models.ImportData{
		InstanceIDs:   []string{"1"},
		Recipe:        "b944be78-f56d-409b-9ebd-ab2b77ffe187",
		Format:        "cantabular",
		UploadedFiles: &[]models.UploadedFile{{AliasName: "aliasCantabular", URL: "s3//aws/000/cantabular.csv"}}}

	Convey("Given a mocked importQueue without a cantabular queue", t, func() {
		importer := CreateImportQueue(nil, nil, nil)

		Convey("Then importing a 'cantabular' recipe results in the expected error being returned", func() {
			err := importer.Queue(ctx, &job)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "cantabular queue (kafka producer) is not available")
		})
	})

	Convey("Given a mocked importQueue with a valid cantabular queue", t, func() {
		cantabularQueue := make(chan []byte, 1)
		importer := CreateImportQueue(nil, nil, cantabularQueue)

		Convey("Then importing an invalid 'cantabular' recipe fails with the expected error", func() {
			err := importer.Queue(ctx, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "job not available")
		})

		Convey("Then importing a 'cantabular' recipe sends the expected import event to the cantabular queue", func() {
			err := importer.Queue(ctx, &job)
			So(err, ShouldBeNil)

			bytes := <-cantabularQueue

			var file events.InputFileAvailable
			events.InputFileAvailableSchema.Unmarshal(bytes, &file)

			So(file.URL, ShouldEqual, (*job.UploadedFiles)[0].URL)
			So(file.InstanceID, ShouldEqual, job.InstanceIDs[0])
		})
	})
}

func TestQueueDefault(t *testing.T) {
	ctx := context.Background()

	job := models.ImportData{
		InstanceIDs:   []string{"1"},
		Recipe:        "b944be78-f56d-409b-9ebd-ab2b77ffe187",
		Format:        "other",
		UploadedFiles: &[]models.UploadedFile{{AliasName: "aliasOther", URL: "s3//aws/000/other.csv"}}}

	Convey("Given a mocked importQueue", t, func() {
		importer := CreateImportQueue(nil, nil, nil)

		Convey("Then importing an 'other' recipe does not return any error and does not trigger any action", func() {
			err := importer.Queue(ctx, &job)
			So(err, ShouldBeNil)
		})
	})
}

func TestValidateJob(t *testing.T) {

	Convey("given a mocked importQueue", t, func() {
		importer := CreateImportQueue(nil, nil, nil)

		Convey("Then calling validateJob with a nil job results in the expected error being returned", func() {
			err := importer.validateJob(nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "job not available")
		})

		Convey("Then calling validateJob with an empty job results in the expected error being returned", func() {
			err := importer.validateJob(&models.ImportData{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid job")
		})

		Convey("Then calling validateJob with a job containing an empty list of uploadedFiles returns the expected error being returned", func() {
			err := importer.validateJob(&models.ImportData{
				InstanceIDs: []string{"testInstanceID"},
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid job")
		})

		Convey("Then calling validateJob with a job containing an empty list of instance IDs returns the expected error being returned", func() {
			err := importer.validateJob(&models.ImportData{
				UploadedFiles: &[]models.UploadedFile{{AliasName: "testUpload"}},
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid job")
		})

		Convey("Then calling validateJob with exactly one instanceID and one uploadedFile succeeds", func() {
			err := importer.validateJob(&models.ImportData{
				InstanceIDs:   []string{"testInstanceID"},
				UploadedFiles: &[]models.UploadedFile{{AliasName: "testUpload"}},
			})
			So(err, ShouldBeNil)
		})

		Convey("Then calling validateJob with more than one instanceID fails with the expected error", func() {
			err := importer.validateJob(&models.ImportData{
				InstanceIDs:   []string{"testInstanceID1", "testInstanceID2"},
				UploadedFiles: &[]models.UploadedFile{{AliasName: "testUpload"}},
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds and uploaded files must be 1")
		})

		Convey("Then calling validateJob with more than one uploaded files fails with the expected error", func() {
			err := importer.validateJob(&models.ImportData{
				InstanceIDs:   []string{"testInstanceID1"},
				UploadedFiles: &[]models.UploadedFile{{AliasName: "testUpload1"}, {AliasName: "testUpload2"}},
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds and uploaded files must be 1")
		})
	})
}
