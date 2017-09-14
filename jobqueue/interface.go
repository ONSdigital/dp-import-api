package jobqueue

import "github.com/ONSdigital/dp-import-api/models"

// JobQueue interface used to queue import jobs
type JobQueue interface {
	Queue(job *models.ImportData) error
}
