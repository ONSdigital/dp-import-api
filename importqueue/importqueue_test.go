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
		JobID:         "jobId",
		InstanceIDs:   []string{"1"},
		Recipe:        "b944be78-f56d-409b-9ebd-ab2b77ffe187",
		Format:        "v4",
		UploadedFiles: &[]models.UploadedFile{{AliasName: "aliasV4", URL: "s3//aws/000/v4.csv"}}}

	Convey("Given a mocked importQueue without a v4 queue", t, func() {
		importer := CreateImportQueue(nil, nil, nil)

		Convey("Then importing a valid 'v4' recipe results in the expected error being returned", func() {
			err := importer.Queue(ctx, &job)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "v4 queue (kafka producer) is not available")
		})
	})

	Convey("Given a mocked importQueue with a valid v4 queue", t, func() {
		v4Queue := make(chan []byte, 1)
		importer := CreateImportQueue(nil, v4Queue, nil)

		Convey("Then importing an nil 'v4' recipe fails with the expected error", func() {
			err := importer.Queue(ctx, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "job not available")
		})

		Convey("Then importing a 'v4' recipe with nil instanceIDs fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs:   nil,
				Recipe:        "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:        "v4",
				UploadedFiles: &[]models.UploadedFile{{AliasName: "aliasV4", URL: "s3//aws/000/v4.csv"}}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds and uploaded files must be 1")
		})

		Convey("Then importing a 'v4' recipe with empty instanceIDs fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs:   []string{},
				Recipe:        "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:        "v4",
				UploadedFiles: &[]models.UploadedFile{{AliasName: "aliasV4", URL: "s3//aws/000/v4.csv"}}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds and uploaded files must be 1")
		})

		Convey("Then importing a 'v4' recipe with multiple instanceIDs fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs:   []string{"1", "2"},
				Recipe:        "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:        "v4",
				UploadedFiles: &[]models.UploadedFile{{AliasName: "aliasV4", URL: "s3//aws/000/v4.csv"}}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds and uploaded files must be 1")
		})

		Convey("Then importing a 'v4' recipe with nil uploadedFiles fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs:   []string{"1"},
				Recipe:        "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:        "v4",
				UploadedFiles: nil})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds and uploaded files must be 1")
		})

		Convey("Then importing a 'v4' recipe with empty uploadedFiles fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs:   []string{"1"},
				Recipe:        "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:        "v4",
				UploadedFiles: &[]models.UploadedFile{}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds and uploaded files must be 1")
		})

		Convey("Then importing a 'v4' recipe with multiple uploadedFiles fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs: []string{"1"},
				Recipe:      "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:      "v4",
				UploadedFiles: &[]models.UploadedFile{
					{AliasName: "aliasV41", URL: "s3//aws/000/v41.csv"},
					{AliasName: "aliasV42", URL: "s3//aws/000/v42.csv"},
				}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds and uploaded files must be 1")
		})

		Convey("Then importing a valid 'v4' recipe sends the expected import event to the v4 queue", func() {
			err := importer.Queue(ctx, &job)
			So(err, ShouldBeNil)

			bytes := <-v4Queue

			var file events.InputFileAvailable
			events.InputFileAvailableSchema.Unmarshal(bytes, &file)

			So(file, ShouldResemble, events.InputFileAvailable{
				JobID:      job.JobID,
				URL:        (*job.UploadedFiles)[0].URL,
				InstanceID: job.InstanceIDs[0],
			})
		})
	})
}

func TestQueueCantabularFile(t *testing.T) {
	ctx := context.Background()

	Convey("Given a mocked importQueue without a cantabular queue", t, func() {
		importer := CreateImportQueue(nil, nil, nil)

		Convey("Then importing a 'cantabular_blob' recipe results in the expected error being returned", func() {
			err := importer.Queue(ctx, &models.ImportData{
				JobID:       "jobId",
				InstanceIDs: []string{"InstanceId"},
				Recipe:      "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:      formatCantabularBlob,
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "cantabular queue (kafka producer) is not available")
		})

		Convey("Then importing a 'cantabular_table' recipe results in the expected error being returned", func() {
			err := importer.Queue(ctx, &models.ImportData{
				JobID:       "jobId",
				InstanceIDs: []string{"InstanceId"},
				Recipe:      "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:      formatCantabularTable,
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "cantabular queue (kafka producer) is not available")
		})
	})

	Convey("Given a mocked importQueue with a valid cantabular queue", t, func() {
		cantabularQueue := make(chan []byte, 1)
		importer := CreateImportQueue(nil, nil, cantabularQueue)

		Convey("Then importing an nil 'cantabular' recipe fails with the expected error", func() {
			err := importer.Queue(ctx, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "job not available")
		})

		Convey("Then importing a 'cantabular_blob' recipe with nil instanceIDs fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs: nil,
				Recipe:      "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:      formatCantabularBlob})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds must be 1")
		})

		Convey("Then importing a 'cantabular_table' recipe with nil instanceIDs fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs: nil,
				Recipe:      "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:      formatCantabularTable})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds must be 1")
		})

		Convey("Then importing a 'cantabular_blob' recipe with empty instanceIDs fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs: []string{},
				Recipe:      "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:      formatCantabularBlob})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds must be 1")
		})

		Convey("Then importing a 'cantabular_table' recipe with empty instanceIDs fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs: []string{},
				Recipe:      "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:      formatCantabularTable})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds must be 1")
		})

		Convey("Then importing a 'cantabular_blob' recipe with multiple instanceIDs fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs: []string{"1", "2"},
				Recipe:      "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:      formatCantabularBlob})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds must be 1")
		})

		Convey("Then importing a 'cantabular_table' recipe with multiple instanceIDs fails with the expected error", func() {
			err := importer.Queue(ctx, &models.ImportData{
				InstanceIDs: []string{"1", "2"},
				Recipe:      "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:      formatCantabularTable})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "InstanceIds must be 1")
		})

		Convey("Then importing a 'cantabular_blob' recipe sends the expected import event to the cantabular queue", func() {
			job := &models.ImportData{
				JobID:       "jobId",
				InstanceIDs: []string{"InstanceId"},
				Recipe:      "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:      formatCantabularBlob,
			}
			err := importer.Queue(ctx, job)
			So(err, ShouldBeNil)

			bytes := <-cantabularQueue

			var file events.CantabularDatasetInstanceStarted
			events.CantabularDatasetInstanceStartedSchema.Unmarshal(bytes, &file)

			So(file, ShouldResemble, events.CantabularDatasetInstanceStarted{
				JobID:          "jobId",
				InstanceID:     job.InstanceIDs[0],
				CantabularType: formatCantabularBlob,
			})
		})

		Convey("Then importing a 'cantabular_table' recipe sends the expected import event to the cantabular queue", func() {
			job := &models.ImportData{
				JobID:       "jobId",
				InstanceIDs: []string{"InstanceId"},
				Recipe:      "b944be78-f56d-409b-9ebd-ab2b77ffe187",
				Format:      formatCantabularTable,
			}
			err := importer.Queue(ctx, job)
			So(err, ShouldBeNil)

			bytes := <-cantabularQueue

			var file events.CantabularDatasetInstanceStarted
			events.CantabularDatasetInstanceStartedSchema.Unmarshal(bytes, &file)

			So(file, ShouldResemble, events.CantabularDatasetInstanceStarted{
				JobID:          job.JobID,
				InstanceID:     job.InstanceIDs[0],
				CantabularType: formatCantabularTable,
			})
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
