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
		message := models.DataBakerEvent{JobId: "123"}
		bytes, avroError := DataBaker.Marshal(message)
		So(avroError, ShouldBeNil)
		var results models.DataBakerEvent
		err := DataBaker.Unmarshal(bytes, &results)
		So(err, ShouldBeNil)
		So(results.JobId, ShouldEqual, message.JobId)
	})
}
