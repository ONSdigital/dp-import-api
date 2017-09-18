package datastore

import (
	"github.com/ONSdigital/dp-import-api/dataset/interface"
	"github.com/ONSdigital/dp-import-api/models"
)

// DataStorer is an interface used to store import jobs
type DataStorer interface {
	AddJob(importJob *models.Job, selfURL string, datasetAPI dataset.DatasetAPIer) (*models.Job, error)
	GetJob(jobID string) (*models.Job, error)
	GetJobs(filters []string) ([]models.Job, error)
	UpdateJob(jobID string, update *models.Job, withOutRestrictions bool) error
	UpdateJobState(jobID string, state string, withOutRestrictions bool) error
	AddUploadedFile(jobID string, message *models.UploadedFile) error
	PrepareJob(datasetAPI dataset.DatasetAPIer, jobID string) (*models.ImportData, error)
}
