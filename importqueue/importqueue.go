package importqueue

import (
	"errors"

	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/schema"
	"github.com/ONSdigital/dp-import/event"
)

// ImportQueue used to send import jobs via kafka topic
type ImportQueue struct {
	v4Queue        chan []byte
	databakerQueue chan []byte
}

// CreateImportQueue used to queue data baker evenets and v4 files
func CreateImportQueue(databakerQueue, v4Queue chan []byte) *ImportQueue {
	return &ImportQueue{databakerQueue: databakerQueue, v4Queue: v4Queue}
}

// Queue an import event
func (q *ImportQueue) Queue(job *models.ImportData) error {

	if job.Format == "v4" {
		if len(job.InstanceIDs) != 1 && len(*job.UploadedFiles) != 1 {
			return errors.New("InstanceIds and uploaded files must be 1")
		}

		inputFileAvailableEvent := event.InputFileAvailable{
			JobID:job.JobID,
			InstanceID: job.InstanceIDs[0],
			URL: (*job.UploadedFiles)[0].URL}

		bytes, avroError := event.InputFileAvailableSchema.Marshal(inputFileAvailableEvent)
		if avroError != nil {
			return avroError
		}

		q.v4Queue <- bytes
		return nil
	}

	bytes, avroError := schema.DataBaker.Marshal(models.DataBakerEvent{JobID: job.JobID})
	if avroError != nil {
		return avroError
	}
	q.databakerQueue <- bytes

	return nil
}
