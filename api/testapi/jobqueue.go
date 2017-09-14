package mockapi

import (
	"fmt"
	"github.com/ONSdigital/dp-import-api/models"
)

// JobQueue interface used to queue import jobs
// type JobQueue interface {
// Queue(job *models.ImportData) error
// }

type JobImporter struct {
	returnError bool
}

type V4File struct {
	InstanceId string `avro:"instance_id"`
	URL        string `avro:"file_url"`
}

func (ji *JobImporter) Queue(job *models.ImportData) error {
	if ji.returnError {
		return fmt.Errorf("Failed to queue import job")
	}
	return nil
}
