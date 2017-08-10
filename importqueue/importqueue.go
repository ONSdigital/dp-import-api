package importqueue

import (
	"errors"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/schema"
	"strings"
)

// ImportQueue used to send import jobs via kafka topic
type ImportQueue struct {
	v4Queue        chan []byte
	databakerQueue chan []byte
}

// V4File to import into a graph database
type V4File struct {
	InstanceID string `avro:"instance_id"`
	URL        string `avro:"file_url"`
}

// CreateImportQueue used to queue data baker evenets and v4 files
func CreateImportQueue(databakerQueue, v4Queue chan []byte) ImportQueue {
	return ImportQueue{databakerQueue: databakerQueue, v4Queue: v4Queue}
}

// Queue an import event
func (q *ImportQueue) Queue(job *models.ImportData) error {
	if strings.ToLower(job.Recipe) == "b944be78-f56d-409b-9ebd-ab2b77ffe187" {
		if len(job.InstanceIDs) != 1 && len(job.UploadedFiles) != 1 {
			return errors.New("InstanceIds and uploaded files must be 1")
		}
		file := V4File{InstanceID: job.InstanceIDs[0], URL: job.UploadedFiles[0].URL}
		bytes, avroError := schema.ImportV4File.Marshal(file)
		if avroError != nil {
			return avroError
		}
		q.v4Queue <- bytes

	} else {

		bytes, avroError := schema.DataBaker.Marshal(models.DataBakerEvent{JobID: job.JobID})
		if avroError != nil {
			return avroError
		}
		q.databakerQueue <- bytes
	}
	return nil
}
