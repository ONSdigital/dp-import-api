package importqueue

import (
	"testing"

	"github.com/ONSdigital/dp-import-api/models"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/ONSdigital/dp-import/event"
)

func TestQueueV4File(t *testing.T) {
	Convey("When a job is imported with a `v4` recipe it is sent directly to the import process", t, func() {
		v4Queue := make(chan []byte, 1)
		dataBakerQueue := make(chan []byte, 1)
		importer := CreateImportQueue(dataBakerQueue, v4Queue)
		job := models.ImportData{
			InstanceIDs:   []string{"1"},
			Recipe:        "b944be78-f56d-409b-9ebd-ab2b77ffe187",
			Format:        "v4",
			UploadedFiles: &[]models.UploadedFile{{AliasName: "v4", URL: "s3//aws/000/v4.csv"}}}
		importError := importer.Queue(&job)
		So(importError, ShouldBeNil)
		bytes := <-v4Queue
		var file event.InputFileAvailable
		event.InputFileAvailableSchema.Unmarshal(bytes, &file)
		So(file.URL, ShouldEqual, (*job.UploadedFiles)[0].URL)
		So(file.InstanceID, ShouldEqual, job.InstanceIDs[0])
	})
}