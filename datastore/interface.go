package datastore

import (
	"context"

	"github.com/ONSdigital/dp-import-api/dataset/interface"
	"github.com/ONSdigital/dp-import-api/models"
)

// DataStorer is an interface used to store import jobs
type DataStorer interface {
	AddJob(ctx context.Context, importJob *models.Job, selfURL string, datasetAPI dataset.DatasetAPIer) (*models.Job, error)
	GetJob(jobID string) (*models.Job, error)
	GetJobs(filters []string) ([]models.Job, error)
	UpdateJob(importID string, update *models.Job, withOutRestrictions bool) error
	UpdateJobState(importID string, state string, withOutRestrictions bool) error
	AddUploadedFile(ctx context.Context, instanceID string, message *models.UploadedFile, datasetAPI dataset.DatasetAPIer, selfURL string) (*models.Instance, error)
	PrepareJob(ctx context.Context, datasetAPI dataset.DatasetAPIer, importID string) (*models.ImportData, error)
}
