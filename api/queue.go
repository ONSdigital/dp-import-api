package api

import "github.com/ONSdigital/dp-import-api/models"

// An interface used to queue import jobs
type JobQueue interface {
	Queue(job *models.ImportData) error
}
