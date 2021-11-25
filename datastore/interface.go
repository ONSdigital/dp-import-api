package datastore

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-api/models"
)

//go:generate moq -out mock/mongo.go -pkg mock . DataStorer

// DataStorer is an interface used to store import jobs
type DataStorer interface {
	AddJob(ctx context.Context, importJob *models.Job) (*models.Job, error)
	GetJob(ctx context.Context, jobID string) (*models.Job, error)
	GetJobs(ctx context.Context, filters []string, offset int, limit int) (*models.JobResults, error)
	UpdateJob(ctx context.Context, jobID string, update *models.Job) error
	UpdateProcessedInstance(ctx context.Context, id string, procInstances []models.ProcessedInstances) error
	AddUploadedFile(ctx context.Context, jobID string, message *models.UploadedFile) error
	Close(context.Context) error
	Checker(context.Context, *healthcheck.CheckState) error
	AcquireInstanceLock(ctx context.Context, jobID string) (lockID string, err error)
	UnlockInstance(ctx context.Context, lockID string)
}
