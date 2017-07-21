package jobimport

import (
	"fmt"
	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import-api/schema"
)

type jobimporter struct {
	v4Queue        chan []byte
	databakerQueue chan []byte
}

type V4File struct {
	InstanceId string `avro:"instance_id"`
	URL        string `avro:"file_url"`
}

func CreateJobImporter(databakerQueue, v4Queue chan []byte) jobimporter {
	return jobimporter{databakerQueue: databakerQueue, v4Queue: v4Queue}
}

func (ji *jobimporter) Queue(job *models.ImportData) error {
	if job.Recipe == "v4" {
		if len(job.InstanceIds) != 1 && len(job.UploadedFiles) != 1 {
			return fmt.Errorf("InstanceIds and uploaded files must be 1")
		}
		file := V4File{InstanceId: job.InstanceIds[0], URL: job.UploadedFiles[0].URL}
		bytes, avroError := schema.ImportV4File.Marshal(file)
		if avroError != nil {
			return avroError
		}
		ji.v4Queue <- bytes


	} else {
		bytes, avroError := schema.DataBaker.Marshal(models.DataBakerEvent{JobId:job.JobId})
		if avroError != nil {
			return avroError
		}
		ji.databakerQueue <- bytes
	}
	return nil
}
