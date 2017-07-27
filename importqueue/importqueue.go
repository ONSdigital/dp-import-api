package importqueue

import (
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/schema"
	"strings"
	"errors"
)

type importQueue struct {
	v4Queue        chan []byte
	databakerQueue chan []byte
}

// A V4 file to import into a graph database
type V4File struct {
	InstanceId string `avro:"instance_id"`
	URL        string `avro:"file_url"`
}

// Create a import queue for databaker evenets and v4 files
func CreateImportQueue(databakerQueue, v4Queue chan []byte) importQueue {
	return importQueue{databakerQueue: databakerQueue, v4Queue: v4Queue}
}

// Queue an import event
func (q *importQueue) Queue(job *models.ImportData) error {
	if strings.ToLower(job.Recipe) == "v4" {
		if len(job.InstanceIDs) != 1 && len(job.UploadedFiles) != 1 {
			return errors.New("InstanceIds and uploaded files must be 1")
		}
		file := V4File{InstanceId: job.InstanceIDs[0], URL: job.UploadedFiles[0].URL}
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
