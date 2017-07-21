package importqueue

import (
	"fmt"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/schema"
	"strings"
)

type importQueue struct {
	v4Queue        chan []byte
	databakerQueue chan []byte
}

type V4File struct {
	InstanceId string `avro:"instance_id"`
	URL        string `avro:"file_url"`
}

func CreateImportQueue(databakerQueue, v4Queue chan []byte) importQueue {
	return importQueue{databakerQueue: databakerQueue, v4Queue: v4Queue}
}

func (q *importQueue) Queue(job *models.ImportData) error {
	if strings.ToLower(job.Recipe) == "v4" {
		if len(job.InstanceIds) != 1 && len(job.UploadedFiles) != 1 {
			return fmt.Errorf("InstanceIds and uploaded files must be 1")
		}
		file := V4File{InstanceId: job.InstanceIds[0], URL: job.UploadedFiles[0].URL}
		fmt.Println("%+v\n", file)
		bytes, avroError := schema.ImportV4File.Marshal(file)
		if avroError != nil {
			return avroError
		}
		q.v4Queue <- bytes

	} else {

		bytes, avroError := schema.DataBaker.Marshal(models.DataBakerEvent{JobId: job.JobId})
		if avroError != nil {
			return avroError
		}
		q.databakerQueue <- bytes
	}
	return nil
}
