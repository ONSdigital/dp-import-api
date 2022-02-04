package importqueue

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import/events"
	"github.com/ONSdigital/log.go/v2/log"
)

// block of constants corresponding to possible job formats
const (
	formatV4                      = "v4"
	formatCantabularBlob          = "cantabular_blob"
	formatCantabularTable         = "cantabular_table"
	formatCantabularFlexibleTable = "cantabular_flexible_table"
)

// ImportQueue used to send import jobs via kafka topic
type ImportQueue struct {
	databakerQueue  chan []byte
	v4Queue         chan []byte
	cantabularQueue chan []byte
}

// CreateImportQueue used to queue data baker evenets and v4 files
func CreateImportQueue(databakerQueue, v4Queue, cantabularQueue chan []byte) *ImportQueue {
	return &ImportQueue{databakerQueue: databakerQueue, v4Queue: v4Queue, cantabularQueue: cantabularQueue}
}

// Queue generates a kafka message for an import event, according to the provided job format
func (q *ImportQueue) Queue(ctx context.Context, job *models.ImportData) error {
	if job == nil {
		return errors.New("job not available")
	}

	switch job.Format {
	case formatV4:
		return q.queueV4(ctx, job)
	case formatCantabularTable, formatCantabularBlob, formatCantabularFlexibleTable:
		return q.queueCantabular(ctx, job)
	default:
		log.Warn(ctx, "unrecognised job format, no action has been taken", log.Data{"job_format": job.Format})
	}
	return nil
}

// queueV4 generates a kafka message for a V4 import
func (q *ImportQueue) queueV4(ctx context.Context, job *models.ImportData) error {
	if q.v4Queue == nil {
		return errors.New("v4 queue (kafka producer) is not available")
	}
	if job.InstanceIDs == nil || len(job.InstanceIDs) != 1 || job.UploadedFiles == nil || len(*job.UploadedFiles) != 1 {
		return errors.New("InstanceIds and uploaded files must have length 1")
	}

	inputFileAvailableEvent := events.InputFileAvailable{
		JobID:      job.JobID,
		InstanceID: job.InstanceIDs[0],
		URL:        (*job.UploadedFiles)[0].URL,
	}

	log.Info(ctx, "producing new input file available event", log.Data{"event": inputFileAvailableEvent, "format": formatV4})

	bytes, avroError := events.InputFileAvailableSchema.Marshal(inputFileAvailableEvent)
	if avroError != nil {
		return avroError
	}

	q.v4Queue <- bytes
	return nil
}

// queueCantabular generates a kafka message for a Cantabular import
func (q *ImportQueue) queueCantabular(ctx context.Context, job *models.ImportData) error {
	if q.cantabularQueue == nil {
		return errors.New("cantabular queue (kafka producer) is not available")
	}
	if job.InstanceIDs == nil || len(job.InstanceIDs) != 1 {
		return errors.New("InstanceIds must have length 1")
	}

	event := events.CantabularDatasetInstanceStarted{
		RecipeID:       job.Recipe,
		JobID:          job.JobID,
		InstanceID:     job.InstanceIDs[0],
		CantabularType: job.Format,
	}

	log.Info(ctx, "producing new cantabular dataset instance started event", log.Data{"event": event})

	bytes, avroError := events.CantabularDatasetInstanceStartedSchema.Marshal(event)
	if avroError != nil {
		return avroError
	}

	q.cantabularQueue <- bytes
	return nil
}
