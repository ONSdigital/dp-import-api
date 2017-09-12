package datastore

import (
	"github.com/ONSdigital/dp-import-api/dataset/interface"
	"github.com/ONSdigital/dp-import-api/models"
)

// DataStore is an interface used to store import jobs
type DataStore interface {
	AddJob(importJob *models.Job, selfURL string, datasetAPI dataset.DatasetAPIer) (*models.Job, error)
	GetJob(jobID string) (*models.Job, error)
	GetJobs(filters []string) ([]models.Job, error)
	UpdateJob(importID string, update *models.Job, withOutRestrictions bool) error
	UpdateJobState(importID string, state string, withOutRestrictions bool) error
	AddUploadedFile(instanceID string, message *models.UploadedFile, datasetAPI dataset.DatasetAPIer, selfURL string) (*models.Instance, error)
	PrepareJob(datasetAPI dataset.DatasetAPIer, importID string) (*models.ImportData, error)
}
