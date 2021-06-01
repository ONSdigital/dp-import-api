package importqueue

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-import/events"
	"github.com/ONSdigital/log.go/log"
)

// block of constants corresponding to possible job formats
const (
	formatV4         = "v4"
	formatCantabular = "cantabular"
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
	case formatCantabular:
		return q.queueCantabular(ctx, job)
	default:
		log.Event(ctx, "unrecognised job format, no action has been taken", log.WARN, log.Data{"job_format": job.Format})
	}
	return nil
}

// queueV4 generates a kafka message for a V4 import
func (q *ImportQueue) queueV4(ctx context.Context, job *models.ImportData) error {
	if q.v4Queue == nil {
		return errors.New("v4 queue (kafka producer) is not available")
	}
	if job.InstanceIDs == nil || len(job.InstanceIDs) != 1 || job.UploadedFiles == nil || len(*job.UploadedFiles) != 1 {
		return errors.New("InstanceIds and uploaded files must be 1")
	}

	inputFileAvailableEvent := events.InputFileAvailable{
		JobID:      job.JobID,
		InstanceID: job.InstanceIDs[0],
		URL:        (*job.UploadedFiles)[0].URL,
	}

	log.Event(ctx, "producing new input file available event", log.INFO, log.Data{"event": inputFileAvailableEvent, "format": formatV4})

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
		return errors.New("InstanceIds must be 1")
	}

	inputFileAvailableEvent := events.InputFileAvailable{
		JobID:      job.JobID,
		InstanceID: job.InstanceIDs[0],
	}

	log.Event(ctx, "producing new input file available event", log.INFO, log.Data{"event": inputFileAvailableEvent, "format": formatCantabular})

	bytes, avroError := events.InputFileAvailableSchema.Marshal(inputFileAvailableEvent)
	if avroError != nil {
		return avroError
	}

	q.cantabularQueue <- bytes
	return nil
}
