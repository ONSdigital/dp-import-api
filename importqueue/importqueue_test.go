package jobimport

import (
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/schema"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestQueueV4File(t *testing.T) {
	Convey("When a job is imported with a `v4` recipe it is sent directly to the import process", t, func() {
		v4Queue := make(chan []byte, 1)
		dataBakerQueue := make(chan []byte, 1)
		importer := CreateJobImporter(dataBakerQueue, v4Queue)
		job := models.ImportData{InstanceIds: []string{"1"}, Recipe: "v4",
			UploadedFiles:                    []models.UploadedFile{models.UploadedFile{AliasName: "v4", URL: "s3//aws/000/v4.csv"}}}
		importError := importer.Queue(&job)
		So(importError, ShouldBeNil)
		bytes := <-v4Queue
		var file V4File
		schema.ImportV4File.Unmarshal(bytes, &file)
		So(file.URL, ShouldEqual, job.UploadedFiles[0].URL)
		So(file.InstanceId, ShouldEqual, job.InstanceIds[0])
	})
}

func TestQueueDataBakerRecipe(t *testing.T) {
	Convey("When a job is imported with a data baker recipe it is sent to the data baker process", t, func() {
		v4Queue := make(chan []byte, 1)
		dataBakerQueue := make(chan []byte, 1)
		importer := CreateJobImporter(dataBakerQueue, v4Queue)
		job := models.ImportData{InstanceIds: []string{"1"}, Recipe: "CPI", JobId: "123",
			UploadedFiles:                    []models.UploadedFile{models.UploadedFile{AliasName: "1", URL: "s3//aws/000/v4.csv"}}}
		importError := importer.Queue(&job)
		So(importError, ShouldBeNil)
		bytes := <- dataBakerQueue
		var task models.DataBakerEvent
		schema.DataBaker.Unmarshal(bytes, &task)
		So(task.JobId, ShouldEqual, "123")
	})
}
