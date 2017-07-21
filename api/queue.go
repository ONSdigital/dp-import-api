package api

import "github.com/ONSdigital/dp-import-api/models"

type JobQueue interface {
	Queue(job *models.ImportData) error
}
