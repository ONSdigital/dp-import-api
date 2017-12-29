package datastore

import (
	"github.com/ONSdigital/dp-import-api/models"
)

//go:generate moq -out testdatastore/datastore.go -pkg testdatastore . DataStorer

// DataStorer is an interface used to store import jobs
type DataStorer interface {
	AddJob(importJob *models.Job) (*models.Job, error)
	GetJob(jobID string) (*models.Job, error)
	GetJobs(filters []string) ([]models.Job, error)
	UpdateJob(jobID string, update *models.Job) error
	UpdateJobState(jobID string, state string) error
	AddUploadedFile(jobID string, message *models.UploadedFile) error
	UpdateInstanceState(jobID, instanceID, taskID, newState string) (err error)
	UpdateInstanceTaskState(jobID, instanceID, taskID, newState string) (err error)
}
