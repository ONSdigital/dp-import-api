package schema

import (
	"github.com/ONSdigital/dp-import-api/models"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAvroSchema(t *testing.T) {
	t.Parallel()
	Convey("When marshalling a PublishDataset message, no errors are returned", t, func() {
		// As the avro files are only checked at run time, this tests just validate the avro schema is valid
		message := models.PublishDataset{Recipe: "test", InstanceIds: []string{"1", "2", "3"},
			UploadedFiles: []models.UploadedFile{models.UploadedFile{URL: "s3//aws/bucket/file.xls", AliasName: "test"}}}
		_, avroError := PublishDataset.Marshal(message)
		So(avroError, ShouldBeNil)
	})
}
