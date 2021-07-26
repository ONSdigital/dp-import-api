package datastore

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-api/models"
)

//go:generate moq -out mock/mongo.go -pkg mock . DataStorer

// DataStorer is an interface used to store import jobs
type DataStorer interface {
	AddJob(importJob *models.Job) (*models.Job, error)
	GetJob(jobID string) (*models.Job, error)
	GetJobs(ctx context.Context, filters []string, offset int, limit int) (*models.JobResults, error)
	UpdateJob(jobID string, update *models.Job) error
	UpdateProcessedInstance(id string, procInstances []models.ProcessedInstances) error
	AddUploadedFile(jobID string, message *models.UploadedFile) error
	Close(context.Context) error
	Checker(context.Context, *healthcheck.CheckState) error
}
