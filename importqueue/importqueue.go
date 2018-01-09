package importqueue

import (
	"errors"

	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import/events"
	"github.com/ONSdigital/go-ns/log"
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

		inputFileAvailableEvent := events.InputFileAvailable{
			JobID:      job.JobID,
			InstanceID: job.InstanceIDs[0],
			URL:        (*job.UploadedFiles)[0].URL}

		log.Debug("producing new input file available event.", log.Data{"event": inputFileAvailableEvent})

		bytes, avroError := events.InputFileAvailableSchema.Marshal(inputFileAvailableEvent)
		if avroError != nil {
			return avroError
		}

		q.v4Queue <- bytes
	}

	return nil
}
