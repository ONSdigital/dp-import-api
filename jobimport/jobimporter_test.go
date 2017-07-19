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
		job := models.PublishDataset{InstanceIds: []string{"1"}, Recipe: "v4",
			UploadedFiles: []models.UploadedFile{models.UploadedFile{AliasName: "v4", URL: "s3//aws/000/v4.csv"}}}
		importError := importer.Queue(&job)
		So(importError, ShouldBeNil)
		bytes := <-v4Queue
		var file V4File
		schema.ImoprtV4File.Unmarshal(bytes, &file)
		So(file.URL, ShouldEqual, job.UploadedFiles[0].URL)
		So(file.InstanceId, ShouldEqual, job.InstanceIds[0])
	})
}

func TestQueueDataBakerRecipe(t *testing.T) {
	Convey("When a job is imported with a data baker recipe it is sent to the data baker process", t, func() {
		v4Queue := make(chan []byte, 1)
		dataBakerQueue := make(chan []byte, 1)
		importer := CreateJobImporter(dataBakerQueue, v4Queue)
		job := models.PublishDataset{InstanceIds: []string{"1"}, Recipe: "CPI",
			UploadedFiles: []models.UploadedFile{models.UploadedFile{AliasName: "1", URL: "s3//aws/000/v4.csv"}}}
		importError := importer.Queue(&job)
		So(importError, ShouldBeNil)
		//bytes := <- dataBakerQueue
		//var task models.PublishDataset
		//schema.ImoprtV4File.Unmarshal(bytes, &task)
		//So(task.InstanceIds, ShouldContain, "1")
		//So(len(task.UploadedFiles), ShouldEqual,1)
	})
}
