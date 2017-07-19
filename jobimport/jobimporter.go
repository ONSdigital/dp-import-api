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

func (ji *jobimporter) Queue(job *models.PublishDataset) error {
	if job.Recipe == "v4" {
		if len(job.InstanceIds) != len(job.UploadedFiles) {
			return fmt.Errorf("InstanceIds and uploaded files need to be the same size")
		}
		for i := 0; i < len(job.UploadedFiles); i++ {
			file := V4File{InstanceId: job.InstanceIds[i], URL: job.UploadedFiles[i].URL}
			bytes, avroError := schema.ImoprtV4File.Marshal(file)
			if avroError != nil {
				return avroError
			}
			ji.v4Queue <- bytes
		}

	} else {
		bytes, avroError := schema.PublishDataset.Marshal(job)
		if avroError != nil {
			return avroError
		}
		ji.databakerQueue <- bytes
	}
	return nil
}
